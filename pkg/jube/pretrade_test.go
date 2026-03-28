package jube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestScreen(t *testing.T, handler http.HandlerFunc, cfg PreTradeConfig) (*PreTradeScreen, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c, err := New(Config{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	screen := NewPreTradeScreen(c, cfg)
	return screen, func() {
		c.Close()
		srv.Close()
	}
}

func TestScreenAllowCleanTransaction(t *testing.T) {
	screen, cleanup := newTestScreen(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TransactionResponse{
			Score:  0.1,
			Action: ActionAllow,
		})
	}, PreTradeConfig{})
	defer cleanup()

	result := screen.Screen(context.Background(), ScreenRequest{
		AccountID: "acct-clean",
		Symbol:    "AAPL",
		Side:      "buy",
		Qty:       "10",
		Price:     "150.00",
		Currency:  "USD",
	})

	if !result.Allowed {
		t.Fatalf("expected allowed=true, got false; errors: %v", result.Errors)
	}
	if result.Action != PreTradeAllow {
		t.Fatalf("action = %q, want %q", result.Action, PreTradeAllow)
	}
	if result.Score != 0.1 {
		t.Fatalf("score = %f, want 0.1", result.Score)
	}
}

func TestScreenBlockHighRisk(t *testing.T) {
	screen, cleanup := newTestScreen(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TransactionResponse{
			Score:  0.95,
			Action: ActionBlock,
			Alerts: []Alert{
				{ID: "a1", RuleName: "structuring", Severity: "critical", Score: 0.95},
			},
		})
	}, PreTradeConfig{})
	defer cleanup()

	result := screen.Screen(context.Background(), ScreenRequest{
		AccountID: "acct-sus",
		Symbol:    "BTC-USD",
		Side:      "buy",
		Qty:       "1",
		Price:     "9500",
		Currency:  "USD",
	})

	if result.Allowed {
		t.Fatal("expected allowed=false for blocked transaction")
	}
	if result.Action != PreTradeBlock {
		t.Fatalf("action = %q, want %q", result.Action, PreTradeBlock)
	}
}

func TestScreenFailOpenOnError(t *testing.T) {
	screen, cleanup := newTestScreen(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}, PreTradeConfig{AllowOnError: true})
	defer cleanup()

	result := screen.Screen(context.Background(), ScreenRequest{
		AccountID: "acct-err",
		Symbol:    "GOOG",
		Side:      "buy",
		Qty:       "5",
		Price:     "100",
		Currency:  "USD",
	})

	if !result.Allowed {
		t.Fatal("expected allowed=true (fail-open) when Jube returns error")
	}
	if result.Action != PreTradeAllow {
		t.Fatalf("action = %q, want %q", result.Action, PreTradeAllow)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected warning about Jube unavailability")
	}
}

func TestScreenFailClosedOnError(t *testing.T) {
	screen, cleanup := newTestScreen(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}, PreTradeConfig{AllowOnError: false})
	defer cleanup()

	result := screen.Screen(context.Background(), ScreenRequest{
		AccountID: "acct-err",
		Symbol:    "GOOG",
		Side:      "buy",
		Qty:       "5",
		Price:     "100",
		Currency:  "USD",
	})

	if result.Allowed {
		t.Fatal("expected allowed=false (fail-closed) when Jube returns error")
	}
	if result.Action != PreTradeBlock {
		t.Fatalf("action = %q, want %q", result.Action, PreTradeBlock)
	}
}

func TestScreenRequestAmount(t *testing.T) {
	tests := []struct {
		qty, price string
		want       float64
	}{
		{"10", "150", 1500},
		{"0.5", "100", 50},
		{"100", "", 100},  // no price = market order, p defaults to 1
		{"", "100", 0},    // no qty = 0
		{"abc", "100", 0}, // invalid qty
	}

	for _, tt := range tests {
		r := ScreenRequest{Qty: tt.qty, Price: tt.price}
		got := r.Amount()
		if got != tt.want {
			t.Errorf("Amount(%q, %q) = %f, want %f", tt.qty, tt.price, got, tt.want)
		}
	}
}

func TestScreenDefaultModelID(t *testing.T) {
	screen := NewPreTradeScreen(&Client{}, PreTradeConfig{})
	if screen.cfg.ModelID != 1 {
		t.Fatalf("default ModelID = %d, want 1", screen.cfg.ModelID)
	}
}
