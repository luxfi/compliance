// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Switzerland implements the Jurisdiction interface for Swiss FINMA regulation.
// Covers the DLT Act 2021, AMLA (Anti-Money Laundering Act), Banking Act,
// and VASP registration requirements.
type Switzerland struct{}

func (ch *Switzerland) Name() string              { return "Switzerland" }
func (ch *Switzerland) Code() string              { return "CH" }
func (ch *Switzerland) RegulatoryFramework() string { return "finma" }
func (ch *Switzerland) PassportableTo() []string  { return nil }

func (ch *Switzerland) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "ch-finma-licence",
			Category:    "licensing",
			Description: "FINMA licence for securities dealing, banking, or fund management",
			Mandatory:   true,
			Reference:   "Financial Institutions Act (FinIA/FINIG), art. 3-5",
		},
		{
			ID:          "ch-dlt-facility",
			Category:    "licensing",
			Description: "DLT trading facility licence under DLT Act 2021 for tokenized securities",
			Mandatory:   false,
			Reference:   "Financial Market Infrastructure Act (FMIA/FinfraG), art. 73a-73i (DLT Act 2021)",
		},
		{
			ID:          "ch-amla-kyc",
			Category:    "kyc",
			Description: "CDD under Anti-Money Laundering Act: verify identity, beneficial owner, purpose",
			Mandatory:   true,
			Reference:   "AMLA (GwG), art. 3-5",
		},
		{
			ID:          "ch-amla-edd",
			Category:    "kyc",
			Description: "Enhanced due diligence for PEPs and high-risk relationships",
			Mandatory:   true,
			Reference:   "AMLA (GwG), art. 6a; AMLO-FINMA art. 13-19",
		},
		{
			ID:          "ch-amla-str",
			Category:    "reporting",
			Description: "Suspicious Activity Report to MROS (Money Laundering Reporting Office Switzerland)",
			Mandatory:   true,
			Reference:   "AMLA (GwG), art. 9",
		},
		{
			ID:          "ch-amla-record",
			Category:    "reporting",
			Description: "Record-keeping obligation: 10 years after relationship ends",
			Mandatory:   true,
			Reference:   "AMLA (GwG), art. 7",
		},
		{
			ID:          "ch-finma-vasp",
			Category:    "licensing",
			Description: "VASP (Virtual Asset Service Provider) registration with FINMA for crypto activities",
			Mandatory:   false,
			Reference:   "AMLA (GwG), art. 2 para. 3bis; FINMA Guidance 02/2019",
		},
		{
			ID:          "ch-cdb-agreement",
			Category:    "aml",
			Description: "Agreement on Due Diligence (CDB 20) obligations via SBA member banks",
			Mandatory:   true,
			Reference:   "Swiss Bankers Association CDB 20 (Convention of Diligence)",
		},
		{
			ID:          "ch-fatca-aeoi",
			Category:    "reporting",
			Description: "FATCA/AEOI (Automatic Exchange of Information) reporting to FTA",
			Mandatory:   true,
			Reference:   "AEOI Act (AIAG), Federal Act of 18 December 2015; FATCA Agreement",
		},
		{
			ID:          "ch-dlt-custodian",
			Category:    "licensing",
			Description: "FINMA-registered custodian requirement for DLT securities (ledger-based securities under CO art. 973d-i)",
			Mandatory:   false,
			Reference:   "Swiss Code of Obligations (CO), art. 973d-973i; DLT Act 2021",
		},
	}
}

func (ch *Switzerland) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	if app.GivenName == "" {
		violations = append(violations, Violation{
			RequirementID: "ch-amla-kyc",
			Field:         "given_name",
			Message:       "Full name is required for AMLA CDD",
			Severity:      "error",
		})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{
			RequirementID: "ch-amla-kyc",
			Field:         "family_name",
			Message:       "Family name is required for AMLA CDD",
			Severity:      "error",
		})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{
			RequirementID: "ch-amla-kyc",
			Field:         "date_of_birth",
			Message:       "Date of birth is required for AMLA CDD",
			Severity:      "error",
		})
	}

	// AHV number for Swiss residents
	if (app.Country == "CH" || app.CountryOfTax == "CH") && app.TaxID == "" {
		violations = append(violations, Violation{
			RequirementID: "ch-amla-kyc",
			Field:         "tax_id",
			Message:       "AHV-Nr (social insurance number) is required for Swiss tax residents",
			Severity:      "error",
		})
	}

	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{
			RequirementID: "ch-amla-kyc",
			Field:         "street",
			Message:       "Residential address is required for AMLA CDD",
			Severity:      "error",
		})
	}
	if app.City == "" {
		violations = append(violations, Violation{
			RequirementID: "ch-amla-kyc",
			Field:         "city",
			Message:       "City is required",
			Severity:      "error",
		})
	}
	if app.PostalCode == "" {
		violations = append(violations, Violation{
			RequirementID: "ch-amla-kyc",
			Field:         "postal_code",
			Message:       "PLZ (postal code) is required",
			Severity:      "error",
		})
	}

	if app.Email == "" {
		violations = append(violations, Violation{
			RequirementID: "ch-amla-kyc",
			Field:         "email",
			Message:       "Email address is required",
			Severity:      "error",
		})
	}

	if app.FundingSource == "" {
		violations = append(violations, Violation{
			RequirementID: "ch-amla-kyc",
			Field:         "funding_source",
			Message:       "Source of funds is required under AMLA art. 4",
			Severity:      "error",
		})
	}

	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "ch-amla-edd",
			Field:         "is_politically_exposed",
			Message:       "PEP status declaration is required under AMLA art. 6a",
			Severity:      "error",
		})
	}
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{
			RequirementID: "ch-amla-edd",
			Field:         "is_politically_exposed",
			Message:       "Enhanced due diligence required: customer is a PEP under AMLA art. 6a",
			Severity:      "warning",
		})
	}

	// DLT securities require FINMA-registered custodian
	if app.AccountType == "entity" && app.Country == "CH" {
		violations = append(violations, Violation{
			RequirementID: "ch-dlt-custodian",
			Field:         "account_type",
			Message:       "Entity accounts for DLT securities require a FINMA-registered custodian (CO art. 973d-i)",
			Severity:      "warning",
		})
	}

	return violations
}

func (ch *Switzerland) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 200_000,   // CHF 200k
		DailyMax:             1_000_000, // CHF 1M
		MonthlyMax:           5_000_000,
		CTRThreshold:         100_000,   // CHF 100k AMLA threshold for enhanced verification
		TravelRuleMin:        1_000,     // CHF 1k FATF travel rule (AMLO-FINMA art. 10)
		Currency:             "CHF",
	}
}
