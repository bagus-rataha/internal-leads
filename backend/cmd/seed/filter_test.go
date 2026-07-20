package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validChain returns one complete, fully-valid province -> city -> district
// -> zip -> village chain. Each test clones these slices and mutates exactly
// what it needs to isolate, so every test proves one specific rule against a
// realistic, otherwise-valid dataset.
func validChain() (provinces, cities, districts, zips, villages []map[string]string) {
	provinces = []map[string]string{
		{"Country/ID": "100", "ID": "1", "External ID": "prov_1", "State Code": "JK", "State Name": "DKI Jakarta"},
	}
	cities = []map[string]string{
		{"Country/ID": "100", "ID": "10", "External ID": "city_10", "State/ID": "1", "Name": "Jakarta Selatan"},
	}
	districts = []map[string]string{
		{"ID": "100", "External ID": "dist_100", "City/ID": "10", "Kecamatan": "Kebayoran Baru"},
	}
	zips = []map[string]string{
		{"ID": "1000", "External ID": "zip_1000", "ZIP": "12110", "City/ID": "10", "Kecamatan/ID": "100"},
	}
	villages = []map[string]string{
		{"ID": "10000", "External ID": "vill_10000", "Kecamatan/ID": "100", "Zip/ID": "1000", "Kelurahan": "Gunung"},
	}
	return
}

func TestBuildSeedData_ValidChain_AllRowsSurvive(t *testing.T) {
	prov, city, dist, zip, vill := validChain()

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	require.Len(t, result.Provinces, 1)
	require.Len(t, result.Cities, 1)
	require.Len(t, result.Districts, 1)
	require.Len(t, result.Zips, 1)
	require.Len(t, result.Villages, 1)
	assert.Equal(t, 1, result.Provinces[0].ID)
	assert.Equal(t, 10, result.Cities[0].ID)
	assert.Equal(t, 1, result.Cities[0].ProvinceID)
	assert.Equal(t, 100, result.Districts[0].ID)
	assert.Equal(t, 10, result.Districts[0].CityID)
	assert.Equal(t, 1000, result.Zips[0].ID)
	assert.Equal(t, 10000, result.Villages[0].ID)
	require.NotNil(t, result.Villages[0].ZipID)
	assert.Equal(t, 1000, *result.Villages[0].ZipID)
	assert.Equal(t, SeedStats{}, result.Stats, "a fully valid chain should not skip or prune anything")
}

func TestBuildSeedData_NonIndonesiaProvince_DropsWholeChain(t *testing.T) {
	prov, city, dist, zip, vill := validChain()
	prov[0]["Country/ID"] = "42" // foreign country code

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	assert.Empty(t, result.Provinces)
	assert.Empty(t, result.Cities, "city's State/ID pointed at a province that no longer exists")
	assert.Equal(t, 1, result.Stats.CitiesSkippedBadParent)
	assert.Equal(t, 1, result.Stats.DistrictsSkippedBadParent, "district's City/ID pointed at a city that no longer exists")
	assert.Equal(t, 1, result.Stats.ZipsSkippedBadParent)
	assert.Equal(t, 1, result.Stats.VillagesSkippedBadParent)
}

func TestBuildSeedData_ForeignCity_SkippedAndCascades(t *testing.T) {
	prov, city, dist, zip, vill := validChain()
	city[0]["Country/ID"] = "42"

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	require.Len(t, result.Provinces, 0, "province's only city was dropped, so it has no surviving city")
	assert.Equal(t, 1, result.Stats.CitiesSkippedBadParent)
	assert.Equal(t, 1, result.Stats.DistrictsSkippedBadParent)
}

func TestBuildSeedData_CityWithBadStateID_Skipped(t *testing.T) {
	prov, city, dist, zip, vill := validChain()
	city[0]["State/ID"] = "not-a-number"

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	assert.Empty(t, result.Cities)
	assert.Equal(t, 1, result.Stats.CitiesSkippedBadParent)
}

func TestBuildSeedData_DistrictWithForeignCityParent_SkippedAndCascades(t *testing.T) {
	prov, city, dist, zip, vill := validChain()
	dist[0]["City/ID"] = "9999" // no such city

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Stats.DistrictsSkippedBadParent)
	assert.Equal(t, 1, result.Stats.ZipsSkippedBadParent, "zip's Kecamatan/ID pointed at a district that was skipped")
	assert.Equal(t, 1, result.Stats.VillagesSkippedBadParent)
	assert.Empty(t, result.Cities, "city has zero surviving districts")
	assert.Empty(t, result.Provinces)
}

func TestBuildSeedData_DeadEndDistrict_PrunedButCityAndProvinceSurviveViaSibling(t *testing.T) {
	prov, city, dist, zip, vill := validChain()
	// Add a second district under the same city with zero villages - it
	// should be pruned as a dead end, while the city/province survive
	// because the FIRST district still has a village.
	dist = append(dist, map[string]string{"ID": "101", "External ID": "dist_101", "City/ID": "10", "Kecamatan": "Dead End"})

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	require.Len(t, result.Districts, 1, "only the district with a village survives")
	assert.Equal(t, 100, result.Districts[0].ID)
	assert.Equal(t, 1, result.Stats.DistrictsPrunedDeadEnd)
	assert.Len(t, result.Cities, 1, "city survives via its other, non-dead-end district")
	assert.Len(t, result.Provinces, 1)
}

func TestBuildSeedData_DeadEndDistrict_CascadesToCityAndProvinceWhenItWasTheOnlyOne(t *testing.T) {
	prov, city, _, _, _ := validChain()
	dist := []map[string]string{
		{"ID": "100", "External ID": "dist_100", "City/ID": "10", "Kecamatan": "Kebayoran Baru"},
	}
	// No villages at all reference district 100 - it's a dead end, and it's
	// the city's only district, so the city and province die too.
	result, err := BuildSeedData(prov, city, dist, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Stats.DistrictsPrunedDeadEnd)
	assert.Equal(t, 1, result.Stats.CitiesPrunedDeadEnd)
	assert.Equal(t, 1, result.Stats.ProvincesPrunedDeadEnd)
	assert.Empty(t, result.Districts)
	assert.Empty(t, result.Cities)
	assert.Empty(t, result.Provinces)
}

func TestBuildSeedData_ZipWithBadParent_Skipped(t *testing.T) {
	prov, city, dist, zip, vill := validChain()
	zip[0]["Kecamatan/ID"] = "9999" // no such district

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	assert.Empty(t, result.Zips)
	assert.Equal(t, 1, result.Stats.ZipsSkippedBadParent)
	// the village's Zip/ID (1000) never existed as a surviving zip, so it's nulled
	require.Len(t, result.Villages, 1)
	assert.Nil(t, result.Villages[0].ZipID)
	assert.Equal(t, 1, result.Stats.VillagesZipNulled)
}

func TestBuildSeedData_ZipContinuationRow_SkippedSilently(t *testing.T) {
	prov, city, dist, _, vill := validChain()
	zip := []map[string]string{
		{"ID": "", "External ID": "", "ZIP": "", "City/ID": "", "Kecamatan/ID": ""}, // Odoo continuation row
	}

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	assert.Empty(t, result.Zips)
	assert.Equal(t, 1, result.Stats.ZipsSkippedContinuation)
	assert.Equal(t, 0, result.Stats.ZipsSkippedBadParent, "a continuation row is expected, not a data error")
}

func TestBuildSeedData_ZipPrunedAsDeadEnd_VillageZipIDNulledNotDropped(t *testing.T) {
	prov, city, _, _, _ := validChain()
	// Two districts under the city: 100 has a village (survives), 101 has no
	// village (dead end, pruned). A zip references district 101 - it passes
	// Pass 1 (101 is a valid candidate district at that point) but must be
	// pruned in Pass 2 once 101 dies. A village under district 100 points its
	// Zip/ID at that doomed zip - it must survive with ZipID nulled.
	dist := []map[string]string{
		{"ID": "100", "External ID": "dist_100", "City/ID": "10", "Kecamatan": "Kebayoran Baru"},
		{"ID": "101", "External ID": "dist_101", "City/ID": "10", "Kecamatan": "Dead End"},
	}
	zip := []map[string]string{
		{"ID": "1000", "External ID": "zip_1000", "ZIP": "12110", "City/ID": "10", "Kecamatan/ID": "101"},
	}
	vill := []map[string]string{
		{"ID": "10000", "External ID": "vill_10000", "Kecamatan/ID": "100", "Zip/ID": "1000", "Kelurahan": "Gunung"},
	}

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Stats.DistrictsPrunedDeadEnd)
	assert.Equal(t, 1, result.Stats.ZipsPrunedDeadEnd)
	assert.Empty(t, result.Zips)
	require.Len(t, result.Villages, 1, "the village itself is fine - only its zip link is broken")
	assert.Nil(t, result.Villages[0].ZipID)
	assert.Equal(t, 1, result.Stats.VillagesZipNulled)
}

func TestBuildSeedData_VillageWithEmptyZipID_NotCountedAsNulled(t *testing.T) {
	prov, city, dist, zip, vill := validChain()
	vill[0]["Zip/ID"] = "" // most villages have no Odoo zip mapping at all - this is normal, not an error

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	require.Len(t, result.Villages, 1)
	assert.Nil(t, result.Villages[0].ZipID)
	assert.Equal(t, 0, result.Stats.VillagesZipNulled, "an empty Zip/ID to begin with is not a correction")
}

func TestBuildSeedData_VillageWithForeignDistrictParent_Skipped(t *testing.T) {
	prov, city, dist, zip, vill := validChain()
	vill[0]["Kecamatan/ID"] = "9999"

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	assert.Empty(t, result.Villages)
	assert.Equal(t, 1, result.Stats.VillagesSkippedBadParent)
	// its district now has zero villages, so it becomes a dead end
	assert.Equal(t, 1, result.Stats.DistrictsPrunedDeadEnd)
}

func TestBuildSeedData_MalformedOwnID_ReturnsError(t *testing.T) {
	prov, city, dist, zip, vill := validChain()
	prov[0]["ID"] = "not-a-number"

	_, err := BuildSeedData(prov, city, dist, zip, vill)

	require.Error(t, err)
}

func TestBuildSeedData_ZipWithMismatchedCityAndDistrict_DroppedWhenCityDies(t *testing.T) {
	prov, city, _, _, _ := validChain()
	// Two cities, each with one district. City 20's district has a village
	// (survives); city 10's district has none (dead end, city 10 dies in
	// Pass 2). A zip claims City/ID=10 but Kecamatan/ID=200 (city 20's
	// district) - an internally inconsistent source row. It must not survive
	// with a city_id (10) that isn't in the final result.
	city = append(city, map[string]string{"Country/ID": "100", "ID": "20", "External ID": "city_20", "State/ID": "1", "Name": "Second City"})
	dist := []map[string]string{
		{"ID": "100", "External ID": "dist_100", "City/ID": "10", "Kecamatan": "Dead End"},
		{"ID": "200", "External ID": "dist_200", "City/ID": "20", "Kecamatan": "Kebayoran Baru"},
	}
	zip := []map[string]string{
		{"ID": "1000", "External ID": "zip_1000", "ZIP": "12110", "City/ID": "10", "Kecamatan/ID": "200"},
	}
	vill := []map[string]string{
		{"ID": "10000", "External ID": "vill_10000", "Kecamatan/ID": "200", "Zip/ID": "", "Kelurahan": "Gunung"},
	}

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	assert.Empty(t, result.Zips, "the zip's own City/ID (10) never survives, even though its district (200) does")
	assert.Equal(t, 1, result.Stats.ZipsPrunedDeadEnd)
	// city 10 must not appear in the final cities either
	for _, c := range result.Cities {
		assert.NotEqual(t, 10, c.ID)
	}
}

func TestBuildSeedData_EveryZipCityIDIsInFinalCities(t *testing.T) {
	prov, city, dist, zip, vill := validChain()

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	finalCityIDs := map[int]bool{}
	for _, c := range result.Cities {
		finalCityIDs[c.ID] = true
	}
	for _, z := range result.Zips {
		assert.True(t, finalCityIDs[z.CityID], "zip %d references city %d which is not in the final result", z.ID, z.CityID)
	}
}

func TestBuildSeedData_VillageWithUnparseableZipID_CountedAsNulled(t *testing.T) {
	prov, city, dist, zip, vill := validChain()
	vill[0]["Zip/ID"] = "not-a-number"

	result, err := BuildSeedData(prov, city, dist, zip, vill)

	require.NoError(t, err)
	require.Len(t, result.Villages, 1)
	assert.Nil(t, result.Villages[0].ZipID)
	assert.Equal(t, 1, result.Stats.VillagesZipNulled, "an unparseable Zip/ID is a real correction, unlike a simply-empty one")
}
