package model

import "time"

type ScheduleStatus string

const (
	ScheduleScheduled   ScheduleStatus = "scheduled"
	ScheduleRescheduled ScheduleStatus = "rescheduled"
	ScheduleCancelled   ScheduleStatus = "cancelled"
)

type PickupSchedule struct {
	ID         int32          `json:"schedule_id" gorm:"column:schedule_id;primaryKey;autoIncrement"`
	RequestID  int32          `json:"request_id" gorm:"column:request_id;not null"`
	AdminID    int32          `json:"admin_id" gorm:"column:admin_id;not null"`
	PickupDate time.Time      `json:"pickup_date" gorm:"column:pickup_date;not null"`
	PickupTime string         `json:"pickup_time" gorm:"column:pickup_time;not null"`
	Status     ScheduleStatus `json:"status" gorm:"column:status;not null;default:scheduled"`
	CreatedAt  time.Time      `json:"created_at" gorm:"column:created_at;not null;default:now()"`

	Request DonationRequest `json:"-" gorm:"foreignKey:RequestID"`
	Pickup  *Pickup         `json:"pickup,omitempty" gorm:"foreignKey:ScheduleID"`
}

func (PickupSchedule) TableName() string {
	return "pickup_schedules"
}
