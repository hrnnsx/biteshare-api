package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"biteshare-api/internal/database"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	var action string
	flag.StringVar(&action, "action", "up", "migration action: up or down")
	flag.Parse()

	if action != "up" && action != "down" {
		log.Fatalf("invalid action: %s. Use 'up' or 'down'", action)
	}

	dbService := database.New()
	defer dbService.Close()

	// Ambil raw *sql.DB dari service database
	var sqlDB *sql.DB
	if err := dbService.GORM().Raw("SELECT 1").Error; err != nil {
		log.Fatalf("database not ready: %v", err)
	}
	rawDB, err := dbService.GORM().DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	sqlDB = rawDB

	// 1. Buat tabel pelacak migrasi jika belum ada
	createTrackerQuery := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`
	if _, err := sqlDB.Exec(createTrackerQuery); err != nil {
		log.Fatalf("failed to create schema_migrations table: %v", err)
	}

	migrationsDir := "schema/migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("failed to read migrations directory: %v", err)
	}

	if action == "up" {
		runUp(sqlDB, migrationsDir, files)
	} else {
		runDown(sqlDB, migrationsDir, files)
	}
}

func runUp(db *sql.DB, dir string, entries []os.DirEntry) {
	var upFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}
	sort.Strings(upFiles)

	for _, filename := range upFiles {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", filename).Scan(&exists)
		if err != nil {
			log.Fatalf("failed to check migration status for %s: %v", filename, err)
		}

		if exists {
			log.Printf("[SKIP] %s already applied", filename)
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			log.Fatalf("failed to read file %s: %v", filename, err)
		}

		log.Printf("[APPLYING] %s ...", filename)
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("failed to start transaction: %v", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			log.Fatalf("failed executing migration %s: %v", filename, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", filename); err != nil {
			_ = tx.Rollback()
			log.Fatalf("failed recording migration %s: %v", filename, err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("failed to commit transaction for %s: %v", filename, err)
		}
		log.Printf("[SUCCESS] %s applied successfully", filename)
	}
}

func runDown(db *sql.DB, dir string, entries []os.DirEntry) {
	var downFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".down.sql") {
			downFiles = append(downFiles, entry.Name())
		}
	}
	// Down dieksekusi secara mundur (reverse order)
	sort.Sort(sort.Reverse(sort.StringSlice(downFiles)))

	for _, filename := range downFiles {
		upEquivalent := strings.TrimSuffix(filename, ".down.sql") + ".up.sql"

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", upEquivalent).Scan(&exists)
		if err != nil {
			log.Fatalf("failed to check status for %s: %v", upEquivalent, err)
		}

		if !exists {
			log.Printf("[SKIP] %s was not applied", upEquivalent)
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			log.Fatalf("failed to read file %s: %v", filename, err)
		}

		log.Printf("[ROLLBACK] %s ...", filename)
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("failed to start transaction: %v", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			log.Fatalf("failed rolling back %s: %v", filename, err)
		}

		if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1)", upEquivalent); err != nil {
			_ = tx.Rollback()
			log.Fatalf("failed removing record %s: %v", upEquivalent, err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("failed to commit transaction: %v", err)
		}
		log.Printf("[SUCCESS] %s rolled back", filename)
	}
}
