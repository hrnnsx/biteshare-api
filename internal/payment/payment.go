package payment

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
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
	SubmitProof(ctx context.Context, orderNumber, transactionReference string, file *multipart.FileHeader, uploader ImageUploader) (*ProofResult, error)
	ListProofs(ctx context.Context) ([]PaymentProof, error)
	GetProof(ctx context.Context, orderNumber string) (*PaymentProofDetail, error)
}

type ImageUploader interface {
	Upload(ctx context.Context, file *multipart.FileHeader) (string, string, error)
}

type ProofResult struct {
	OrderNumber string `json:"order_number"`
	ImageURL    string `json:"image_url"`
	FileID      string `json:"file_id"`
}

type PaymentProof struct {
	PaymentID             int32   `json:"payment_id"`
	OrderNumber           string  `json:"order_number"`
	BuyerName             string  `json:"buyer_name"`
	BuyerEmail            string  `json:"buyer_email"`
	Amount                float64 `json:"amount"`
	PaymentMethod         string  `json:"payment_method"`
	PaymentStatus         string  `json:"payment_status"`
	TransactionReference  *string `json:"transaction_reference,omitempty"`
	ProofImageURL         string  `json:"proof_image_url"`
	ProofFileID           string  `json:"proof_file_id"`
}

type PaymentProofItem struct {
	ProductID   int32   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
}

type PaymentProofDetail struct {
	PaymentProof
	OrderStatus     string             `json:"order_status"`
	DeliveryAddress string             `json:"delivery_address"`
	Items           []PaymentProofItem `json:"items"`
}

var (
	ErrPaymentNotFound      = fmt.Errorf("payment not found")
	ErrPaymentAlreadyClosed = fmt.Errorf("payment is already completed")
	ErrPaymentProofNotFound = fmt.Errorf("payment proof not found")
)

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

// SubmitProof uploads a manual-transfer proof and stores its ImageKit details.
func (s *paymentService) SubmitProof(ctx context.Context, orderNumber, transactionReference string, file *multipart.FileHeader, uploader ImageUploader) (*ProofResult, error) {
	imageURL, fileID, err := uploader.Upload(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("upload payment proof: %w", err)
	}

	result := &ProofResult{
		OrderNumber: orderNumber,
		ImageURL:    imageURL,
		FileID:      fileID,
	}

	err = s.db.GORM().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var payment struct {
			PaymentID   int32  `gorm:"column:payment_id"`
			PaymentStatus string `gorm:"column:payment_status"`
		}
		query := tx.Table("payments AS p").
			Select("p.payment_id, p.payment_status").
			Joins("JOIN orders AS o ON o.order_id = p.order_id").
			Where("o.order_number = ?", orderNumber).
			First(&payment)
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return ErrPaymentNotFound
		}
		if query.Error != nil {
			return query.Error
		}
		if payment.PaymentStatus == "paid" || payment.PaymentStatus == "failed" {
			return ErrPaymentAlreadyClosed
		}

		values := map[string]interface{}{
			"proof_image_url": imageURL,
			"proof_file_id":   fileID,
		}
		if transactionReference != "" {
			values["transaction_reference"] = transactionReference
		}
		return tx.Table("payments").Where("payment_id = ?", payment.PaymentID).Updates(values).Error
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *paymentService) ListProofs(ctx context.Context) ([]PaymentProof, error) {
	var proofs []PaymentProof
	err := s.db.GORM().WithContext(ctx).
		Table("payments AS p").
		Select(`
			p.payment_id,
			o.order_number,
			b.name AS buyer_name,
			b.email AS buyer_email,
			p.amount,
			p.payment_method,
			p.payment_status,
			p.transaction_reference,
			p.proof_image_url,
			p.proof_file_id`).
		Joins("JOIN orders AS o ON o.order_id = p.order_id").
		Joins("JOIN buyers AS b ON b.buyer_id = o.buyer_id").
		Where("p.proof_image_url IS NOT NULL AND p.proof_image_url <> ''").
		Order("p.payment_id DESC").
		Scan(&proofs).Error
	if err != nil {
		return nil, fmt.Errorf("list payment proofs: %w", err)
	}
	if proofs == nil {
		proofs = []PaymentProof{}
	}
	return proofs, nil
}

func (s *paymentService) GetProof(ctx context.Context, orderNumber string) (*PaymentProofDetail, error) {
	var proof PaymentProofDetail
	result := s.db.GORM().WithContext(ctx).
		Table("payments AS p").
		Select(`
			p.payment_id,
			o.order_number,
			b.name AS buyer_name,
			b.email AS buyer_email,
			p.amount,
			p.payment_method,
			p.payment_status,
			p.transaction_reference,
			p.proof_image_url,
			p.proof_file_id,
			o.order_status,
			o.delivery_address`).
		Joins("JOIN orders AS o ON o.order_id = p.order_id").
		Joins("JOIN buyers AS b ON b.buyer_id = o.buyer_id").
		Where("o.order_number = ? AND p.proof_image_url IS NOT NULL AND p.proof_image_url <> ''", orderNumber).
		First(&proof)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrPaymentProofNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get payment proof: %w", result.Error)
	}

	err := s.db.GORM().WithContext(ctx).
		Table("order_items AS oi").
		Select(`
			oi.product_id,
			fp.product_name,
			oi.quantity,
			oi.unit_price,
			oi.subtotal`).
		Joins("JOIN orders AS o ON o.order_id = oi.order_id").
		Joins("JOIN food_products AS fp ON fp.product_id = oi.product_id").
		Where("o.order_number = ?", orderNumber).
		Order("oi.order_item_id ASC").
		Scan(&proof.Items).Error
	if err != nil {
		return nil, fmt.Errorf("get payment proof items: %w", err)
	}
	if proof.Items == nil {
		proof.Items = []PaymentProofItem{}
	}
	return &proof, nil
}
