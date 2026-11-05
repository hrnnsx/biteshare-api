package dto

import "time"

type RecordPickupRequest struct {
	Status     string     `json:"status" binding:"required,oneof=SUCCESSFUL FAILED"`
	PickedUpAt *time.Time `json:"picked_up_at,omitempty"`
	Notes      string     `json:"notes,omitempty"`
}

type PickupResponse struct {
	ID         uint64     `json:"pickup_id"`
	ScheduleID uint64     `json:"schedule_id"`
	PickedUpAt *time.Time `json:"picked_up_at,omitempty"`
	Status     string     `json:"status"`
	Notes      string     `json:"notes,omitempty"`
}
