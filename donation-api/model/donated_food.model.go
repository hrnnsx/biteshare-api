package model

import "time"

type DonatedFood struct {
	ID          int32      `json:"food_id" gorm:"column:food_id;primaryKey;autoIncrement"`
	RequestID   int32      `json:"request_id" gorm:"column:request_id;not null"`
	FoodName    string     `json:"food_name" gorm:"column:food_name;not null"`
	Quantity    float64    `json:"quantity" gorm:"column:quantity;type:numeric;not null"`
	Unit        string     `json:"unit" gorm:"column:unit;not null"`
	ExpiryDate  *time.Time `json:"expiry_date,omitempty" gorm:"column:expiry_date;type:date"`
	Description *string    `json:"description,omitempty" gorm:"column:description;type:text"`

	Request DonationRequest `json:"-" gorm:"foreignKey:RequestID"`
	Images  []DonationImage `json:"images,omitempty" gorm:"foreignKey:FoodID"`
}

func (DonatedFood) TableName() string {
	return "donated_foods"
}
