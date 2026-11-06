package server

import (
	"net/http"

	"biteshare-api/donation-api/handler"
	"biteshare-api/donation-api/repository"
	"biteshare-api/donation-api/service"
	"biteshare-api/internal/auth"
	"biteshare-api/internal/imagekit"
	"biteshare-api/internal/mailer"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *DonationServer) RegisterDonationRoutes() http.Handler {
	r := gin.Default()

	// CORS configuration.
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
			"PATCH",
		},
		AllowHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
		},
		AllowCredentials: true,
	}))

	// Health check routes.
	r.GET("/", s.DonationHelloWorldHandler)
	r.GET("/health", s.donationHealthHandler)

	// Initialize repositories.
	donationRepo := repository.NewDonationRepository(
		s.db.GORM(),
	)

	imageKitClient, err := imagekit.NewClient()
	if err != nil {
		panic(err)
	}

	pickupScheduleRepo := repository.NewPickupScheduleRepository(
		s.db.GORM(),
	)

	emailMailer := mailer.New()

	// Initialize services.
	donationService := service.NewDonationService(
		donationRepo, imageKitClient, emailMailer,
	)

	pickupScheduleService := service.NewPickupScheduleService(
		pickupScheduleRepo, emailMailer,
	)

	// Initialize handlers.
	donationHandler := handler.NewDonationHandler(
		donationService,
	)

	pickupScheduleHandler := handler.NewPickupScheduleHandler(
		pickupScheduleService,
	)

	// Public donation routes.
	donationRoutes := r.Group("/api/v1/donations")
	{
		donationRoutes.POST("", donationHandler.Create)
	}

	// Admin donation routes.
	adminDonationRoutes := r.Group("/api/v1/donations")
	adminDonationRoutes.Use(auth.RequireAdmin(s.jwtManager))
	{
		adminDonationRoutes.GET("", donationHandler.GetAll)
		adminDonationRoutes.GET("/:id", donationHandler.GetByID)
		adminDonationRoutes.PATCH("/:id/review", donationHandler.Review)

		// Pickup schedule route.
		adminDonationRoutes.POST(
			"/:id/schedules",
			pickupScheduleHandler.SchedulePickup,
		)
	}

	return r
}

// DonationHelloWorldHandler handles the root endpoint.
func (s *DonationServer) DonationHelloWorldHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Donation API is running",
	})
}

// donationHealthHandler handles the health check endpoint.
func (s *DonationServer) donationHealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
