package main

import (
	"context"

	"github.com/lmtani/rinha-de-backend-2025/internal/adapter/http_client"
	"github.com/lmtani/rinha-de-backend-2025/internal/adapter/http_server"
	"github.com/lmtani/rinha-de-backend-2025/internal/adapter/in_memory_repository"
	"github.com/lmtani/rinha-de-backend-2025/internal/config"
	"github.com/lmtani/rinha-de-backend-2025/internal/domain/service"
	"github.com/lmtani/rinha-de-backend-2025/internal/port"
	"github.com/lmtani/rinha-de-backend-2025/internal/usecase"
)

// Container holds all application dependencies
type Container struct {
	Config *config.Config

	// Ports/Interfaces
	Repository        port.PaymentRepository
	Queue             port.PaymentQueue
	Store             port.InMemoryStore
	DefaultProcessor  port.PaymentProcessor
	FallbackProcessor port.PaymentProcessor
	CircuitBreaker    port.CircuitBreaker

	// Domain Services
	PaymentProcessorService *service.PaymentProcessorService

	// Use Cases
	RequestPaymentUC  *usecase.RequestPaymentUseCase
	AuditPaymentsUC   *usecase.AuditPaymentsUseCase
	ProcessPaymentsUC *usecase.ProcessPaymentsUseCase

	// Infrastructure
	HTTPServer *http_server.Server
}

// NewContainer creates and wires all dependencies
func NewContainer() *Container {
	c := &Container{}

	// Load configuration
	c.Config = config.Load()

	// Initialize infrastructure adapters
	c.Repository = in_memory_repository.NewInMemoryRepository()
	c.Queue = in_memory_repository.NewInMemoryQueue(c.Config.Processor.QueueBufferSize)
	c.Store = in_memory_repository.NewInMemoryStore()

	// Initialize HTTP clients
	c.DefaultProcessor = http_client.NewPaymentProcessorClient(
		c.Config.Processor.DefaultURL,
		c.Config.Processor.Timeout,
	)
	c.FallbackProcessor = http_client.NewPaymentProcessorClient(
		c.Config.Processor.FallbackURL,
		c.Config.Processor.Timeout,
	)

	// Initialize circuit breaker
	c.CircuitBreaker = http_client.NewCircuitBreakerAdapter(
		"payment-processor",
		c.Config.Processor.CircuitBreaker.MaxRequests,
		c.Config.Processor.CircuitBreaker.Interval,
		c.Config.Processor.CircuitBreaker.Timeout,
		c.Config.Processor.CircuitBreaker.FailureRatio,
		c.Config.Processor.CircuitBreaker.MinRequests,
	)

	// Initialize domain services
	c.PaymentProcessorService = service.NewPaymentProcessorService(
		c.DefaultProcessor,
		c.FallbackProcessor,
		c.CircuitBreaker,
		c.Repository,
	)

	// Initialize use cases
	c.RequestPaymentUC = usecase.NewRequestPaymentUseCase(c.Queue, c.Store)
	c.AuditPaymentsUC = usecase.NewAuditPaymentsUseCase(c.Repository)
	c.ProcessPaymentsUC = usecase.NewProcessPaymentsUseCase(c.Queue, c.PaymentProcessorService)

	// Initialize HTTP server
	c.HTTPServer = http_server.NewServer(c.RequestPaymentUC, c.AuditPaymentsUC, &c.Config.Server)

	return c
}

// Start starts all background services
func (c *Container) Start(ctx context.Context) {
	c.ProcessPaymentsUC.Start(ctx)
}

// Stop gracefully stops all services
func (c *Container) Stop() error {
	return c.ProcessPaymentsUC.Stop()
}
