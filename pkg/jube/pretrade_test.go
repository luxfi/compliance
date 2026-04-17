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
	// RED-09: Fail-open requires both AllowOnError=true AND ENVIRONMENT in FailOpenEnvironments.
	t.Setenv("ENVIRONMENT", "local")
	screen, cleanup := newTestScreen(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}, PreTradeConfig{
		AllowOnError:         true,
		FailOpenEnvironments: []string{"local", "localnet"},
	})
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
		t.Fatal("expected allowed=true (fail-open) when Jube returns error in local env")
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

// TestPreTradeFailClosed (RED-09) verifies that when AML service errors,
// the default behavior blocks the order (fail-closed).
func TestPreTradeFailClosed(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	screen, cleanup := newTestScreen(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}, PreTradeConfig{
		// AllowOnError not set — defaults to false (fail-closed).
	})
	defer cleanup()

	result := screen.Screen(context.Background(), ScreenRequest{
		AccountID: "acct-fail",
		Symbol:    "AAPL",
		Side:      "buy",
		Qty:       "10",
		Price:     "150",
		Currency:  "USD",
	})

	if result.Allowed {
		t.Fatal("RED-09: expected fail-closed (allowed=false) when AML errors in production")
	}
	if result.Action != PreTradeBlock {
		t.Fatalf("action = %q, want %q", result.Action, PreTradeBlock)
	}
}

// TestPreTradeFailOpenOnlyLocal (RED-09) verifies that fail-open is only
// permitted when ENVIRONMENT matches FailOpenEnvironments.
func TestPreTradeFailOpenOnlyLocal(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"down"}`))
	}

	// Case 1: ENVIRONMENT=local with local in FailOpenEnvironments → fail-open
	t.Run("local_permitted", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "local")
		screen, cleanup := newTestScreen(t, handler, PreTradeConfig{
			AllowOnError:         true,
			FailOpenEnvironments: []string{"local", "localnet"},
		})
		defer cleanup()

		result := screen.Screen(context.Background(), ScreenRequest{
			AccountID: "a1", Symbol: "AAPL", Side: "buy", Qty: "1", Price: "1", Currency: "USD",
		})
		if !result.Allowed {
			t.Fatal("RED-09: expected fail-open in local environment")
		}
	})

	// Case 2: ENVIRONMENT=dev with local-only FailOpenEnvironments → fail-closed
	t.Run("dev_not_permitted", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "dev")
		screen, cleanup := newTestScreen(t, handler, PreTradeConfig{
			AllowOnError:         true,
			FailOpenEnvironments: []string{"local", "localnet"},
		})
		defer cleanup()

		result := screen.Screen(context.Background(), ScreenRequest{
			AccountID: "a2", Symbol: "AAPL", Side: "buy", Qty: "1", Price: "1", Currency: "USD",
		})
		if result.Allowed {
			t.Fatal("RED-09: expected fail-closed in dev environment (not in FailOpenEnvironments)")
		}
	})

	// Case 3: ENVIRONMENT=production, AllowOnError=true but no FailOpenEnvironments → fail-closed
	t.Run("production_forced_closed", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "production")
		screen, cleanup := newTestScreen(t, handler, PreTradeConfig{
			AllowOnError: true,
			// FailOpenEnvironments empty → fail-closed everywhere
		})
		defer cleanup()

		result := screen.Screen(context.Background(), ScreenRequest{
			AccountID: "a3", Symbol: "AAPL", Side: "buy", Qty: "1", Price: "1", Currency: "USD",
		})
		if result.Allowed {
			t.Fatal("RED-09: expected fail-closed in production (empty FailOpenEnvironments)")
		}
	})
}
