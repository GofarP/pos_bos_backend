package main

import (
	"database/sql"
	"log"
	"os"
	"time"
)

func seedRBAC(db *sql.DB) {
	log.Println("Menjalankan seeder untuk RBAC (permissions, roles)...")

	// 1. Seed Permissions
	permissions := []string{
		"view.user", "create.user", "edit.user", "delete.user",
		"view.role", "create.role", "edit.role", "delete.role",
		"view.permission", "create.permission", "edit.permission", "delete.permission",
		"view.category", "create.category", "edit.category", "delete.category",
		"view.product", "create.product", "edit.product", "delete.product",
		"view.transaction", "create.transaction", "edit.transaction", "delete.transaction",
	}

	now := time.Now()
	for _, p := range permissions {
		_, err := db.Exec("INSERT IGNORE INTO permissions (name, created_at, updated_at) VALUES (?, ?, ?)", p, now, now)
		if err != nil {
			log.Fatalf("Gagal insert permission %s: %v", p, err)
		}
	}
	log.Println("✅ Permissions berhasil di-seed.")

	// 2. Seed Roles
	roles := []string{"Superadmin", "Cashier"}
	for _, r := range roles {
		_, err := db.Exec("INSERT IGNORE INTO roles (name, created_at, updated_at) VALUES (?, ?, ?)", r, now, now)
		if err != nil {
			log.Fatalf("Gagal insert role %s: %v", r, err)
		}
	}
	log.Println("✅ Roles berhasil di-seed.")

	// 3. Attach Permissions to Roles
	var superadminID int
	if err := db.QueryRow("SELECT id FROM roles WHERE name = 'Superadmin'").Scan(&superadminID); err == nil {
		for _, p := range permissions {
			var permID int
			if err := db.QueryRow("SELECT id FROM permissions WHERE name = ?", p).Scan(&permID); err == nil {
				db.Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)", superadminID, permID)
			}
		}
	}

	var cashierID int
	if err := db.QueryRow("SELECT id FROM roles WHERE name = 'Cashier'").Scan(&cashierID); err == nil {
		cashierPerms := []string{
			"view.category",
			"view.product",
			"view.transaction", "create.transaction",
		}
		for _, p := range cashierPerms {
			var permID int
			if err := db.QueryRow("SELECT id FROM permissions WHERE name = ?", p).Scan(&permID); err == nil {
				db.Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)", cashierID, permID)
			}
		}
	}
	log.Println("✅ Role Permissions berhasil di-seed.")

	// 4. Attach Superadmin to Admin User
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@posbos.com" // fallback matching users.go
	}

	var adminUserID int
	if err := db.QueryRow("SELECT id FROM users WHERE email = ?", adminEmail).Scan(&adminUserID); err == nil {
		db.Exec("INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)", adminUserID, superadminID)
		log.Println("✅ Superadmin role attached to Admin User.")
	}
}
