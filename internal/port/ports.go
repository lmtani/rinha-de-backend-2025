package port

import (
	"context"
	"time"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
)

// PaymentRepository defines the interface for payment statistics storage
type PaymentRepository interface {
	Add(channel domain.ProcessorChannel, amount float64) error
	GetSummary() (domain.PaymentsSummary, error)
	// GetSummaryInRange returns the summary filtered by the given time range.
	// If from or to are nil, the respective bound is ignored.
	GetSummaryInRange(from, to *time.Time) (domain.PaymentsSummary, error)
}

// PaymentProcessor defines the interface for external payment processors
type PaymentProcessor interface {
	ProcessPayment(ctx context.Context, payment domain.Payment) error
}

// PaymentQueue defines the interface for payment message queue
type PaymentQueue interface {
	Send(payment domain.Payment) error
	Receive() <-chan domain.Payment
	Close() error
}

// CircuitBreaker defines the interface for circuit breaker functionality
type CircuitBreaker interface {
	Execute(func() error) error
	State() string
}

// Store defines the interface for UUID storage
type Store interface {
	Add(uuid string) error
	Exists(uuid string) bool
}
