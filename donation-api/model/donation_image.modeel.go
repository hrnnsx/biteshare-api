package model

import "time"

type DonationImage struct {
	ID           uint64    `json:"image_id"`
	RequestID    uint64    `json:"request_id"`
	FoodID       *uint64   `json:"food_id,omitempty"`
	OriginalName string    `json:"original_name"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	StorageKey   string    `json:"storage_key,omitempty"`
	UploadedAt   time.Time `json:"uploaded_at"`
}
