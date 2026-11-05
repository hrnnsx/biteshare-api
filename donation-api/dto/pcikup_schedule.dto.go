package dto

import "time"

type CreatePickupScheduleRequest struct {
	PickupAt time.Time `json:"pickup_at" binding:"required"`
}

type PickupScheduleResponse struct {
	ID        uint64    `json:"schedule_id"`
	RequestID uint64    `json:"request_id"`
	AdminID   uint64    `json:"admin_id"`
	PickupAt  time.Time `json:"pickup_at"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
