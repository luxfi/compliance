// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package payments

import (
	"context"
	"testing"
	"time"

	"github.com/luxfi/compliance/pkg/aml"
)

func TestValidatePayinApproved(t *testing.T) {
	engine := NewComplianceEngine(nil)
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:        "pay-1",
		Amount:    2_000, // below travel rule threshold ($3k)
		Currency:  "USD",
		Country:   "US",
		AccountID: "acct-1",
		Type:      "ach",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionApprove {
		t.Fatalf("expected approve, got %q", result.Decision)
	}
}

func TestValidatePayoutApproved(t *testing.T) {
	engine := NewComplianceEngine(nil)
	result, err := engine.ValidatePayout(context.Background(), &PaymentRequest{
		ID:        "pay-2",
		Amount:    1_000,
		Currency:  "USD",
		Country:   "US",
		AccountID: "acct-1",
		Type:      "wire",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionApprove {
		t.Fatalf("expected approve, got %q", result.Decision)
	}
}

func TestValidateOverSingleLimit(t *testing.T) {
	engine := NewComplianceEngine(nil)
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:        "pay-big",
		Amount:    500_000, // exceeds US $250k limit
		Currency:  "USD",
		Country:   "US",
		AccountID: "acct-1",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionDecline {
		t.Fatalf("expected decline for over-limit, got %q", result.Decision)
	}
	if len(result.Reasons) == 0 {
		t.Fatal("expected decline reasons")
	}
}

func TestValidateCTRRequired(t *testing.T) {
	engine := NewComplianceEngine(nil)
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:        "pay-ctr",
		Amount:    15_000, // above $10k CTR threshold
		Currency:  "USD",
		Country:   "US",
		AccountID: "acct-1",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.RequiresCTR {
		t.Fatal("expected CTR required for $15k transaction")
	}
}

func TestValidateCTRNotRequired(t *testing.T) {
	engine := NewComplianceEngine(nil)
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:        "pay-small",
		Amount:    5_000,
		Currency:  "USD",
		Country:   "US",
		AccountID: "acct-1",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RequiresCTR {
		t.Fatal("did not expect CTR for $5k transaction")
	}
}

func TestTravelRuleCompliant(t *testing.T) {
	engine := NewComplianceEngine(nil)
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:                 "pay-tr",
		Amount:             5_000, // above $3k travel rule threshold
		Currency:           "USD",
		Country:            "US",
		AccountID:          "acct-1",
		Type:               "wire",
		OriginatorName:     "John Doe",
		OriginatorAccount:  "12345",
		OriginatorAddress:  "123 Main St",
		BeneficiaryName:    "Jane Smith",
		BeneficiaryAccount: "67890",
		Timestamp:          time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TravelRule == nil {
		t.Fatal("expected travel rule result")
	}
	if !result.TravelRule.Applicable {
		t.Fatal("expected travel rule to be applicable")
	}
	if !result.TravelRule.Compliant {
		t.Fatalf("expected travel rule compliant, detail: %s", result.TravelRule.Detail)
	}
}

func TestTravelRuleNonCompliant(t *testing.T) {
	engine := NewComplianceEngine(nil)
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:        "pay-tr-nc",
		Amount:    5_000,
		Currency:  "USD",
		Country:   "US",
		AccountID: "acct-1",
		Type:      "wire",
		// Missing originator and beneficiary info
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TravelRule == nil {
		t.Fatal("expected travel rule result")
	}
	if !result.TravelRule.Applicable {
		t.Fatal("expected travel rule applicable for $5k")
	}
	if result.TravelRule.Compliant {
		t.Fatal("expected travel rule non-compliant without party info")
	}
	if result.Decision != DecisionReview {
		t.Fatalf("expected review for non-compliant travel rule, got %q", result.Decision)
	}
}

func TestTravelRuleBelowThreshold(t *testing.T) {
	engine := NewComplianceEngine(nil)
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:        "pay-tr-small",
		Amount:    1_000, // below $3k
		Currency:  "USD",
		Country:   "US",
		AccountID: "acct-1",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TravelRule == nil {
		t.Fatal("expected travel rule result")
	}
	if result.TravelRule.Applicable {
		t.Fatal("travel rule should not apply below threshold")
	}
}

func TestTravelRuleUKThreshold(t *testing.T) {
	engine := NewComplianceEngine(nil)
	// UK threshold is EUR 1k
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:        "pay-uk",
		Amount:    2_000,
		Currency:  "GBP",
		Country:   "GB",
		AccountID: "acct-uk",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TravelRule == nil || !result.TravelRule.Applicable {
		t.Fatal("expected travel rule applicable for GBP 2k in UK")
	}
}

func TestSanctionsScreeningDecline(t *testing.T) {
	screening := aml.NewScreeningService(aml.DefaultScreeningConfig())
	screening.AddEntry(aml.SanctionEntry{
		ID:   "ofac-001",
		List: aml.ListOFAC,
		Name: "Viktor Bout",
	})

	engine := NewComplianceEngine(screening)
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:             "pay-sanctions",
		Amount:         1_000,
		Currency:       "USD",
		Country:        "US",
		AccountID:      "acct-1",
		BeneficiaryName: "Viktor Bout",
		Timestamp:      time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionDecline {
		t.Fatalf("expected decline for sanctioned beneficiary, got %q", result.Decision)
	}
}

func TestSanctionsScreeningClean(t *testing.T) {
	screening := aml.NewScreeningService(aml.DefaultScreeningConfig())
	screening.AddEntry(aml.SanctionEntry{
		ID:   "ofac-001",
		List: aml.ListOFAC,
		Name: "Viktor Bout",
	})

	engine := NewComplianceEngine(screening)
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:             "pay-clean",
		Amount:         1_000,
		Currency:       "USD",
		Country:        "US",
		AccountID:      "acct-1",
		BeneficiaryName: "John Smith",
		Timestamp:      time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision == DecisionDecline {
		t.Fatal("did not expect decline for clean beneficiary")
	}
}

func TestValidateZeroAmount(t *testing.T) {
	engine := NewComplianceEngine(nil)
	_, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:     "pay-zero",
		Amount: 0,
	})
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestValidateNegativeAmount(t *testing.T) {
	engine := NewComplianceEngine(nil)
	_, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:     "pay-neg",
		Amount: -100,
	})
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestValidateUnknownJurisdiction(t *testing.T) {
	engine := NewComplianceEngine(nil)
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:        "pay-unknown",
		Amount:    1_000,
		Country:   "XX",
		AccountID: "acct-1",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still approve (no jurisdiction limits to check)
	if result.Decision != DecisionApprove {
		t.Fatalf("expected approve for unknown jurisdiction, got %q", result.Decision)
	}
}

func TestWireReporting(t *testing.T) {
	engine := NewComplianceEngine(nil)
	result, err := engine.ValidatePayin(context.Background(), &PaymentRequest{
		ID:        "pay-wire",
		Amount:    15_000,
		Currency:  "USD",
		Country:   "US",
		AccountID: "acct-1",
		Type:      "wire",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundWire := false
	for _, r := range result.Rules {
		if r.ID == "wire_reporting" {
			foundWire = true
		}
	}
	if !foundWire {
		t.Fatal("expected wire_reporting rule to be applied")
	}
}

// --- Stablecoin ---

func TestStablecoinValidateTransfer(t *testing.T) {
	engine := NewStablecoinEngine()
	engine.SetPolicy(StablecoinPolicy{
		Country:       "US",
		AllowedTokens: []string{"USDC", "USDT"},
	})

	result, err := engine.ValidateTransfer(context.Background(), &StablecoinTransfer{
		ID:          "sc-1",
		TokenSymbol: "USDC",
		Amount:      1_000,
		Country:     "US",
		Direction:   "transfer",
		Timestamp:   time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionApprove {
		t.Fatalf("expected approve for USDC in US, got %q", result.Decision)
	}
}

func TestStablecoinProhibitedToken(t *testing.T) {
	engine := NewStablecoinEngine()
	engine.SetPolicy(StablecoinPolicy{
		Country:          "US",
		ProhibitedTokens: []string{"BUSD"},
	})

	result, err := engine.ValidateTransfer(context.Background(), &StablecoinTransfer{
		ID:          "sc-2",
		TokenSymbol: "BUSD",
		Amount:      1_000,
		Country:     "US",
		Direction:   "transfer",
		Timestamp:   time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionDecline {
		t.Fatalf("expected decline for prohibited token, got %q", result.Decision)
	}
}

func TestStablecoinNotAllowed(t *testing.T) {
	engine := NewStablecoinEngine()
	engine.SetPolicy(StablecoinPolicy{
		Country:       "GB",
		AllowedTokens: []string{"USDC"}, // only USDC allowed
	})

	result, err := engine.ValidateTransfer(context.Background(), &StablecoinTransfer{
		ID:          "sc-3",
		TokenSymbol: "USDT",
		Amount:      1_000,
		Country:     "GB",
		Direction:   "transfer",
		Timestamp:   time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionDecline {
		t.Fatalf("expected decline for non-allowed token, got %q", result.Decision)
	}
}

func TestStablecoinSanctionedAddress(t *testing.T) {
	engine := NewStablecoinEngine()
	engine.FlagAddress("0xBAD", "sanctioned", "ofac", "sanctioned wallet")

	result, err := engine.ValidateTransfer(context.Background(), &StablecoinTransfer{
		ID:              "sc-4",
		TokenSymbol:     "USDC",
		Amount:          1_000,
		SenderAddress:   "0xBAD",
		ReceiverAddress: "0xGOOD",
		Country:         "US",
		Direction:       "transfer",
		Timestamp:       time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionDecline {
		t.Fatalf("expected decline for sanctioned address, got %q", result.Decision)
	}
	if result.AddressRisk != "sanctioned" {
		t.Fatalf("expected sanctioned risk, got %q", result.AddressRisk)
	}
}

func TestStablecoinFlaggedAddress(t *testing.T) {
	engine := NewStablecoinEngine()
	engine.FlagAddress("0xSUS", "flagged", "chainalysis", "suspicious activity")

	result, err := engine.ValidateTransfer(context.Background(), &StablecoinTransfer{
		ID:              "sc-5",
		TokenSymbol:     "USDC",
		Amount:          1_000,
		SenderAddress:   "0xOK",
		ReceiverAddress: "0xSUS",
		Country:         "US",
		Direction:       "transfer",
		Timestamp:       time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionReview {
		t.Fatalf("expected review for flagged address, got %q", result.Decision)
	}
}

func TestStablecoinMintRequiresReview(t *testing.T) {
	engine := NewStablecoinEngine()
	result, err := engine.ValidateTransfer(context.Background(), &StablecoinTransfer{
		ID:          "sc-mint",
		TokenSymbol: "USDC",
		Amount:      1_000_000,
		Direction:   "mint",
		Timestamp:   time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionReview {
		t.Fatalf("expected review for mint operation, got %q", result.Decision)
	}
}

func TestStablecoinCheckAddress(t *testing.T) {
	engine := NewStablecoinEngine()
	engine.FlagAddress("0xBAD", "sanctioned", "ofac", "test")

	if engine.CheckAddress("0xBAD") != "sanctioned" {
		t.Fatal("expected sanctioned for flagged address")
	}
	if engine.CheckAddress("0xGOOD") != "clean" {
		t.Fatal("expected clean for unknown address")
	}
}

func TestStablecoinGetPolicy(t *testing.T) {
	engine := NewStablecoinEngine()
	engine.SetPolicy(StablecoinPolicy{Country: "US", AllowedTokens: []string{"USDC"}})

	p, err := engine.GetPolicy("US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.AllowedTokens) != 1 || p.AllowedTokens[0] != "USDC" {
		t.Fatal("expected USDC in allowed tokens")
	}

	_, err = engine.GetPolicy("XX")
	if err == nil {
		t.Fatal("expected error for unknown country policy")
	}
}

func TestStablecoinZeroAmount(t *testing.T) {
	engine := NewStablecoinEngine()
	_, err := engine.ValidateTransfer(context.Background(), &StablecoinTransfer{
		ID:     "sc-zero",
		Amount: 0,
	})
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestStablecoinMaxAmount(t *testing.T) {
	engine := NewStablecoinEngine()
	engine.SetPolicy(StablecoinPolicy{
		Country:           "US",
		MaxTransferAmount: 100_000,
	})

	result, err := engine.ValidateTransfer(context.Background(), &StablecoinTransfer{
		ID:          "sc-max",
		TokenSymbol: "USDC",
		Amount:      200_000,
		Country:     "US",
		Direction:   "transfer",
		Timestamp:   time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionDecline {
		t.Fatalf("expected decline for over max, got %q", result.Decision)
	}
}
