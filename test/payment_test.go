package test

import (
	"context"
	"testing"

	"github.com/lmtani/rinha-de-backend-2025/internal/adapter/in_memory_repository"
	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
	"github.com/lmtani/rinha-de-backend-2025/internal/usecase"
)

func TestPaymentFlow(t *testing.T) {
	// Arrange
	repository := in_memory_repository.NewInMemoryRepository()
	queue := in_memory_repository.NewInMemoryQueue(10)
	store := in_memory_repository.NewInMemoryStore()

	requestUC := usecase.NewRequestPaymentUseCase(queue, store)
	auditUC := usecase.NewAuditPaymentsUseCase(repository)

	payment := domain.Payment{
		CorrelationId: "test-123",
		Amount:        "100.50",
	}

	ctx := context.Background()

	// Act
	err := requestUC.Execute(ctx, payment)
	if err != nil {
		t.Fatalf("Failed to execute request payment: %v", err)
	}

	// Verify payment was queued
	select {
	case receivedPayment := <-queue.Receive():
		if receivedPayment.CorrelationId != payment.CorrelationId {
			t.Errorf("Expected correlation ID %s, got %s", payment.CorrelationId, receivedPayment.CorrelationId)
		}
	default:
		t.Error("Payment was not queued")
	}

	// Add payment to repository for audit test
	err = repository.Add(domain.DefaultProcessor, 100.50)
	if err != nil {
		t.Fatalf("Failed to add payment to repository: %v", err)
	}

	// Test audit
	summary, err := auditUC.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute audit: %v", err)
	}

	if summary.Default.TotalRequests != 1 {
		t.Errorf("Expected 1 request, got %d", summary.Default.TotalRequests)
	}

	if summary.Default.TotalAmount != 100.50 {
		t.Errorf("Expected amount 100.50, got %f", summary.Default.TotalAmount)
	}
}

func TestPaymentValidation(t *testing.T) {
	tests := []struct {
		name    string
		payment domain.Payment
		wantErr bool
	}{
		{
			name: "valid payment",
			payment: domain.Payment{
				CorrelationId: "test-123",
				Amount:        "100.50",
			},
			wantErr: false,
		},
		{
			name: "missing correlation ID",
			payment: domain.Payment{
				Amount: "100.50",
			},
			wantErr: true,
		},
		{
			name: "missing amount",
			payment: domain.Payment{
				CorrelationId: "test-123",
			},
			wantErr: true,
		},
		{
			name: "invalid amount",
			payment: domain.Payment{
				CorrelationId: "test-123",
				Amount:        "invalid",
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			payment: domain.Payment{
				CorrelationId: "test-123",
				Amount:        "-100.50",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payment.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Payment.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
