package server

import (
	"net/http"
	"strings"

	"biteshare-api/ecommerce-api/model"
	"biteshare-api/internal/auth"
	"github.com/gin-gonic/gin"
)

// adminLoginRequest contains the credentials submitted by an admin.
type adminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// adminLoginResponse contains only data safe to return to the client.
type adminLoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// adminLoginHandler verifies admin credentials and creates an access token.
func (s *Server) adminLoginHandler(c *gin.Context) {
	// Do not attempt login when the application has no JWT secret configured.
	if s.jwtManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "authentication is not configured"})
		return
	}

	// Parse and validate the JSON request body.
	var request adminLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	// Load the admin by email without exposing whether the email exists.
	var admin model.Admin
	result := s.db.GORM().Where("email = ?", strings.TrimSpace(request.Email)).First(&admin)
	if result.Error != nil || auth.ComparePassword(admin.PasswordHash, request.Password) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Create an access token only after the password has been verified.
	accessToken, err := s.jwtManager.CreateAccessToken(admin.ID, admin.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create access token"})
		return
	}

	// Return a standard bearer-token response.
	c.JSON(http.StatusOK, adminLoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	})
}
