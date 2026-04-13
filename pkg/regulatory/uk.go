// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// UK implements the Jurisdiction interface for United Kingdom FCA compliance.
// Covers FCA registration, 5AMLD, CDD/EDD requirements.
type UK struct{}

func (u *UK) Name() string              { return "United Kingdom" }
func (u *UK) Code() string              { return "GB" }
func (u *UK) RegulatoryFramework() string { return "uk_fca" }
func (u *UK) PassportableTo() []string  { return nil }

func (u *UK) Requirements() []Requirement {
	return []Requirement{
		// FCA registration
		{
			ID:          "uk-fca-registration",
			Category:    "licensing",
			Description: "FCA registration under Part 4A of FSMA 2000",
			Mandatory:   true,
			Reference:   "Financial Services and Markets Act 2000",
		},
		// 5AMLD / MLR 2017
		{
			ID:          "uk-5amld-cdd",
			Category:    "kyc",
			Description: "Customer Due Diligence: verify identity before establishing business relationship",
			Mandatory:   true,
			Reference:   "Money Laundering Regulations 2017, Reg 28",
		},
		{
			ID:          "uk-5amld-edd",
			Category:    "kyc",
			Description: "Enhanced Due Diligence for high-risk customers (PEPs, high-risk countries)",
			Mandatory:   true,
			Reference:   "MLR 2017, Reg 33-35",
		},
		{
			ID:          "uk-5amld-ongoing",
			Category:    "aml",
			Description: "Ongoing monitoring of business relationship and transactions",
			Mandatory:   true,
			Reference:   "MLR 2017, Reg 28(11)",
		},
		{
			ID:          "uk-5amld-pep",
			Category:    "kyc",
			Description: "PEP screening: identify politically exposed persons and family/close associates",
			Mandatory:   true,
			Reference:   "MLR 2017, Reg 35",
		},
		{
			ID:          "uk-5amld-sanctions",
			Category:    "aml",
			Description: "HM Treasury sanctions list screening",
			Mandatory:   true,
			Reference:   "Sanctions and Anti-Money Laundering Act 2018",
		},
		// CDD requirements
		{
			ID:          "uk-cdd-identity",
			Category:    "kyc",
			Description: "Verify customer identity: full name, date of birth, residential address",
			Mandatory:   true,
			Reference:   "MLR 2017, Reg 28(2)",
		},
		{
			ID:          "uk-cdd-purpose",
			Category:    "kyc",
			Description: "Establish purpose and intended nature of business relationship",
			Mandatory:   true,
			Reference:   "MLR 2017, Reg 28(6)",
		},
		{
			ID:          "uk-cdd-source-funds",
			Category:    "kyc",
			Description: "Determine source of funds for the business relationship",
			Mandatory:   true,
			Reference:   "MLR 2017, Reg 28(3)",
		},
		// Reporting
		{
			ID:          "uk-sar",
			Category:    "reporting",
			Description: "Suspicious Activity Reports to NCA",
			Mandatory:   true,
			Reference:   "Proceeds of Crime Act 2002, s330-332",
		},
		// FCA consumer duty
		{
			ID:          "uk-fca-consumer-duty",
			Category:    "kyc",
			Description: "Consumer Duty: act to deliver good outcomes for retail customers",
			Mandatory:   true,
			Reference:   "FCA PS22/9 - Consumer Duty",
		},
	}
}

func (u *UK) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	// CDD: full name
	if app.GivenName == "" {
		violations = append(violations, Violation{
			RequirementID: "uk-cdd-identity",
			Field:         "given_name",
			Message:       "Given name is required for UK CDD",
			Severity:      "error",
		})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{
			RequirementID: "uk-cdd-identity",
			Field:         "family_name",
			Message:       "Family name is required for UK CDD",
			Severity:      "error",
		})
	}

	// CDD: DOB
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{
			RequirementID: "uk-cdd-identity",
			Field:         "date_of_birth",
			Message:       "Date of birth is required for UK CDD",
			Severity:      "error",
		})
	}

	// CDD: residential address
	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{
			RequirementID: "uk-cdd-identity",
			Field:         "street",
			Message:       "Residential address is required for UK CDD",
			Severity:      "error",
		})
	}
	if app.City == "" {
		violations = append(violations, Violation{
			RequirementID: "uk-cdd-identity",
			Field:         "city",
			Message:       "City is required for UK CDD",
			Severity:      "error",
		})
	}
	if app.PostalCode == "" {
		violations = append(violations, Violation{
			RequirementID: "uk-cdd-identity",
			Field:         "postal_code",
			Message:       "Postcode is required for UK CDD",
			Severity:      "error",
		})
	}
	if app.Country == "" {
		violations = append(violations, Violation{
			RequirementID: "uk-cdd-identity",
			Field:         "country",
			Message:       "Country of residence is required",
			Severity:      "error",
		})
	}

	// Source of funds
	if app.FundingSource == "" {
		violations = append(violations, Violation{
			RequirementID: "uk-cdd-source-funds",
			Field:         "funding_source",
			Message:       "Source of funds is required under UK CDD",
			Severity:      "error",
		})
	}

	// PEP screening
	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "uk-5amld-pep",
			Field:         "is_politically_exposed",
			Message:       "PEP status declaration is required under 5AMLD",
			Severity:      "error",
		})
	}
	if app.ImmediateFamilyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "uk-5amld-pep",
			Field:         "immediate_family_exposed",
			Message:       "Family member PEP status is required under 5AMLD",
			Severity:      "error",
		})
	}

	// EDD triggers
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{
			RequirementID: "uk-5amld-edd",
			Field:         "is_politically_exposed",
			Message:       "Enhanced Due Diligence required: customer is a PEP",
			Severity:      "warning",
		})
	}
	if app.ImmediateFamilyExposed != nil && *app.ImmediateFamilyExposed {
		violations = append(violations, Violation{
			RequirementID: "uk-5amld-edd",
			Field:         "immediate_family_exposed",
			Message:       "Enhanced Due Diligence required: customer is a family member of a PEP",
			Severity:      "warning",
		})
	}

	// Email
	if app.Email == "" {
		violations = append(violations, Violation{
			RequirementID: "uk-cdd-identity",
			Field:         "email",
			Message:       "Email address is required",
			Severity:      "error",
		})
	}

	// UK tax ID (UTR or NINO) for UK residents
	if app.Country == "GB" && app.TaxID == "" {
		violations = append(violations, Violation{
			RequirementID: "uk-cdd-identity",
			Field:         "tax_id",
			Message:       "National Insurance Number or UTR is recommended for UK residents",
			Severity:      "warning",
		})
	}

	return violations
}

func (u *UK) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 200_000,   // GBP 200k
		DailyMax:             500_000,   // GBP 500k
		MonthlyMax:           2_000_000, // GBP 2M
		CTRThreshold:         15_000,    // EUR 15k equivalent (5AMLD threshold)
		TravelRuleMin:        1_000,     // EUR 1k (EU wire transfer regulation)
		Currency:             "GBP",
	}
}
