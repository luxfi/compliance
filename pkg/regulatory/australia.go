// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Australia implements the Jurisdiction interface for Australian regulation.
// Covers ASIC (Australian Securities and Investments Commission) AFS licensing,
// AUSTRAC AML/CTF Act 2006, and RG 209 credit guidance.
type Australia struct{}

func (a *Australia) Name() string              { return "Australia" }
func (a *Australia) Code() string              { return "AU" }
func (a *Australia) RegulatoryFramework() string { return "asic" }
func (a *Australia) PassportableTo() []string  { return nil }

func (a *Australia) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "au-asic-afsl",
			Category:    "licensing",
			Description: "Australian Financial Services Licence (AFSL) for dealing in financial products",
			Mandatory:   true,
			Reference:   "Corporations Act 2001 (Cth), s.911A",
		},
		{
			ID:          "au-asic-rg209",
			Category:    "kyc",
			Description: "ASIC RG 209 responsible lending and product suitability for retail clients",
			Mandatory:   true,
			Reference:   "ASIC Regulatory Guide 209",
		},
		{
			ID:          "au-austrac-registration",
			Category:    "aml",
			Description: "AUSTRAC registration as reporting entity under AML/CTF Act 2006",
			Mandatory:   true,
			Reference:   "Anti-Money Laundering and Counter-Terrorism Financing Act 2006 (Cth), s.6",
		},
		{
			ID:          "au-austrac-cdd",
			Category:    "kyc",
			Description: "Customer identification: verify full name, DOB, residential address (100-point ID check)",
			Mandatory:   true,
			Reference:   "AML/CTF Act 2006, Part 2; AML/CTF Rules Chapter 4",
		},
		{
			ID:          "au-austrac-edd",
			Category:    "kyc",
			Description: "Enhanced CDD for high-risk customers (PEPs, correspondent banking, high-risk countries)",
			Mandatory:   true,
			Reference:   "AML/CTF Act 2006, s.36; AML/CTF Rules Chapter 15",
		},
		{
			ID:          "au-austrac-smr",
			Category:    "reporting",
			Description: "Suspicious Matter Report (SMR) to AUSTRAC within 24 hours (terrorism) or 3 days",
			Mandatory:   true,
			Reference:   "AML/CTF Act 2006, s.41",
		},
		{
			ID:          "au-austrac-ttr",
			Category:    "reporting",
			Description: "Threshold Transaction Report for cash transactions AUD 10,000+ to AUSTRAC",
			Mandatory:   true,
			Reference:   "AML/CTF Act 2006, s.43",
		},
		{
			ID:          "au-austrac-ifti",
			Category:    "reporting",
			Description: "International Funds Transfer Instruction report to AUSTRAC",
			Mandatory:   true,
			Reference:   "AML/CTF Act 2006, s.45",
		},
		{
			ID:          "au-asic-ddo",
			Category:    "kyc",
			Description: "Design and Distribution Obligations: target market determination for retail products",
			Mandatory:   true,
			Reference:   "Corporations Act 2001 (Cth), Part 7.8A",
		},
		{
			ID:          "au-fatca-crs",
			Category:    "reporting",
			Description: "FATCA/CRS reporting to ATO (Australian Taxation Office)",
			Mandatory:   true,
			Reference:   "Tax Administration Act 1953, Subdivision 396-A",
		},
	}
}

func (a *Australia) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	if app.GivenName == "" {
		violations = append(violations, Violation{
			RequirementID: "au-austrac-cdd",
			Field:         "given_name",
			Message:       "Full name is required for AUSTRAC 100-point ID check",
			Severity:      "error",
		})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{
			RequirementID: "au-austrac-cdd",
			Field:         "family_name",
			Message:       "Family name is required for AUSTRAC 100-point ID check",
			Severity:      "error",
		})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{
			RequirementID: "au-austrac-cdd",
			Field:         "date_of_birth",
			Message:       "Date of birth is required for AUSTRAC 100-point ID check",
			Severity:      "error",
		})
	}

	// TFN (Tax File Number) for Australian residents
	if (app.Country == "AU" || app.CountryOfTax == "AU") && app.TaxID == "" {
		violations = append(violations, Violation{
			RequirementID: "au-austrac-cdd",
			Field:         "tax_id",
			Message:       "Tax File Number (TFN) is required for Australian tax residents",
			Severity:      "error",
		})
	}

	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{
			RequirementID: "au-austrac-cdd",
			Field:         "street",
			Message:       "Residential address is required for AUSTRAC CDD",
			Severity:      "error",
		})
	}
	if app.City == "" {
		violations = append(violations, Violation{
			RequirementID: "au-austrac-cdd",
			Field:         "city",
			Message:       "City/suburb is required",
			Severity:      "error",
		})
	}
	if app.PostalCode == "" {
		violations = append(violations, Violation{
			RequirementID: "au-austrac-cdd",
			Field:         "postal_code",
			Message:       "Postcode is required",
			Severity:      "error",
		})
	}

	if app.Email == "" {
		violations = append(violations, Violation{
			RequirementID: "au-austrac-cdd",
			Field:         "email",
			Message:       "Email address is required",
			Severity:      "error",
		})
	}

	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "au-austrac-edd",
			Field:         "is_politically_exposed",
			Message:       "PEP status declaration is required under AML/CTF Act 2006",
			Severity:      "error",
		})
	}
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{
			RequirementID: "au-austrac-edd",
			Field:         "is_politically_exposed",
			Message:       "Enhanced CDD required: customer is a PEP under AML/CTF Act 2006 s.36",
			Severity:      "warning",
		})
	}

	return violations
}

func (a *Australia) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 200_000,   // AUD 200k
		DailyMax:             1_000_000, // AUD 1M
		MonthlyMax:           5_000_000,
		CTRThreshold:         10_000, // AUD 10k threshold transaction report
		TravelRuleMin:        1_000,  // AUD 1k FATF travel rule (IFTI threshold)
		Currency:             "AUD",
	}
}
