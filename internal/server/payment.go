package server

import (
	"biteshare-api/internal/payment"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) paymentWebhookHandler(c *gin.Context) {
	var payload payment.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payService := payment.NewService(s.db)
	if err := payService.ProcessWebhook(payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "payment webhook processed successfully",
	})
}
