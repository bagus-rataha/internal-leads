package dto

import (
	"time"

	"github.com/google/uuid"
)

// DashboardQuery is the parsed, typed form of every dashboard endpoint's
// shared query params (date_from, date_to, team_id, owner_id). DateFrom/
// DateTo are always populated - parseDashboardQuery defaults them to a
// trailing 30-day window when the caller omits both, per ARCHITECTURE.md
// §9's "Default rentang: 30 hari terakhir".
type DashboardQuery struct {
	DateFrom time.Time
	DateTo   time.Time
	TeamID   *uuid.UUID
	OwnerID  *uuid.UUID
}

// MetricCard is a single dashboard/summary metric: a current-period value
// plus its period-over-period change. ChangePct is nil when the comparison
// period's value was 0 (ARCHITECTURE.md §10's zero-state contract) - the
// frontend derives "—" (both zero) vs "baru" (newly nonzero) from Value
// together with a nil ChangePct; it never computes a percentage itself.
type MetricCard struct {
	Value     int64 `json:"value"`
	ChangePct *int  `json:"change_pct"`
}

// FunnelStage is one row of the conversion funnel snapshot (widget B). Pct
// is nil when the funnel's own total is 0 - a caller with zero leads in
// scope has no meaningful percentage to show.
type FunnelStage struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
	Pct   *int   `json:"pct"`
}

// FunnelResponse is the widget B payload: a snapshot of the caller's scope
// (after team_id/owner_id narrowing), NOT bound to date_from/date_to - see
// the spec for why (date-windowing by creation date biases a funnel toward
// "not enough time to convert yet" for anything created recently).
type FunnelResponse struct {
	Stages         []FunnelStage `json:"stages"`
	BaruToFuPct    *int          `json:"baru_to_fu_pct"`
	FuToHandoffPct *int          `json:"fu_to_handoff_pct"`
	LostCount      int64         `json:"lost_count"`
	LostPct        *int          `json:"lost_pct"`
}

// DashboardSummaryResponse is GET /dashboard/summary's payload: widgets A + B.
type DashboardSummaryResponse struct {
	LeadBaru  MetricCard     `json:"lead_baru"`
	FollowUp  MetricCard     `json:"follow_up"`
	Handoff   MetricCard     `json:"handoff"`
	Terlantar MetricCard     `json:"terlantar"`
	Funnel    FunnelResponse `json:"funnel"`
}

// ActivityBucket is one day of widget C's trend chart (GET /dashboard/activity).
type ActivityBucket struct {
	Date     string `json:"date"` // YYYY-MM-DD
	LeadBaru int64  `json:"lead_baru"`
	FollowUp int64  `json:"follow_up"`
}

type DashboardActivityResponse struct {
	Buckets []ActivityBucket `json:"buckets"`
}

// StaleLeadRow is one row of widget E's table (GET /dashboard/stale-leads).
type StaleLeadRow struct {
	Code        string `json:"code"`
	CompanyName string `json:"company_name"`
	OwnerName   string `json:"owner_name"`
	DaysSince   int    `json:"days_since"`
}

type DashboardStaleLeadsResponse struct {
	Items []StaleLeadRow `json:"items"`
}

// SalesActivityRow is one row of widget D's table (GET /dashboard/sales-activity,
// 403 for SALES).
type SalesActivityRow struct {
	UserID       uuid.UUID  `json:"user_id"`
	Name         string     `json:"name"`
	TeamName     *string    `json:"team_name"`
	LeadBaru     int64      `json:"lead_baru"`
	FollowUp     int64      `json:"follow_up"`
	AvgFuPerLead float64    `json:"avg_fu_per_lead"`
	Terlantar    int64      `json:"terlantar"`
	Handoff      int64      `json:"handoff"`
	ConvPct      *int       `json:"conv_pct"`
	LastActivity *time.Time `json:"last_activity"`
	AttentionTag *string    `json:"attention_tag"`
}

type DashboardSalesActivityResponse struct {
	Items []SalesActivityRow `json:"items"`
}

// SegmentRow is one row of widgets F (Sumber Lead) and I (Bidang Usaha) -
// same shape, keyed by a different grouping field.
type SegmentRow struct {
	Name          string `json:"name"`
	Count         int64  `json:"count"`
	ConversionPct *int   `json:"conversion_pct"`
	Warn          bool   `json:"warn"`
}

// RegionRow is one row of widget G (Penetrasi Wilayah) - a flat list the
// frontend groups by Level for display (a "city" row's Parent names its
// province).
type RegionRow struct {
	Level        string `json:"level"` // "province" | "city"
	Name         string `json:"name"`
	Parent       string `json:"parent,omitempty"`
	LeadCount    int64  `json:"lead_count"`
	HandoffCount int64  `json:"handoff_count"`
}

// CompetitorStats is widget H's 3 stat tiles - nil (not 0) when no
// qualifying lead exists, matching the null-not-zero zero-state contract.
type CompetitorStats struct {
	AvgPricePerMbps          *float64 `json:"avg_price_per_mbps"`
	AvgPricePerMbpsDedicated *float64 `json:"avg_price_per_mbps_dedicated"`
	AvgPricePerMbpsBroadband *float64 `json:"avg_price_per_mbps_broadband"`
}

// IspRow is one row of widget H's existing-ISP distribution table.
type IspRow struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// DashboardSegmentsResponse is GET /dashboard/segments's payload: widgets F+G+H+I.
type DashboardSegmentsResponse struct {
	AnyHandoff     bool            `json:"any_handoff"`
	Sources        []SegmentRow    `json:"sources"`
	Regions        []RegionRow     `json:"regions"`
	Competitor     CompetitorStats `json:"competitor"`
	Isps           []IspRow        `json:"isps"`
	BusinessFields []SegmentRow    `json:"business_fields"`
}
