package model

import "time"

type DonationStatus string

const (
	DonationPending         DonationStatus = "pending"
	DonationApproved        DonationStatus = "approved"
	DonationRejected        DonationStatus = "rejected"
	DonationPickupScheduled DonationStatus = "pickup_scheduled"
	DonationPickedUp        DonationStatus = "picked_up"
	DonationReceived        DonationStatus = "received"
)

type DonationRequest struct {
	ID          int32          `json:"request_id" gorm:"column:request_id;primaryKey;autoIncrement"`
	DonorID     int32          `json:"donor_id" gorm:"column:donor_id;not null"`
	Status      DonationStatus `json:"status" gorm:"column:status;not null;default:pending"`
	Notes       *string        `json:"notes,omitempty" gorm:"column:notes;type:text"`
	SubmittedAt time.Time      `json:"submitted_at" gorm:"column:submitted_at;not null;default:now()"`

	Donor           Donor            `json:"donor,omitempty" gorm:"foreignKey:DonorID"`
	Foods           []DonatedFood    `json:"foods,omitempty" gorm:"foreignKey:RequestID"`
	PickupSchedules []PickupSchedule `json:"pickup_schedules,omitempty" gorm:"foreignKey:RequestID"`
}

func (DonationRequest) TableName() string {
	return "donation_requests"
}
