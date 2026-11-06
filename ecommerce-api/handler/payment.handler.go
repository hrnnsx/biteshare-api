package handler

import (
	"errors"
	"net/http"
	"strings"

	ecomservice "biteshare-api/ecommerce-api/service"
	"biteshare-api/internal/payment"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	service  payment.PaymentService
	uploader *ecomservice.ImageKitUploader
}

func NewPaymentHandler(paymentService payment.PaymentService, uploader *ecomservice.ImageKitUploader) *PaymentHandler {
	return &PaymentHandler{service: paymentService, uploader: uploader}
}

// SubmitProof handles POST /payments/proof.
// SubmitProof submits a manual payment proof.
//
// @Summary Submit payment proof
// @Description Buyer submits a payment proof image for an existing order. The image is uploaded to ImageKit and its URL is stored in the database.
// @Tags Payments
// @Accept multipart/form-data
// @Produce json
// @Param order_number formData string true "Order number"
// @Param transaction_reference formData string false "Bank transaction reference"
// @Param image formData file true "Payment proof image"
// @Success 200 {object} map[string]interface{} "Payment proof submitted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or missing payment proof"
// @Failure 404 {object} map[string]interface{} "Payment not found"
// @Failure 409 {object} map[string]interface{} "Payment is already completed"
// @Failure 502 {object} map[string]interface{} "Image upload or payment proof submission failed"
// @Failure 503 {object} map[string]interface{} "ImageKit is not configured"
// @Router /payments/proof [post]
func (h *PaymentHandler) SubmitProof(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request must be multipart/form-data"})
		return
	}

	orderNumber := strings.TrimSpace(c.PostForm("order_number"))
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_number is required"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image is required"})
		return
	}
	if h.uploader == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ImageKit is not configured"})
		return
	}

	result, err := h.service.SubmitProof(
		c.Request.Context(),
		orderNumber,
		strings.TrimSpace(c.PostForm("transaction_reference")),
		file,
		h.uploader,
	)
	if err != nil {
		switch {
		case errors.Is(err, payment.ErrPaymentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		case errors.Is(err, payment.ErrPaymentAlreadyClosed):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "payment proof submitted successfully",
		"data":    result,
	})
}

// ListProofs handles GET /payments/proofs for admins.
// ListProofs retrieves all submitted payment proofs.
//
// @Summary Get all payment proofs
// @Description Admin retrieves a list of payment proofs submitted by buyers.
// @Tags Payments
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of payment proofs"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Could not list payment proofs"
// @Router /payments/proofs [get]
func (h *PaymentHandler) ListProofs(c *gin.Context) {
	proofs, err := h.service.ListProofs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list payment proofs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": proofs})
}

// GetProof handles GET /payments/proofs/:order_number for admins.
// GetProof retrieves payment proof details by order number.
//
// @Summary Get payment proof details
// @Description Admin retrieves payment proof information and related order details using the order number.
// @Tags Payments
// @Produce json
// @Security BearerAuth
// @Param order_number path string true "Order number"
// @Success 200 {object} map[string]interface{} "Payment proof details"
// @Failure 400 {object} map[string]interface{} "Order number is required"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Payment proof not found"
// @Failure 500 {object} map[string]interface{} "Could not get payment proof"
// @Router /payments/proofs/{order_number} [get]
func (h *PaymentHandler) GetProof(c *gin.Context) {
	orderNumber := strings.TrimSpace(c.Param("order_number"))
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_number is required"})
		return
	}

	proof, err := h.service.GetProof(c.Request.Context(), orderNumber)
	if err != nil {
		if errors.Is(err, payment.ErrPaymentProofNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment proof not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get payment proof"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": proof})
}
