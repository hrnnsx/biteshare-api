package model

import "time"

type FoodProduct struct {
	ID          uint64        `json:"product_id"`
	AdminID     uint64        `json:"admin_id"`
	SupplierID  *uint64       `json:"supplier_id,omitempty"`
	ProductName string        `json:"product_name"`
	Description *string       `json:"description,omitempty"`
	Price       float64       `json:"price"`
	Stock       float64       `json:"stock"`
	Unit        string        `json:"unit"`
	ExpiryDate  *time.Time    `json:"expiry_date,omitempty"`
	Status      ProductStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
}
