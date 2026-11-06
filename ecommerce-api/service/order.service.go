package service

import (
	"context"
	"log"

	"biteshare-api/ecommerce-api/dto"
	"biteshare-api/ecommerce-api/model"
	"biteshare-api/ecommerce-api/repository"
	"biteshare-api/internal/mailer"
)

type OrderService struct {
	repository *repository.OrderRepository
	mailer     mailer.Mailer
}

func NewOrderService(orderRepository *repository.OrderRepository, orderMailer mailer.Mailer) *OrderService {
	return &OrderService{repository: orderRepository, mailer: orderMailer}
}

func (s *OrderService) Checkout(ctx context.Context, input dto.CheckoutRequest) (*model.Order, error) {
	order, err := s.repository.CreateCheckout(ctx, input)
	if err != nil {
		return nil, err
	}

	// Email delivery happens after the transaction succeeds. An email failure
	// must not undo a successfully created order.
	if s.mailer != nil {
		if err := s.mailer.SendOrderReceiptEmail(
			order.Buyer.Email,
			order.Buyer.Name,
			order.OrderNumber,
			order.TotalAmount,
		); err != nil {
			log.Printf("could not send order receipt for %s: %v", order.OrderNumber, err)
		}
	}

	return order, nil
}
