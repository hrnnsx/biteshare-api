package model

// Admin represents an administrator who can sign in to the API.
type Admin struct {
	ID           uint64    `json:"admin_id" gorm:"column:admin_id;primaryKey"`
	Name         string    `json:"name" gorm:"column:name"`
	Email        string    `json:"email" gorm:"column:email"`
	PasswordHash string    `json:"-" gorm:"column:password"`
}

// TableName tells GORM which database table stores admins.
func (Admin) TableName() string {
	return "admins"
}
