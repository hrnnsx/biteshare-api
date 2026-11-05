package model

type Admin struct {
	ID           uint64 `json:"admin_id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}
