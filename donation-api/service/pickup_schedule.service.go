package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"biteshare-api/donation-api/dto"
	"biteshare-api/donation-api/model"
	"biteshare-api/donation-api/repository"
	"biteshare-api/internal/mailer"
)

var (
	ErrInvalidScheduleData = errors.New("invalid pickup schedule data")
	ErrInvalidAdminID      = errors.New("invalid admin ID")
)

type PickupScheduleService interface {
	SchedulePickup(
		ctx context.Context,
		requestID int32,
		adminID uint64,
		input dto.CreatePickupScheduleRequest,
	) (*model.PickupSchedule, error)
}

type pickupScheduleService struct {
	repo   repository.PickupScheduleRepository
	mailer mailer.Mailer
}

func NewPickupScheduleService(
	repo repository.PickupScheduleRepository,
	emailMailer mailer.Mailer,
) PickupScheduleService {
	return &pickupScheduleService{
		repo:   repo,
		mailer: emailMailer,
	}
}

func (s *pickupScheduleService) SchedulePickup(
	ctx context.Context,
	requestID int32,
	adminID uint64,
	input dto.CreatePickupScheduleRequest,
) (*model.PickupSchedule, error) {
	if requestID <= 0 {
		return nil, ErrInvalidScheduleData
	}

	if adminID == 0 || adminID > 2147483647 {
		return nil, ErrInvalidAdminID
	}

	pickupDate, err := time.Parse(
		"2006-01-02",
		strings.TrimSpace(input.PickupDate),
	)
	if err != nil {
		return nil, ErrInvalidScheduleData
	}

	pickupTime := strings.TrimSpace(input.PickupTime)

	if _, err := time.Parse("15:04", pickupTime); err != nil {
		return nil, ErrInvalidScheduleData
	}

	schedule := &model.PickupSchedule{
		RequestID:  requestID,
		AdminID:    int32(adminID),
		PickupDate: pickupDate,
		PickupTime: pickupTime,
		Status:     model.ScheduleScheduled,
		CreatedAt:  time.Now(),
	}

	// Save schedule and update donation status.
	if err := s.repo.CreateSchedule(ctx, schedule); err != nil {
		return nil, err
	}

	// Send email after the database transaction succeeds.
	if s.mailer != nil {
		donor := schedule.Request.Donor

		if err := s.mailer.SendPickupScheduledEmail(
			donor.Email,
			donor.Name,
			requestID,
			schedule.PickupDate,
			schedule.PickupTime,
		); err != nil {
			log.Printf(
				"[PickupScheduleService] Failed to send scheduling email for request #%d: %v",
				requestID,
				err,
			)
		}
	}

	return schedule, nil
}

var _ PickupScheduleService = (*pickupScheduleService)(nil)
