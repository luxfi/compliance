package jube

import (
	"context"
	"fmt"
	"strconv"
	"time"

	cextypes "github.com/luxfi/cex/pkg/types"
)

// PreTradeScreen adapts the Jube client into a CEX engine PreTradeCheck.
type PreTradeScreen struct {
	client *Client
}

// NewPreTradeScreen creates a pre-trade compliance screen backed by Jube.
func NewPreTradeScreen(client *Client) *PreTradeScreen {
	return &PreTradeScreen{client: client}
}

// Check returns a function compatible with engine.AddPreTradeCheck.
// It sends the order to Jube for real-time AML/fraud scoring.
// If Jube blocks the transaction (ResponseElevation >= 3), the order is rejected.
// If Jube is unreachable and FailOpen is true, the order is allowed through.
func (s *PreTradeScreen) Check() func(ctx context.Context, order *cextypes.Order) error {
	return func(ctx context.Context, order *cextypes.Order) error {
		// Build Jube transaction from order
		var amount string
		if order.Qty != "" && order.LimitPrice != "" {
			qty, _ := strconv.ParseFloat(order.Qty, 64)
			price, _ := strconv.ParseFloat(order.LimitPrice, 64)
			if qty > 0 && price > 0 {
				amount = fmt.Sprintf("%.2f", qty*price)
			}
		}
		if amount == "" && order.Notional != "" {
			amount = order.Notional
		}
		if amount == "" {
			amount = "0"
		}

		tx := &Transaction{
			AccountID:      order.AccountID,
			TxnID:          order.ID,
			TxnDateTime:    time.Now().UTC().Format("2006-01-02T15:04:05.000"),
			Currency:       "USD",
			CurrencyAmount: amount,
			AmountUSD:      amount,
			ServiceCode:    "TRADE",
			ChannelID:      "ATS",
			OrderID:        order.ID,
		}

		resp, err := s.client.Screen(ctx, tx)
		if err != nil {
			if s.client.FailOpen() {
				// Allow through — in-process AML rules still apply as fallback
				return nil
			}
			return fmt.Errorf("jube pre-trade screen unavailable: %w", err)
		}

		if resp.IsBlocked() {
			return fmt.Errorf("order rejected by AML screening: %s (score=%.2f, elevation=%d)",
				resp.ResponseElevationContent, resp.Score, resp.ResponseElevation)
		}

		return nil
	}
}

// PostTradeHook returns a function compatible with engine.AddPostTradeHook.
// It sends executed trades to Jube for post-trade monitoring and surveillance.
func (s *PreTradeScreen) PostTradeHook() func(ctx context.Context, trade *cextypes.Trade) {
	return func(ctx context.Context, trade *cextypes.Trade) {
		qty, _ := strconv.ParseFloat(trade.Qty, 64)
		price, _ := strconv.ParseFloat(trade.Price, 64)
		amount := fmt.Sprintf("%.2f", qty*price)

		tx := &Transaction{
			AccountID:      trade.AccountID,
			TxnID:          trade.ID,
			TxnDateTime:    time.Now().UTC().Format("2006-01-02T15:04:05.000"),
			Currency:       "USD",
			CurrencyAmount: amount,
			AmountUSD:      amount,
			ServiceCode:    "TRADE_SETTLE",
			ChannelID:      "ATS",
			OrderID:        trade.OrderID,
		}

		// Fire and forget — post-trade monitoring should not block
		_, _ = s.client.Screen(ctx, tx)
	}
}
