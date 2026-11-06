package server

import (
	"net/http"

	"biteshare-api/donation-api/handler"
	"biteshare-api/donation-api/repository"
	"biteshare-api/donation-api/service"
	"biteshare-api/internal/auth"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *DonationServer) RegisterDonationRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Basic routes
	r.GET("/", s.DonationHelloWorldHandler)
	r.GET("/health", s.donationHealthHandler)

	// Initialize donation dependencies
	donationRepository := repository.NewDonationRepository(s.db.GORM())
	donationService := service.NewDonationService(donationRepository)
	donationHandler := handler.NewDonationHandler(donationService)

	// Public routes
	donationRoutes := r.Group("/api/v1/donations")
	{
		donationRoutes.POST("", donationHandler.Create)
	}

	// Admin routes require authentication
	adminDonationRoutes := r.Group("/api/v1/donations")
	adminDonationRoutes.Use(auth.RequireAdmin(s.jwtManager))
	{
		adminDonationRoutes.GET("", donationHandler.GetAll)
		adminDonationRoutes.GET("/:id", donationHandler.GetByID)
		adminDonationRoutes.PATCH("/:id/review", donationHandler.Review)
	}

	return r
}

func (s *DonationServer) DonationHelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Donation API is running"

	c.JSON(http.StatusOK, resp)
}

func (s *DonationServer) donationHealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
