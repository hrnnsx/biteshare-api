package model

import "time"

type Delivery struct {
	ID              uint64         `json:"delivery_id"`
	OrderID         uint64         `json:"order_id"`
	CourierName     *string        `json:"courier_name,omitempty"`
	CourierPhone    *string        `json:"courier_phone,omitempty"`
	DeliveryAddress string         `json:"delivery_address"`
	Status          DeliveryStatus `json:"delivery_status"`
	ShippedAt       *time.Time     `json:"shipped_at,omitempty"`
	DeliveredAt     *time.Time     `json:"delivered_at,omitempty"`
	Notes           *string        `json:"notes,omitempty"`
}
