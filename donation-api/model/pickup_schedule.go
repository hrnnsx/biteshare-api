package model

import "time"

type PickupSchedule struct {
	ID         uint64         `json:"schedule_id"`
	RequestID  uint64         `json:"request_id"`
	AdminID    uint64         `json:"admin_id"`
	PickupDate time.Time      `json:"pickup_date"`
	PickupTime string         `json:"pickup_time"`
	Status     ScheduleStatus `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
}
