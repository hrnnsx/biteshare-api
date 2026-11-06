package repository

import (
	"context"
	"errors"

	"biteshare-api/donation-api/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Pickup schedule repository errors.
var (
	ErrDonationNotApproved = errors.New("donation request is not approved")
)

// PickupScheduleRepository defines pickup schedule database operations.
type PickupScheduleRepository interface {
	CreateSchedule(
		ctx context.Context,
		schedule *model.PickupSchedule,
	) error
}

// pickupScheduleRepository implements PickupScheduleRepository using GORM.
type pickupScheduleRepository struct {
	db *gorm.DB
}

// NewPickupScheduleRepository creates a new pickup schedule repository.
func NewPickupScheduleRepository(db *gorm.DB) PickupScheduleRepository {
	return &pickupScheduleRepository{
		db: db,
	}
}

// CreateSchedule creates a pickup schedule and updates donation status.
// Both operations are executed within one database transaction.
func (r *pickupScheduleRepository) CreateSchedule(
	ctx context.Context,
	schedule *model.PickupSchedule,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var donation model.DonationRequest

		// Lock the donation request and load donor information.
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Donor").
			First(&donation, "request_id = ?", schedule.RequestID).
			Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDonationNotFound
		}

		if err != nil {
			return err
		}

		// Only approved donations can be scheduled.
		if donation.Status != model.DonationApproved {
			return ErrDonationNotApproved
		}

		// Create pickup schedule.
		if err := tx.Create(schedule).Error; err != nil {
			return err
		}

		// Update donation status.
		if err := tx.
			Model(&donation).
			Update("status", model.DonationPickupScheduled).
			Error; err != nil {
			return err
		}

		// Attach donation and donor information for the service.
		// This assignment happens after Create, so GORM will not
		// attempt to insert the existing donation as a new record.
		schedule.Request = donation

		return nil
	})
}

// Compile-time interface implementation check.
var _ PickupScheduleRepository = (*pickupScheduleRepository)(nil)
