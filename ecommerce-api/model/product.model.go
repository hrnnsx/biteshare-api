package model

import "time"

// Product represents a food product that can be shown in the public catalog.
type Product struct {
	ID          int32          `json:"product_id" gorm:"column:product_id;primaryKey;autoIncrement"`
	AdminID     int32          `json:"-" gorm:"column:admin_id;not null"`
	Name        string         `json:"product_name" gorm:"column:product_name;not null"`
	Description *string        `json:"description,omitempty" gorm:"column:description;type:text"`
	Price       float64        `json:"price" gorm:"column:price;not null"`
	Stock       float64        `json:"stock" gorm:"column:stock;not null"`
	Unit        string         `json:"unit" gorm:"column:unit;not null"`
	ExpiryDate  *time.Time     `json:"expiry_date,omitempty" gorm:"column:expiry_date;type:date"`
	Status      string         `json:"status" gorm:"column:status;not null"`
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at;not null"`
	Images      []ProductImage `json:"images,omitempty" gorm:"foreignKey:ProductID"`
}

func (Product) TableName() string {
	return "food_products"
}

// ProductImage represents an image associated with a product.
type ProductImage struct {
	ID        int32     `json:"image_id" gorm:"column:image_id;primaryKey;autoIncrement"`
	ProductID int32     `json:"-" gorm:"column:product_id;not null"`
	ImageURL  string    `json:"image_url" gorm:"column:image_url;not null"`
	FileID    *string   `json:"file_id,omitempty" gorm:"column:file_id"`
	IsPrimary bool      `json:"is_primary" gorm:"column:is_primary;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (ProductImage) TableName() string {
	return "product_images"
}
