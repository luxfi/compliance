// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package regulatory provides multi-jurisdiction compliance rules for USA (FinCEN/SEC/FINRA),
// UK (FCA), and Isle of Man (IOMFSA). Each jurisdiction defines KYC requirements,
// transaction limits, and application validation rules.
package regulatory

// Jurisdiction defines the compliance requirements for a specific regulatory regime.
type Jurisdiction interface {
	Name() string
	Code() string // ISO 3166-1 alpha-2 (or ISO 3166-2 for subdivisions like AE-DIFC)
	RegulatoryFramework() string
	PassportableTo() []string // jurisdiction codes this one can passport to (empty for most)
	Requirements() []Requirement
	ValidateApplication(app *ApplicationData) []Violation
	TransactionLimits() *Limits
}

// Requirement describes a single regulatory requirement.
type Requirement struct {
	ID          string `json:"id"`
	Category    string `json:"category"`    // kyc, aml, reporting, capital, licensing
	Description string `json:"description"`
	Mandatory   bool   `json:"mandatory"`
	Reference   string `json:"reference,omitempty"` // regulation citation
}

// Violation is a compliance rule violation found during validation.
type Violation struct {
	RequirementID string `json:"requirement_id"`
	Field         string `json:"field"`
	Message       string `json:"message"`
	Severity      string `json:"severity"` // error, warning
}

// Limits defines transaction limits for a jurisdiction.
type Limits struct {
	SingleTransactionMax float64 `json:"single_transaction_max"` // max single tx (in local currency)
	DailyMax             float64 `json:"daily_max"`
	MonthlyMax           float64 `json:"monthly_max"`
	AnnualMax            float64 `json:"annual_max,omitempty"`
	CTRThreshold         float64 `json:"ctr_threshold,omitempty"` // Currency Transaction Report threshold
	TravelRuleMin        float64 `json:"travel_rule_min"`         // FATF travel rule applies above this
	Currency             string  `json:"currency"`
}

// ApplicationData is a normalized view of an application for validation.
// Mirrors the fields from kyc.Application so the regulatory package stays decoupled.
type ApplicationData struct {
	GivenName   string
	FamilyName  string
	DateOfBirth string
	Email       string
	Phone       string

	Street     []string
	City       string
	State      string
	PostalCode string
	Country    string

	TaxID        string
	TaxIDType    string
	CountryOfTax string

	IsControlPerson        *bool
	IsAffiliatedExchange   *bool
	IsPoliticallyExposed   *bool
	ImmediateFamilyExposed *bool

	EmploymentStatus string
	EmployerName     string

	AnnualIncome        string
	NetWorth            string
	LiquidNetWorth      string
	FundingSource       string
	InvestmentObjective string

	AccountType string // individual, joint, ira, entity
}

// GetJurisdiction returns a jurisdiction implementation by ISO country code.
// Uses ISO 3166-1 alpha-2 codes, or ISO 3166-2 subdivision codes for
// jurisdictions with distinct regulatory regimes (e.g. AE-DIFC, AE-ADGM, AE-VARA).
func GetJurisdiction(code string) Jurisdiction {
	switch code {
	case "US":
		return &USA{}
	case "GB":
		return &UK{}
	case "IM":
		return &IOM{}
	case "CA":
		return &Canada{}
	case "BR":
		return &Brazil{}
	case "IN":
		return &India{}
	case "SG":
		return &Singapore{}
	case "AU":
		return &Australia{}
	case "CH":
		return &Switzerland{}
	case "AE":
		return &UAE{}
	case "AE-DIFC":
		return &UAEDIFC{}
	case "AE-ADGM":
		return &UAEADGM{}
	case "AE-VARA":
		return &UAEVARA{}
	case "LU":
		return &Luxembourg{}
	case "DE":
		return &Germany{}
	case "FR":
		return &France{}
	case "NL":
		return &Netherlands{}
	case "IE":
		return &Ireland{}
	case "IT":
		return &Italy{}
	case "ES":
		return &Spain{}
	default:
		return nil
	}
}

// AllJurisdictions returns all supported jurisdictions.
func AllJurisdictions() []Jurisdiction {
	return []Jurisdiction{
		&USA{}, &UK{}, &IOM{},
		&Canada{}, &Brazil{}, &India{}, &Singapore{}, &Australia{}, &Switzerland{},
		&UAE{}, &UAEDIFC{}, &UAEADGM{}, &UAEVARA{},
		&Luxembourg{}, &Germany{}, &France{}, &Netherlands{}, &Ireland{}, &Italy{}, &Spain{},
	}
}
