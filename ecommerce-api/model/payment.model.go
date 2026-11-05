package model

import "time"

type Payment struct {
	ID                   uint64        `json:"payment_id"`
	OrderID              uint64        `json:"order_id"`
	PaymentMethod        string        `json:"payment_method"`
	Amount               float64       `json:"amount"`
	Status               PaymentStatus `json:"payment_status"`
	TransactionReference *string       `json:"transaction_reference,omitempty"`
	PaidAt               *time.Time    `json:"paid_at,omitempty"`
}
