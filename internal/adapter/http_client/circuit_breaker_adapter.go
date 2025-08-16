package http_client

import (
	"fmt"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreakerAdapter wraps the gobreaker library to implement our CircuitBreaker port
type CircuitBreakerAdapter struct {
	cb *gobreaker.CircuitBreaker
}

// NewCircuitBreakerAdapter creates a new circuit breaker adapter
func NewCircuitBreakerAdapter(name string, maxRequests uint32, interval, timeout time.Duration, failureRatio float64, minRequests uint32) *CircuitBreakerAdapter {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: maxRequests,
		Interval:    interval,
		Timeout:     timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < minRequests {
				return false
			}
			ratio := float64(counts.TotalFailures) / float64(counts.Requests)
			return ratio >= failureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit breaker %s changed from %v to %v\n", name, from, to)
		},
	}

	return &CircuitBreakerAdapter{
		cb: gobreaker.NewCircuitBreaker(settings),
	}
}

// Execute executes the given function with circuit breaker protection
func (c *CircuitBreakerAdapter) Execute(fn func() error) error {
	_, err := c.cb.Execute(func() (interface{}, error) {
		return nil, fn()
	})
	return err
}

// State returns the current state of the circuit breaker
func (c *CircuitBreakerAdapter) State() string {
	return c.cb.State().String()
}
