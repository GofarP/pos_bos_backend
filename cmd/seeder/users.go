package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func seedAdminUser(db *sql.DB) {
	log.Println("Menjalankan seeder untuk tabel users...")

	adminName := "Super Admin"
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@posbos.com"
	}
	
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Gagal melakukan hashing password: %v", err)
	}

	var adminCount int
	db.QueryRow("SELECT COUNT(id) FROM users WHERE email = ?", adminEmail).Scan(&adminCount)

	now := time.Now()
	if adminCount == 0 {
		query := `INSERT INTO users (name, email, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
		_, err = db.Exec(query, adminName, adminEmail, string(hashedPassword), now, now)
		if err != nil {
			log.Fatalf("Gagal memasukkan data admin: %v", err)
		}
		log.Println("✅ Sukses! Akun admin telah dibuat (admin@posbos.com / admin123).")
	}

	cashierName := "Kasir Toko"
	cashierEmail := "kasir@posbos.com"
	cashierPassword := "kasir123"

	var cashierCount int
	db.QueryRow("SELECT COUNT(id) FROM users WHERE email = ?", cashierEmail).Scan(&cashierCount)

	if cashierCount == 0 {
		hashedCashierPassword, err := bcrypt.GenerateFromPassword([]byte(cashierPassword), bcrypt.DefaultCost)
		if err == nil {
			query := `INSERT INTO users (name, email, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
			_, err = db.Exec(query, cashierName, cashierEmail, string(hashedCashierPassword), now, now)
			if err == nil {
				log.Println("✅ Sukses! Akun kasir telah dibuat (kasir@posbos.com / kasir123).")
			}
		}
	}
}
