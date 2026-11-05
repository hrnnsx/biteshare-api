package model

import "time"

type DonationImage struct {
	ID        int32     `json:"image_id" gorm:"column:image_id;primaryKey;autoIncrement"`
	FoodID    int32     `json:"food_id" gorm:"column:food_id;not null"`
	ImageURL  string    `json:"image_url" gorm:"column:image_url;type:text;not null"`
	FileID    *string   `json:"file_id,omitempty" gorm:"column:file_id"`
	IsPrimary bool      `json:"is_primary" gorm:"column:is_primary;not null;default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;default:now()"`

	Food DonatedFood `json:"-" gorm:"foreignKey:FoodID"`
}

func (DonationImage) TableName() string {
	return "donation_images"
}
