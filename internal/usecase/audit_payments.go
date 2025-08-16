package usecase

import (
	"context"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
	"github.com/lmtani/rinha-de-backend-2025/internal/port"
)

// AuditPaymentsUseCase handles payment audit operations
type AuditPaymentsUseCase struct {
	repository port.PaymentRepository
}

// NewAuditPaymentsUseCase creates a new audit payments use case
func NewAuditPaymentsUseCase(repository port.PaymentRepository) *AuditPaymentsUseCase {
	return &AuditPaymentsUseCase{
		repository: repository,
	}
}

// Execute retrieves payment summary for auditing
func (uc *AuditPaymentsUseCase) Execute(ctx context.Context) (domain.PaymentsSummary, error) {
	return uc.repository.GetSummary()
}
