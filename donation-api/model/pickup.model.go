package model

import "time"

type Pickup struct {
	ID         uint64       `json:"pickup_id"`
	ScheduleID uint64       `json:"schedule_id"`
	PickedUpAt *time.Time   `json:"picked_up_at,omitempty"`
	Status     PickupStatus `json:"status"`
	Notes      string       `json:"notes,omitempty"`
}
