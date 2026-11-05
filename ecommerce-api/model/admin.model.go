package model

import "time"

// Admin represents an administrator who can sign in to the API.
type Admin struct {
	ID           uint64    `json:"admin_id" gorm:"column:admin_id;primaryKey"`
	Name         string    `json:"name" gorm:"column:name"`
	Email        string    `json:"email" gorm:"column:email"`
	PasswordHash string    `json:"-" gorm:"column:password"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
}

// TableName tells GORM which database table stores admins.
func (Admin) TableName() string {
	return "admins"
}
