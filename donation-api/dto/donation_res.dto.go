package dto

import "time"

type DonationResponse struct {
	ID            uint64                `json:"request_id"`
	Donor         DonorResponse         `json:"donor"`
	PickupAddress string                `json:"pickup_address"`
	Status        string                `json:"status"`
	Notes         string                `json:"notes,omitempty"`
	SubmittedAt   time.Time             `json:"submitted_at"`
	Foods         []DonatedFoodResponse `json:"foods"`
}

type DonorResponse struct {
	ID      uint64 `json:"donor_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type DonationListResponse struct {
	Donations  []DonationResponse `json:"donations"`
	Pagination PaginationResponse `json:"pagination"`
}
