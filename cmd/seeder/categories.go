package main

import (
	"database/sql"
	"log"
	"time"
)

func seedCategories(db *sql.DB) {
	log.Println("Menjalankan seeder untuk categories...")

	categories := []struct {
		Name        string
		Description string
	}{
		{"Food", "Makanan Utama"},
		{"Beverage", "Minuman Ringan"},
		{"Snack", "Makanan Ringan"},
	}

	now := time.Now()
	for _, c := range categories {
		_, err := db.Exec("INSERT IGNORE INTO categories (name, description, created_at, updated_at) VALUES (?, ?, ?, ?)", c.Name, c.Description, now, now)
		if err != nil {
			log.Fatalf("Gagal insert category %s: %v", c.Name, err)
		}
	}
	log.Println("✅ Categories berhasil di-seed.")
}
