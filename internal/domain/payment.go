package domain

import (
	"errors"
	"strconv"
)

// Payment represents a payment request in the domain
type Payment struct {
	CorrelationId string
	Amount        string
}

// AmountAsFloat returns the amount as a float64
func (p Payment) AmountAsFloat() (float64, error) {
	if p.Amount == "" {
		return 0, errors.New("amount is required")
	}

	amount, err := strconv.ParseFloat(p.Amount, 64)
	if err != nil {
		return 0, errors.New("invalid amount format")
	}

	if amount <= 0 {
		return 0, errors.New("amount must be positive")
	}

	return amount, nil
}

// Validate validates the payment data
func (p Payment) Validate() error {
	if p.CorrelationId == "" {
		return errors.New("correlation ID is required")
	}

	_, err := p.AmountAsFloat()
	return err
}

// PaymentsChannelStats represents statistics for a payment channel
type PaymentsChannelStats struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

// PaymentsSummary represents a summary of all payment channels
type PaymentsSummary struct {
	Default  PaymentsChannelStats `json:"default"`
	Fallback PaymentsChannelStats `json:"fallback"`
}

// ProcessorChannel represents the different payment processor channels
type ProcessorChannel string

const (
	DefaultProcessor  ProcessorChannel = "default"
	FallbackProcessor ProcessorChannel = "fallback"
)

// String returns the string representation of the processor channel
func (p ProcessorChannel) String() string {
	return string(p)
}
