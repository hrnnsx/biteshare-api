package repository

import (
	"context"
	"errors"

	"biteshare-api/donation-api/model"

	"gorm.io/gorm"
)

var (
	ErrDonationNotFound = errors.New("donation request not found")
)

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

type donationRepository struct {
	db *gorm.DB
}

func NewDonationRepository(db *gorm.DB) DonationRepository {
	return &donationRepository{
		db: db,
	}
}

// Create stores donor, donation request, donated foods,
// and their image metadata in one database transaction.
func (r *donationRepository) Create(
	ctx context.Context,
	donor *model.Donor,
	request *model.DonationRequest,
	foods []model.DonatedFood,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(donor).Error; err != nil {
			return err
		}

		request.DonorID = donor.ID

		if err := tx.Create(request).Error; err != nil {
			return err
		}

		for i := range foods {
			foods[i].RequestID = request.ID

			// Keep images temporarily before creating the food.
			images := foods[i].Images
			foods[i].Images = nil

			if err := tx.Create(&foods[i]).Error; err != nil {
				return err
			}

			// The food ID is now available.
			for j := range images {
				images[j].FoodID = foods[i].ID

				if err := tx.Create(&images[j]).Error; err != nil {
					return err
				}
			}

			foods[i].Images = images
		}

		return nil
	})
}

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

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	err := query.
		Preload("Donor").
		Preload("Foods").
		Preload("Foods.Images").
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

var _ DonationRepository = (*donationRepository)(nil)
