// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package webhook provides unified webhook handling with signature validation,
// idempotency tracking, and retry/dead-letter support.
package webhook

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// WebhookStatus describes the processing outcome.
type WebhookStatus string

const (
	StatusProcessed   WebhookStatus = "processed"
	StatusDuplicate   WebhookStatus = "duplicate"
	StatusFailed      WebhookStatus = "failed"
	StatusDeadLetter  WebhookStatus = "dead_letter"
)

// WebhookResult is the outcome of processing a webhook.
type WebhookResult struct {
	ID         string        `json:"id"`
	Provider   string        `json:"provider"`
	Status     WebhookStatus `json:"status"`
	EventID    string        `json:"event_id,omitempty"`
	Detail     string        `json:"detail,omitempty"`
	ProcessedAt time.Time    `json:"processed_at"`
}

// WebhookConfig holds configuration for a webhook provider.
type WebhookConfig struct {
	Provider        string `json:"provider"`
	Secret          string `json:"-"` // signing secret (never serialize)
	SignatureHeader string `json:"signature_header"`
	MaxRetries      int    `json:"max_retries"`
}

// ProcessorFunc is a function that processes a webhook payload.
// It receives the provider name, raw body, and parsed headers.
// It returns an event ID (for idempotency) and any error.
type ProcessorFunc func(provider string, body []byte, headers map[string]string) (eventID string, err error)

// Handler routes and processes webhooks from multiple providers.
type Handler struct {
	mu          sync.RWMutex
	configs     map[string]*WebhookConfig   // provider -> config
	processors  map[string]ProcessorFunc    // provider -> processor
	processed   map[string]*WebhookResult   // eventKey -> result (idempotency)
	deadLetters []deadLetter
	maxRetries  int
}

type deadLetter struct {
	Provider  string
	Body      []byte
	Headers   map[string]string
	Error     string
	Attempts  int
	CreatedAt time.Time
}

// NewHandler creates a new webhook handler.
func NewHandler() *Handler {
	return &Handler{
		configs:     make(map[string]*WebhookConfig),
		processors:  make(map[string]ProcessorFunc),
		processed:   make(map[string]*WebhookResult),
		maxRetries:  3,
	}
}

// RegisterProvider registers a webhook provider with its config and processor.
func (h *Handler) RegisterProvider(cfg WebhookConfig, processor ProcessorFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.configs[cfg.Provider] = &cfg
	h.processors[cfg.Provider] = processor
}

// Handle processes an incoming webhook.
func (h *Handler) Handle(provider string, body []byte, headers map[string]string) (*WebhookResult, error) {
	h.mu.RLock()
	cfg, hasCfg := h.configs[provider]
	processor, hasProc := h.processors[provider]
	h.mu.RUnlock()

	if !hasCfg || !hasProc {
		return nil, fmt.Errorf("webhook provider %q not registered", provider)
	}

	// Validate signature
	if cfg.Secret != "" {
		if !h.validateSignature(body, headers, cfg) {
			return &WebhookResult{
				ID:          newWebhookID(),
				Provider:    provider,
				Status:      StatusFailed,
				Detail:      "invalid webhook signature",
				ProcessedAt: time.Now(),
			}, fmt.Errorf("invalid webhook signature for provider %s", provider)
		}
	}

	// Process with retries
	var eventID string
	var err error
	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = h.maxRetries
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		eventID, err = processor(provider, body, headers)
		if err == nil {
			break
		}
	}

	resultID := newWebhookID()
	now := time.Now()

	if err != nil {
		// Dead letter
		h.mu.Lock()
		h.deadLetters = append(h.deadLetters, deadLetter{
			Provider:  provider,
			Body:      body,
			Headers:   headers,
			Error:     err.Error(),
			Attempts:  maxRetries + 1,
			CreatedAt: now,
		})
		h.mu.Unlock()

		return &WebhookResult{
			ID:          resultID,
			Provider:    provider,
			Status:      StatusDeadLetter,
			Detail:      err.Error(),
			ProcessedAt: now,
		}, fmt.Errorf("webhook processing failed after %d attempts: %w", maxRetries+1, err)
	}

	// Idempotency check
	eventKey := provider + ":" + eventID
	h.mu.Lock()
	if existing, ok := h.processed[eventKey]; ok {
		h.mu.Unlock()
		return &WebhookResult{
			ID:          existing.ID,
			Provider:    provider,
			Status:      StatusDuplicate,
			EventID:     eventID,
			Detail:      "already processed",
			ProcessedAt: existing.ProcessedAt,
		}, nil
	}

	result := &WebhookResult{
		ID:          resultID,
		Provider:    provider,
		Status:      StatusProcessed,
		EventID:     eventID,
		ProcessedAt: now,
	}
	h.processed[eventKey] = result
	h.mu.Unlock()

	return result, nil
}

func (h *Handler) validateSignature(body []byte, headers map[string]string, cfg *WebhookConfig) bool {
	sigHeader := cfg.SignatureHeader
	if sigHeader == "" {
		sigHeader = "X-Webhook-Signature"
	}
	sig := headers[sigHeader]
	if sig == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(cfg.Secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}

// GetDeadLetters returns all dead letter entries.
func (h *Handler) GetDeadLetters() []deadLetter {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]deadLetter, len(h.deadLetters))
	copy(result, h.deadLetters)
	return result
}

// DeadLetterCount returns the number of dead letter entries.
func (h *Handler) DeadLetterCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.deadLetters)
}

// ProcessedCount returns the number of successfully processed webhooks.
func (h *Handler) ProcessedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.processed)
}

func newWebhookID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return "wh_" + hex.EncodeToString(b)
}
