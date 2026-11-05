package model

type Donor struct {
	ID      uint64 `json:"donor_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}
