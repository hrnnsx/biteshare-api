package model

type Donor struct {
	ID      int32  `json:"donor_id" gorm:"column:donor_id;primaryKey;autoIncrement"`
	Name    string `json:"name" gorm:"column:name;not null"`
	Email   string `json:"email" gorm:"column:email;not null"`
	Phone   string `json:"phone" gorm:"column:phone;not null"`
	Address string `json:"address" gorm:"column:address;type:text;not null"`

	DonationRequests []DonationRequest `json:"-" gorm:"foreignKey:DonorID"`
}

func (Donor) TableName() string {
	return "donors"
}
