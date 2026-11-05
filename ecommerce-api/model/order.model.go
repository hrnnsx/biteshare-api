package model

import "time"

type Order struct {
	ID              uint64      `json:"order_id"`
	BuyerID         uint64      `json:"buyer_id"`
	OrderNumber     string      `json:"order_number"`
	DeliveryAddress string      `json:"delivery_address"`
	TotalAmount     float64     `json:"total_amount"`
	DeliveryFee     float64     `json:"delivery_fee"`
	Status          OrderStatus `json:"order_status"`
	OrderedAt       time.Time   `json:"ordered_at"`
}
