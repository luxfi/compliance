package jube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	c, err := New(Config{})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer c.Close()

	if c.baseURL != DefaultBaseURL {
		t.Fatalf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.httpClient.Timeout != DefaultTimeout {
		t.Fatalf("timeout = %v, want %v", c.httpClient.Timeout, DefaultTimeout)
	}
}

func TestNewClientCustomBaseURL(t *testing.T) {
	c, err := New(Config{BaseURL: "http://localhost:9999", Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer c.Close()

	if c.baseURL != "http://localhost:9999" {
		t.Fatalf("baseURL = %q, want http://localhost:9999", c.baseURL)
	}
	if c.httpClient.Timeout != 5*time.Second {
		t.Fatalf("timeout = %v, want 5s", c.httpClient.Timeout)
	}
}

func TestScreenTransaction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/EntityAnalysisModel/Invoke" {
			t.Errorf("path = %s, want /api/EntityAnalysisModel/Invoke", r.URL.Path)
		}

		var req TransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.EntityAnalysisModelID != 1 {
			t.Errorf("modelId = %d, want 1", req.EntityAnalysisModelID)
		}

		json.NewEncoder(w).Encode(TransactionResponse{
			Score:  0.85,
			Action: ActionBlock,
			Alerts: []Alert{{ID: "a1", RuleName: "high-value", Severity: "high"}},
		})
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer c.Close()

	resp, err := c.ScreenTransaction(context.Background(), TransactionRequest{
		EntityAnalysisModelID: 1,
		EntityInstanceEntryPayload: map[string]interface{}{
			"AccountId": "acct-123",
			"Amount":    50000,
			"Currency":  "USD",
		},
	})
	if err != nil {
		t.Fatalf("ScreenTransaction() error: %v", err)
	}
	if resp.Score != 0.85 {
		t.Fatalf("score = %f, want 0.85", resp.Score)
	}
	if resp.Action != ActionBlock {
		t.Fatalf("action = %q, want %q", resp.Action, ActionBlock)
	}
	if len(resp.Alerts) != 1 {
		t.Fatalf("alerts len = %d, want 1", len(resp.Alerts))
	}
}

func TestCheckSanctions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		name := r.URL.Query().Get("name")
		if name != "John Doe" {
			t.Errorf("name param = %q, want 'John Doe'", name)
		}

		json.NewEncoder(w).Encode(SanctionResult{
			Hit: true,
			Matches: []SanctionMatch{
				{ListName: "OFAC SDN", EntityName: "John Doe", Score: 0.95, Country: "US"},
			},
		})
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer c.Close()

	result, err := c.CheckSanctions(context.Background(), "John Doe", "US")
	if err != nil {
		t.Fatalf("CheckSanctions() error: %v", err)
	}
	if !result.Hit {
		t.Fatal("expected sanctions hit")
	}
	if len(result.Matches) != 1 {
		t.Fatalf("matches len = %d, want 1", len(result.Matches))
	}
}

func TestScreenTransactionHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"model not found"}`))
	}))
	defer srv.Close()

	c, err := New(Config{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer c.Close()

	_, err = c.ScreenTransaction(context.Background(), TransactionRequest{EntityAnalysisModelID: 999})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Fatalf("error should mention status 500, got: %v", err)
	}
}

func TestVerifySignature(t *testing.T) {
	payload := []byte(`{"event":"aml.flagged","data":{"accountId":"123"}}`)
	secret := "my-hmac-secret"

	sig := SignPayload(payload, secret)

	if !VerifySignature(payload, sig, secret) {
		t.Fatal("valid signature rejected")
	}
	if VerifySignature(payload, sig, "wrong-secret") {
		t.Fatal("invalid secret accepted")
	}
	if VerifySignature(payload, "deadbeef", secret) {
		t.Fatal("invalid signature accepted")
	}
}

func TestActionConstants(t *testing.T) {
	if ActionAllow != "allow" {
		t.Fatalf("ActionAllow = %q", ActionAllow)
	}
	if ActionBlock != "block" {
		t.Fatalf("ActionBlock = %q", ActionBlock)
	}
	if ActionReview != "review" {
		t.Fatalf("ActionReview = %q", ActionReview)
	}
}
