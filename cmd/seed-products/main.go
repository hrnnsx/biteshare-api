package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"biteshare-api/internal/database"
	"gorm.io/gorm"
)

type productSeed struct {
	name        string
	description string
	price       float64
	stock       float64
	unit        string
	imageURL    string
}

var productSeeds = []productSeed{
	{
		name:        "Dedak",
		description: "Dedak untuk pakan ternak atau hewan peliharaan sesuai kebutuhan.",
		price:       5000,
		stock:       25,
		unit:        "kg",
		imageURL:    "https://ik.imagekit.io/ihzrpibzw/dedak_-E9CJcq3z",
	},
	{
		name:        "Jagung",
		description: "Jagung untuk pakan ternak atau hewan peliharaan sesuai kebutuhan.",
		price:       3000,
		stock:       20,
		unit:        "kg",
		imageURL:    "https://ik.imagekit.io/ihzrpibzw/jagung_5IdCwfCWh",
	},
	{
		name:        "Roti",
		description: "Roti sisa untuk pakan hewan atau kebutuhan pertanian sesuai pengolahan yang aman.",
		price:       2000,
		stock:       15,
		unit:        "kg",
		imageURL:    "https://ik.imagekit.io/ihzrpibzw/roti_SMXLWeWLB",
	},
	{
		name:        "Tomat",
		description: "Tomat sisa untuk pakan hewan atau kebutuhan pertanian sesuai pengolahan yang aman.",
		price:       2500,
		stock:       10,
		unit:        "kg",
		imageURL:    "https://ik.imagekit.io/ihzrpibzw/tomat_qjZsUOE-N",
	},
}

func main() {
	db := database.New()
	defer db.Close()

	adminID, err := findAdminID(db.GORM())
	if err != nil {
		log.Fatal(err)
	}

	err = db.GORM().Transaction(func(tx *gorm.DB) error {
		for _, seed := range productSeeds {
			var productID int32
			result := tx.Raw(`
				SELECT product_id
				FROM food_products
				WHERE admin_id = ? AND product_name = ?
				LIMIT 1
			`, adminID, seed.name).Scan(&productID)
			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				result = tx.Exec(`
					INSERT INTO food_products
						(admin_id, product_name, description, price, stock, unit, status)
					VALUES (?, ?, ?, ?, ?, ?, 'available')
				`, adminID, seed.name, seed.description, seed.price, seed.stock, seed.unit)
				if result.Error != nil {
					return result.Error
				}

				result = tx.Raw(`
					SELECT product_id
					FROM food_products
					WHERE admin_id = ? AND product_name = ?
					LIMIT 1
				`, adminID, seed.name).Scan(&productID)
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return fmt.Errorf("could not find seeded product %q", seed.name)
				}
			}

			var imageID int32
			result = tx.Raw(`
				SELECT image_id
				FROM product_images
				WHERE product_id = ? AND image_url = ?
				LIMIT 1
			`, productID, seed.imageURL).Scan(&imageID)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				result = tx.Exec(`
					INSERT INTO product_images (product_id, image_url, is_primary)
					VALUES (?, ?, TRUE)
				`, productID, seed.imageURL)
				if result.Error != nil {
					return result.Error
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Fatalf("could not seed products: %v", err)
	}

	log.Printf("seeded %d products", len(productSeeds))
}

func findAdminID(db *gorm.DB) (int32, error) {
	email := strings.TrimSpace(os.Getenv("SEED_ADMIN_EMAIL"))
	if email == "" {
		email = strings.TrimSpace(os.Getenv("ADMIN_1_EMAIL"))
	}

	var adminID int32
	var result *gorm.DB
	if email != "" {
		result = db.Raw(`SELECT admin_id FROM admins WHERE email = ? LIMIT 1`, email).Scan(&adminID)
	} else {
		result = db.Raw(`SELECT admin_id FROM admins ORDER BY admin_id ASC LIMIT 1`).Scan(&adminID)
	}
	if result.Error != nil {
		return 0, fmt.Errorf("could not find seed admin: %w", result.Error)
	}
	if result.RowsAffected == 0 || adminID == 0 {
		return 0, fmt.Errorf("no admin found; run go run ./cmd/seed-admin first")
	}

	return adminID, nil
}
