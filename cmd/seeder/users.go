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

	// Data Admin yang akan di-seed (diambil dari .env)
	adminName := "Super Admin"
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@posbos.com" // fallback
	}
	
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123" // fallback
	}

	// Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Gagal melakukan hashing password: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(id) FROM users WHERE email = ?", adminEmail).Scan(&count)
	if err != nil {
		log.Fatalf("Gagal mengecek data user: %v", err)
	}

	if count > 0 {
		log.Println("Seeder dibatalkan: Admin sudah ada di database.")
		return
	}

	// Insert ke Database
	query := `INSERT INTO users (name, email, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	now := time.Now()

	_, err = db.Exec(query, adminName, adminEmail, string(hashedPassword), now, now)
	if err != nil {
		log.Fatalf("Gagal memasukkan data admin: %v", err)
	}

	log.Println("✅ Sukses! Akun admin telah dibuat.")
	log.Printf("Email: %s | Password: %s\n", adminEmail, adminPassword)
}
