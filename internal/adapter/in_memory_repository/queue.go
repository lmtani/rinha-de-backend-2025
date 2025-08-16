package in_memory_repository

import (
	"fmt"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
)

// InMemoryQueue implements the PaymentQueue port using Go channels
type InMemoryQueue struct {
	queue   chan domain.Payment
	closed  bool
	closeCh chan struct{}
}

// NewInMemoryQueue creates a new in-memory payment queue
func NewInMemoryQueue(bufferSize int) *InMemoryQueue {
	return &InMemoryQueue{
		queue:   make(chan domain.Payment, bufferSize),
		closeCh: make(chan struct{}),
	}
}

// Send adds a payment to the queue
func (q *InMemoryQueue) Send(payment domain.Payment) error {
	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	select {
	case q.queue <- payment:
		return nil
	case <-q.closeCh:
		return fmt.Errorf("queue is closed")
	default:
		return fmt.Errorf("queue is full")
	}
}

// Receive returns a channel to receive payments from the queue
func (q *InMemoryQueue) Receive() <-chan domain.Payment {
	return q.queue
}

// Close closes the queue
func (q *InMemoryQueue) Close() error {
	if q.closed {
		return nil
	}

	q.closed = true
	close(q.closeCh)
	close(q.queue)
	return nil
}
