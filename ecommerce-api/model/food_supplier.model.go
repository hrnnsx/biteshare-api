package model

import "time"

type FoodSupplier struct {
	ID           uint64    `json:"supplier_id"`
	SupplierName string    `json:"supplier_name"`
	Email        *string   `json:"email,omitempty"`
	Phone        *string   `json:"phone,omitempty"`
	Address      *string   `json:"address,omitempty"`
	SupplierType *string   `json:"supplier_type,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
