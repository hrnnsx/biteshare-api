package model

import "time"

type Buyer struct {
	ID      int32  `json:"buyer_id" gorm:"column:buyer_id;primaryKey;autoIncrement"`
	Name    string `json:"name" gorm:"column:name;not null"`
	Email   string `json:"email" gorm:"column:email;not null"`
	Phone   string `json:"phone" gorm:"column:phone;not null"`
	Address string `json:"address" gorm:"column:address;not null"`
}

func (Buyer) TableName() string { return "buyers" }

type Order struct {
	ID              int32       `json:"order_id" gorm:"column:order_id;primaryKey;autoIncrement"`
	BuyerID         int32       `json:"buyer_id" gorm:"column:buyer_id;not null"`
	OrderNumber     string      `json:"order_number" gorm:"column:order_number;unique;not null"`
	DeliveryAddress string      `json:"delivery_address" gorm:"column:delivery_address;not null"`
	TotalAmount     float64     `json:"total_amount" gorm:"column:total_amount;not null"`
	DeliveryFee     float64     `json:"delivery_fee" gorm:"column:delivery_fee;not null"`
	Status          string      `json:"order_status" gorm:"column:order_status;not null"`
	OrderedAt       time.Time   `json:"ordered_at" gorm:"column:ordered_at;not null"`
	Buyer           Buyer       `json:"buyer" gorm:"foreignKey:BuyerID"`
	Items           []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	ID        int32   `json:"order_item_id" gorm:"column:order_item_id;primaryKey;autoIncrement"`
	OrderID   int32   `json:"order_id" gorm:"column:order_id;not null"`
	ProductID int32   `json:"product_id" gorm:"column:product_id;not null"`
	Quantity  float64 `json:"quantity" gorm:"column:quantity;not null"`
	UnitPrice float64 `json:"unit_price" gorm:"column:unit_price;not null"`
	Subtotal  float64 `json:"subtotal" gorm:"column:subtotal;->"`
	Product   Product `json:"product" gorm:"foreignKey:ProductID"`
}

func (OrderItem) TableName() string { return "order_items" }
