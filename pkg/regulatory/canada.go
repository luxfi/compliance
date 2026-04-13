// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Canada implements the Jurisdiction interface for Canadian securities regulation.
// Covers CIRO (merged IIROC+MFDA, Jan 2023), CSA provincial regulation, and FINTRAC AML.
type Canada struct{}

func (c *Canada) Name() string              { return "Canada" }
func (c *Canada) Code() string              { return "CA" }
func (c *Canada) RegulatoryFramework() string { return "ciro" }
func (c *Canada) PassportableTo() []string  { return nil }

func (c *Canada) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "ca-ciro-registration",
			Category:    "licensing",
			Description: "CIRO dealer member registration (successor to IIROC/MFDA, effective Jan 1 2023)",
			Mandatory:   true,
			Reference:   "National Instrument 31-103, Part 2",
		},
		{
			ID:          "ca-csa-passport",
			Category:    "licensing",
			Description: "CSA passport system registration in principal jurisdiction; deemed registered in other provinces",
			Mandatory:   true,
			Reference:   "National Policy 11-202 (CSA Passport System)",
		},
		{
			ID:          "ca-fintrac-registration",
			Category:    "aml",
			Description: "FINTRAC registration as money service business or securities dealer",
			Mandatory:   true,
			Reference:   "Proceeds of Crime (Money Laundering) and Terrorist Financing Act, s.11.1",
		},
		{
			ID:          "ca-fintrac-kyc",
			Category:    "kyc",
			Description: "Verify identity of every individual: name, DOB, address; confirm existence of entities",
			Mandatory:   true,
			Reference:   "PCMLTFA Reg SOR/2002-184, s.64",
		},
		{
			ID:          "ca-fintrac-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Reports to FINTRAC",
			Mandatory:   true,
			Reference:   "PCMLTFA, s.7",
		},
		{
			ID:          "ca-fintrac-lcttr",
			Category:    "reporting",
			Description: "Large Cash Transaction Report for CAD 10,000+ cash receipts",
			Mandatory:   true,
			Reference:   "PCMLTFA Reg SOR/2002-184, s.12",
		},
		{
			ID:          "ca-fintrac-eft",
			Category:    "reporting",
			Description: "Electronic Funds Transfer Report for international transfers CAD 10,000+",
			Mandatory:   true,
			Reference:   "PCMLTFA Reg SOR/2002-184, s.12.1",
		},
		{
			ID:          "ca-ciro-kyc",
			Category:    "kyc",
			Description: "Know Your Client: financial situation, investment knowledge, risk tolerance, investment objectives",
			Mandatory:   true,
			Reference:   "CIRO Rule 3200 (formerly IIROC Rule 1300)",
		},
		{
			ID:          "ca-ciro-suitability",
			Category:    "kyc",
			Description: "Suitability determination before executing trades",
			Mandatory:   true,
			Reference:   "CIRO Rule 3400 (formerly IIROC Rule 1300.1)",
		},
		{
			ID:          "ca-ni31-103-proficiency",
			Category:    "licensing",
			Description: "Individual proficiency requirements (CSC, CPH examinations)",
			Mandatory:   true,
			Reference:   "NI 31-103, Part 3",
		},
	}
}

func (c *Canada) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	if app.GivenName == "" {
		violations = append(violations, Violation{
			RequirementID: "ca-fintrac-kyc",
			Field:         "given_name",
			Message:       "Given name is required for FINTRAC identity verification",
			Severity:      "error",
		})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{
			RequirementID: "ca-fintrac-kyc",
			Field:         "family_name",
			Message:       "Family name is required for FINTRAC identity verification",
			Severity:      "error",
		})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{
			RequirementID: "ca-fintrac-kyc",
			Field:         "date_of_birth",
			Message:       "Date of birth is required for FINTRAC identity verification",
			Severity:      "error",
		})
	}
	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{
			RequirementID: "ca-fintrac-kyc",
			Field:         "street",
			Message:       "Residential address is required for FINTRAC identity verification",
			Severity:      "error",
		})
	}
	if app.City == "" {
		violations = append(violations, Violation{
			RequirementID: "ca-fintrac-kyc",
			Field:         "city",
			Message:       "City is required",
			Severity:      "error",
		})
	}
	if app.PostalCode == "" {
		violations = append(violations, Violation{
			RequirementID: "ca-fintrac-kyc",
			Field:         "postal_code",
			Message:       "Postal code is required",
			Severity:      "error",
		})
	}

	// Canadian SIN for residents
	if (app.Country == "CA" || app.CountryOfTax == "CA") && app.TaxID == "" {
		violations = append(violations, Violation{
			RequirementID: "ca-fintrac-kyc",
			Field:         "tax_id",
			Message:       "Social Insurance Number (SIN) is required for Canadian tax residents",
			Severity:      "error",
		})
	}

	// CIRO KYC: investment knowledge and objectives
	if app.InvestmentObjective == "" {
		violations = append(violations, Violation{
			RequirementID: "ca-ciro-kyc",
			Field:         "investment_objective",
			Message:       "Investment objective is required for CIRO KYC",
			Severity:      "error",
		})
	}
	if app.AnnualIncome == "" {
		violations = append(violations, Violation{
			RequirementID: "ca-ciro-kyc",
			Field:         "annual_income",
			Message:       "Annual income is required for CIRO suitability assessment",
			Severity:      "error",
		})
	}

	if app.Email == "" {
		violations = append(violations, Violation{
			RequirementID: "ca-fintrac-kyc",
			Field:         "email",
			Message:       "Email address is required",
			Severity:      "error",
		})
	}

	// PEP screening (FINTRAC requirement)
	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "ca-fintrac-kyc",
			Field:         "is_politically_exposed",
			Message:       "PEP status declaration is required under PCMLTFA",
			Severity:      "error",
		})
	}
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{
			RequirementID: "ca-fintrac-kyc",
			Field:         "is_politically_exposed",
			Message:       "Enhanced Due Diligence required: customer is a politically exposed person under PCMLTFA s.9.3",
			Severity:      "warning",
		})
	}

	return violations
}

func (c *Canada) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 200_000,   // CAD 200k
		DailyMax:             1_000_000, // CAD 1M
		MonthlyMax:           5_000_000, // CAD 5M
		CTRThreshold:         10_000,    // FINTRAC LCTR at CAD 10k
		TravelRuleMin:        1_000,     // FATF travel rule at CAD 1k
		Currency:             "CAD",
	}
}
