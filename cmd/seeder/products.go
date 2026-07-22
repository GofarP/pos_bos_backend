package main

import (
	"database/sql"
	"log"
	"time"
)

func seedProducts(db *sql.DB) {
	log.Println("Menjalankan seeder untuk products...")
	products := []struct {
		SKU         string
		Name        string
		Description string
		Price       float64
		Stock       int
		Category    string
	}{
		{"NGR001", "Nasi Goreng Spesial", "Nasi goreng ayam dengan telur ceplok", 25000, 100, "Food"},
		{"MGR001", "Mie Goreng", "Mie goreng ayam", 20000, 50, "Food"},
		{"EST001", "Es Teh Manis", "Teh manis dingin", 5000, 200, "Beverage"},
		{"ESJ001", "Es Jeruk", "Jeruk peras dingin", 10000, 100, "Beverage"},
		{"KRP001", "Kerupuk Udang", "Kerupuk udang renyah", 2000, 500, "Snack"},
	}

	now := time.Now()
	for _, p := range products {
		var catID int
		err := db.QueryRow("SELECT id FROM categories WHERE name = ?", p.Category).Scan(&catID)
		if err == nil {
			_, err := db.Exec("INSERT IGNORE INTO products (category_id, sku, name, description, price, stock, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				catID, p.SKU, p.Name, p.Description, p.Price, p.Stock, now, now)
			if err != nil {
				log.Fatalf("Gagal insert product %s: %v", p.SKU, err)
			}
		} else {
			log.Printf("Gagal mencari category %s untuk product %s: %v\n", p.Category, p.SKU, err)
		}
	}
	log.Println("Products berhasil di-seed.")
}
