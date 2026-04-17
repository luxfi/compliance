// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package kyc provides KYC orchestration across multiple identity verification
// providers. It manages the full lifecycle: initiate, webhook callback, status
// tracking, and application linkage.
package kyc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hanzoai/idv/provider"
)

// Verification tracks the state of a single verification attempt.
type Verification struct {
	ID            string                 `json:"id"`
	ApplicationID string                 `json:"application_id"`
	OrgID         string                 `json:"org_id,omitempty"`
	Provider      string                 `json:"provider"`
	Status        idv.VerificationStatus `json:"status"`
	RedirectURL   string                 `json:"redirect_url,omitempty"`
	RiskScore     float64                `json:"risk_score,omitempty"`
	Checks        []idv.Check            `json:"checks,omitempty"`
	RawResult     json.RawMessage        `json:"raw_result,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
}

// Service manages KYC verification across multiple providers.
type Service struct {
	mu              sync.RWMutex
	providers       map[string]idv.Provider
	verifications   map[string]*Verification // verificationID -> Verification
	appIndex        map[string][]string      // applicationID -> []verificationID
	orgIndex        map[string][]string      // orgID -> []verificationID
	defaultProvider string
	webhookSecrets  map[string]string // provider -> webhook secret
}

// NewService creates a new KYC service.
func NewService() *Service {
	return &Service{
		providers:      make(map[string]idv.Provider),
		verifications:  make(map[string]*Verification),
		appIndex:       make(map[string][]string),
		orgIndex:       make(map[string][]string),
		webhookSecrets: make(map[string]string),
	}
}

// RegisterProvider adds a KYC provider.
func (s *Service) RegisterProvider(p idv.Provider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[p.Name()] = p
	if s.defaultProvider == "" {
		s.defaultProvider = p.Name()
	}
}

// SetDefault sets the default KYC provider.
func (s *Service) SetDefault(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.defaultProvider = name
}

// SetWebhookSecret sets the webhook signing secret for a provider.
func (s *Service) SetWebhookSecret(provider, secret string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.webhookSecrets[provider] = secret
}

// InitiateKYC initiates a KYC verification for an application.
func (s *Service) InitiateKYC(ctx context.Context, req *idv.VerificationRequest) (*idv.VerificationResponse, error) {
	s.mu.RLock()
	providerName := req.Provider
	if providerName == "" {
		providerName = s.defaultProvider
	}
	p, ok := s.providers[providerName]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("KYC provider %q not registered", providerName)
	}

	resp, err := p.InitiateVerification(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("KYC initiation failed: %w", err)
	}

	v := &Verification{
		ID:            resp.VerificationID,
		ApplicationID: req.ApplicationID,
		Provider:      providerName,
		Status:        idv.StatusPending,
		RedirectURL:   resp.RedirectURL,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.mu.Lock()
	s.verifications[v.ID] = v
	s.appIndex[req.ApplicationID] = append(s.appIndex[req.ApplicationID], v.ID)
	s.mu.Unlock()

	return resp, nil
}

// HandleWebhook processes an incoming webhook from a KYC provider.
func (s *Service) HandleWebhook(providerName string, body []byte, headers map[string]string) (*idv.WebhookEvent, error) {
	s.mu.RLock()
	p, ok := s.providers[providerName]
	secret := s.webhookSecrets[providerName]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("KYC provider %q not registered", providerName)
	}

	// Validate webhook signature if secret is configured
	if secret != "" {
		if !validateWebhookSignature(body, headers, secret, providerName) {
			return nil, fmt.Errorf("invalid webhook signature for provider %s", providerName)
		}
	}

	event, err := p.ParseWebhook(body, headers)
	if err != nil {
		return nil, fmt.Errorf("webhook parse failed: %w", err)
	}

	// Update verification record
	s.mu.Lock()
	if v, ok := s.verifications[event.VerificationID]; ok {
		v.Status = event.Status
		v.RiskScore = event.RiskScore
		v.Checks = event.Checks
		v.RawResult = event.RawPayload
		v.UpdatedAt = time.Now()
		if event.Status != idv.StatusPending {
			now := time.Now()
			v.CompletedAt = &now
		}
		event.ApplicationID = v.ApplicationID
	}
	s.mu.Unlock()

	return event, nil
}

// GetStatus returns a specific verification by ID.
func (s *Service) GetStatus(id string) (*Verification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.verifications[id]
	if !ok {
		return nil, fmt.Errorf("verification %s not found", id)
	}
	return v, nil
}

// ListByOrg returns all verifications for an organization.
func (s *Service) ListByOrg(orgID string) []*Verification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.orgIndex[orgID]
	result := make([]*Verification, 0, len(ids))
	for _, id := range ids {
		if v, ok := s.verifications[id]; ok {
			result = append(result, v)
		}
	}
	return result
}

// GetByApplication returns all verifications for an application.
func (s *Service) GetByApplication(applicationID string) []*Verification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.appIndex[applicationID]
	result := make([]*Verification, 0, len(ids))
	for _, id := range ids {
		if v, ok := s.verifications[id]; ok {
			result = append(result, v)
		}
	}
	return result
}

// ListProviders returns the names of registered KYC providers.
func (s *Service) ListProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}

func validateWebhookSignature(body []byte, headers map[string]string, secret, provider string) bool {
	var sig string
	switch provider {
	case idv.ProviderJumio:
		sig = headers["X-Jumio-Signature"]
	case idv.ProviderOnfido:
		sig = headers["X-SHA2-Signature"]
	case idv.ProviderPlaid:
		sig = headers["Plaid-Verification"]
	default:
		sig = headers["X-Webhook-Signature"]
	}
	if sig == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}
