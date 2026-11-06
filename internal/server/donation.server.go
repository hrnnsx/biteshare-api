package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"biteshare-api/internal/auth"
	"biteshare-api/internal/database"
)

type DonationServer struct {
	port       int
	jwtManager *auth.JWTManager
	db         database.Service
}

func NewDonationServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("DONATION_PORT"))

	if port <= 0 {
		port = 8081
	}

	tokenTTL := 24 * time.Hour

	// Use a configured token lifetime when one is provided.
	if configuredTTL := os.Getenv("JWT_ACCESS_TOKEN_TTL"); configuredTTL != "" {
		parsedTTL, err := time.ParseDuration(configuredTTL)
		if err != nil {
			log.Fatalf("invalid JWT_ACCESS_TOKEN_TTL: %v", err)
		}
		tokenTTL = parsedTTL
	}

	// Keep login disabled until a secure JWT secret is configured.
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

	newDonationServer := &DonationServer{
		port:       port,
		jwtManager: jwtManager,
		db:         database.New(),
	}

	// Declare Donation Server config.
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newDonationServer.port),
		Handler:      newDonationServer.RegisterDonationRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
