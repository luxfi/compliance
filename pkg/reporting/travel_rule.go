package reporting

import (
	"time"

	"github.com/luxfi/compliance/pkg/regulatory"
)

// TravelRuleData is an IVMS101-compatible travel rule message
// per FATF Recommendation 16.
type TravelRuleData struct {
	Originator  TravelRuleParty `json:"originator"`
	Beneficiary TravelRuleParty `json:"beneficiary"`
	Amount      float64         `json:"amount"`
	Currency    string          `json:"currency"`
	Timestamp   time.Time       `json:"timestamp"`
}

// TravelRuleParty is one side of a travel rule message.
type TravelRuleParty struct {
	Name        string `json:"name"`
	AccountID   string `json:"account_id"`
	Address     string `json:"address,omitempty"`
	Institution string `json:"institution,omitempty"` // VASP name or bank
	LEI         string `json:"lei,omitempty"`
}

// RequiresTravelRule checks whether a transaction requires travel rule
// compliance based on jurisdiction-specific thresholds.
// US: >$3,000; EU: >EUR 1,000 (MiCA); SG: >SGD 1,500.
func RequiresTravelRule(amount float64, currency, jurisdiction string) bool {
	j := regulatory.GetJurisdiction(jurisdiction)
	if j == nil {
		// Unknown jurisdiction: apply the most conservative threshold ($1,000).
		return amount > 1000
	}

	limits := j.TransactionLimits()
	if limits == nil {
		return amount > 1000
	}

	if limits.TravelRuleMin > 0 {
		return amount > limits.TravelRuleMin
	}

	// Fallback: USD $3,000.
	return amount > 3000
}

// BuildTravelRuleMessage constructs an IVMS101-format message.
func BuildTravelRuleMessage(data TravelRuleData) *TravelRuleMessage {
	msg := &TravelRuleMessage{
		Version:   "IVMS101",
		Timestamp: data.Timestamp,
		Amount:    data.Amount,
		Currency:  data.Currency,
		Originator: IVMS101Party{
			Name:        data.Originator.Name,
			AccountID:   data.Originator.AccountID,
			Address:     data.Originator.Address,
			Institution: data.Originator.Institution,
			LEI:         data.Originator.LEI,
		},
		Beneficiary: IVMS101Party{
			Name:        data.Beneficiary.Name,
			AccountID:   data.Beneficiary.AccountID,
			Address:     data.Beneficiary.Address,
			Institution: data.Beneficiary.Institution,
			LEI:         data.Beneficiary.LEI,
		},
	}

	return msg
}

// TravelRuleMessage is the wire format for inter-VASP messaging.
type TravelRuleMessage struct {
	Version     string       `json:"version"`
	Timestamp   time.Time    `json:"timestamp"`
	Amount      float64      `json:"amount"`
	Currency    string       `json:"currency"`
	Originator  IVMS101Party `json:"originator"`
	Beneficiary IVMS101Party `json:"beneficiary"`
}

// IVMS101Party is a party in the IVMS101 message standard.
type IVMS101Party struct {
	Name        string `json:"name"`
	AccountID   string `json:"account_id"`
	Address     string `json:"address,omitempty"`
	Institution string `json:"institution,omitempty"`
	LEI         string `json:"lei,omitempty"`
}

// ValidateTravelRule returns validation errors.
func ValidateTravelRule(data TravelRuleData) []string {
	var errs []string

	if data.Amount <= 0 {
		errs = append(errs, "amount must be positive")
	}
	if data.Currency == "" {
		errs = append(errs, "currency is required")
	}

	// Originator completeness.
	if data.Originator.Name == "" {
		errs = append(errs, "originator name is required")
	}
	if data.Originator.AccountID == "" {
		errs = append(errs, "originator account_id is required")
	}

	// Beneficiary completeness.
	if data.Beneficiary.Name == "" {
		errs = append(errs, "beneficiary name is required")
	}
	if data.Beneficiary.AccountID == "" {
		errs = append(errs, "beneficiary account_id is required")
	}

	return errs
}
