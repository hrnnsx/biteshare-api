package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"biteshare-api/donation-api/dto"
	"biteshare-api/donation-api/model"
	"biteshare-api/donation-api/repository"
)

var (
	ErrInvalidDonationData = errors.New("invalid donation data")
	ErrInvalidStatus       = errors.New("invalid donation status")
	ErrInvalidTransition   = errors.New("invalid donation status transition")
)

type DonationService interface {
	Create(ctx context.Context, input dto.CreateDonationRequest) (*model.DonationRequest, error)
	GetByID(ctx context.Context, requestID int32) (*model.DonationRequest, error)
	GetAll(ctx context.Context, status string, page, limit int) ([]model.DonationRequest, int64, error)
	Review(ctx context.Context, requestID int32, input dto.ReviewDonationRequest) (*model.DonationRequest, error)
}

type donationService struct {
	repo repository.DonationRepository
}

func NewDonationService(repo repository.DonationRepository) DonationService {
	return &donationService{repo: repo}
}

func (s *donationService) Create(
	ctx context.Context,
	input dto.CreateDonationRequest,
) (*model.DonationRequest, error) {
	if strings.TrimSpace(input.Donor.Name) == "" ||
		strings.TrimSpace(input.Donor.Email) == "" ||
		strings.TrimSpace(input.Donor.Phone) == "" ||
		strings.TrimSpace(input.Donor.Address) == "" ||
		len(input.Foods) == 0 {
		return nil, ErrInvalidDonationData
	}

	donor := &model.Donor{
		Name:    strings.TrimSpace(input.Donor.Name),
		Email:   strings.TrimSpace(input.Donor.Email),
		Phone:   strings.TrimSpace(input.Donor.Phone),
		Address: strings.TrimSpace(input.Donor.Address),
	}

	request := &model.DonationRequest{
		Status:      model.DonationPending,
		Notes:       stringToPointer(input.Notes),
		SubmittedAt: time.Now(),
	}

	foods := make([]model.DonatedFood, 0, len(input.Foods))

	for _, foodInput := range input.Foods {
		if strings.TrimSpace(foodInput.FoodName) == "" ||
			foodInput.Quantity <= 0 ||
			strings.TrimSpace(foodInput.Unit) == "" {
			return nil, ErrInvalidDonationData
		}

		var expiryDate *time.Time

		if foodInput.ExpiryDate != nil &&
			strings.TrimSpace(*foodInput.ExpiryDate) != "" {
			parsedDate, err := time.Parse(
				"2006-01-02",
				strings.TrimSpace(*foodInput.ExpiryDate),
			)
			if err != nil {
				return nil, ErrInvalidDonationData
			}
			expiryDate = &parsedDate
		}

		foods = append(foods, model.DonatedFood{
			FoodName:    strings.TrimSpace(foodInput.FoodName),
			Quantity:    foodInput.Quantity,
			Unit:        strings.TrimSpace(foodInput.Unit),
			ExpiryDate:  expiryDate,
			Description: stringToPointer(&foodInput.Description),
		})
	}

	if err := s.repo.Create(ctx, donor, request, foods); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, request.ID)
}

func (s *donationService) GetByID(
	ctx context.Context,
	requestID int32,
) (*model.DonationRequest, error) {
	if requestID <= 0 {
		return nil, ErrInvalidDonationData
	}

	return s.repo.FindByID(ctx, requestID)
}

func (s *donationService) GetAll(
	ctx context.Context,
	status string,
	page, limit int,
) ([]model.DonationRequest, int64, error) {
	if status != "" {
		validStatuses := map[string]bool{
			string(model.DonationPending):         true,
			string(model.DonationApproved):        true,
			string(model.DonationRejected):        true,
			string(model.DonationPickupScheduled): true,
			string(model.DonationPickedUp):        true,
			string(model.DonationReceived):        true,
		}

		if !validStatuses[status] {
			return nil, 0, ErrInvalidStatus
		}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.FindAll(ctx, status, page, limit)
}

func (s *donationService) Review(
	ctx context.Context,
	requestID int32,
	input dto.ReviewDonationRequest,
) (*model.DonationRequest, error) {
	if requestID <= 0 {
		return nil, ErrInvalidDonationData
	}

	var newStatus model.DonationStatus

	switch input.Status {
	case dto.ReviewApproved:
		newStatus = model.DonationApproved
	case dto.ReviewRejected:
		newStatus = model.DonationRejected
	default:
		return nil, ErrInvalidStatus
	}

	donation, err := s.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if donation.Status != model.DonationPending {
		return nil, ErrInvalidTransition
	}

	if err := s.repo.UpdateStatus(ctx, requestID, newStatus); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, requestID)
}

func stringToPointer(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

var _ DonationService = (*donationService)(nil)
