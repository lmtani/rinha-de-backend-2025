package service

import (
	"context"
	"fmt"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
	"github.com/lmtani/rinha-de-backend-2025/internal/port"
)

// PaymentProcessorService handles the business logic for payment processing
type PaymentProcessorService struct {
	defaultProcessor  port.PaymentProcessor
	fallbackProcessor port.PaymentProcessor
	circuitBreaker    port.CircuitBreaker
	repository        port.PaymentRepository
}

// NewPaymentProcessorService creates a new payment processor service
func NewPaymentProcessorService(
	defaultProcessor, fallbackProcessor port.PaymentProcessor,
	circuitBreaker port.CircuitBreaker,
	repository port.PaymentRepository,
) *PaymentProcessorService {
	return &PaymentProcessorService{
		defaultProcessor:  defaultProcessor,
		fallbackProcessor: fallbackProcessor,
		circuitBreaker:    circuitBreaker,
		repository:        repository,
	}
}

// ProcessPayment processes a payment using the default processor with fallback
func (s *PaymentProcessorService) ProcessPayment(ctx context.Context, payment domain.Payment) error {
	if err := payment.Validate(); err != nil {
		return fmt.Errorf("invalid payment: %w", err)
	}

	amount, _ := payment.AmountAsFloat() // Already validated above

	// Try default processor with circuit breaker
	err := s.circuitBreaker.Execute(func() error {
		return s.defaultProcessor.ProcessPayment(ctx, payment)
	})

	if err == nil {
		// Success with default processor
		if err := s.repository.Add(domain.DefaultProcessor, amount); err != nil {
			// Log error but don't fail the payment
			fmt.Printf("Failed to record payment stats: %v\n", err)
		}
		return nil
	}

	// Try fallback processor
	if err := s.fallbackProcessor.ProcessPayment(ctx, payment); err != nil {
		return fmt.Errorf("both default and fallback processors failed: %w", err)
	}

	// Success with fallback processor
	if err := s.repository.Add(domain.FallbackProcessor, amount); err != nil {
		// Log error but don't fail the payment
		fmt.Printf("Failed to record payment stats: %v\n", err)
	}

	return nil
}
