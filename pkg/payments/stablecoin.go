// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package payments

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StablecoinTransfer represents a stablecoin transfer to validate.
type StablecoinTransfer struct {
	ID              string  `json:"id"`
	ChainID         string  `json:"chain_id"`          // ethereum, lux, solana, etc.
	TokenSymbol     string  `json:"token_symbol"`      // USDC, USDT, DAI, LUSD, etc.
	Amount          float64 `json:"amount"`
	SenderAddress   string  `json:"sender_address"`
	ReceiverAddress string  `json:"receiver_address"`
	TxHash          string  `json:"tx_hash,omitempty"`
	Country         string  `json:"country"`            // jurisdiction of the account
	AccountID       string  `json:"account_id"`
	Direction       string  `json:"direction"`          // mint, burn, transfer
	Timestamp       time.Time `json:"timestamp"`
}

// StablecoinResult is the compliance evaluation of a stablecoin transfer.
type StablecoinResult struct {
	TransferID string          `json:"transfer_id"`
	Decision   PaymentDecision `json:"decision"`
	Reasons    []string        `json:"reasons,omitempty"`
	Warnings   []string        `json:"warnings,omitempty"`
	AddressRisk string         `json:"address_risk,omitempty"` // clean, flagged, sanctioned
}

// StablecoinPolicy defines per-jurisdiction stablecoin rules.
type StablecoinPolicy struct {
	Country           string   `json:"country"`
	AllowedTokens     []string `json:"allowed_tokens"`      // which stablecoins are permitted
	ProhibitedTokens  []string `json:"prohibited_tokens"`   // explicitly banned
	RequiresReserveAttestation bool `json:"requires_reserve_attestation"`
	MaxTransferAmount float64  `json:"max_transfer_amount,omitempty"`
	MinTransferAmount float64  `json:"min_transfer_amount,omitempty"`
}

// AddressRisk describes the risk assessment of a blockchain address.
type AddressRisk struct {
	Address   string `json:"address"`
	Risk      string `json:"risk"` // clean, flagged, sanctioned
	Source    string `json:"source,omitempty"`
	Detail    string `json:"detail,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StablecoinEngine validates stablecoin transfers.
type StablecoinEngine struct {
	mu         sync.RWMutex
	policies   map[string]*StablecoinPolicy // country -> policy
	addresses  map[string]*AddressRisk      // address -> risk
}

// NewStablecoinEngine creates a stablecoin compliance engine.
func NewStablecoinEngine() *StablecoinEngine {
	return &StablecoinEngine{
		policies:  make(map[string]*StablecoinPolicy),
		addresses: make(map[string]*AddressRisk),
	}
}

// SetPolicy sets the stablecoin policy for a jurisdiction.
func (e *StablecoinEngine) SetPolicy(policy StablecoinPolicy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.policies[policy.Country] = &policy
}

// FlagAddress marks a blockchain address with a risk level.
func (e *StablecoinEngine) FlagAddress(addr, risk, source, detail string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.addresses[addr] = &AddressRisk{
		Address:   addr,
		Risk:      risk,
		Source:    source,
		Detail:    detail,
		UpdatedAt: time.Now(),
	}
}

// GetPolicy returns the stablecoin policy for a jurisdiction.
func (e *StablecoinEngine) GetPolicy(country string) (*StablecoinPolicy, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	p, ok := e.policies[country]
	if !ok {
		return nil, fmt.Errorf("no stablecoin policy for country %s", country)
	}
	return p, nil
}

// ValidateTransfer validates a stablecoin transfer against compliance rules.
func (e *StablecoinEngine) ValidateTransfer(ctx context.Context, tx *StablecoinTransfer) (*StablecoinResult, error) {
	if tx.Amount <= 0 {
		return nil, fmt.Errorf("transfer amount must be positive")
	}

	result := &StablecoinResult{
		TransferID:  tx.ID,
		Decision:    DecisionApprove,
		AddressRisk: "clean",
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	// Check address risk
	senderRisk := e.checkAddressRisk(tx.SenderAddress)
	receiverRisk := e.checkAddressRisk(tx.ReceiverAddress)

	if senderRisk == "sanctioned" || receiverRisk == "sanctioned" {
		result.Decision = DecisionDecline
		result.AddressRisk = "sanctioned"
		result.Reasons = append(result.Reasons, "Address is on sanctions list")
		return result, nil
	}
	if senderRisk == "flagged" || receiverRisk == "flagged" {
		result.Decision = DecisionReview
		result.AddressRisk = "flagged"
		result.Warnings = append(result.Warnings, "Address flagged for review by chain analysis")
	}

	// Check jurisdiction policy
	policy := e.policies[tx.Country]
	if policy != nil {
		// Token allowed?
		if len(policy.AllowedTokens) > 0 && !contains(policy.AllowedTokens, tx.TokenSymbol) {
			result.Decision = DecisionDecline
			result.Reasons = append(result.Reasons,
				fmt.Sprintf("Token %s is not allowed in %s", tx.TokenSymbol, tx.Country))
			return result, nil
		}

		// Token prohibited?
		if contains(policy.ProhibitedTokens, tx.TokenSymbol) {
			result.Decision = DecisionDecline
			result.Reasons = append(result.Reasons,
				fmt.Sprintf("Token %s is prohibited in %s", tx.TokenSymbol, tx.Country))
			return result, nil
		}

		// Amount limits
		if policy.MaxTransferAmount > 0 && tx.Amount > policy.MaxTransferAmount {
			result.Decision = DecisionDecline
			result.Reasons = append(result.Reasons,
				fmt.Sprintf("Amount $%.2f exceeds max $%.2f for %s", tx.Amount, policy.MaxTransferAmount, tx.Country))
		}
		if policy.MinTransferAmount > 0 && tx.Amount < policy.MinTransferAmount {
			result.Decision = DecisionDecline
			result.Reasons = append(result.Reasons,
				fmt.Sprintf("Amount $%.2f below minimum $%.2f for %s", tx.Amount, policy.MinTransferAmount, tx.Country))
		}

		// Reserve attestation
		if policy.RequiresReserveAttestation {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Token %s requires reserve attestation in %s", tx.TokenSymbol, tx.Country))
		}
	}

	// Mint/burn specific checks
	if tx.Direction == "mint" || tx.Direction == "burn" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Stablecoin %s operation requires enhanced compliance review", tx.Direction))
		if result.Decision == DecisionApprove {
			result.Decision = DecisionReview
		}
	}

	return result, nil
}

// CheckAddress returns the risk assessment for a blockchain address.
func (e *StablecoinEngine) CheckAddress(addr string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.checkAddressRisk(addr)
}

func (e *StablecoinEngine) checkAddressRisk(addr string) string {
	if ar, ok := e.addresses[addr]; ok {
		return ar.Risk
	}
	return "clean"
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
