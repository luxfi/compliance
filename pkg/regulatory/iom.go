// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// IOM implements the Jurisdiction interface for Isle of Man IOMFSA compliance.
// Covers the Anti-Money Laundering and Countering the Financing of Terrorism
// Code 2019, Designated Business registration, and CDD/EDD requirements.
type IOM struct{}

func (iom *IOM) Name() string              { return "Isle of Man" }
func (iom *IOM) Code() string              { return "IM" }
func (iom *IOM) RegulatoryFramework() string { return "iom" }
func (iom *IOM) PassportableTo() []string  { return nil }

func (iom *IOM) Requirements() []Requirement {
	return []Requirement{
		// IOMFSA Designated Business registration
		{
			ID:          "im-iomfsa-registration",
			Category:    "licensing",
			Description: "IOMFSA Designated Business registration under the Proceeds of Crime Act 2008",
			Mandatory:   true,
			Reference:   "Proceeds of Crime Act 2008 (Isle of Man)",
		},
		// AML/CFT Code 2019
		{
			ID:          "im-amlcft-cdd",
			Category:    "kyc",
			Description: "Customer Due Diligence: verify identity, address, and beneficial ownership",
			Mandatory:   true,
			Reference:   "AML/CFT Code 2019, Part 3",
		},
		{
			ID:          "im-amlcft-edd",
			Category:    "kyc",
			Description: "Enhanced Due Diligence for high-risk relationships",
			Mandatory:   true,
			Reference:   "AML/CFT Code 2019, Part 4",
		},
		{
			ID:          "im-amlcft-source-wealth",
			Category:    "kyc",
			Description: "Source of wealth verification for all customers",
			Mandatory:   true,
			Reference:   "AML/CFT Code 2019, Para 13",
		},
		{
			ID:          "im-amlcft-source-funds",
			Category:    "kyc",
			Description: "Source of funds verification for each transaction/relationship",
			Mandatory:   true,
			Reference:   "AML/CFT Code 2019, Para 14",
		},
		{
			ID:          "im-amlcft-pep",
			Category:    "kyc",
			Description: "PEP screening and ongoing monitoring",
			Mandatory:   true,
			Reference:   "AML/CFT Code 2019, Para 16",
		},
		{
			ID:          "im-amlcft-sanctions",
			Category:    "aml",
			Description: "Sanctions screening against applicable lists (UK, EU, UN)",
			Mandatory:   true,
			Reference:   "AML/CFT Code 2019, Para 17",
		},
		{
			ID:          "im-amlcft-ongoing",
			Category:    "aml",
			Description: "Ongoing monitoring of business relationships and transactions",
			Mandatory:   true,
			Reference:   "AML/CFT Code 2019, Part 5",
		},
		{
			ID:          "im-amlcft-record",
			Category:    "reporting",
			Description: "Record keeping: maintain CDD records for at least 5 years after relationship ends",
			Mandatory:   true,
			Reference:   "AML/CFT Code 2019, Part 6",
		},
		// STR reporting
		{
			ID:          "im-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Reports to Isle of Man FIU",
			Mandatory:   true,
			Reference:   "Proceeds of Crime Act 2008, s21",
		},
		// Risk assessment
		{
			ID:          "im-risk-assessment",
			Category:    "aml",
			Description: "Business-wide and customer risk assessments",
			Mandatory:   true,
			Reference:   "AML/CFT Code 2019, Part 2",
		},
	}
}

func (iom *IOM) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	// CDD: identity
	if app.GivenName == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-cdd",
			Field:         "given_name",
			Message:       "Given name is required for IOM CDD",
			Severity:      "error",
		})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-cdd",
			Field:         "family_name",
			Message:       "Family name is required for IOM CDD",
			Severity:      "error",
		})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-cdd",
			Field:         "date_of_birth",
			Message:       "Date of birth is required for IOM CDD",
			Severity:      "error",
		})
	}

	// CDD: address
	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-cdd",
			Field:         "street",
			Message:       "Residential address is required for IOM CDD",
			Severity:      "error",
		})
	}
	if app.City == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-cdd",
			Field:         "city",
			Message:       "City/town is required for IOM CDD",
			Severity:      "error",
		})
	}
	if app.PostalCode == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-cdd",
			Field:         "postal_code",
			Message:       "Postal code is required for IOM CDD",
			Severity:      "error",
		})
	}
	if app.Country == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-cdd",
			Field:         "country",
			Message:       "Country of residence is required for IOM CDD",
			Severity:      "error",
		})
	}

	// Source of wealth (mandatory for IOM)
	if app.NetWorth == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-source-wealth",
			Field:         "net_worth",
			Message:       "Net worth is required for IOM source of wealth verification",
			Severity:      "error",
		})
	}
	if app.AnnualIncome == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-source-wealth",
			Field:         "annual_income",
			Message:       "Annual income is required for IOM source of wealth verification",
			Severity:      "error",
		})
	}

	// Source of funds
	if app.FundingSource == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-source-funds",
			Field:         "funding_source",
			Message:       "Source of funds is required under IOM AML/CFT Code",
			Severity:      "error",
		})
	}

	// PEP screening
	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-pep",
			Field:         "is_politically_exposed",
			Message:       "PEP status declaration is required under IOM AML/CFT Code",
			Severity:      "error",
		})
	}
	if app.ImmediateFamilyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-pep",
			Field:         "immediate_family_exposed",
			Message:       "Family member PEP status is required under IOM AML/CFT Code",
			Severity:      "error",
		})
	}

	// EDD triggers
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-edd",
			Field:         "is_politically_exposed",
			Message:       "Enhanced Due Diligence required: customer is a PEP",
			Severity:      "warning",
		})
	}

	// Employment for source of wealth context
	if app.EmploymentStatus == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-source-wealth",
			Field:         "employment_status",
			Message:       "Employment status is required for source of wealth context",
			Severity:      "error",
		})
	}

	// Email
	if app.Email == "" {
		violations = append(violations, Violation{
			RequirementID: "im-amlcft-cdd",
			Field:         "email",
			Message:       "Email address is required",
			Severity:      "error",
		})
	}

	return violations
}

func (iom *IOM) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 150_000,   // GBP 150k
		DailyMax:             500_000,   // GBP 500k
		MonthlyMax:           2_000_000, // GBP 2M
		CTRThreshold:         15_000,    // EUR 15k equivalent
		TravelRuleMin:        1_000,     // EUR 1k (aligned with EU)
		Currency:             "GBP",
	}
}
