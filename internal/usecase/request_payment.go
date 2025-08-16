package usecase

import (
	"context"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
	"github.com/lmtani/rinha-de-backend-2025/internal/port"
)

// RequestPaymentUseCase handles payment request operations
type RequestPaymentUseCase struct {
	queue port.PaymentQueue
	store port.InMemoryStore
}

// NewRequestPaymentUseCase creates a new request payment use case
func NewRequestPaymentUseCase(queue port.PaymentQueue, store port.InMemoryStore) *RequestPaymentUseCase {
	return &RequestPaymentUseCase{
		queue: queue,
		store: store,
	}
}

// Execute processes a payment request by adding it to the queue
func (uc *RequestPaymentUseCase) Execute(ctx context.Context, payment domain.Payment) error {
	if err := payment.Validate(); err != nil {
		return err
	}

	if err := uc.store.Add(payment.CorrelationId); err != nil {
		return err
	}
	return uc.queue.Send(payment)
}
