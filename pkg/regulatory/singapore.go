// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Singapore implements the Jurisdiction interface for MAS regulation.
// Covers the Securities and Futures Act (SFA), Payment Services Act 2019 (PS Act),
// and MAS AML/CFT notices.
type Singapore struct{}

func (s *Singapore) Name() string              { return "Singapore" }
func (s *Singapore) Code() string              { return "SG" }
func (s *Singapore) RegulatoryFramework() string { return "mas" }
func (s *Singapore) PassportableTo() []string  { return nil }

func (s *Singapore) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "sg-mas-cml",
			Category:    "licensing",
			Description: "Capital Markets Services (CMS) licence for dealing in securities",
			Mandatory:   true,
			Reference:   "Securities and Futures Act 2001 (SFA), s.82",
		},
		{
			ID:          "sg-mas-dpt",
			Category:    "licensing",
			Description: "Digital Payment Token (DPT) service licence under Payment Services Act 2019",
			Mandatory:   false,
			Reference:   "Payment Services Act 2019 (PS Act), s.6; MAS Notice PSN02",
		},
		{
			ID:          "sg-mas-sfa04n12",
			Category:    "aml",
			Description: "AML/CFT measures per MAS Notice SFA 04-N12 for capital markets intermediaries",
			Mandatory:   true,
			Reference:   "MAS Notice SFA 04-N12",
		},
		{
			ID:          "sg-mas-cdd",
			Category:    "kyc",
			Description: "Customer Due Diligence: verify identity, beneficial ownership, purpose of relationship",
			Mandatory:   true,
			Reference:   "MAS Notice SFA 04-N12, Part IV",
		},
		{
			ID:          "sg-mas-edd",
			Category:    "kyc",
			Description: "Enhanced CDD for higher-risk customers (PEPs, high-risk countries, complex structures)",
			Mandatory:   true,
			Reference:   "MAS Notice SFA 04-N12, Part V",
		},
		{
			ID:          "sg-mas-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Report to Suspicious Transaction Reporting Office (STRO)",
			Mandatory:   true,
			Reference:   "Corruption, Drug Trafficking and Other Serious Crimes Act (CDSA), s.39",
		},
		{
			ID:          "sg-mas-ctr",
			Category:    "reporting",
			Description: "Cash Transaction Report for cash transactions SGD 20,000+ to STRO",
			Mandatory:   true,
			Reference:   "MAS Notice SFA 04-N12, Part IX",
		},
		{
			ID:          "sg-mas-fatca-crs",
			Category:    "reporting",
			Description: "FATCA/CRS reporting to IRAS (Inland Revenue Authority of Singapore)",
			Mandatory:   true,
			Reference:   "Income Tax Act s.105M; Income Tax (International Tax Compliance Agreements) Regulations",
		},
		{
			ID:          "sg-mas-risk-assessment",
			Category:    "aml",
			Description: "Enterprise-wide risk assessment for ML/TF risks",
			Mandatory:   true,
			Reference:   "MAS Notice SFA 04-N12, Part III",
		},
		{
			ID:          "sg-mas-fit-proper",
			Category:    "licensing",
			Description: "Fit and proper criteria for substantial shareholders and key officers",
			Mandatory:   true,
			Reference:   "MAS Guidelines on Fit and Proper Criteria [FSG-G01]",
		},
	}
}

func (s *Singapore) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	if app.GivenName == "" {
		violations = append(violations, Violation{
			RequirementID: "sg-mas-cdd",
			Field:         "given_name",
			Message:       "Full name is required for MAS CDD",
			Severity:      "error",
		})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{
			RequirementID: "sg-mas-cdd",
			Field:         "family_name",
			Message:       "Family name is required for MAS CDD",
			Severity:      "error",
		})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{
			RequirementID: "sg-mas-cdd",
			Field:         "date_of_birth",
			Message:       "Date of birth is required for MAS CDD",
			Severity:      "error",
		})
	}

	// NRIC/FIN for Singapore residents
	if (app.Country == "SG" || app.CountryOfTax == "SG") && app.TaxID == "" {
		violations = append(violations, Violation{
			RequirementID: "sg-mas-cdd",
			Field:         "tax_id",
			Message:       "NRIC or FIN is required for Singapore residents",
			Severity:      "error",
		})
	}

	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{
			RequirementID: "sg-mas-cdd",
			Field:         "street",
			Message:       "Residential address is required for MAS CDD",
			Severity:      "error",
		})
	}
	if app.City == "" {
		violations = append(violations, Violation{
			RequirementID: "sg-mas-cdd",
			Field:         "city",
			Message:       "City is required",
			Severity:      "error",
		})
	}
	if app.PostalCode == "" {
		violations = append(violations, Violation{
			RequirementID: "sg-mas-cdd",
			Field:         "postal_code",
			Message:       "Postal code is required",
			Severity:      "error",
		})
	}

	if app.Email == "" {
		violations = append(violations, Violation{
			RequirementID: "sg-mas-cdd",
			Field:         "email",
			Message:       "Email address is required",
			Severity:      "error",
		})
	}

	if app.FundingSource == "" {
		violations = append(violations, Violation{
			RequirementID: "sg-mas-cdd",
			Field:         "funding_source",
			Message:       "Source of funds is required under MAS CDD",
			Severity:      "error",
		})
	}

	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "sg-mas-edd",
			Field:         "is_politically_exposed",
			Message:       "PEP status declaration is required under MAS Notice SFA 04-N12",
			Severity:      "error",
		})
	}
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{
			RequirementID: "sg-mas-edd",
			Field:         "is_politically_exposed",
			Message:       "Enhanced CDD required: customer is a PEP under MAS Notice SFA 04-N12 Part V",
			Severity:      "warning",
		})
	}

	return violations
}

func (s *Singapore) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 200_000,   // SGD 200k
		DailyMax:             1_000_000, // SGD 1M
		MonthlyMax:           5_000_000,
		CTRThreshold:         20_000, // SGD 20k cash reporting
		TravelRuleMin:        1_500,  // SGD 1.5k FATF travel rule
		Currency:             "SGD",
	}
}
