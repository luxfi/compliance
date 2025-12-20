// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func computeHMAC(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestHandleSuccess(t *testing.T) {
	h := NewHandler()
	h.RegisterProvider(WebhookConfig{
		Provider: "jumio",
	}, func(provider string, body []byte, headers map[string]string) (string, error) {
		return "event-001", nil
	})

	result, err := h.Handle("jumio", []byte(`{}`), map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusProcessed {
		t.Fatalf("expected processed, got %q", result.Status)
	}
	if result.EventID != "event-001" {
		t.Fatalf("expected event-001, got %q", result.EventID)
	}
}

func TestHandleIdempotency(t *testing.T) {
	h := NewHandler()
	h.RegisterProvider(WebhookConfig{
		Provider: "onfido",
	}, func(provider string, body []byte, headers map[string]string) (string, error) {
		return "event-dup", nil
	})

	// First call
	result1, err := h.Handle("onfido", []byte(`{}`), map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result1.Status != StatusProcessed {
		t.Fatalf("expected processed, got %q", result1.Status)
	}

	// Second call with same event ID
	result2, err := h.Handle("onfido", []byte(`{}`), map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result2.Status != StatusDuplicate {
		t.Fatalf("expected duplicate, got %q", result2.Status)
	}
}

func TestHandleSignatureValidation(t *testing.T) {
	secret := "my-webhook-secret"
	h := NewHandler()
	h.RegisterProvider(WebhookConfig{
		Provider:        "plaid",
		Secret:          secret,
		SignatureHeader: "X-Plaid-Sig",
	}, func(provider string, body []byte, headers map[string]string) (string, error) {
		return "event-sig", nil
	})

	body := []byte(`{"test": true}`)

	// Missing signature
	_, err := h.Handle("plaid", body, map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing signature")
	}

	// Wrong signature
	_, err = h.Handle("plaid", body, map[string]string{"X-Plaid-Sig": "wrong"})
	if err == nil {
		t.Fatal("expected error for wrong signature")
	}

	// Correct signature
	sig := computeHMAC(body, secret)
	result, err := h.Handle("plaid", body, map[string]string{"X-Plaid-Sig": sig})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusProcessed {
		t.Fatalf("expected processed, got %q", result.Status)
	}
}

func TestHandleUnregisteredProvider(t *testing.T) {
	h := NewHandler()
	_, err := h.Handle("unknown", []byte(`{}`), nil)
	if err == nil {
		t.Fatal("expected error for unregistered provider")
	}
}

func TestHandleDeadLetter(t *testing.T) {
	h := NewHandler()
	callCount := 0
	h.RegisterProvider(WebhookConfig{
		Provider:   "failing",
		MaxRetries: 2,
	}, func(provider string, body []byte, headers map[string]string) (string, error) {
		callCount++
		return "", fmt.Errorf("always fails")
	})

	result, err := h.Handle("failing", []byte(`{}`), map[string]string{})
	if err == nil {
		t.Fatal("expected error for failed processing")
	}
	if result.Status != StatusDeadLetter {
		t.Fatalf("expected dead_letter, got %q", result.Status)
	}
	// Should have retried: 1 initial + 2 retries = 3
	if callCount != 3 {
		t.Fatalf("expected 3 attempts, got %d", callCount)
	}

	if h.DeadLetterCount() != 1 {
		t.Fatalf("expected 1 dead letter, got %d", h.DeadLetterCount())
	}
}

func TestHandleRetrySuccess(t *testing.T) {
	h := NewHandler()
	callCount := 0
	h.RegisterProvider(WebhookConfig{
		Provider:   "retry-ok",
		MaxRetries: 3,
	}, func(provider string, body []byte, headers map[string]string) (string, error) {
		callCount++
		if callCount < 3 {
			return "", fmt.Errorf("transient error")
		}
		return "event-retry", nil
	})

	result, err := h.Handle("retry-ok", []byte(`{}`), map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusProcessed {
		t.Fatalf("expected processed after retry, got %q", result.Status)
	}
	if callCount != 3 {
		t.Fatalf("expected 3 calls, got %d", callCount)
	}
}

func TestProcessedCount(t *testing.T) {
	h := NewHandler()
	h.RegisterProvider(WebhookConfig{Provider: "test"}, func(provider string, body []byte, headers map[string]string) (string, error) {
		return "evt-1", nil
	})

	if h.ProcessedCount() != 0 {
		t.Fatalf("expected 0, got %d", h.ProcessedCount())
	}

	h.Handle("test", []byte(`{}`), map[string]string{})

	if h.ProcessedCount() != 1 {
		t.Fatalf("expected 1, got %d", h.ProcessedCount())
	}
}

func TestDeadLetterDetails(t *testing.T) {
	h := NewHandler()
	h.RegisterProvider(WebhookConfig{
		Provider:   "fail",
		MaxRetries: 0,
	}, func(provider string, body []byte, headers map[string]string) (string, error) {
		return "", fmt.Errorf("processing error")
	})

	h.Handle("fail", []byte(`{"data": 1}`), map[string]string{"X-Test": "val"})

	dls := h.GetDeadLetters()
	if len(dls) != 1 {
		t.Fatalf("expected 1 dead letter, got %d", len(dls))
	}
	if dls[0].Provider != "fail" {
		t.Fatalf("expected provider 'fail', got %q", dls[0].Provider)
	}
	if dls[0].Error != "processing error" {
		t.Fatalf("expected 'processing error', got %q", dls[0].Error)
	}
}

func TestNoSecretSkipsSignatureCheck(t *testing.T) {
	h := NewHandler()
	h.RegisterProvider(WebhookConfig{
		Provider: "no-secret",
		// No secret configured
	}, func(provider string, body []byte, headers map[string]string) (string, error) {
		return "evt-ns", nil
	})

	result, err := h.Handle("no-secret", []byte(`{}`), map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusProcessed {
		t.Fatalf("expected processed, got %q", result.Status)
	}
}

func TestStatusConstants(t *testing.T) {
	if StatusProcessed != "processed" {
		t.Fatalf("StatusProcessed = %q", StatusProcessed)
	}
	if StatusDuplicate != "duplicate" {
		t.Fatalf("StatusDuplicate = %q", StatusDuplicate)
	}
	if StatusFailed != "failed" {
		t.Fatalf("StatusFailed = %q", StatusFailed)
	}
	if StatusDeadLetter != "dead_letter" {
		t.Fatalf("StatusDeadLetter = %q", StatusDeadLetter)
	}
}
