package dto

type CreateDonationRequest struct {
	Donor CreateDonorRequest         `json:"donor" binding:"required"`
	Notes *string                    `json:"notes"`
	Foods []CreateDonatedFoodRequest `json:"foods" binding:"required,min=1,dive"`
}

type CreateDonorRequest struct {
	Name    string `json:"name" binding:"required,min=2,max=100"`
	Email   string `json:"email" binding:"required,email,max=100"`
	Phone   string `json:"phone" binding:"required,max=20"`
	Address string `json:"address" binding:"required"`
}

type ReviewDonationRequest struct {
	Status DonationReviewStatus `json:"status" binding:"required,oneof=approved rejected"`
}

type DonationReviewStatus string

const (
	ReviewApproved DonationReviewStatus = "approved"
	ReviewRejected DonationReviewStatus = "rejected"
)
