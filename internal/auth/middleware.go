package auth

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// adminIDContextKey is the private key used to store the authenticated admin ID.
const adminIDContextKey = "auth.admin_id"

// RequireAdmin protects a Gin route with a valid admin bearer token.
func RequireAdmin(manager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Reject protected requests when JWT configuration is missing.
		if manager == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "authentication is not configured",
			})
			return
		}

		// Read the Authorization header in the form: Bearer <token>.
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "bearer token is required",
			})
			return
		}

		// Validate the JWT signature, issuer, and expiration time.
		claims, err := manager.ValidateAccessToken(parts[1])
		if err != nil || claims.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid access token",
			})
			return
		}

		// Convert the JWT subject into the admin ID used by repositories.
		adminID, err := strconv.ParseUint(claims.Subject, 10, 64)
		if err != nil || adminID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid admin identity",
			})
			return
		}

		// Make the authenticated admin ID available to downstream handlers.
		c.Set(adminIDContextKey, adminID)
		c.Next()
	}
}

// AdminID returns the authenticated admin ID stored by RequireAdmin.
func AdminID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(adminIDContextKey)
	if !exists {
		return 0, false
	}

	adminID, ok := value.(uint64)
	return adminID, ok && adminID > 0
}
