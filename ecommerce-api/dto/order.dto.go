package dto

// CheckoutRequest is the data submitted by a buyer during checkout.
type CheckoutRequest struct {
	Name          string             `json:"name" binding:"required"`
	Email         string             `json:"email" binding:"required,email"`
	Phone         string             `json:"phone" binding:"required"`
	Address       string             `json:"address" binding:"required"`
	PaymentMethod string             `json:"payment_method" binding:"required"`
	DeliveryFee   float64            `json:"delivery_fee"`
	Items         []CheckoutItem `json:"items" binding:"required,min=1,dive"`
}

type CheckoutItem struct {
	ProductID int32   `json:"product_id" binding:"required,gt=0"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
}
