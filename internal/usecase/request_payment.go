package usecase

import (
	"context"
	"fmt"
	"os"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
	"github.com/lmtani/rinha-de-backend-2025/internal/port"
)

// RequestPaymentUseCase handles payment request operations
type RequestPaymentUseCase struct {
	queue      port.PaymentQueue
	store      port.Store
	instanceID string
}

// NewRequestPaymentUseCase creates a new request payment use case
func NewRequestPaymentUseCase(queue port.PaymentQueue, store port.Store) *RequestPaymentUseCase {
	// Get instance ID from environment or generate a default
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "default-instance"
	}

	return &RequestPaymentUseCase{
		queue:      queue,
		store:      store,
		instanceID: instanceID,
	}
}

// Execute processes a payment request by adding it to the queue
func (uc *RequestPaymentUseCase) Execute(ctx context.Context, payment domain.Payment) error {
	if err := payment.Validate(); err != nil {
		return err
	}

	fmt.Printf("[%s] Received payment request: %s\n", uc.instanceID, payment.CorrelationId)

	if err := uc.store.Add(payment.CorrelationId); err != nil {
		fmt.Printf("[%s] Failed to add payment to store: %v\n", uc.instanceID, err)
		return err
	}

	if err := uc.queue.Send(payment); err != nil {
		fmt.Printf("[%s] Failed to send payment to queue: %v\n", uc.instanceID, err)
		return err
	}

	fmt.Printf("[%s] Successfully queued payment: %s\n", uc.instanceID, payment.CorrelationId)
	return nil
}
