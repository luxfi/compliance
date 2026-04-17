// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package kyc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/hanzoai/idv/provider"
)

// mockProvider is a minimal idv.Provider for testing the Service layer.
type mockProvider struct {
	name        string
	verifyID    string
	redirectURL string
	verifyErr   error
	statusResult *idv.VerificationStatusResult
	statusErr   error
	parseEvent  *idv.WebhookEvent
	parseErr    error
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) InitiateVerification(_ context.Context, _ *idv.VerificationRequest) (*idv.VerificationResponse, error) {
	if m.verifyErr != nil {
		return nil, m.verifyErr
	}
	return &idv.VerificationResponse{
		VerificationID: m.verifyID,
		Provider:       m.name,
		Status:         idv.StatusPending,
		RedirectURL:    m.redirectURL,
	}, nil
}

func (m *mockProvider) CheckStatus(_ context.Context, _ string) (*idv.VerificationStatusResult, error) {
	if m.statusErr != nil {
		return nil, m.statusErr
	}
	return m.statusResult, nil
}

func (m *mockProvider) ParseWebhook(body []byte, headers map[string]string) (*idv.WebhookEvent, error) {
	if m.parseErr != nil {
		return nil, m.parseErr
	}
	return m.parseEvent, nil
}

// --- Service tests ---

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if len(svc.ListProviders()) != 0 {
		t.Fatalf("expected 0 providers, got %d", len(svc.ListProviders()))
	}
}

func TestRegisterProvider(t *testing.T) {
	svc := NewService()
	svc.RegisterProvider(&mockProvider{name: "jumio"})
	if len(svc.ListProviders()) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(svc.ListProviders()))
	}
	svc.RegisterProvider(&mockProvider{name: "onfido"})
	svc.RegisterProvider(&mockProvider{name: "plaid"})
	names := svc.ListProviders()
	sort.Strings(names)
	if len(names) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(names))
	}
}

func TestSetDefault(t *testing.T) {
	svc := NewService()
	svc.RegisterProvider(&mockProvider{name: "jumio"})
	svc.RegisterProvider(&mockProvider{name: "onfido"})
	svc.SetDefault("onfido")
	if svc.defaultProvider != "onfido" {
		t.Fatalf("expected default 'onfido', got %q", svc.defaultProvider)
	}
}

func TestInitiateKYCUsesDefault(t *testing.T) {
	svc := NewService()
	svc.RegisterProvider(&mockProvider{
		name:        "jumio",
		verifyID:    "ver-001",
		redirectURL: "https://verify.example.com",
	})

	resp, err := svc.InitiateKYC(context.Background(), &idv.VerificationRequest{
		ApplicationID: "app-1",
		Email:         "test@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.VerificationID != "ver-001" {
		t.Fatalf("expected 'ver-001', got %q", resp.VerificationID)
	}
}

func TestInitiateKYCExplicitProvider(t *testing.T) {
	svc := NewService()
	svc.RegisterProvider(&mockProvider{name: "jumio", verifyID: "j-1"})
	svc.RegisterProvider(&mockProvider{name: "onfido", verifyID: "o-1"})

	resp, err := svc.InitiateKYC(context.Background(), &idv.VerificationRequest{
		ApplicationID: "app-2",
		Provider:      "onfido",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.VerificationID != "o-1" {
		t.Fatalf("expected 'o-1', got %q", resp.VerificationID)
	}
}

func TestInitiateKYCUnregistered(t *testing.T) {
	svc := NewService()
	svc.RegisterProvider(&mockProvider{name: "jumio"})
	_, err := svc.InitiateKYC(context.Background(), &idv.VerificationRequest{Provider: "unknown"})
	if err == nil {
		t.Fatal("expected error for unregistered provider")
	}
}

func TestInitiateKYCNoProviders(t *testing.T) {
	svc := NewService()
	_, err := svc.InitiateKYC(context.Background(), &idv.VerificationRequest{})
	if err == nil {
		t.Fatal("expected error when no providers registered")
	}
}

func TestInitiateKYCProviderError(t *testing.T) {
	svc := NewService()
	svc.RegisterProvider(&mockProvider{name: "jumio", verifyErr: fmt.Errorf("api timeout")})
	_, err := svc.InitiateKYC(context.Background(), &idv.VerificationRequest{ApplicationID: "app-3"})
	if err == nil {
		t.Fatal("expected error from provider")
	}
}

func TestGetStatus(t *testing.T) {
	svc := NewService()
	svc.RegisterProvider(&mockProvider{name: "jumio", verifyID: "ver-100"})
	_, err := svc.InitiateKYC(context.Background(), &idv.VerificationRequest{ApplicationID: "app-10"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, err := svc.GetStatus("ver-100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != "ver-100" {
		t.Fatalf("expected 'ver-100', got %q", v.ID)
	}
	if v.Status != idv.StatusPending {
		t.Fatalf("expected pending, got %q", v.Status)
	}
}

func TestGetStatusNotFound(t *testing.T) {
	svc := NewService()
	_, err := svc.GetStatus("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent verification")
	}
}

func TestGetByApplication(t *testing.T) {
	svc := NewService()
	svc.RegisterProvider(&mockProvider{name: "jumio", verifyID: "v1"})
	_, _ = svc.InitiateKYC(context.Background(), &idv.VerificationRequest{ApplicationID: "app-20"})
	results := svc.GetByApplication("app-20")
	if len(results) != 1 {
		t.Fatalf("expected 1 verification, got %d", len(results))
	}
}

func TestGetByApplicationEmpty(t *testing.T) {
	svc := NewService()
	results := svc.GetByApplication("no-such-app")
	if len(results) != 0 {
		t.Fatalf("expected 0 verifications, got %d", len(results))
	}
}

func TestHandleWebhookUpdatesVerification(t *testing.T) {
	svc := NewService()
	svc.RegisterProvider(&mockProvider{
		name:     "jumio",
		verifyID: "ver-wh",
		parseEvent: &idv.WebhookEvent{
			Provider:       "jumio",
			VerificationID: "ver-wh",
			Status:         idv.StatusApproved,
			RiskScore:      0.95,
			Checks:         []idv.Check{{Type: "document", Status: "clear"}},
		},
	})

	_, err := svc.InitiateKYC(context.Background(), &idv.VerificationRequest{ApplicationID: "app-wh"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	event, err := svc.HandleWebhook("jumio", []byte(`{}`), map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Status != idv.StatusApproved {
		t.Fatalf("expected approved, got %q", event.Status)
	}
	if event.ApplicationID != "app-wh" {
		t.Fatalf("expected 'app-wh', got %q", event.ApplicationID)
	}

	v, _ := svc.GetStatus("ver-wh")
	if v.Status != idv.StatusApproved {
		t.Fatalf("expected stored status approved, got %q", v.Status)
	}
	if v.CompletedAt == nil {
		t.Fatal("expected CompletedAt set")
	}
}

func TestHandleWebhookUnregistered(t *testing.T) {
	svc := NewService()
	_, err := svc.HandleWebhook("unknown", []byte(`{}`), nil)
	if err == nil {
		t.Fatal("expected error for unregistered provider")
	}
}

func TestHandleWebhookSignatureValidation(t *testing.T) {
	svc := NewService()
	svc.RegisterProvider(&mockProvider{
		name: "jumio",
		parseEvent: &idv.WebhookEvent{
			Provider:       "jumio",
			VerificationID: "v-sig",
			Status:         idv.StatusApproved,
		},
	})
	svc.SetWebhookSecret("jumio", "my-secret")

	body := []byte(`{"test": true}`)

	// Missing signature
	_, err := svc.HandleWebhook("jumio", body, map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing signature")
	}

	// Wrong signature
	_, err = svc.HandleWebhook("jumio", body, map[string]string{"X-Jumio-Signature": "wrong"})
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}

	// Correct signature
	mac := hmac.New(sha256.New, []byte("my-secret"))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	event, err := svc.HandleWebhook("jumio", body, map[string]string{"X-Jumio-Signature": sig})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Status != idv.StatusApproved {
		t.Fatalf("expected approved, got %q", event.Status)
	}
}

func TestHandleWebhookNoSecretSkipsValidation(t *testing.T) {
	svc := NewService()
	svc.RegisterProvider(&mockProvider{
		name: "jumio",
		parseEvent: &idv.WebhookEvent{
			Provider:       "jumio",
			VerificationID: "v-nosec",
			Status:         idv.StatusApproved,
		},
	})
	_, err := svc.HandleWebhook("jumio", []byte(`{}`), map[string]string{})
	if err != nil {
		t.Fatalf("expected no error when secret not configured: %v", err)
	}
}

// --- Application Store tests ---

func TestStoreCreate(t *testing.T) {
	store := NewStore()
	app := &Application{
		GivenName:  "John",
		FamilyName: "Doe",
		Email:      "john@example.com",
	}
	if err := store.Create(app); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.ID == "" {
		t.Fatal("expected ID to be assigned")
	}
	if app.Status != StatusDraft {
		t.Fatalf("expected draft, got %q", app.Status)
	}
	if app.KYCStatus != KYCNotStarted {
		t.Fatalf("expected not_started, got %q", app.KYCStatus)
	}
}

func TestStoreGet(t *testing.T) {
	store := NewStore()
	app := &Application{GivenName: "Jane"}
	store.Create(app)
	got, err := store.Get(app.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GivenName != "Jane" {
		t.Fatalf("expected 'Jane', got %q", got.GivenName)
	}
}

func TestStoreGetNotFound(t *testing.T) {
	store := NewStore()
	_, err := store.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent app")
	}
}

func TestStoreUpdate(t *testing.T) {
	store := NewStore()
	app := &Application{GivenName: "Alice"}
	store.Create(app)
	app.Status = StatusApproved
	if err := store.Update(app); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := store.Get(app.ID)
	if got.Status != StatusApproved {
		t.Fatalf("expected approved, got %q", got.Status)
	}
}

func TestStoreUpdateNotFound(t *testing.T) {
	store := NewStore()
	err := store.Update(&Application{ID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent app")
	}
}

func TestStoreList(t *testing.T) {
	store := NewStore()
	store.Create(&Application{GivenName: "A", Status: StatusDraft})
	store.Create(&Application{GivenName: "B", Status: StatusApproved})
	store.Create(&Application{GivenName: "C", Status: StatusDraft})

	all := store.List("")
	if len(all) != 3 {
		t.Fatalf("expected 3 apps, got %d", len(all))
	}

	drafts := store.List("draft")
	if len(drafts) != 2 {
		t.Fatalf("expected 2 drafts, got %d", len(drafts))
	}

	approved := store.List("approved")
	if len(approved) != 1 {
		t.Fatalf("expected 1 approved, got %d", len(approved))
	}
}

func TestStoreStats(t *testing.T) {
	store := NewStore()
	store.Create(&Application{Status: StatusDraft, KYCStatus: KYCNotStarted})
	store.Create(&Application{Status: StatusApproved, KYCStatus: KYCVerified})

	stats := store.Stats()
	if stats["total"] != 2 {
		t.Fatalf("expected total=2, got %v", stats["total"])
	}
	if stats["draft"] != 1 {
		t.Fatalf("expected draft=1, got %v", stats["draft"])
	}
	if stats["kyc_verified"] != 1 {
		t.Fatalf("expected kyc_verified=1, got %v", stats["kyc_verified"])
	}
}

func TestStoreCount(t *testing.T) {
	store := NewStore()
	if store.Count() != 0 {
		t.Fatalf("expected 0, got %d", store.Count())
	}
	store.Create(&Application{})
	store.Create(&Application{})
	if store.Count() != 2 {
		t.Fatalf("expected 2, got %d", store.Count())
	}
}

// --- Application JSON ---

func TestApplicationJSON(t *testing.T) {
	app := &Application{
		ID:         "test-id",
		GivenName:  "John",
		FamilyName: "Doe",
		Email:      "john@example.com",
		Status:     StatusPending,
		KYCStatus:  KYCPending,
	}
	data, err := json.Marshal(app)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded Application
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.GivenName != "John" {
		t.Fatalf("expected 'John', got %q", decoded.GivenName)
	}
	if decoded.KYCStatus != KYCPending {
		t.Fatalf("expected pending, got %q", decoded.KYCStatus)
	}
}
