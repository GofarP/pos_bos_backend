#!/bin/bash

# Baca koneksi database langsung dari file .env
DB_DSN=$(grep "^DB_DSN=" .env | cut -d "=" -f2-)

echo "Menjalankan database migrations..."
go run github.com/pressly/goose/v3/cmd/goose@latest -dir database/migrations mysql "$DB_DSN" up
echo "Selesai!"
