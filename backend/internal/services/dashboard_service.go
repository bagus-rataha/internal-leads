package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DashboardService has no repository - ARCHITECTURE.md's explicit exception
// for dashboard/export (read-only, aggregate-only queries, a repository
// layer here would be an empty abstraction). userRepo is kept only because
// buildLeadScope needs it for a LEADER's team lookup.
type DashboardService struct {
	db       *gorm.DB
	userRepo userRepositoryForLead
}

func NewDashboardService(db *gorm.DB, userRepo userRepositoryForLead) *DashboardService {
	return &DashboardService{db: db, userRepo: userRepo}
}

// scopedLeads is every dashboard query's starting point: the caller's scope,
// narrowed by the query's own team_id/owner_id params (applied after scope,
// never replacing it - identical narrowing SQL to LeadRepository.
// applyLeadFilter's TeamID/OwnerID clauses).
func (s *DashboardService) scopedLeads(scope repository.LeadScope, teamID, ownerID *uuid.UUID) *gorm.DB {
	tx := repository.ApplyLeadScope(s.db.Model(&models.Lead{}), scope)
	if teamID != nil {
		tx = tx.Where("leads.owner_id IN (SELECT id FROM users WHERE team_id = ?)", *teamID)
	}
	if ownerID != nil {
		tx = tx.Where("leads.owner_id = ?", *ownerID)
	}
	return tx
}

// followUpsInScope mirrors scopedLeads for follow_ups: joins to leads so the
// same ApplyLeadScope/team/owner clauses (which reference leads.owner_id)
// apply unchanged.
func (s *DashboardService) followUpsInScope(scope repository.LeadScope, teamID, ownerID *uuid.UUID) *gorm.DB {
	tx := s.db.Model(&models.FollowUp{}).Joins("JOIN leads ON leads.id = follow_ups.lead_id")
	tx = repository.ApplyLeadScope(tx, scope)
	if teamID != nil {
		tx = tx.Where("leads.owner_id IN (SELECT id FROM users WHERE team_id = ?)", *teamID)
	}
	if ownerID != nil {
		tx = tx.Where("leads.owner_id = ?", *ownerID)
	}
	return tx
}

// changePct implements MetricCard's zero-state contract: nil when the
// comparison period was 0. A caller distinguishes "—" from "baru" using
// Value alongside a nil ChangePct (see dto.MetricCard's doc comment) -
// this function never needs to know which of the two cases it is.
func changePct(cur, prev int64) *int {
	if prev == 0 {
		return nil
	}
	pct := int(math.Round(float64(cur-prev) / float64(prev) * 100))
	return &pct
}

// percentOf is the same zero-state rule applied to a part/total ratio
// (funnel stages, segment conversion rates): nil when total is 0.
func percentOf(part, total int64) *int {
	if total == 0 {
		return nil
	}
	pct := int(math.Round(float64(part) / float64(total) * 100))
	return &pct
}

// comparisonPeriod returns the same-length period immediately preceding
// [from, to] (inclusive both ends, calendar days) - ARCHITECTURE.md §9's
// "pembanding: rentang yang sama sebelumnya".
func comparisonPeriod(from, to time.Time) (prevFrom, prevTo time.Time) {
	days := int(to.Sub(from).Hours()/24) + 1
	prevTo = from.AddDate(0, 0, -1)
	prevFrom = prevTo.AddDate(0, 0, -(days - 1))
	return prevFrom, prevTo
}

// Summary answers GET /dashboard/summary: the metric cards and the
// conversion funnel. The funnel is a scope snapshot (team_id/owner_id
// narrowed, but NOT date_from/date_to-bound) - see FunnelResponse's doc comment.
func (s *DashboardService) Summary(callerID uuid.UUID, role string, q dto.DashboardQuery) (*dto.DashboardSummaryResponse, error) {
	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}
	base := func() *gorm.DB { return s.scopedLeads(scope, q.TeamID, q.OwnerID) }
	prevFrom, prevTo := comparisonPeriod(q.DateFrom, q.DateTo)

	var leadBaruCur, leadBaruPrev int64
	if err := base().Where("leads.created_at >= ? AND leads.created_at < ?", q.DateFrom, q.DateTo.AddDate(0, 0, 1)).Count(&leadBaruCur).Error; err != nil {
		return nil, err
	}
	if err := base().Where("leads.created_at >= ? AND leads.created_at < ?", prevFrom, prevTo.AddDate(0, 0, 1)).Count(&leadBaruPrev).Error; err != nil {
		return nil, err
	}

	var fuCur, fuPrev int64
	if err := s.followUpsInScope(scope, q.TeamID, q.OwnerID).
		Where("follow_ups.created_at >= ? AND follow_ups.created_at < ?", q.DateFrom, q.DateTo.AddDate(0, 0, 1)).
		Count(&fuCur).Error; err != nil {
		return nil, err
	}
	if err := s.followUpsInScope(scope, q.TeamID, q.OwnerID).
		Where("follow_ups.created_at >= ? AND follow_ups.created_at < ?", prevFrom, prevTo.AddDate(0, 0, 1)).
		Count(&fuPrev).Error; err != nil {
		return nil, err
	}

	// Handoff period-window uses updated_at as the transition-timestamp
	// proxy: HANDOFF_ODOO is a terminal status (no legal further mutation
	// once a lead reaches it), so updated_at on a HANDOFF_ODOO lead is
	// reliably "when it got there". No dedicated column exists - documented
	// approximation, not silently assumed.
	var handCur, handPrev int64
	if err := base().Where("leads.status = 'HANDOFF_ODOO' AND leads.updated_at >= ? AND leads.updated_at < ?", q.DateFrom, q.DateTo.AddDate(0, 0, 1)).Count(&handCur).Error; err != nil {
		return nil, err
	}
	if err := base().Where("leads.status = 'HANDOFF_ODOO' AND leads.updated_at >= ? AND leads.updated_at < ?", prevFrom, prevTo.AddDate(0, 0, 1)).Count(&handPrev).Error; err != nil {
		return nil, err
	}

	// Stale is a snapshot metric (not period-bound). stalePrev mirrors the
	// mockup's own approximation: leads that were ALREADY stale twice-over
	// as of today - there's no historical snapshot to compare against, so
	// this proxies "was this already stale in the prior window" using a
	// doubled threshold against current data.
	now := time.Now()
	var staleNow, stalePrev int64
	if err := base().Where("leads.status IN ('BARU','FOLLOW_UP') AND COALESCE(leads.last_follow_up_at, leads.created_at) < ?",
		now.AddDate(0, 0, -repository.StaleLeadThresholdDays)).Count(&staleNow).Error; err != nil {
		return nil, err
	}
	if err := base().Where("leads.status IN ('BARU','FOLLOW_UP') AND COALESCE(leads.last_follow_up_at, leads.created_at) < ?",
		now.AddDate(0, 0, -2*repository.StaleLeadThresholdDays)).Count(&stalePrev).Error; err != nil {
		return nil, err
	}

	var total, reachedFu, reachedHand, lost int64
	if err := base().Count(&total).Error; err != nil {
		return nil, err
	}
	if err := base().Where("leads.follow_up_count > 0").Count(&reachedFu).Error; err != nil {
		return nil, err
	}
	if err := base().Where("leads.status = 'HANDOFF_ODOO'").Count(&reachedHand).Error; err != nil {
		return nil, err
	}
	if err := base().Where("leads.status = 'LOST'").Count(&lost).Error; err != nil {
		return nil, err
	}

	funnel := dto.FunnelResponse{
		Stages: []dto.FunnelStage{
			{Name: "Lead Baru (masuk)", Count: total, Pct: percentOf(total, total)},
			{Name: "Sudah di-follow-up", Count: reachedFu, Pct: percentOf(reachedFu, total)},
			{Name: "Handoff ke Odoo", Count: reachedHand, Pct: percentOf(reachedHand, total)},
		},
		BaruToFuPct:    percentOf(reachedFu, total),
		FuToHandoffPct: percentOf(reachedHand, reachedFu),
		LostCount:      lost,
		LostPct:        percentOf(lost, total),
	}

	return &dto.DashboardSummaryResponse{
		LeadBaru:  dto.MetricCard{Value: leadBaruCur, ChangePct: changePct(leadBaruCur, leadBaruPrev)},
		FollowUp:  dto.MetricCard{Value: fuCur, ChangePct: changePct(fuCur, fuPrev)},
		Handoff:   dto.MetricCard{Value: handCur, ChangePct: changePct(handCur, handPrev)},
		Terlantar: dto.MetricCard{Value: staleNow, ChangePct: changePct(staleNow, stalePrev)},
		Funnel:    funnel,
	}, nil
}

// Activity answers GET /dashboard/activity: daily lead-baru vs
// follow-up counts. Two GROUP BY aggregate queries total, regardless of the
// requested range length - never a query-per-day loop (ARCHITECTURE.md §10's
// "dilarang N+1" applies to this per-day breakdown too, not just per-entity).
func (s *DashboardService) Activity(callerID uuid.UUID, role string, q dto.DashboardQuery) (*dto.DashboardActivityResponse, error) {
	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}

	type dayCount struct {
		Day   time.Time
		Count int64
	}

	var leadRows []dayCount
	if err := s.scopedLeads(scope, q.TeamID, q.OwnerID).
		Select("DATE(leads.created_at) AS day, COUNT(*) AS count").
		Where("leads.created_at >= ? AND leads.created_at < ?", q.DateFrom, q.DateTo.AddDate(0, 0, 1)).
		Group("DATE(leads.created_at)").
		Scan(&leadRows).Error; err != nil {
		return nil, err
	}

	var fuRows []dayCount
	if err := s.followUpsInScope(scope, q.TeamID, q.OwnerID).
		Select("DATE(follow_ups.created_at) AS day, COUNT(*) AS count").
		Where("follow_ups.created_at >= ? AND follow_ups.created_at < ?", q.DateFrom, q.DateTo.AddDate(0, 0, 1)).
		Group("DATE(follow_ups.created_at)").
		Scan(&fuRows).Error; err != nil {
		return nil, err
	}

	leadByDay := make(map[string]int64, len(leadRows))
	for _, r := range leadRows {
		leadByDay[r.Day.Format("2006-01-02")] = r.Count
	}
	fuByDay := make(map[string]int64, len(fuRows))
	for _, r := range fuRows {
		fuByDay[r.Day.Format("2006-01-02")] = r.Count
	}

	days := int(q.DateTo.Sub(q.DateFrom).Hours()/24) + 1
	buckets := make([]dto.ActivityBucket, days)
	for i := 0; i < days; i++ {
		key := q.DateFrom.AddDate(0, 0, i).Format("2006-01-02")
		buckets[i] = dto.ActivityBucket{Date: key, LeadBaru: leadByDay[key], FollowUp: fuByDay[key]}
	}

	return &dto.DashboardActivityResponse{Buckets: buckets}, nil
}

// StaleLeads answers GET /dashboard/stale-leads: the top-7
// longest-overdue table. Snapshot (not date_from/date_to-bound, matching
// the "terlantar" concept elsewhere in this app), team_id/owner_id still
// narrow.
func (s *DashboardService) StaleLeads(callerID uuid.UUID, role string, q dto.DashboardQuery) (*dto.DashboardStaleLeadsResponse, error) {
	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	threshold := now.AddDate(0, 0, -repository.StaleLeadThresholdDays)

	var leads []models.Lead
	err = s.scopedLeads(scope, q.TeamID, q.OwnerID).
		Preload("Owner").
		Where("leads.status IN ('BARU','FOLLOW_UP') AND COALESCE(leads.last_follow_up_at, leads.created_at) < ?", threshold).
		Order("COALESCE(leads.last_follow_up_at, leads.created_at) ASC").
		Limit(7).
		Find(&leads).Error
	if err != nil {
		return nil, err
	}

	items := make([]dto.StaleLeadRow, len(leads))
	for i, lead := range leads {
		lastActivity := lead.CreatedAt
		if lead.LastFollowUpAt != nil {
			lastActivity = *lead.LastFollowUpAt
		}
		ownerName := ""
		if lead.Owner != nil {
			ownerName = lead.Owner.Name
		}
		items[i] = dto.StaleLeadRow{
			Code:        lead.Code,
			CompanyName: lead.CompanyName,
			OwnerName:   ownerName,
			DaysSince:   int(now.Sub(lastActivity).Hours() / 24),
		}
	}

	return &dto.DashboardStaleLeadsResponse{Items: items}, nil
}

// countByUUID runs one GROUP BY aggregate over `table`, keyed by `groupCol`,
// narrowed to `ids` plus an optional extra WHERE clause - the shared helper
// behind every per-user metric in SalesActivity, so adding a new metric
// never means adding a query-per-user loop.
func (s *DashboardService) countByUUID(table, groupCol string, ids []uuid.UUID, extraWhere string, args ...interface{}) (map[uuid.UUID]int64, error) {
	type row struct {
		ID    uuid.UUID
		Count int64
	}
	q := s.db.Table(table).Select(groupCol+" AS id, COUNT(*) AS count").Where(groupCol+" IN ?", ids)
	if extraWhere != "" {
		q = q.Where(extraWhere, args...)
	}
	var rows []row
	if err := q.Group(groupCol).Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]int64, len(rows))
	for _, r := range rows {
		result[r.ID] = r.Count
	}
	return result, nil
}

// SalesActivity answers GET /dashboard/sales-activity: the sales activity
// roster table. Fixed small number of GROUP BY aggregate queries regardless
// of roster size - never a query-per-sales-person loop (ARCHITECTURE.md §10).
func (s *DashboardService) SalesActivity(callerID uuid.UUID, role string, q dto.DashboardQuery) (*dto.DashboardSalesActivityResponse, error) {
	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}

	rosterQuery := s.db.Model(&models.User{}).Where("role = 'SALES' AND is_active = true")
	if scope.TeamID != nil {
		rosterQuery = rosterQuery.Where("team_id = ?", *scope.TeamID)
	}
	if q.TeamID != nil {
		rosterQuery = rosterQuery.Where("team_id = ?", *q.TeamID)
	}
	if q.OwnerID != nil {
		rosterQuery = rosterQuery.Where("id = ?", *q.OwnerID)
	}
	var roster []models.User
	if err := rosterQuery.Preload("Team").Find(&roster).Error; err != nil {
		return nil, err
	}
	if len(roster) == 0 {
		return &dto.DashboardSalesActivityResponse{Items: []dto.SalesActivityRow{}}, nil
	}
	ids := make([]uuid.UUID, len(roster))
	for i, u := range roster {
		ids[i] = u.ID
	}

	leadBaruByOwner, err := s.countByUUID("leads", "owner_id", ids, "created_at >= ? AND created_at < ?", q.DateFrom, q.DateTo.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	fuWrittenByAuthor, err := s.countByUUID("follow_ups", "created_by_id", ids, "created_at >= ? AND created_at < ?", q.DateFrom, q.DateTo.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	terlantarByOwner, err := s.countByUUID("leads", "owner_id", ids, "status IN ('BARU','FOLLOW_UP') AND COALESCE(last_follow_up_at, created_at) < ?", time.Now().AddDate(0, 0, -repository.StaleLeadThresholdDays))
	if err != nil {
		return nil, err
	}
	handoffByOwner, err := s.countByUUID("leads", "owner_id", ids, "status = 'HANDOFF_ODOO'")
	if err != nil {
		return nil, err
	}

	type ownedAgg struct {
		OwnerID uuid.UUID
		Total   int64
		AvgFu   float64
	}
	var ownedRows []ownedAgg
	if err := s.db.Model(&models.Lead{}).
		Select("owner_id, COUNT(*) AS total, AVG(follow_up_count) AS avg_fu").
		Where("owner_id IN ?", ids).
		Group("owner_id").
		Scan(&ownedRows).Error; err != nil {
		return nil, err
	}
	ownedByOwner := make(map[uuid.UUID]ownedAgg, len(ownedRows))
	for _, r := range ownedRows {
		ownedByOwner[r.OwnerID] = r
	}

	type lastActivityAgg struct {
		CreatedByID uuid.UUID
		Last        time.Time
	}
	var lastRows []lastActivityAgg
	if err := s.db.Model(&models.FollowUp{}).
		Select("created_by_id, MAX(created_at) AS last").
		Where("created_by_id IN ?", ids).
		Group("created_by_id").
		Scan(&lastRows).Error; err != nil {
		return nil, err
	}
	lastByAuthor := make(map[uuid.UUID]time.Time, len(lastRows))
	for _, r := range lastRows {
		lastByAuthor[r.CreatedByID] = r.Last
	}

	// sortableRow keeps each row's `inactive` flag traveling alongside the
	// row itself through every swap. Sorting []dto.SalesActivityRow directly
	// while indexing a separate inactiveFlags[] slice by comparator index
	// breaks the moment the first swap happens: SliceStable's a/b are
	// positions in the slice being reordered, not positions in the original
	// roster, so a same-indexed side slice silently desyncs from the rows
	// it was supposed to describe.
	type sortableRow struct {
		row      dto.SalesActivityRow
		inactive bool
	}

	now := time.Now()
	rows := make([]sortableRow, len(roster))
	for i, u := range roster {
		owned := ownedByOwner[u.ID]
		terlantar := terlantarByOwner[u.ID]

		var lastActivity *time.Time
		inactive := true
		if last, ok := lastByAuthor[u.ID]; ok {
			l := last
			lastActivity = &l
			inactive = now.Sub(last).Hours()/24 > 7
		}

		var attentionTag *string
		switch {
		case inactive:
			tag := "Tanpa aktivitas 7 hari"
			attentionTag = &tag
		case terlantar >= 3:
			tag := "Banyak lead terlantar"
			attentionTag = &tag
		}

		var teamName *string
		if u.Team != nil {
			teamName = &u.Team.Name
		}

		rows[i] = sortableRow{
			row: dto.SalesActivityRow{
				UserID:       u.ID,
				Name:         u.Name,
				TeamName:     teamName,
				LeadBaru:     leadBaruByOwner[u.ID],
				FollowUp:     fuWrittenByAuthor[u.ID],
				AvgFuPerLead: owned.AvgFu,
				Terlantar:    terlantar,
				Handoff:      handoffByOwner[u.ID],
				ConvPct:      percentOf(handoffByOwner[u.ID], owned.Total),
				LastActivity: lastActivity,
				AttentionTag: attentionTag,
			},
			inactive: inactive,
		}
	}

	sort.SliceStable(rows, func(a, b int) bool {
		if rows[a].inactive != rows[b].inactive {
			return rows[a].inactive
		}
		return rows[a].row.Terlantar > rows[b].row.Terlantar
	})

	items := make([]dto.SalesActivityRow, len(rows))
	for i, r := range rows {
		items[i] = r.row
	}

	return &dto.DashboardSalesActivityResponse{Items: items}, nil
}

// nilIfNoHandoff enforces the segments zero-state rule: when the caller's
// entire scope has zero handoffs anywhere, every row's conversion is null -
// not just rows with their own zero denominator, which never happens here
// (a row only exists for a category with >=1 lead). Extends the same
// null-not-zero contract ARCHITECTURE.md already defines for change_pct,
// applied at the scope level instead of the per-row level.
func nilIfNoHandoff(pct *int, anyHandoff bool) *int {
	if !anyHandoff {
		return nil
	}
	return pct
}

// avgPricePerMbps computes one competitor-pricing stat tile. serviceType
// empty means "all service types"; otherwise it joins service_types by
// name. Returns nil (SQL NULL, via the pointer scan target) when no lead
// qualifies - never a fake 0.
func (s *DashboardService) avgPricePerMbps(base func() *gorm.DB, serviceType string) (*float64, error) {
	tx := base().Where("leads.price IS NOT NULL AND leads.price > 0 AND leads.capacity_mbps IS NOT NULL AND leads.capacity_mbps > 0")
	if serviceType != "" {
		tx = tx.Joins("JOIN service_types ON service_types.id = leads.service_type_id").Where("service_types.name = ?", serviceType)
	}
	var row struct{ Avg *float64 }
	if err := tx.Select("AVG(leads.price / leads.capacity_mbps) AS avg").Scan(&row).Error; err != nil {
		return nil, err
	}
	return row.Avg, nil
}

// Segments answers GET /dashboard/segments: the lead-source,
// region-penetration, competitor-intel, and business-field breakdowns - all
// snapshots of the caller's scope (team_id/owner_id narrowed, not
// date_from/date_to-bound), bundled into one response since they're all
// cheap "current portfolio state" breakdowns over the same scoped set.
func (s *DashboardService) Segments(callerID uuid.UUID, role string, q dto.DashboardQuery) (*dto.DashboardSegmentsResponse, error) {
	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}
	base := func() *gorm.DB { return s.scopedLeads(scope, q.TeamID, q.OwnerID) }

	var reachedHand int64
	if err := base().Where("leads.status = 'HANDOFF_ODOO'").Count(&reachedHand).Error; err != nil {
		return nil, err
	}
	anyHandoff := reachedHand > 0

	type nameAgg struct {
		Name    string
		Count   int64
		Handoff int64
	}

	// Sumber Lead
	var sourceRows []nameAgg
	if err := base().
		Joins("JOIN lead_sources ON lead_sources.id = leads.lead_source_id").
		Select("lead_sources.name AS name, COUNT(*) AS count, COUNT(*) FILTER (WHERE leads.status = 'HANDOFF_ODOO') AS handoff").
		Group("lead_sources.name").
		Order("count DESC").
		Scan(&sourceRows).Error; err != nil {
		return nil, err
	}
	sources := make([]dto.SegmentRow, len(sourceRows))
	for i, r := range sourceRows {
		convPct := percentOf(r.Handoff, r.Count)
		warn := anyHandoff && r.Count >= 3 && *convPct < 20
		sources[i] = dto.SegmentRow{Name: r.Name, Count: r.Count, ConversionPct: nilIfNoHandoff(convPct, anyHandoff), Warn: warn}
	}

	// Bidang Usaha (leads.business_field is a plain nullable text column, no join)
	var fieldRows []nameAgg
	if err := base().
		Where("leads.business_field IS NOT NULL AND leads.business_field <> ''").
		Select("leads.business_field AS name, COUNT(*) AS count, COUNT(*) FILTER (WHERE leads.status = 'HANDOFF_ODOO') AS handoff").
		Group("leads.business_field").
		Order("count DESC").
		Scan(&fieldRows).Error; err != nil {
		return nil, err
	}
	businessFields := make([]dto.SegmentRow, len(fieldRows))
	for i, r := range fieldRows {
		businessFields[i] = dto.SegmentRow{Name: r.Name, Count: r.Count, ConversionPct: nilIfNoHandoff(percentOf(r.Handoff, r.Count), anyHandoff)}
	}

	// Penetrasi Wilayah: province rollup + city rollup, merged into one
	// flat, frontend-groupable list.
	type regionAgg struct {
		ProvinceName string
		CityName     string
		Count        int64
		Handoff      int64
	}
	var provRows []regionAgg
	if err := base().
		Joins("JOIN provinces ON provinces.id = leads.province_id").
		Select("provinces.name AS province_name, COUNT(*) AS count, COUNT(*) FILTER (WHERE leads.status = 'HANDOFF_ODOO') AS handoff").
		Group("provinces.name").
		Order("count DESC").
		Scan(&provRows).Error; err != nil {
		return nil, err
	}
	var cityRows []regionAgg
	if err := base().
		Joins("JOIN cities ON cities.id = leads.city_id").
		Joins("JOIN provinces ON provinces.id = cities.province_id").
		Select("provinces.name AS province_name, cities.name AS city_name, COUNT(*) AS count, COUNT(*) FILTER (WHERE leads.status = 'HANDOFF_ODOO') AS handoff").
		Group("provinces.name, cities.name").
		Scan(&cityRows).Error; err != nil {
		return nil, err
	}
	citiesByProvince := make(map[string][]regionAgg, len(cityRows))
	for _, c := range cityRows {
		citiesByProvince[c.ProvinceName] = append(citiesByProvince[c.ProvinceName], c)
	}
	var regions []dto.RegionRow
	for _, p := range provRows {
		regions = append(regions, dto.RegionRow{Level: "province", Name: p.ProvinceName, LeadCount: p.Count, HandoffCount: p.Handoff})
		cities := citiesByProvince[p.ProvinceName]
		sort.SliceStable(cities, func(a, b int) bool { return cities[a].Count > cities[b].Count })
		for _, c := range cities {
			regions = append(regions, dto.RegionRow{Level: "city", Name: c.CityName, Parent: p.ProvinceName, LeadCount: c.Count, HandoffCount: c.Handoff})
		}
	}

	// Intel Kompetitor
	avgAll, err := s.avgPricePerMbps(base, "")
	if err != nil {
		return nil, err
	}
	avgDedicated, err := s.avgPricePerMbps(base, "Dedicated")
	if err != nil {
		return nil, err
	}
	avgBroadband, err := s.avgPricePerMbps(base, "Broadband")
	if err != nil {
		return nil, err
	}
	stats := dto.CompetitorStats{
		AvgPricePerMbps:          avgAll,
		AvgPricePerMbpsDedicated: avgDedicated,
		AvgPricePerMbpsBroadband: avgBroadband,
	}
	var ispRows []struct {
		Name  string
		Count int64
	}
	if err := base().
		Where("leads.existing_isp IS NOT NULL AND leads.existing_isp <> ''").
		Select("leads.existing_isp AS name, COUNT(*) AS count").
		Group("leads.existing_isp").
		Order("count DESC").
		Scan(&ispRows).Error; err != nil {
		return nil, err
	}
	isps := make([]dto.IspRow, len(ispRows))
	for i, r := range ispRows {
		isps[i] = dto.IspRow{Name: r.Name, Count: r.Count}
	}

	return &dto.DashboardSegmentsResponse{
		AnyHandoff:     anyHandoff,
		Sources:        sources,
		Regions:        regions,
		Competitor:     stats,
		Isps:           isps,
		BusinessFields: businessFields,
	}, nil
}
