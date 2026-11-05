package server

import (
	"net/http"

	"biteshare-api/ecommerce-api/model"
	"github.com/gin-gonic/gin"
)

// listAdminsHandler returns all admins without exposing password hashes.
func (s *Server) listAdminsHandler(c *gin.Context) {
	// Select only public columns from the admins table.
	var admins []model.Admin
	result := s.db.GORM().
		Select("admin_id, name, email").
		Order("admin_id ASC").
		Find(&admins)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not list admins",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"admins": admins,
	})
}
