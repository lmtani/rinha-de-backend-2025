package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ProcessorClient handles communication with payment processor admin endpoints
type ProcessorClient struct {
	BaseURL    string
	Token      string
	httpClient *http.Client
}

// NewProcessorClient creates a new processor admin client
func NewProcessorClient(baseURL, token string) *ProcessorClient {
	return &ProcessorClient{
		BaseURL: baseURL,
		Token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PaymentsSummary represents the payments summary response
type PaymentsSummary struct {
	TotalRequests     int     `json:"totalRequests"`
	TotalAmount       float64 `json:"totalAmount"`
	TotalFee          float64 `json:"totalFee"`
	FeePerTransaction float64 `json:"feePerTransaction"`
}

// TokenConfig represents the token configuration request
type TokenConfig struct {
	Token string `json:"token"`
}

// DelayConfig represents the delay configuration request
type DelayConfig struct {
	Delay int `json:"delay"`
}

// FailureConfig represents the failure configuration request
type FailureConfig struct {
	Failure bool `json:"failure"`
}

// PurgeResponse represents the purge response
type PurgeResponse struct {
	Message string `json:"message"`
}

// GetPaymentsSummary retrieves payments summary from processor
func (c *ProcessorClient) GetPaymentsSummary(ctx context.Context, from, to *time.Time) (*PaymentsSummary, error) {
	url := fmt.Sprintf("%s/admin/payments-summary", c.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters if provided
	q := req.URL.Query()
	if from != nil {
		q.Add("from", from.Format(time.RFC3339))
	}
	if to != nil {
		q.Add("to", to.Format(time.RFC3339))
	}
	req.URL.RawQuery = q.Encode()

	// Add token header
	req.Header.Set("X-Rinha-Token", c.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var summary PaymentsSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &summary, nil
}

// SetToken configures the token for admin endpoints
func (c *ProcessorClient) SetToken(ctx context.Context, token string) error {
	config := TokenConfig{Token: token}
	return c.makeConfigRequest(ctx, "/admin/configurations/token", config)
}

// SetDelay configures the delay for payments endpoint
func (c *ProcessorClient) SetDelay(ctx context.Context, delay int) error {
	config := DelayConfig{Delay: delay}
	return c.makeConfigRequest(ctx, "/admin/configurations/delay", config)
}

// SetFailure configures failure mode for payments endpoint
func (c *ProcessorClient) SetFailure(ctx context.Context, failure bool) error {
	config := FailureConfig{Failure: failure}
	return c.makeConfigRequest(ctx, "/admin/configurations/failure", config)
}

// PurgePayments deletes all payments from the processor database
func (c *ProcessorClient) PurgePayments(ctx context.Context) (*PurgeResponse, error) {
	url := fmt.Sprintf("%s/admin/purge-payments", c.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Rinha-Token", c.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response PurgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// makeConfigRequest is a helper method for configuration requests
func (c *ProcessorClient) makeConfigRequest(ctx context.Context, endpoint string, config interface{}) error {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	body, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Rinha-Token", c.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Accept both 200 OK and 204 No Content as success
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateToken updates the client's token
func (c *ProcessorClient) UpdateToken(token string) {
	c.Token = token
}
