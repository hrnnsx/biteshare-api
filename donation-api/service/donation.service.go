package service

import (
	"context"
	"errors"
	"fmt"
	"math"
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
	Review(ctx context.Context, requestID int32, input dto.ReviewDonationRequest) error
}

type donationService struct {
	repo repository.DonationRepository
}

func NewDonationService(repo repository.DonationRepository) DonationService {
	return &donationService{
		repo: repo,
	}
}

// Create creates a new donation request along with its donor and donated foods.
func (s *donationService) Create(
	ctx context.Context,
	input dto.CreateDonationRequest,
) (*model.DonationRequest, error) {

	// Validate donor information.
	if strings.TrimSpace(input.Donor.Name) == "" ||
		strings.TrimSpace(input.Donor.Email) == "" ||
		strings.TrimSpace(input.Donor.Phone) == "" ||
		strings.TrimSpace(input.Donor.Address) == "" {
		return nil, ErrInvalidDonationData
	}

	// A donation request must contain at least one food item.
	if len(input.Foods) == 0 {
		return nil, ErrInvalidDonationData
	}

	// Convert donor DTO into donor model.
	donor := &model.Donor{
		Name:    strings.TrimSpace(input.Donor.Name),
		Email:   strings.TrimSpace(input.Donor.Email),
		Phone:   strings.TrimSpace(input.Donor.Phone),
		Address: strings.TrimSpace(input.Donor.Address),
	}

	// Create donation request model.
	request := &model.DonationRequest{
		Status:      model.DonationPending,
		Notes:       input.Notes,
		SubmittedAt: time.Now(),
	}

	// Convert food DTOs into food models.
	foods := make([]model.DonatedFood, 0, len(input.Foods))

	for _, item := range input.Foods {
		foodName := strings.TrimSpace(item.FoodName)
		unit := strings.TrimSpace(item.Unit)

		if foodName == "" || unit == "" ||
			math.IsNaN(item.Quantity) ||
			math.IsInf(item.Quantity, 0) ||
			item.Quantity <= 0 {
			return nil, ErrInvalidDonationData
		}

		// Convert expiry date from YYYY-MM-DD to time.Time.
		var expiryDate *time.Time

		if item.ExpiryDate != nil && strings.TrimSpace(*item.ExpiryDate) != "" {
			parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(*item.ExpiryDate))
			if err != nil {
				return nil, fmt.Errorf("%w: invalid expiry date format", ErrInvalidDonationData)
			}

			expiryDate = &parsedDate
		}

		foods = append(foods, model.DonatedFood{
			FoodName:    foodName,
			Quantity:    item.Quantity,
			Unit:        unit,
			ExpiryDate:  expiryDate,
			Description: stringToPointer(item.Description),
		})
	}

	// Save donor, request, and foods through the repository.
	if err := s.repo.Create(ctx, donor, request, foods); err != nil {
		return nil, err
	}

	// Retrieve the created request with its relationships.
	createdRequest, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return nil, err
	}

	return createdRequest, nil
}

// GetByID retrieves a donation request by its ID.
func (s *donationService) GetByID(
	ctx context.Context,
	requestID int32,
) (*model.DonationRequest, error) {

	if requestID <= 0 {
		return nil, ErrInvalidDonationData
	}

	return s.repo.FindByID(ctx, requestID)
}

// GetAll retrieves donation requests with optional status filtering and pagination.
func (s *donationService) GetAll(
	ctx context.Context,
	status string,
	page, limit int,
) ([]model.DonationRequest, int64, error) {

	status = strings.TrimSpace(status)

	if status != "" {
		switch model.DonationStatus(status) {
		case model.DonationPending,
			model.DonationApproved,
			model.DonationRejected,
			model.DonationPickupScheduled,
			model.DonationPickedUp,
			model.DonationReceived:
			// Valid status.
		default:
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

// Review approves or rejects a pending donation request.
func (s *donationService) Review(
	ctx context.Context,
	requestID int32,
	input dto.ReviewDonationRequest,
) error {

	if requestID <= 0 {
		return ErrInvalidDonationData
	}

	var status model.DonationStatus

	switch input.Status {
	case dto.ReviewApproved:
		status = model.DonationApproved

	case dto.ReviewRejected:
		status = model.DonationRejected

	default:
		return ErrInvalidStatus
	}

	// Retrieve the request before reviewing it.
	request, err := s.repo.FindByID(ctx, requestID)
	if err != nil {
		return err
	}

	// Only pending requests can be reviewed.
	if request.Status != model.DonationPending {
		return ErrInvalidTransition
	}

	return s.repo.UpdateStatus(ctx, requestID, status)
}

// stringToPointer converts a non-empty string into a string pointer.
// Empty strings are represented as nil (NULL in the database).
func stringToPointer(value string) *string {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	return &value
}

var _ DonationService = (*donationService)(nil)
