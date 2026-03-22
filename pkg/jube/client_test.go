package jube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient_RequiresBaseURL(t *testing.T) {
	_, err := NewClient(Config{ModelID: "test"})
	if err == nil {
		t.Fatal("expected error for empty BaseURL")
	}
}

func TestNewClient_RequiresModelID(t *testing.T) {
	_, err := NewClient(Config{BaseURL: "http://localhost:5001"})
	if err == nil {
		t.Fatal("expected error for empty ModelID")
	}
}

func TestScreen_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/invoke/EntityAnalysisModel/test-model-id" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var tx Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			t.Fatal(err)
		}
		if tx.AccountID != "acct-123" {
			t.Fatalf("expected acct-123, got %s", tx.AccountID)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"ActivationsRaised":                    0,
			"Score":                                0.12,
			"ResponseElevation":                    0,
			"ResponseElevationContent":             "",
			"EntityAnalysisModelInstanceEntryGuid": "abc-def",
		})
	}))
	defer srv.Close()

	c, err := NewClient(Config{
		BaseURL: srv.URL,
		ModelID: "test-model-id",
	})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.Screen(context.Background(), &Transaction{
		AccountID:      "acct-123",
		TxnID:          "txn-456",
		CurrencyAmount: "1000.00",
	})
	if err != nil {
		t.Fatal(err)
	}

	if resp.IsBlocked() {
		t.Fatal("expected not blocked")
	}
	if resp.NeedsReview() {
		t.Fatal("expected no review needed")
	}
	if resp.Score != 0.12 {
		t.Fatalf("expected score 0.12, got %f", resp.Score)
	}
}

func TestScreen_Blocked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"ActivationsRaised":        2,
			"Score":                    0.95,
			"ResponseElevation":        3,
			"ResponseElevationContent": "High risk: sanctions match",
		})
	}))
	defer srv.Close()

	c, err := NewClient(Config{BaseURL: srv.URL, ModelID: "m"})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.Screen(context.Background(), &Transaction{
		AccountID:      "acct-bad",
		TxnID:          "txn-999",
		CurrencyAmount: "50000.00",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsBlocked() {
		t.Fatal("expected blocked")
	}
	if !resp.NeedsReview() {
		t.Fatal("expected review")
	}
}

func TestScreen_FailOpen(t *testing.T) {
	c, err := NewClient(Config{
		BaseURL:  "http://127.0.0.1:1", // unreachable
		ModelID:  "m",
		FailOpen: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !c.FailOpen() {
		t.Fatal("expected FailOpen=true")
	}

	// Screen should return error (caller decides fail-open behavior)
	_, err = c.Screen(context.Background(), &Transaction{
		AccountID: "test",
		TxnID:     "test",
	})
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestScreen_FailClosed(t *testing.T) {
	c, err := NewClient(Config{
		BaseURL:  "http://127.0.0.1:1",
		ModelID:  "m",
		FailOpen: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	if c.FailOpen() {
		t.Fatal("expected FailOpen=false")
	}
}

func TestScreen_HTTP503(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c, err := NewClient(Config{BaseURL: srv.URL, ModelID: "m"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Screen(context.Background(), &Transaction{AccountID: "test", TxnID: "test"})
	if err == nil {
		t.Fatal("expected error for 503")
	}
}

func TestResponse_Elevation(t *testing.T) {
	tests := []struct {
		elevation int
		blocked   bool
		review    bool
	}{
		{0, false, false},
		{1, false, true},
		{2, false, true},
		{3, true, true},
		{4, true, true},
	}
	for _, tt := range tests {
		r := &Response{ResponseElevation: tt.elevation}
		if r.IsBlocked() != tt.blocked {
			t.Errorf("elevation %d: IsBlocked=%v, want %v", tt.elevation, r.IsBlocked(), tt.blocked)
		}
		if r.NeedsReview() != tt.review {
			t.Errorf("elevation %d: NeedsReview=%v, want %v", tt.elevation, r.NeedsReview(), tt.review)
		}
	}
}
