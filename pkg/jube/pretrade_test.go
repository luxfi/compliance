package jube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cextypes "github.com/luxfi/cex/pkg/types"
)

func TestPreTradeScreen_AllowsCleanOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"ActivationsRaised":    0,
			"Score":                0.05,
			"ResponseElevation":    0,
		})
	}))
	defer srv.Close()

	c, _ := NewClient(Config{BaseURL: srv.URL, ModelID: "m"})
	screen := NewPreTradeScreen(c)
	check := screen.Check()

	err := check(context.Background(), &cextypes.Order{
		ID:         "ord-1",
		AccountID:  "acct-1",
		Symbol:     "AAPL",
		Qty:        "10",
		LimitPrice: "150.00",
	})
	if err != nil {
		t.Fatalf("expected order allowed, got: %v", err)
	}
}

func TestPreTradeScreen_BlocksHighRiskOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"ActivationsRaised":        3,
			"Score":                    0.99,
			"ResponseElevation":        3,
			"ResponseElevationContent": "Sanctions match detected",
		})
	}))
	defer srv.Close()

	c, _ := NewClient(Config{BaseURL: srv.URL, ModelID: "m"})
	screen := NewPreTradeScreen(c)
	check := screen.Check()

	err := check(context.Background(), &cextypes.Order{
		ID:         "ord-2",
		AccountID:  "acct-bad",
		Symbol:     "BTC-USD",
		Qty:        "100",
		LimitPrice: "50000.00",
	})
	if err == nil {
		t.Fatal("expected order rejected")
	}
}

func TestPreTradeScreen_FailOpenAllows(t *testing.T) {
	c, _ := NewClient(Config{
		BaseURL:  "http://127.0.0.1:1",
		ModelID:  "m",
		FailOpen: true,
	})
	screen := NewPreTradeScreen(c)
	check := screen.Check()

	err := check(context.Background(), &cextypes.Order{
		ID:        "ord-3",
		AccountID: "acct-1",
		Symbol:    "AAPL",
		Qty:       "1",
		LimitPrice: "100.00",
	})
	if err != nil {
		t.Fatalf("expected fail-open to allow order, got: %v", err)
	}
}

func TestPreTradeScreen_FailClosedRejects(t *testing.T) {
	c, _ := NewClient(Config{
		BaseURL:  "http://127.0.0.1:1",
		ModelID:  "m",
		FailOpen: false,
	})
	screen := NewPreTradeScreen(c)
	check := screen.Check()

	err := check(context.Background(), &cextypes.Order{
		ID:        "ord-4",
		AccountID: "acct-1",
		Symbol:    "AAPL",
		Qty:       "1",
		LimitPrice: "100.00",
	})
	if err == nil {
		t.Fatal("expected fail-closed to reject order")
	}
}

func TestPreTradeScreen_NotionalOrder(t *testing.T) {
	var receivedTx Transaction
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedTx)
		json.NewEncoder(w).Encode(map[string]any{
			"ResponseElevation": 0,
			"Score":             0.01,
		})
	}))
	defer srv.Close()

	c, _ := NewClient(Config{BaseURL: srv.URL, ModelID: "m"})
	screen := NewPreTradeScreen(c)
	check := screen.Check()

	// Notional order (no qty/price, just dollar amount)
	err := check(context.Background(), &cextypes.Order{
		ID:        "ord-5",
		AccountID: "acct-1",
		Symbol:    "AAPL",
		Notional:  "5000.00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedTx.CurrencyAmount != "5000.00" {
		t.Fatalf("expected notional 5000.00, got %s", receivedTx.CurrencyAmount)
	}
}
