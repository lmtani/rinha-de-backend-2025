package in_memory_repository

import (
	"sync"
	"time"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
)

// InMemoryRepository implements the PaymentRepository port using in-memory storage
type InMemoryRepository struct {
	mu       sync.RWMutex
	channels map[string]*channelStats
	// keep a simple append-only log to support range queries for auditing
	events []paymentEvent
}

type channelStats struct {
	totalRequests int
	totalAmount   float64
}

type paymentEvent struct {
	when    time.Time
	channel domain.ProcessorChannel
	amount  float64
}

// NewInMemoryRepository creates a new in-memory payment repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		channels: map[string]*channelStats{
			domain.DefaultProcessor.String():  {},
			domain.FallbackProcessor.String(): {},
		},
		events: make([]paymentEvent, 0, 1024),
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
	// record event timestamped in UTC to align with API expectations
	r.events = append(r.events, paymentEvent{when: time.Now().UTC(), channel: channel, amount: amount})
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

// GetSummaryInRange returns a summary filtered by the provided time range.
// If from is nil, it is treated as the beginning of time. If to is nil, it is treated as now.
func (r *InMemoryRepository) GetSummaryInRange(from, to *time.Time) (domain.PaymentsSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var start time.Time
	var end time.Time
	if from != nil {
		start = from.UTC()
	} else {
		start = time.Time{}
	}
	if to != nil {
		end = to.UTC()
	} else {
		end = time.Now().UTC()
	}

	var defReq int
	var defAmt float64
	var fbReq int
	var fbAmt float64

	for _, e := range r.events {
		if (e.when.Equal(start) || e.when.After(start)) && (e.when.Before(end) || e.when.Equal(end)) {
			if e.channel == domain.DefaultProcessor {
				defReq++
				defAmt += e.amount
			} else if e.channel == domain.FallbackProcessor {
				fbReq++
				fbAmt += e.amount
			}
		}
	}

	return domain.PaymentsSummary{
		Default: domain.PaymentsChannelStats{
			TotalRequests: defReq,
			TotalAmount:   defAmt,
		},
		Fallback: domain.PaymentsChannelStats{
			TotalRequests: fbReq,
			TotalAmount:   fbAmt,
		},
	}, nil
}
