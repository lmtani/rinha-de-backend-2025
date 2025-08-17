package usecase

import (
	"context"
	"time"

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

// Execute retrieves payment summary for auditing.
// If from or to are nil, the respective bound is ignored. Times are expected in UTC.
func (uc *AuditPaymentsUseCase) Execute(ctx context.Context, from, to *time.Time) (domain.PaymentsSummary, error) { //nolint:revive // ctx reserved for future use
	// Delegate to repository which handles nil bounds and UTC coercion
	return uc.repository.GetSummaryInRange(from, to)
}
