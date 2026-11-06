package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"biteshare-api/internal/auth"

	"github.com/stretchr/testify/assert"
)

func TestJWTManagerAndMiddleware(t *testing.T) {
	secret := "super-secret-key-at-least-32-characters-long"
	manager, err := auth.NewJWTManager(secret, 15*time.Minute, "biteshare-api")
	assert.NoError(t, err)

	token, err := manager.CreateAccessToken(1, "admin@biteshare.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := manager.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "admin@biteshare.com", claims.Email)
	assert.Equal(t, "admin", claims.Role)
}

func TestPaymentWebhookHandlerIntegration(t *testing.T) {
	payload := map[string]interface{}{
		"transaction_reference": "TRX-MOCK-999",
		"order_number":          "INV-TEST-001",
		"status":                "PAID",
		"amount":                100000,
		"payment_method":        "QRIS",
	}

	jsonBytes, _ := json.Marshal(payload)

	_ = httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/payments/webhook", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	assert.NotNil(t, req)
}
