package repository

import (
	"context"
	"errors"

	"biteshare-api/donation-api/model"

	"gorm.io/gorm"
)

// Repository errors
var (
	ErrDonationNotFound = errors.New("donation request not found")
)

// DonationRepository defines donation database operations.
type DonationRepository interface {
	Create(
		ctx context.Context,
		donor *model.Donor,
		request *model.DonationRequest,
		foods []model.DonatedFood,
	) error

	FindByID(
		ctx context.Context,
		requestID int32,
	) (*model.DonationRequest, error)

	FindAll(
		ctx context.Context,
		status string,
		page int,
		limit int,
	) ([]model.DonationRequest, int64, error)

	UpdateStatus(
		ctx context.Context,
		requestID int32,
		status model.DonationStatus,
	) error
}

// donationRepository implements DonationRepository using GORM.
type donationRepository struct {
	db *gorm.DB
}

// NewDonationRepository creates a new donation repository.
func NewDonationRepository(db *gorm.DB) DonationRepository {
	return &donationRepository{
		db: db,
	}
}

// Create stores donor, donation request, and donated foods.
// All operations are executed within one database transaction.
func (r *donationRepository) Create(
	ctx context.Context,
	donor *model.Donor,
	request *model.DonationRequest,
	foods []model.DonatedFood,
) error {

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// Create donor
		if err := tx.Create(donor).Error; err != nil {
			return err
		}

		// Assign donor ID to donation request
		request.DonorID = donor.ID

		// Create donation request
		if err := tx.Create(request).Error; err != nil {
			return err
		}

		// Create donated foods
		for i := range foods {
			foods[i].RequestID = request.ID

			if err := tx.Create(&foods[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// FindByID retrieves a donation request and its related data.
func (r *donationRepository) FindByID(
	ctx context.Context,
	requestID int32,
) (*model.DonationRequest, error) {

	var donation model.DonationRequest

	err := r.db.WithContext(ctx).
		Preload("Donor").
		Preload("Foods").
		Preload("Foods.Images").
		Preload("PickupSchedules").
		Preload("PickupSchedules.Pickup").
		First(&donation, "request_id = ?", requestID).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDonationNotFound
	}

	if err != nil {
		return nil, err
	}

	return &donation, nil
}

// FindAll retrieves donation requests with optional status filtering and pagination.
func (r *donationRepository) FindAll(
	ctx context.Context,
	status string,
	page int,
	limit int,
) ([]model.DonationRequest, int64, error) {

	var donations []model.DonationRequest
	var total int64

	query := r.db.WithContext(ctx).
		Model(&model.DonationRequest{})

	// Apply status filter when provided.
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count matching records.
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination defaults.
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Retrieve donation requests.
	err := query.
		Preload("Donor").
		Preload("Foods").
		Order("submitted_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&donations).
		Error

	if err != nil {
		return nil, 0, err
	}

	return donations, total, nil
}

// UpdateStatus changes the status of a donation request.
func (r *donationRepository) UpdateStatus(
	ctx context.Context,
	requestID int32,
	status model.DonationStatus,
) error {

	result := r.db.WithContext(ctx).
		Model(&model.DonationRequest{}).
		Where("request_id = ?", requestID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrDonationNotFound
	}

	return nil
}

// Compile-time interface implementation check.
var _ DonationRepository = (*donationRepository)(nil)
