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
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop()

	// Give the server 5 seconds to finish ongoing requests.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that shutdown is complete.
	done <- true
}

func main() {
	apiServer := server.NewDonationServer()

	// Create a done channel to signal when shutdown is complete.
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine.
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
