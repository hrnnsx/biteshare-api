package model

import "time"

type Admin struct {
	ID           uint64    `json:"admin_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-" gorm:"column:password"`
	CreatedAt    time.Time `json:"created_at"`
}
