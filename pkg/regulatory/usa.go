// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

import "strings"

// USA implements the Jurisdiction interface for United States compliance.
// Covers FinCEN BSA, SEC/FINRA, and state money transmission requirements.
type USA struct{}

func (u *USA) Name() string { return "United States" }
func (u *USA) Code() string { return "US" }

func (u *USA) Requirements() []Requirement {
	return []Requirement{
		// FinCEN BSA requirements
		{
			ID:          "us-bsa-cip",
			Category:    "kyc",
			Description: "Customer Identification Program (CIP): collect name, DOB, address, SSN/TIN",
			Mandatory:   true,
			Reference:   "31 CFR 1020.220",
		},
		{
			ID:          "us-bsa-ctr",
			Category:    "reporting",
			Description: "Currency Transaction Report for transactions over $10,000",
			Mandatory:   true,
			Reference:   "31 CFR 1010.311",
		},
		{
			ID:          "us-bsa-sar",
			Category:    "reporting",
			Description: "Suspicious Activity Report for suspicious transactions $5,000+",
			Mandatory:   true,
			Reference:   "31 CFR 1020.320",
		},
		{
			ID:          "us-bsa-ofac",
			Category:    "aml",
			Description: "OFAC sanctions screening against SDN list",
			Mandatory:   true,
			Reference:   "31 CFR Part 501",
		},
		{
			ID:          "us-bsa-recordkeeping",
			Category:    "reporting",
			Description: "Maintain records of transactions $3,000+ for 5 years",
			Mandatory:   true,
			Reference:   "31 CFR 1010.410",
		},
		// SEC/FINRA requirements
		{
			ID:          "us-sec-accredited",
			Category:    "kyc",
			Description: "Accredited investor verification for Reg D offerings",
			Mandatory:   false,
			Reference:   "SEC Rule 501(a), Regulation D",
		},
		{
			ID:          "us-finra-suitability",
			Category:    "kyc",
			Description: "Customer suitability: investment objectives, financial status, risk tolerance",
			Mandatory:   true,
			Reference:   "FINRA Rule 2111",
		},
		{
			ID:          "us-finra-pdt",
			Category:    "kyc",
			Description: "Pattern Day Trader rules: $25,000 minimum equity",
			Mandatory:   false,
			Reference:   "FINRA Rule 4210",
		},
		{
			ID:          "us-finra-disclosures",
			Category:    "kyc",
			Description: "Regulatory disclosures: control person, affiliation, PEP status",
			Mandatory:   true,
			Reference:   "FINRA Rule 3210",
		},
		// State requirements
		{
			ID:          "us-state-mtl",
			Category:    "licensing",
			Description: "Money Transmitter License required per state",
			Mandatory:   true,
			Reference:   "State Money Transmission Laws",
		},
		// Address verification
		{
			ID:          "us-address",
			Category:    "kyc",
			Description: "US residential address verification",
			Mandatory:   true,
			Reference:   "31 CFR 1020.220(a)(2)",
		},
	}
}

func (u *USA) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	// CIP: name required
	if app.GivenName == "" {
		violations = append(violations, Violation{
			RequirementID: "us-bsa-cip",
			Field:         "given_name",
			Message:       "Given name is required for US CIP compliance",
			Severity:      "error",
		})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{
			RequirementID: "us-bsa-cip",
			Field:         "family_name",
			Message:       "Family name is required for US CIP compliance",
			Severity:      "error",
		})
	}

	// CIP: DOB required
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{
			RequirementID: "us-bsa-cip",
			Field:         "date_of_birth",
			Message:       "Date of birth is required for US CIP compliance",
			Severity:      "error",
		})
	}

	// CIP: SSN/TIN required for US residents
	if app.Country == "US" || app.CountryOfTax == "US" {
		if app.TaxID == "" {
			violations = append(violations, Violation{
				RequirementID: "us-bsa-cip",
				Field:         "tax_id",
				Message:       "SSN or TIN is required for US persons",
				Severity:      "error",
			})
		}
		if app.TaxIDType == "" {
			violations = append(violations, Violation{
				RequirementID: "us-bsa-cip",
				Field:         "tax_id_type",
				Message:       "Tax ID type (SSN/ITIN/EIN) is required",
				Severity:      "error",
			})
		}
	}

	// Address required
	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{
			RequirementID: "us-address",
			Field:         "street",
			Message:       "Street address is required",
			Severity:      "error",
		})
	}
	if app.City == "" {
		violations = append(violations, Violation{
			RequirementID: "us-address",
			Field:         "city",
			Message:       "City is required",
			Severity:      "error",
		})
	}
	if app.State == "" {
		violations = append(violations, Violation{
			RequirementID: "us-address",
			Field:         "state",
			Message:       "State is required for US address",
			Severity:      "error",
		})
	}
	if app.PostalCode == "" {
		violations = append(violations, Violation{
			RequirementID: "us-address",
			Field:         "postal_code",
			Message:       "ZIP code is required",
			Severity:      "error",
		})
	}

	// FINRA suitability
	if app.InvestmentObjective == "" {
		violations = append(violations, Violation{
			RequirementID: "us-finra-suitability",
			Field:         "investment_objective",
			Message:       "Investment objective is required for FINRA suitability",
			Severity:      "error",
		})
	}
	if app.AnnualIncome == "" {
		violations = append(violations, Violation{
			RequirementID: "us-finra-suitability",
			Field:         "annual_income",
			Message:       "Annual income range is required for FINRA suitability",
			Severity:      "error",
		})
	}

	// FINRA disclosures: must be explicitly answered (not nil)
	if app.IsControlPerson == nil {
		violations = append(violations, Violation{
			RequirementID: "us-finra-disclosures",
			Field:         "is_control_person",
			Message:       "Control person disclosure is required",
			Severity:      "error",
		})
	}
	if app.IsAffiliatedExchange == nil {
		violations = append(violations, Violation{
			RequirementID: "us-finra-disclosures",
			Field:         "is_affiliated_exchange_or_finra",
			Message:       "Exchange/FINRA affiliation disclosure is required",
			Severity:      "error",
		})
	}
	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "us-finra-disclosures",
			Field:         "is_politically_exposed",
			Message:       "Politically exposed person disclosure is required",
			Severity:      "error",
		})
	}
	if app.ImmediateFamilyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "us-finra-disclosures",
			Field:         "immediate_family_exposed",
			Message:       "Immediate family PEP disclosure is required",
			Severity:      "error",
		})
	}

	// PEP flagging
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{
			RequirementID: "us-bsa-ofac",
			Field:         "is_politically_exposed",
			Message:       "Enhanced Due Diligence required for politically exposed persons",
			Severity:      "warning",
		})
	}

	// Funding source
	if app.FundingSource == "" {
		violations = append(violations, Violation{
			RequirementID: "us-finra-suitability",
			Field:         "funding_source",
			Message:       "Source of funds is required",
			Severity:      "error",
		})
	}

	// Email required for electronic delivery
	if app.Email == "" {
		violations = append(violations, Violation{
			RequirementID: "us-bsa-cip",
			Field:         "email",
			Message:       "Email address is required",
			Severity:      "error",
		})
	}

	// Employment status
	if app.EmploymentStatus == "" {
		violations = append(violations, Violation{
			RequirementID: "us-finra-suitability",
			Field:         "employment_status",
			Message:       "Employment status is required",
			Severity:      "error",
		})
	}
	if app.EmploymentStatus == "employed" && app.EmployerName == "" {
		violations = append(violations, Violation{
			RequirementID: "us-finra-suitability",
			Field:         "employer_name",
			Message:       "Employer name is required when employed",
			Severity:      "error",
		})
	}

	// Validate state code format (2 letters)
	if app.State != "" && (len(app.State) != 2 || strings.ToUpper(app.State) != app.State) {
		violations = append(violations, Violation{
			RequirementID: "us-address",
			Field:         "state",
			Message:       "State must be a 2-letter abbreviation (e.g. CA, NY)",
			Severity:      "warning",
		})
	}

	return violations
}

func (u *USA) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 250_000,   // $250k default single tx
		DailyMax:             1_000_000, // $1M daily
		MonthlyMax:           5_000_000, // $5M monthly
		CTRThreshold:         10_000,    // FinCEN CTR at $10k
		TravelRuleMin:        3_000,     // FATF travel rule at $3k
		Currency:             "USD",
	}
}
