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

// Checkout creates a new customer order.
//
// @Summary Checkout products
// @Description Creates an order for a buyer, validates product availability and stock, calculates the order total, and initiates the order process.
// @Tags Orders
// @Accept json
// @Produce json
// @Param request body dto.CheckoutRequest true "Buyer and order information"
// @Success 201 {object} map[string]interface{} "Checkout completed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid checkout information"
// @Failure 409 {object} map[string]interface{} "Product unavailable, duplicate product, or insufficient stock"
// @Failure 500 {object} map[string]interface{} "Could not create order"
// @Router /ecom/checkout [post]
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
