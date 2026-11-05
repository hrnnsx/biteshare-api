package main

import (
	"log"
	"os"
	"strings"

	"biteshare-api/internal/auth"
	"biteshare-api/internal/database"
)

// adminSeed contains the credentials needed to create one admin account.
type adminSeed struct {
	name     string
	email    string
	password string
}

func main() {
	// Read all three admin accounts from the environment.
	admins := []adminSeed{
		loadAdminSeed("ADMIN_1"),
		loadAdminSeed("ADMIN_2"),
		loadAdminSeed("ADMIN_3"),
	}

	// Open the shared database connection.
	db := database.New()
	defer db.Close()

	// Insert each admin only when the email is not already registered.
	for _, admin := range admins {
		// Hash each password before it is sent to PostgreSQL.
		passwordHash, err := auth.HashPassword(admin.password)
		if err != nil {
			log.Fatalf("could not hash password for %s: %v", admin.email, err)
		}

		result := db.GORM().Exec(`
			INSERT INTO admins (name, email, password)
			VALUES (?, ?, ?)
			ON CONFLICT (email) DO NOTHING
		`, admin.name, admin.email, passwordHash)
		if result.Error != nil {
			log.Fatalf("could not seed admin %s: %v", admin.email, result.Error)
		}

		if result.RowsAffected == 0 {
			log.Printf("admin already exists: %s", admin.email)
			continue
		}

		log.Printf("admin seeded successfully: %s", admin.email)
	}
}

// loadAdminSeed reads and validates one admin's seed values.
func loadAdminSeed(prefix string) adminSeed {
	// Read credentials from the environment instead of committing them.
	admin := adminSeed{
		name:     strings.TrimSpace(os.Getenv(prefix + "_NAME")),
		email:    strings.TrimSpace(os.Getenv(prefix + "_EMAIL")),
		password: os.Getenv(prefix + "_PASSWORD"),
	}

	// Fail early when a required value is missing.
	if admin.name == "" || admin.email == "" || admin.password == "" {
		log.Fatalf("%s_NAME, %s_EMAIL, and %s_PASSWORD are required", prefix, prefix, prefix)
	}

	return admin
}
