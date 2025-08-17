package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
	"github.com/lmtani/rinha-de-backend-2025/internal/domain/service"
	"github.com/lmtani/rinha-de-backend-2025/internal/port"
)

// ProcessPaymentsUseCase handles the background processing of payments from the queue
type ProcessPaymentsUseCase struct {
	queue            port.PaymentQueue
	processorService *service.PaymentProcessorService
	running          bool
	mu               sync.Mutex
	instanceID       string
	workerCount      int
}

func NewProcessPaymentsUseCase(
	queue port.PaymentQueue,
	processorService *service.PaymentProcessorService,
	instanceID string,
	workerCount int,
) *ProcessPaymentsUseCase {
	if workerCount <= 0 {
		workerCount = 4 // default worker count
	}

	return &ProcessPaymentsUseCase{
		queue:            queue,
		processorService: processorService,
		instanceID:       instanceID,
		workerCount:      workerCount,
	}
}

// Start begins processing payments from the queue with multiple workers
func (uc *ProcessPaymentsUseCase) Start(ctx context.Context) {
	uc.mu.Lock()
	if uc.running {
		uc.mu.Unlock()
		return
	}
	uc.running = true
	uc.mu.Unlock()

	fmt.Printf("[%s] Starting payment processor with %d workers\n", uc.instanceID, uc.workerCount)

	// Create a channel to receive payments
	paymentChan := uc.queue.Receive()

	// Start multiple worker goroutines
	for i := 0; i < uc.workerCount; i++ {
		workerID := fmt.Sprintf("%s-worker-%d", uc.instanceID, i)
		go uc.startWorker(ctx, paymentChan, workerID)
	}
}

// startWorker starts a single worker goroutine
func (uc *ProcessPaymentsUseCase) startWorker(ctx context.Context, paymentChan <-chan domain.Payment, workerID string) {
	fmt.Printf("[%s] Worker started\n", workerID)

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[%s] Worker recovered from panic: %v\n", workerID, r)
			// Restart the worker after a short delay
			time.Sleep(time.Second)
			go uc.startWorker(ctx, paymentChan, workerID)
		} else {
			fmt.Printf("[%s] Worker stopped\n", workerID)
		}
	}()

	for {
		select {
		case payment, ok := <-paymentChan:
			if !ok {
				fmt.Printf("[%s] Payment channel closed, stopping worker\n", workerID)
				return
			}

			// Process payment with timeout
			processingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := uc.processorService.ProcessPayment(processingCtx, payment)
			cancel()

			if err != nil {
				fmt.Printf("[%s] Failed to process payment %s: %v\n", workerID, payment.CorrelationId, err)
				fmt.Printf("[%s] Re-enqueuing payment %s\n", workerID, payment.CorrelationId)
				// We could add a exponential backoff for retries
				// But I think we need to connect to a redis tracker of what failed.
				uc.queue.Send(payment)
			} else {
				fmt.Printf("[%s] Successfully processed payment %s\n", workerID, payment.CorrelationId)
			}

		case <-ctx.Done():
			fmt.Printf("[%s] Context cancelled, stopping worker\n", workerID)
			return
		}

	}
}

// Stop stops the payment processing
func (uc *ProcessPaymentsUseCase) Stop() error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if !uc.running {
		return nil
	}

	fmt.Printf("[%s] Stopping payment processor\n", uc.instanceID)
	uc.running = false
	return uc.queue.Close()
}
