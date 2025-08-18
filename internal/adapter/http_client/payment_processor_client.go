package http_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
)

// PaymentProcessorClient implements the PaymentProcessor port using HTTP
type PaymentProcessorClient struct {
	baseURL string
	client  *http.Client
}

// NewPaymentProcessorClient creates a new HTTP payment processor client
func NewPaymentProcessorClient(baseURL string, timeout time.Duration) *PaymentProcessorClient {
	return &PaymentProcessorClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// ProcessPayment sends a payment request to the external processor
func (p *PaymentProcessorClient) ProcessPayment(ctx context.Context, payment domain.Payment) error {
	url := fmt.Sprintf("%s/payments", p.baseURL)

	paymentData := map[string]interface{}{
		"correlationId": payment.CorrelationId,
		"amount":        payment.Amount,
	}

	paymentJSON, err := json.Marshal(paymentData)
	if err != nil {
		return fmt.Errorf("failed to marshal payment data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(paymentJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send payment request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("payment processor returned error status: %d", resp.StatusCode)
	}

	return nil
}
