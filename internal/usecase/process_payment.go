package usecase

import (
	"context"
	"fmt"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain/service"
	"github.com/lmtani/rinha-de-backend-2025/internal/port"
)

// ProcessPaymentsUseCase handles the background processing of payments from the queue
type ProcessPaymentsUseCase struct {
	queue            port.PaymentQueue
	processorService *service.PaymentProcessorService
	running          bool
}

func NewProcessPaymentsUseCase(
	queue port.PaymentQueue,
	processorService *service.PaymentProcessorService,
) *ProcessPaymentsUseCase {
	return &ProcessPaymentsUseCase{
		queue:            queue,
		processorService: processorService,
	}
}

// Start begins processing payments from the queue
func (uc *ProcessPaymentsUseCase) Start(ctx context.Context) {
	if uc.running {
		return
	}

	uc.running = true

	go func() {
		defer func() {
			uc.running = false
		}()

		for {
			select {
			case payment, ok := <-uc.queue.Receive():

				if !ok {
					fmt.Println("Payment queue closed, stopping processor")
					return
				}

				if err := uc.processorService.ProcessPayment(ctx, payment); err != nil {
					fmt.Printf("Failed to process payment %s: %v\n", payment.CorrelationId, err)
				} else {
					fmt.Printf("Successfully processed payment %s\n", payment.CorrelationId)
				}

			case <-ctx.Done():
				fmt.Println("Context cancelled, stopping payment processor")
				return
			}
		}
	}()
}

// Stop stops the payment processing
func (uc *ProcessPaymentsUseCase) Stop() error {
	return uc.queue.Close()
}
