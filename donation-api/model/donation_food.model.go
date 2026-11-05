package model

import "time"

type DonatedFood struct {
	ID          uint64     `json:"food_id"`
	RequestID   uint64     `json:"request_id"`
	FoodName    string     `json:"food_name"`
	Quantity    float64    `json:"quantity"`
	Unit        string     `json:"unit"`
	ExpiryDate  *time.Time `json:"expiry_date,omitempty"`
	Description string     `json:"description,omitempty"`
}
