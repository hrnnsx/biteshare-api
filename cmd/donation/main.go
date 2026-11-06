package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"biteshare-api/internal/server"

	_ "github.com/joho/godotenv/autoload"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()
	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")
	done <- true
}

// @title BiteShare Donation API
// @version 1.0
// @description API untuk pengelolaan donasi makanan, mulai dari pengajuan donasi, peninjauan admin, hingga penjadwalan pickup.
// @termsOfService http://swagger.io/terms/

// @contact.name BiteShare Development Team

// @host localhost:8081
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan JWT dengan format: Bearer {token}
func main() {
	apiServer := server.NewDonationServer()

	done := make(chan bool, 1)
	go gracefulShutdown(apiServer, done)

	log.Printf("Donation API is running on %s", apiServer.Addr)
	err := apiServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for graceful shutdown to complete.
	<-done
	log.Println("Graceful shutdown complete.")
}
