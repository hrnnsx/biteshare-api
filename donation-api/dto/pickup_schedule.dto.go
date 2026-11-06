package dto

type CreatePickupScheduleRequest struct {
	PickupDate string `json:"pickup_date" binding:"required"`
	PickupTime string `json:"pickup_time" binding:"required"`
}
