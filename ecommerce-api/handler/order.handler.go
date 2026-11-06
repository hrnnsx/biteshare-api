package handler

import (
	"errors"
	"net/http"

	"biteshare-api/ecommerce-api/dto"
	"biteshare-api/ecommerce-api/repository"
	"biteshare-api/ecommerce-api/service"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{service: orderService}
}

// Checkout handles POST /ecom/checkout.
func (h *OrderHandler) Checkout(c *gin.Context) {
	var input dto.CheckoutRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.service.Checkout(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateProduct),
			errors.Is(err, repository.ErrProductUnavailable),
			errors.Is(err, repository.ErrInsufficientStock):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create order"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "checkout completed successfully",
		"order":   order,
	})
}
