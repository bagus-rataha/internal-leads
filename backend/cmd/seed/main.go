package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"

	"fiber-api-boilerplate/internal/config"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// seedDir is where the Odoo wilayah CSV exports live. Overridable via SEED_DIR
// for the real full export; defaults to seeds/ relative to the backend dir.
func seedDir() string {
	if d := os.Getenv("SEED_DIR"); d != "" {
		return d
	}
	return "seeds"
}

// csvRows reads a CSV and returns each data row as a name->value map keyed by
// the header row, so columns are matched by name (a reorder in the real export
// won't silently corrupt data).
func csvRows(name string) []map[string]string {
	path := filepath.Join(seedDir(), name)
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		log.Fatalf("read header %s: %v", path, err)
	}

	var rows []map[string]string
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("read %s: %v", path, err)
		}
		m := make(map[string]string, len(header))
		for i, h := range header {
			if i < len(rec) {
				m[h] = rec[i]
			}
		}
		rows = append(rows, m)
	}
	return rows
}

// upsert writes a slice in batches, updating on external_id conflict so the
// command is idempotent (safe to re-run against real data).
func upsert[T any](db *gorm.DB, rows []T) {
	if len(rows) == 0 {
		return
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "external_id"}},
		UpdateAll: true,
	}).CreateInBatches(rows, 500).Error; err != nil {
		log.Fatalf("upsert: %v", err)
	}
}

func main() {
	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)

	provinceRows := csvRows("Country state (res.country.state).csv")
	cityRows := csvRows("City (res.city).csv")
	districtRows := csvRows("Kecamatan (location.district).csv")
	zipRows := csvRows("Citylocations completion object (res.city.zip).csv")
	villageRows := csvRows("Kelurahan (location.sub.district).csv")

	result, err := BuildSeedData(provinceRows, cityRows, districtRows, zipRows, villageRows)
	if err != nil {
		log.Fatal(err)
	}

	upsert(db, result.Provinces)
	upsert(db, result.Cities)
	upsert(db, result.Districts)
	upsert(db, result.Zips)
	upsert(db, result.Villages)

	s := result.Stats
	log.Printf(
		"seeded: provinces=%d (pruned %d) cities=%d (skipped %d bad-parent, pruned %d) "+
			"districts=%d (skipped %d bad-parent, pruned %d) "+
			"zips=%d (skipped %d continuation, %d bad-parent, pruned %d) "+
			"villages=%d (skipped %d bad-parent, %d zip-nulled)",
		len(result.Provinces), s.ProvincesPrunedDeadEnd,
		len(result.Cities), s.CitiesSkippedBadParent, s.CitiesPrunedDeadEnd,
		len(result.Districts), s.DistrictsSkippedBadParent, s.DistrictsPrunedDeadEnd,
		len(result.Zips), s.ZipsSkippedContinuation, s.ZipsSkippedBadParent, s.ZipsPrunedDeadEnd,
		len(result.Villages), s.VillagesSkippedBadParent, s.VillagesZipNulled,
	)
}
