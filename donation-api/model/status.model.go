package model

// Donation Request Status

type DonationStatus string

const (
	DonationPending  DonationStatus = "PENDING"
	DonationApproved DonationStatus = "APPROVED"
	DonationRejected DonationStatus = "REJECTED"
)

// Pickup Schedule Status

type ScheduleStatus string

const (
	ScheduleScheduled   ScheduleStatus = "SCHEDULED"
	ScheduleRescheduled ScheduleStatus = "RESCHEDULED"
	ScheduleCancelled   ScheduleStatus = "CANCELLED"
)

// Pickup Result Status

type PickupStatus string

const (
	PickupPending    PickupStatus = "PENDING"
	PickupSuccessful PickupStatus = "SUCCESSFUL"
	PickupFailed     PickupStatus = "FAILED"
)
