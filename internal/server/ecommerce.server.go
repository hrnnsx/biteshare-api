package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	producthandler "biteshare-api/ecommerce-api/handler"
	"biteshare-api/ecommerce-api/service"
	"biteshare-api/internal/auth"
	"biteshare-api/internal/database"
)

type Server struct {
	port       int
	jwtManager *auth.JWTManager
	products   *producthandler.ProductHandler
	db         database.Service
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080
	}
	return NewCustomServer(port)
}

func NewCustomServer(port int) *http.Server {
	tokenTTL := 24 * time.Hour
	if configuredTTL := os.Getenv("JWT_ACCESS_TOKEN_TTL"); configuredTTL != "" {
		parsedTTL, err := time.ParseDuration(configuredTTL)
		if err != nil {
			log.Fatalf("invalid JWT_ACCESS_TOKEN_TTL: %v", err)
		}
		tokenTTL = parsedTTL
	}

	var jwtManager *auth.JWTManager
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		issuer := os.Getenv("JWT_ISSUER")
		if issuer == "" {
			issuer = "biteshare-api"
		}
		var err error
		jwtManager, err = auth.NewJWTManager(secret, tokenTTL, issuer)
		if err != nil {
			log.Fatalf("invalid JWT configuration: %v", err)
		}
	}

	s := &Server{
		port:       port,
		jwtManager: jwtManager,
		db:         database.New(),
	}
	s.products = producthandler.NewProductHandler(s.db, service.NewImageKitUploader())

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
