package server

import (
	"net/http"

	"biteshare-api/internal/auth"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.GET("/", s.HelloWorldHandler)
	r.GET("/health", s.healthHandler)

	// Admin login & listing
	r.POST("/auth/login", s.adminLoginHandler)
	r.GET("/admins", auth.RequireAdmin(s.jwtManager), s.listAdminsHandler)

	// Payment endpoints require an authenticated admin JWT.
	r.POST("/payments/webhook", auth.RequireAdmin(s.jwtManager), s.paymentWebhookHandler)
	r.POST("/payments/proof", s.payments.SubmitProof)
	r.GET("/payments/proofs", auth.RequireAdmin(s.jwtManager), s.payments.ListProofs)
	r.GET("/payments/proofs/:order_number", auth.RequireAdmin(s.jwtManager), s.payments.GetProof)

	// Public product catalog endpoints
	r.GET("/ecom/products", s.products.ListProducts)
	r.GET("/ecom/product/:id", s.products.GetProduct)
	r.POST("/ecom/checkout", s.orders.Checkout)

	// Product mutations require an authenticated admin JWT
	r.POST("/ecom/products", auth.RequireAdmin(s.jwtManager), s.products.CreateProduct)
	r.PATCH("/ecom/products/:id", auth.RequireAdmin(s.jwtManager), s.products.UpdateProduct)
	r.DELETE("/ecom/product/:id", auth.RequireAdmin(s.jwtManager), s.products.DeleteProduct)

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"
	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
