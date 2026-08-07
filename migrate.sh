#!/bin/bash

# Baca koneksi database langsung dari file .env
DB_DSN=$(grep "^DB_DSN=" .env | cut -d "=" -f2-)

if [ ! -f "./goose" ]; then
  echo "Mengunduh goose binary..."
  wget -q -O goose https://github.com/pressly/goose/releases/latest/download/goose_linux_x86_64
  chmod +x goose
fi

echo "Menjalankan database migrations..."
./goose -dir database/migrations mysql "$DB_DSN" up
echo "Selesai!"
