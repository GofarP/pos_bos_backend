package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	// 1. Connect to Database
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Gagal menyambung ke database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database tidak merespons: %v", err)
	}

	log.Println("Memulai proses seeding...")

	seedAdminUser(db)
	seedRBAC(db)
	seedCategories(db)
	seedProducts(db)

	log.Println("✅ Semua seeder selesai dijalankan!")
}
