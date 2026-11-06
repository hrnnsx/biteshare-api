package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"biteshare-api/internal/server"

	_ "github.com/joho/godotenv/autoload"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("[Donation API] shutting down gracefully, press Ctrl+C again to force")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("[Donation API] forced shutdown error: %v", err)
	}
	log.Println("[Donation API] server exiting")
	done <- true
}

func main() {
	// Donation API berjalan di port yang berbeda (default 8081)
	portStr := os.Getenv("DONATION_PORT")
	if portStr == "" {
		portStr = "8081"
	}
	port, _ := strconv.Atoi(portStr)

	// Inisialisasi server dengan port donation
	srv := server.NewCustomServer(port)

	done := make(chan bool, 1)
	go gracefulShutdown(srv, done)

	log.Printf("[Donation API] running on port :%d\n", port)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("[Donation API] server error: %s", err))
	}

	<-done
	log.Println("[Donation API] Graceful shutdown complete.")
}
