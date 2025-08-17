package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lmtani/rinha-de-backend-2025/internal/admin/handler"
)

func main() {
	// Get configuration from environment variables
	defaultProcessorURL := os.Getenv("PROCESSOR_DEFAULT_URL")
	if defaultProcessorURL == "" {
		defaultProcessorURL = "http://localhost:8001"
	}

	fallbackProcessorURL := os.Getenv("PROCESSOR_FALLBACK_URL")
	if fallbackProcessorURL == "" {
		fallbackProcessorURL = "http://localhost:8002"
	}

	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		adminToken = "123" // Default token
	}

	port := os.Getenv("ADMIN_PORT")
	if port == "" {
		port = ":8081"
	}

	// Create Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Create admin handler
	adminHandler := handler.NewAdminHandler(defaultProcessorURL, fallbackProcessorURL, adminToken)

	// Register routes
	adminHandler.RegisterRoutes(r)

	// Start server
	log.Printf("Starting admin server on %s", port)
	log.Printf("Default Processor: %s", defaultProcessorURL)
	log.Printf("Fallback Processor: %s", fallbackProcessorURL)
	log.Printf("Admin Token: %s", adminToken)

	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start admin server:", err)
	}
}
