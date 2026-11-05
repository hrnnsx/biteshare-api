package model

type OrderItem struct {
	ID        uint64  `json:"order_item_id"`
	OrderID   uint64  `json:"order_id"`
	ProductID uint64  `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Subtotal  float64 `json:"subtotal"`
}
