package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"biteshare-api/ecommerce-api/dto"
	"biteshare-api/ecommerce-api/model"
	"biteshare-api/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrProductUnavailable = errors.New("one or more products are unavailable")
	ErrInsufficientStock  = errors.New("insufficient product stock")
	ErrDuplicateProduct  = errors.New("the same product was added more than once")
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db database.Service) *OrderRepository {
	return &OrderRepository{db: db.GORM()}
}

// CreateCheckout creates all checkout records and updates stock atomically.
func (r *OrderRepository) CreateCheckout(ctx context.Context, input dto.CheckoutRequest) (*model.Order, error) {
	var order model.Order

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		buyer := model.Buyer{
			Name:    input.Name,
			Email:   input.Email,
			Phone:   input.Phone,
			Address: input.Address,
		}
		var existingBuyer model.Buyer
		buyerResult := tx.Where("email = ?", input.Email).First(&existingBuyer)
		switch {
		case errors.Is(buyerResult.Error, gorm.ErrRecordNotFound):
			if err := tx.Create(&buyer).Error; err != nil {
				return err
			}
		case buyerResult.Error != nil:
			return buyerResult.Error
		default:
			buyer = existingBuyer
			if err := tx.Model(&buyer).Updates(map[string]interface{}{
				"name": input.Name, "phone": input.Phone, "address": input.Address,
			}).Error; err != nil {
				return err
			}
		}

		quantities := make(map[int32]float64, len(input.Items))
		productIDs := make([]int32, 0, len(input.Items))
		for _, item := range input.Items {
			if _, exists := quantities[item.ProductID]; exists {
				return ErrDuplicateProduct
			}
			quantities[item.ProductID] = item.Quantity
			productIDs = append(productIDs, item.ProductID)
		}

		var products []model.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("product_id IN ? AND status = ?", productIDs, "available").
			Find(&products).Error; err != nil {
			return err
		}
		if len(products) != len(productIDs) {
			return ErrProductUnavailable
		}

		productsByID := make(map[int32]model.Product, len(products))
		for _, product := range products {
			productsByID[product.ID] = product
		}

		order = model.Order{
			BuyerID:         buyer.ID,
			OrderNumber:     newOrderNumber(),
			DeliveryAddress: input.Address,
			DeliveryFee:     input.DeliveryFee,
			Status:          "pending",
			OrderedAt:       time.Now(),
			Buyer:           buyer,
		}
		if order.DeliveryFee < 0 {
			return fmt.Errorf("delivery_fee must be non-negative")
		}

		for _, item := range input.Items {
			product := productsByID[item.ProductID]
			if product.Stock < item.Quantity {
				return fmt.Errorf("%w: %s", ErrInsufficientStock, product.Name)
			}
			order.TotalAmount += item.Quantity * product.Price
		}
		order.TotalAmount += order.DeliveryFee

		if err := tx.Omit("Buyer", "Items").Create(&order).Error; err != nil {
			return err
		}

		for _, item := range input.Items {
			product := productsByID[item.ProductID]
			orderItem := model.OrderItem{
				OrderID:   order.ID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				UnitPrice: product.Price,
				Subtotal:  item.Quantity * product.Price,
				Product:   product,
			}
			if err := tx.Omit("Product").Create(&orderItem).Error; err != nil {
				return err
			}
			order.Items = append(order.Items, orderItem)

			result := tx.Model(&model.Product{}).
				Where("product_id = ? AND stock >= ?", item.ProductID, item.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return fmt.Errorf("%w: %s", ErrInsufficientStock, product.Name)
			}
		}

		payment := map[string]interface{}{
			"order_id": order.ID, "payment_method": input.PaymentMethod,
			"amount": order.TotalAmount, "payment_status": "pending",
		}
		if err := tx.Table("payments").Create(payment).Error; err != nil {
			return err
		}

		delivery := map[string]interface{}{
			"order_id": order.ID, "delivery_address": input.Address,
			"delivery_status": "pending",
		}
		if err := tx.Table("deliveries").Create(delivery).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func newOrderNumber() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("INV-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("INV-%s-%s", time.Now().Format("20060102150405"), hex.EncodeToString(bytes))
}
