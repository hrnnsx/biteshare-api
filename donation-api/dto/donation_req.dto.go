package dto

type CreateDonationRequest struct {
	Donor         CreateDonorRequest         `json:"donor" binding:"required"`
	PickupAddress string                     `json:"pickup_address" binding:"required"`
	Notes         string                     `json:"notes,omitempty"`
	Foods         []CreateDonatedFoodRequest `json:"foods" binding:"required,min=1,dive"`
}

type CreateDonorRequest struct {
	Name    string `json:"name" binding:"required,max=100"`
	Email   string `json:"email" binding:"required,email,max=100"`
	Phone   string `json:"phone" binding:"required,max=20"`
	Address string `json:"address" binding:"required"`
}

type ReviewDonationRequest struct {
	Status string `json:"status" binding:"required,oneof=APPROVED REJECTED"`
	Notes  string `json:"notes,omitempty"`
}
