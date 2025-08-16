package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Create dependency injection container
	container := NewContainer()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background services
	container.Start(ctx)

	// Setup graceful shutdown
	setupGracefulShutdown(cancel, container)

	// Start HTTP server
	log.Printf("Starting payment service on %s", container.Config.Server.Port)
	if err := container.HTTPServer.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupGracefulShutdown(cancel context.CancelFunc, container *Container) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Received shutdown signal, shutting down gracefully...")

		// Cancel context to stop background processes
		cancel()

		// Stop services
		if err := container.Stop(); err != nil {
			log.Printf("Error stopping services: %v", err)
		}

		log.Println("Shutdown complete")
		os.Exit(0)
	}()
}
