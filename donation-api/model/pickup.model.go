package model

import "time"

type PickupStatus string

const (
	PickupPending    PickupStatus = "pending"
	PickupSuccessful PickupStatus = "successful"
	PickupFailed     PickupStatus = "failed"
)

type Pickup struct {
	ID         int32        `json:"pickup_id" gorm:"column:pickup_id;primaryKey;autoIncrement"`
	ScheduleID int32        `json:"schedule_id" gorm:"column:schedule_id;not null"`
	PickedUpAt *time.Time   `json:"picked_up_at,omitempty" gorm:"column:picked_up_at"`
	Status     PickupStatus `json:"status" gorm:"column:status;not null;default:pending"`
	Notes      *string      `json:"notes,omitempty" gorm:"column:notes;type:text"`

	Schedule PickupSchedule `json:"-" gorm:"foreignKey:ScheduleID"`
}

func (Pickup) TableName() string {
	return "pickups"
}
