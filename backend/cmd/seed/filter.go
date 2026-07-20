package main

import (
	"fmt"
	"strconv"

	"fiber-api-boilerplate/internal/models"
)

// SeedStats counts what BuildSeedData skipped or pruned at each level, broken
// down by reason, so a bad export is diagnosable from the log alone.
type SeedStats struct {
	CitiesSkippedBadParent    int
	DistrictsSkippedBadParent int
	DistrictsPrunedDeadEnd    int
	ZipsSkippedContinuation   int
	ZipsSkippedBadParent      int
	ZipsPrunedDeadEnd         int
	VillagesSkippedBadParent  int
	VillagesZipNulled         int
	CitiesPrunedDeadEnd       int
	ProvincesPrunedDeadEnd    int
}

// SeedResult is BuildSeedData's output: the final, referentially-consistent
// rows ready to upsert, plus stats on what was filtered out along the way.
type SeedResult struct {
	Provinces []models.Province
	Cities    []models.City
	Districts []models.District
	Zips      []models.Zip
	Villages  []models.Village
	Stats     SeedStats
}

// mustInt parses a row's own identity field (ID). A failure here means the
// CSV's structure itself is broken, not a data-quality issue - unlike parent-
// reference fields (see parseParentID), it's a hard error.
func mustInt(level, col, s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%s: bad int in %s: %q", level, col, s)
	}
	return n, nil
}

// parseParentID parses a field that references a parent row. Empty or
// unparseable is never fatal here - the caller treats it as "no valid
// parent" and skips or nulls accordingly.
func parseParentID(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

// BuildSeedData runs the two-pass filter/prune pipeline. Pass 1 keeps each
// level's rows only if their parent survived the level above (or, for
// provinces/cities, if they're Indonesia). Pass 2 prunes dead-end branches
// bottom-up (a district with no villages, a city with no surviving
// districts, a province with no surviving cities) and nulls - never drops -
// a village's Zip/ID if that zip didn't survive.
func BuildSeedData(
	provinceRows []map[string]string,
	cityRows []map[string]string,
	districtRows []map[string]string,
	zipRows []map[string]string,
	villageRows []map[string]string,
) (SeedResult, error) {
	var stats SeedStats

	// --- Pass 1, level 1: provinces (Indonesia only) ---
	var provinces []models.Province
	validProvinceIDs := map[int]bool{}
	for _, r := range provinceRows {
		if r["Country/ID"] != "100" {
			continue
		}
		id, err := mustInt("provinces", "ID", r["ID"])
		if err != nil {
			return SeedResult{}, err
		}
		code := r["State Code"]
		provinces = append(provinces, models.Province{
			ID:         id,
			ExternalID: r["External ID"],
			Code:       &code,
			Name:       r["State Name"],
		})
		validProvinceIDs[id] = true
	}

	// --- Pass 1, level 2: cities (Indonesia + valid province parent) ---
	var cities []models.City
	validCityIDs := map[int]bool{}
	for _, r := range cityRows {
		if r["Country/ID"] != "100" {
			stats.CitiesSkippedBadParent++
			continue
		}
		provinceID, ok := parseParentID(r["State/ID"])
		if !ok || !validProvinceIDs[provinceID] {
			stats.CitiesSkippedBadParent++
			continue
		}
		id, err := mustInt("cities", "ID", r["ID"])
		if err != nil {
			return SeedResult{}, err
		}
		cities = append(cities, models.City{
			ID:         id,
			ExternalID: r["External ID"],
			ProvinceID: provinceID,
			Name:       r["Name"],
		})
		validCityIDs[id] = true
	}

	// --- Pass 1, level 3: districts (valid city parent) ---
	var candidateDistricts []models.District
	candidateDistrictIDs := map[int]bool{}
	for _, r := range districtRows {
		cityID, ok := parseParentID(r["City/ID"])
		if !ok || !validCityIDs[cityID] {
			stats.DistrictsSkippedBadParent++
			continue
		}
		id, err := mustInt("districts", "ID", r["ID"])
		if err != nil {
			return SeedResult{}, err
		}
		candidateDistricts = append(candidateDistricts, models.District{
			ID:         id,
			ExternalID: r["External ID"],
			CityID:     cityID,
			Name:       r["Kecamatan"],
		})
		candidateDistrictIDs[id] = true
	}

	// --- Pass 1, level 4: zips (valid city AND valid candidate district) ---
	var candidateZips []models.Zip
	for _, r := range zipRows {
		if r["ID"] == "" {
			stats.ZipsSkippedContinuation++ // Odoo "continuation" row, expected, not a data error
			continue
		}
		cityID, cityOK := parseParentID(r["City/ID"])
		districtID, distOK := parseParentID(r["Kecamatan/ID"])
		if !cityOK || !validCityIDs[cityID] || !distOK || !candidateDistrictIDs[districtID] {
			stats.ZipsSkippedBadParent++
			continue
		}
		id, err := mustInt("zips", "ID", r["ID"])
		if err != nil {
			return SeedResult{}, err
		}
		candidateZips = append(candidateZips, models.Zip{
			ID:         id,
			ExternalID: r["External ID"],
			Code:       r["ZIP"],
			CityID:     cityID,
			DistrictID: districtID,
		})
	}

	// --- Pass 1, level 5: villages (valid candidate district) ---
	var candidateVillages []models.Village
	villageCountByDistrict := map[int]int{}
	for _, r := range villageRows {
		districtID, ok := parseParentID(r["Kecamatan/ID"])
		if !ok || !candidateDistrictIDs[districtID] {
			stats.VillagesSkippedBadParent++
			continue
		}
		id, err := mustInt("villages", "ID", r["ID"])
		if err != nil {
			return SeedResult{}, err
		}
		var zipID *int
		if raw := r["Zip/ID"]; raw != "" {
			if zid, zok := parseParentID(raw); zok {
				zipID = &zid
			} else {
				stats.VillagesZipNulled++
			}
		}
		candidateVillages = append(candidateVillages, models.Village{
			ID:         id,
			ExternalID: r["External ID"],
			DistrictID: districtID,
			ZipID:      zipID,
			Name:       r["Kelurahan"],
		})
		villageCountByDistrict[districtID]++
	}

	// --- Pass 2, step 1: prune districts with zero villages ---
	validDistrictIDs := map[int]bool{}
	var districts []models.District
	for _, d := range candidateDistricts {
		if villageCountByDistrict[d.ID] == 0 {
			stats.DistrictsPrunedDeadEnd++
			continue
		}
		validDistrictIDs[d.ID] = true
		districts = append(districts, d)
	}

	// cities with at least one surviving district (needed below to re-filter
	// zips, and reused by step 4 to prune cities - both checks mean the same
	// thing and are computed from the same districts slice)
	cityHasSurvivingDistrict := map[int]bool{}
	for _, d := range districts {
		cityHasSurvivingDistrict[d.CityID] = true
	}

	// --- Pass 2, step 2: re-filter zips against the pruned district set ---
	// (a zip can pass Pass 1's per-parent check yet still dangle if its
	// district was pruned here as a dead end). City/ID and Kecamatan/ID are
	// independent fields in the source CSV, so also check the zip's own
	// City/ID directly - a district surviving doesn't guarantee the zip's
	// (possibly inconsistent) city did too.
	zips := make([]models.Zip, 0, len(candidateZips))
	validZipIDs := map[int]bool{}
	for _, z := range candidateZips {
		if !validDistrictIDs[z.DistrictID] || !cityHasSurvivingDistrict[z.CityID] {
			stats.ZipsPrunedDeadEnd++
			continue
		}
		zips = append(zips, z)
		validZipIDs[z.ID] = true
	}

	// --- Pass 2, step 3: null (never drop) a village's Zip/ID if its zip didn't survive ---
	villages := make([]models.Village, 0, len(candidateVillages))
	for _, v := range candidateVillages {
		if v.ZipID != nil && !validZipIDs[*v.ZipID] {
			v.ZipID = nil
			stats.VillagesZipNulled++
		}
		villages = append(villages, v)
	}

	// --- Pass 2, step 4: prune cities with zero surviving districts ---
	// (cityHasSurvivingDistrict was already computed above, before step 2)
	finalCities := make([]models.City, 0, len(cities))
	for _, c := range cities {
		if !cityHasSurvivingDistrict[c.ID] {
			stats.CitiesPrunedDeadEnd++
			continue
		}
		finalCities = append(finalCities, c)
	}

	// --- Pass 2, step 5: prune provinces with zero surviving cities ---
	provinceHasSurvivingCity := map[int]bool{}
	for _, c := range finalCities {
		provinceHasSurvivingCity[c.ProvinceID] = true
	}
	finalProvinces := make([]models.Province, 0, len(provinces))
	for _, p := range provinces {
		if !provinceHasSurvivingCity[p.ID] {
			stats.ProvincesPrunedDeadEnd++
			continue
		}
		finalProvinces = append(finalProvinces, p)
	}

	return SeedResult{
		Provinces: finalProvinces,
		Cities:    finalCities,
		Districts: districts,
		Zips:      zips,
		Villages:  villages,
		Stats:     stats,
	}, nil
}
