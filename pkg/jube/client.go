// Package jube provides an HTTP client for the Jube AML/fraud detection API.
//
// Jube is a C# transaction monitoring engine that exposes an HTTP API at
// POST /api/invoke/EntityAnalysisModel/{modelGUID}. It scores transactions
// in real time against configurable rules and ML models, returning activation
// alerts, sanctions matches, and response elevations.
package jube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config holds Jube client configuration.
type Config struct {
	// BaseURL is the Jube API base URL (e.g. "http://jube.liquidity.svc.cluster.local:5001").
	BaseURL string

	// ModelID is the EntityAnalysisModel GUID to invoke for transaction screening.
	ModelID string

	// FailOpen determines behavior when Jube is unreachable.
	// If true, transactions are allowed through when Jube is down.
	// If false, transactions are rejected when Jube is unavailable.
	FailOpen bool

	// Timeout for HTTP requests to Jube. Defaults to 5 seconds.
	Timeout time.Duration
}

// Client is an HTTP client for the Jube AML/fraud API.
type Client struct {
	cfg    Config
	http   *http.Client
}

// NewClient creates a Jube API client.
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("jube: BaseURL is required")
	}
	if cfg.ModelID == "" {
		return nil, fmt.Errorf("jube: ModelID is required")
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Client{
		cfg: cfg,
		http: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Transaction is the payload sent to Jube for scoring.
// Fields map to Jube's EntityAnalysisModel request XPaths.
type Transaction struct {
	AccountID     string  `json:"AccountId"`
	TxnID         string  `json:"TxnId"`
	TxnDateTime   string  `json:"TxnDateTime"`   // ISO 8601
	Currency      string  `json:"Currency"`
	CurrencyAmount string `json:"CurrencyAmount"`
	AmountUSD     string  `json:"AmountUSD,omitempty"`
	ResponseCode  string  `json:"ResponseCode,omitempty"`
	IP            string  `json:"IP,omitempty"`
	ChannelID     string  `json:"ChannelId,omitempty"`
	ServiceCode   string  `json:"ServiceCode,omitempty"`
	Email         string  `json:"Email,omitempty"`
	ToAccountID   string  `json:"ToAccountId,omitempty"`
	OrderID       string  `json:"OrderId,omitempty"`
}

// Response is the decoded Jube scoring response.
type Response struct {
	// ActivationsRaised indicates the number of rule activations triggered.
	ActivationsRaised int `json:"ActivationsRaised"`

	// Score is the overall risk score (0-1, higher = riskier).
	Score float64 `json:"Score"`

	// ResponseElevation is Jube's recommended action:
	//   0 = allow, 1 = review, 2 = suspend, 3 = reject
	ResponseElevation int `json:"ResponseElevation"`

	// ResponseElevationContent is a human-readable reason.
	ResponseElevationContent string `json:"ResponseElevationContent"`

	// EntityAnalysisModelInstanceEntryGUID is the unique instance entry for callback.
	EntryGUID string `json:"EntityAnalysisModelInstanceEntryGuid"`

	// Raw holds the full JSON response for inspection.
	Raw json.RawMessage `json:"-"`
}

// Screen sends a transaction to Jube for real-time scoring.
// Returns the scoring response or an error.
func (c *Client) Screen(ctx context.Context, tx *Transaction) (*Response, error) {
	body, err := json.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("jube: marshal transaction: %w", err)
	}

	url := c.cfg.BaseURL + "/api/invoke/EntityAnalysisModel/" + c.cfg.ModelID
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("jube: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jube: request failed: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("jube: read response: %w", err)
	}

	if resp.StatusCode == http.StatusServiceUnavailable {
		return nil, fmt.Errorf("jube: service unavailable (503)")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("jube: HTTP %d: %s", resp.StatusCode, string(rawBody))
	}

	var result Response
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return nil, fmt.Errorf("jube: decode response: %w", err)
	}
	result.Raw = rawBody

	return &result, nil
}

// IsBlocked returns true if Jube recommends rejecting the transaction.
func (r *Response) IsBlocked() bool {
	return r.ResponseElevation >= 3
}

// NeedsReview returns true if the transaction needs manual review.
func (r *Response) NeedsReview() bool {
	return r.ResponseElevation >= 1
}

// FailOpen returns the client's fail-open setting.
func (c *Client) FailOpen() bool {
	return c.cfg.FailOpen
}
