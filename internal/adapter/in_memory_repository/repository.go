package in_memory_repository

import (
	"sync"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
)

// InMemoryRepository implements the PaymentRepository port using in-memory storage
type InMemoryRepository struct {
	mu       sync.RWMutex
	channels map[string]*channelStats
}

type channelStats struct {
	totalRequests int
	totalAmount   float64
}

// NewInMemoryRepository creates a new in-memory payment repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		channels: map[string]*channelStats{
			domain.DefaultProcessor.String():  {},
			domain.FallbackProcessor.String(): {},
		},
	}
}

// Add records a payment in the specified channel
func (r *InMemoryRepository) Add(channel domain.ProcessorChannel, amount float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	channelKey := channel.String()
	stats, ok := r.channels[channelKey]
	if !ok {
		stats = &channelStats{}
		r.channels[channelKey] = stats
	}

	stats.totalRequests++
	stats.totalAmount += amount
	return nil
}

// GetSummary returns a summary of all payment channels
func (r *InMemoryRepository) GetSummary() (domain.PaymentsSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	getStats := func(channel domain.ProcessorChannel) domain.PaymentsChannelStats {
		stats, ok := r.channels[channel.String()]
		if !ok || stats == nil {
			return domain.PaymentsChannelStats{}
		}
		return domain.PaymentsChannelStats{
			TotalRequests: stats.totalRequests,
			TotalAmount:   stats.totalAmount,
		}
	}

	return domain.PaymentsSummary{
		Default:  getStats(domain.DefaultProcessor),
		Fallback: getStats(domain.FallbackProcessor),
	}, nil
}
