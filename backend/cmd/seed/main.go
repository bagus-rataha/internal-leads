package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/models"

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

// atoi parses an Odoo integer id; empty/invalid is a hard error since these are
// required foreign keys or primary keys.
func atoi(file, col, s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("%s: bad int in %q: %q", file, col, s)
	}
	return n
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

	// provinces <- res.country.state, only Indonesia (Country/ID == 100)
	const provFile = "Country state (res.country.state).csv"
	var provinces []models.Province
	for _, r := range csvRows(provFile) {
		if r["Country/ID"] != "100" {
			continue
		}
		code := r["State Code"]
		provinces = append(provinces, models.Province{
			ID:         atoi(provFile, "ID", r["ID"]),
			ExternalID: r["External ID"],
			Code:       &code,
			Name:       r["State Name"],
		})
	}
	upsert(db, provinces)

	// cities <- res.city, only Indonesia
	const cityFile = "City (res.city).csv"
	var cities []models.City
	for _, r := range csvRows(cityFile) {
		if r["Country/ID"] != "100" {
			continue
		}
		cities = append(cities, models.City{
			ID:         atoi(cityFile, "ID", r["ID"]),
			ExternalID: r["External ID"],
			ProvinceID: atoi(cityFile, "State/ID", r["State/ID"]),
			Name:       r["Name"],
		})
	}
	upsert(db, cities)

	// districts <- location.district
	const distFile = "Kecamatan (location.district).csv"
	var districts []models.District
	for _, r := range csvRows(distFile) {
		districts = append(districts, models.District{
			ID:         atoi(distFile, "ID", r["ID"]),
			ExternalID: r["External ID"],
			CityID:     atoi(distFile, "City/ID", r["City/ID"]),
			Name:       r["Kecamatan"],
		})
	}
	upsert(db, districts)

	// zips <- res.city.zip; skip Odoo continuation rows (empty ID)
	const zipFile = "Citylocations completion object (res.city.zip).csv"
	var zips []models.Zip
	skippedZips := 0
	for _, r := range csvRows(zipFile) {
		if r["ID"] == "" {
			skippedZips++
			continue
		}
		zips = append(zips, models.Zip{
			ID:         atoi(zipFile, "ID", r["ID"]),
			ExternalID: r["External ID"],
			Code:       r["ZIP"],
			CityID:     atoi(zipFile, "City/ID", r["City/ID"]),
			DistrictID: atoi(zipFile, "Kecamatan/ID", r["Kecamatan/ID"]),
		})
	}
	upsert(db, zips)

	// villages <- location.sub.district; Zip/ID is nullable
	const villFile = "Kelurahan (location.sub.district).csv"
	var villages []models.Village
	for _, r := range csvRows(villFile) {
		var zipID *int
		if s := r["Zip/ID"]; s != "" {
			id := atoi(villFile, "Zip/ID", s)
			zipID = &id
		}
		villages = append(villages, models.Village{
			ID:         atoi(villFile, "ID", r["ID"]),
			ExternalID: r["External ID"],
			DistrictID: atoi(villFile, "Kecamatan/ID", r["Kecamatan/ID"]),
			ZipID:      zipID,
			Name:       r["Kelurahan"],
		})
	}
	upsert(db, villages)

	log.Printf("seeded: provinces=%d cities=%d districts=%d zips=%d (skipped %d continuation) villages=%d",
		len(provinces), len(cities), len(districts), len(zips), skippedZips, len(villages))
}
