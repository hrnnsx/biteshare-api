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
func (h *PaymentHandler) ListProofs(c *gin.Context) {
	proofs, err := h.service.ListProofs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list payment proofs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": proofs})
}

// GetProof handles GET /payments/proofs/:order_number for admins.
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
