package dto

type CreateDonatedFoodRequest struct {
	FoodName    string  `json:"food_name" binding:"required,max=100"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	Unit        string  `json:"unit" binding:"required,max=30"`
	ExpiryDate  *string `json:"expiry_date"`
	Description string  `json:"description"`
}

type DonatedFoodResponse struct {
	ID          uint64  `json:"food_id"`
	FoodName    string  `json:"food_name"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	ExpiryDate  *string `json:"expiry_date,omitempty"`
	Description string  `json:"description,omitempty"`
}
