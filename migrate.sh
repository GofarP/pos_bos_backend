#!/bin/bash

# Baca koneksi database langsung dari file .env
DB_DSN=$(grep "^DB_DSN=" .env | cut -d "=" -f2-)

echo "Menjalankan database migrations..."
./goose -dir database/migrations mysql "$DB_DSN" up
echo "Selesai!"
