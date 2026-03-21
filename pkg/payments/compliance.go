// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package payments provides payment compliance validation including travel rule,
// sanctions screening, and per-jurisdiction transaction limits.
package payments

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/luxfi/compliance/pkg/aml"
	"github.com/luxfi/compliance/pkg/regulatory"
)

// PaymentDirection is the direction of a payment.
type PaymentDirection string

const (
	PaymentPayin  PaymentDirection = "payin"
	PaymentPayout PaymentDirection = "payout"
)

// PaymentDecision is the compliance outcome for a payment.
type PaymentDecision string

const (
	DecisionApprove PaymentDecision = "approve"
	DecisionDecline PaymentDecision = "decline"
	DecisionReview  PaymentDecision = "review"
)

// PaymentRequest represents a payment to validate.
type PaymentRequest struct {
	ID            string           `json:"id"`
	Direction     PaymentDirection `json:"direction"`
	Amount        float64          `json:"amount"`
	Currency      string           `json:"currency"`
	Country       string           `json:"country"` // jurisdiction
	AccountID     string           `json:"account_id"`
	Type          string           `json:"type"` // wire, ach, crypto, card

	// Originator info (for travel rule)
	OriginatorName    string `json:"originator_name,omitempty"`
	OriginatorAccount string `json:"originator_account,omitempty"`
	OriginatorAddress string `json:"originator_address,omitempty"`
	OriginatorCountry string `json:"originator_country,omitempty"`

	// Beneficiary info (for travel rule)
	BeneficiaryName    string `json:"beneficiary_name,omitempty"`
	BeneficiaryAccount string `json:"beneficiary_account,omitempty"`
	BeneficiaryAddress string `json:"beneficiary_address,omitempty"`
	BeneficiaryCountry string `json:"beneficiary_country,omitempty"`

	Timestamp time.Time `json:"timestamp"`
}

// PaymentResult is the compliance evaluation of a payment.
type PaymentResult struct {
	PaymentID  string          `json:"payment_id"`
	Decision   PaymentDecision `json:"decision"`
	Reasons    []string        `json:"reasons,omitempty"`
	Warnings   []string        `json:"warnings,omitempty"`
	Rules      []PaymentRule   `json:"rules_applied"`
	RequiresCTR bool           `json:"requires_ctr"`
	RequiresSAR bool           `json:"requires_sar"`
	TravelRule  *TravelRuleResult `json:"travel_rule,omitempty"`
}

// PaymentRule describes a compliance rule that was evaluated.
type PaymentRule struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Detail   string `json:"detail,omitempty"`
}

// TravelRuleResult captures travel rule compliance status.
type TravelRuleResult struct {
	Applicable         bool   `json:"applicable"`
	OriginatorComplete bool   `json:"originator_complete"`
	BeneficiaryComplete bool  `json:"beneficiary_complete"`
	Compliant          bool   `json:"compliant"`
	Detail             string `json:"detail,omitempty"`
}

// ComplianceEngine validates payments against regulatory rules.
type ComplianceEngine struct {
	mu        sync.RWMutex
	screening *aml.ScreeningService
	daily     map[string]float64 // accountID -> daily total
	lastReset time.Time
}

// NewComplianceEngine creates a payment compliance engine.
func NewComplianceEngine(screening *aml.ScreeningService) *ComplianceEngine {
	return &ComplianceEngine{
		screening: screening,
		daily:     make(map[string]float64),
		lastReset: time.Now(),
	}
}

// ValidatePayin validates an incoming payment.
func (e *ComplianceEngine) ValidatePayin(ctx context.Context, req *PaymentRequest) (*PaymentResult, error) {
	req.Direction = PaymentPayin
	return e.validate(ctx, req)
}

// ValidatePayout validates an outgoing payment.
func (e *ComplianceEngine) ValidatePayout(ctx context.Context, req *PaymentRequest) (*PaymentResult, error) {
	req.Direction = PaymentPayout
	return e.validate(ctx, req)
}

func (e *ComplianceEngine) validate(ctx context.Context, req *PaymentRequest) (*PaymentResult, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("payment amount must be positive")
	}

	result := &PaymentResult{
		PaymentID: req.ID,
		Decision:  DecisionApprove,
		Rules:     []PaymentRule{},
	}

	// Get jurisdiction limits
	jurisdiction := regulatory.GetJurisdiction(req.Country)
	var limits *regulatory.Limits
	if jurisdiction != nil {
		limits = jurisdiction.TransactionLimits()
	}

	// Rule 1: Single transaction limit
	if limits != nil {
		rule := PaymentRule{
			ID:   "single_tx_limit",
			Name: "Single Transaction Limit",
		}
		if req.Amount > limits.SingleTransactionMax {
			rule.Passed = false
			rule.Detail = fmt.Sprintf("Amount $%.2f exceeds limit $%.2f", req.Amount, limits.SingleTransactionMax)
			result.Decision = DecisionDecline
			result.Reasons = append(result.Reasons, rule.Detail)
		} else {
			rule.Passed = true
		}
		result.Rules = append(result.Rules, rule)
	}

	// Rule 2: Daily aggregate limit
	if limits != nil {
		e.mu.Lock()
		// Reset daily totals if needed
		if time.Since(e.lastReset) > 24*time.Hour {
			e.daily = make(map[string]float64)
			e.lastReset = time.Now()
		}
		dailyTotal := e.daily[req.AccountID] + req.Amount
		e.daily[req.AccountID] = dailyTotal
		e.mu.Unlock()

		rule := PaymentRule{
			ID:   "daily_limit",
			Name: "Daily Aggregate Limit",
		}
		if dailyTotal > limits.DailyMax {
			rule.Passed = false
			rule.Detail = fmt.Sprintf("Daily total $%.2f exceeds limit $%.2f", dailyTotal, limits.DailyMax)
			result.Decision = DecisionDecline
			result.Reasons = append(result.Reasons, rule.Detail)
		} else {
			rule.Passed = true
		}
		result.Rules = append(result.Rules, rule)
	}

	// Rule 3: CTR threshold
	if limits != nil && limits.CTRThreshold > 0 && req.Amount >= limits.CTRThreshold {
		result.RequiresCTR = true
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Transaction $%.2f requires Currency Transaction Report (threshold $%.2f)",
				req.Amount, limits.CTRThreshold))
	}

	// Rule 4: Travel Rule compliance
	travelRule := e.applyTravelRule(req, limits)
	result.TravelRule = travelRule
	trRule := PaymentRule{
		ID:   "travel_rule",
		Name: "FATF Travel Rule (Recommendation 16)",
	}
	if travelRule.Applicable {
		trRule.Passed = travelRule.Compliant
		if !travelRule.Compliant {
			trRule.Detail = travelRule.Detail
			if result.Decision == DecisionApprove {
				result.Decision = DecisionReview
			}
			result.Warnings = append(result.Warnings, travelRule.Detail)
		}
	} else {
		trRule.Passed = true
		trRule.Detail = "Below travel rule threshold"
	}
	result.Rules = append(result.Rules, trRule)

	// Rule 5: Sanctions screening on counterparty names
	screenRule := PaymentRule{
		ID:   "sanctions_screening",
		Name: "Sanctions Screening",
	}
	if e.screening != nil {
		screenHit, screenErr := e.screenCounterparties(ctx, req)
		if screenErr != nil {
			// Screening failed — cannot confirm clean, default to manual review
			screenRule.Passed = false
			screenRule.Detail = screenErr.Error()
			if result.Decision == DecisionApprove {
				result.Decision = DecisionReview
			}
			result.Warnings = append(result.Warnings, screenErr.Error())
		} else if screenHit != "" {
			screenRule.Passed = false
			screenRule.Detail = screenHit
			result.Decision = DecisionDecline
			result.Reasons = append(result.Reasons, screenHit)
		} else {
			screenRule.Passed = true
		}
	} else {
		screenRule.Passed = true
		screenRule.Detail = "No screening service configured"
	}
	result.Rules = append(result.Rules, screenRule)

	// Rule 6: Wire transfer specific reporting
	if req.Type == "wire" {
		wireRule := PaymentRule{
			ID:   "wire_reporting",
			Name: "Wire Transfer Reporting",
		}
		wireRule.Passed = true
		if limits != nil && req.Amount >= limits.CTRThreshold {
			wireRule.Detail = "Wire transfer requires enhanced record keeping"
			result.Warnings = append(result.Warnings, wireRule.Detail)
		}
		result.Rules = append(result.Rules, wireRule)
	}

	return result, nil
}

// ApplyTravelRule evaluates FATF Recommendation 16 compliance.
func (e *ComplianceEngine) ApplyTravelRule(req *PaymentRequest) *TravelRuleResult {
	jurisdiction := regulatory.GetJurisdiction(req.Country)
	var limits *regulatory.Limits
	if jurisdiction != nil {
		limits = jurisdiction.TransactionLimits()
	}
	return e.applyTravelRule(req, limits)
}

func (e *ComplianceEngine) applyTravelRule(req *PaymentRequest, limits *regulatory.Limits) *TravelRuleResult {
	tr := &TravelRuleResult{}

	threshold := 3000.0 // USD default
	if limits != nil && limits.TravelRuleMin > 0 {
		threshold = limits.TravelRuleMin
	}

	if req.Amount < threshold {
		tr.Applicable = false
		return tr
	}

	tr.Applicable = true

	// Check originator info
	tr.OriginatorComplete = req.OriginatorName != "" &&
		req.OriginatorAccount != "" &&
		req.OriginatorAddress != ""

	// Check beneficiary info
	tr.BeneficiaryComplete = req.BeneficiaryName != "" &&
		req.BeneficiaryAccount != ""

	tr.Compliant = tr.OriginatorComplete && tr.BeneficiaryComplete

	if !tr.Compliant {
		missing := []string{}
		if !tr.OriginatorComplete {
			missing = append(missing, "originator info incomplete")
		}
		if !tr.BeneficiaryComplete {
			missing = append(missing, "beneficiary info incomplete")
		}
		tr.Detail = fmt.Sprintf("Travel Rule non-compliant: %s", joinStrings(missing, ", "))
	}

	return tr
}

func (e *ComplianceEngine) screenCounterparties(ctx context.Context, req *PaymentRequest) (string, error) {
	names := []string{}
	if req.OriginatorName != "" {
		names = append(names, req.OriginatorName)
	}
	if req.BeneficiaryName != "" {
		names = append(names, req.BeneficiaryName)
	}

	for _, name := range names {
		parts := splitName(name)
		result, err := e.screening.Screen(ctx, &aml.ScreeningRequest{
			GivenName:  parts[0],
			FamilyName: parts[1],
		})
		if err != nil {
			return "", fmt.Errorf("screening error for %q: %w", name, err)
		}
		if !result.Clear {
			return fmt.Sprintf("Sanctions match found for %q (risk: %s)", name, result.Risk), nil
		}
	}
	return "", nil
}

func splitName(full string) [2]string {
	for i, c := range full {
		if c == ' ' {
			return [2]string{full[:i], full[i+1:]}
		}
	}
	return [2]string{full, ""}
}

func joinStrings(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for _, s := range ss[1:] {
		result += sep + s
	}
	return result
}
