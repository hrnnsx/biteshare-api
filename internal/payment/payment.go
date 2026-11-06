package payment

import (
	"fmt"
	"time"

	"biteshare-api/internal/database"

	"gorm.io/gorm"
)

type WebhookPayload struct {
	TransactionReference string  `json:"transaction_reference" binding:"required"`
	OrderNumber          string  `json:"order_number" binding:"required"`
	Status               string  `json:"status" binding:"required"` // "PAID", "FAILED", "EXPIRED"
	Amount               float64 `json:"amount" binding:"required"`
	PaymentMethod        string  `json:"payment_method"`
}

type PaymentService interface {
	ProcessWebhook(payload WebhookPayload) error
}

type paymentService struct {
	db database.Service
}

func NewService(db database.Service) PaymentService {
	return &paymentService{db: db}
}

func (s *paymentService) ProcessWebhook(payload WebhookPayload) error {
	return s.db.GORM().Transaction(func(tx *gorm.DB) error {
		// 1. Cari Order berdasarkan order_number
		var order struct {
			OrderID     int32   `gorm:"column:order_id"`
			TotalAmount float64 `gorm:"column:total_amount"`
			OrderStatus string  `gorm:"column:order_status"`
		}
		if err := tx.Table("orders").Where("order_number = ?", payload.OrderNumber).First(&order).Error; err != nil {
			return fmt.Errorf("order not found: %w", err)
		}

		now := time.Now()
		var paymentStatus string
		var nextOrderStatus string

		switch payload.Status {
		case "PAID", "SUCCESS":
			paymentStatus = "paid"
			nextOrderStatus = "confirmed"
		case "FAILED":
			paymentStatus = "failed"
			nextOrderStatus = "cancelled"
		default:
			paymentStatus = "pending"
			nextOrderStatus = order.OrderStatus
		}

		// 2. Update atau Insert data di tabel payments
		updatePayment := map[string]interface{}{
			"payment_status": paymentStatus,
			"paid_at":        now,
		}
		if payload.PaymentMethod != "" {
			updatePayment["payment_method"] = payload.PaymentMethod
		}

		res := tx.Table("payments").
			Where("order_id = ? OR transaction_reference = ?", order.OrderID, payload.TransactionReference).
			Updates(updatePayment)

		if res.Error != nil {
			return fmt.Errorf("failed to update payment: %w", res.Error)
		}

		// 3. Update status pesanan di tabel orders
		if err := tx.Table("orders").Where("order_id = ?", order.OrderID).Update("order_status", nextOrderStatus).Error; err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		return nil
	})
}
