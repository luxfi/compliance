// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// UAEDIFC implements the Jurisdiction interface for the Dubai International
// Financial Centre (DIFC), regulated by the DFSA (Dubai Financial Services Authority).
type UAEDIFC struct{}

func (d *UAEDIFC) Name() string              { return "UAE - DIFC" }
func (d *UAEDIFC) Code() string              { return "AE-DIFC" }
func (d *UAEDIFC) RegulatoryFramework() string { return "dfsa" }
func (d *UAEDIFC) PassportableTo() []string  { return nil }

func (d *UAEDIFC) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "ae-difc-dfsa-licence",
			Category:    "licensing",
			Description: "DFSA licence for financial services: dealing in investments, managing assets, arranging deals",
			Mandatory:   true,
			Reference:   "DFSA Rulebook, General (GEN) Module, s.2.2",
		},
		{
			ID:          "ae-difc-dfsa-crypto",
			Category:    "licensing",
			Description: "DFSA crypto token framework: recognised/accepted crypto token categories",
			Mandatory:   false,
			Reference:   "DFSA Rulebook, GEN Module s.2.2.2(j); Investment Tokens regime (2022 amendment)",
		},
		{
			ID:          "ae-difc-dfsa-aml",
			Category:    "aml",
			Description: "AML/CFT compliance per DFSA AML Module and Federal Decree-Law 20/2018",
			Mandatory:   true,
			Reference:   "DFSA Rulebook, Anti-Money Laundering (AML) Module",
		},
		{
			ID:          "ae-difc-dfsa-cdd",
			Category:    "kyc",
			Description: "Customer due diligence: verify identity, beneficial ownership, source of wealth/funds",
			Mandatory:   true,
			Reference:   "DFSA AML Module, s.7",
		},
		{
			ID:          "ae-difc-dfsa-edd",
			Category:    "kyc",
			Description: "Enhanced due diligence for PEPs and higher-risk customers",
			Mandatory:   true,
			Reference:   "DFSA AML Module, s.8",
		},
		{
			ID:          "ae-difc-dfsa-str",
			Category:    "reporting",
			Description: "Suspicious Activity Report to DFSA and UAE FIU",
			Mandatory:   true,
			Reference:   "DFSA AML Module, s.11",
		},
		{
			ID:          "ae-difc-dfsa-conduct",
			Category:    "kyc",
			Description: "Conduct of Business rules: client classification, suitability, disclosure",
			Mandatory:   true,
			Reference:   "DFSA Rulebook, Conduct of Business (COB) Module",
		},
	}
}

func (d *UAEDIFC) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	if app.GivenName == "" {
		violations = append(violations, Violation{RequirementID: "ae-difc-dfsa-cdd", Field: "given_name", Message: "Full name is required for DFSA CDD", Severity: "error"})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{RequirementID: "ae-difc-dfsa-cdd", Field: "family_name", Message: "Family name is required for DFSA CDD", Severity: "error"})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{RequirementID: "ae-difc-dfsa-cdd", Field: "date_of_birth", Message: "Date of birth is required for DFSA CDD", Severity: "error"})
	}
	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{RequirementID: "ae-difc-dfsa-cdd", Field: "street", Message: "Residential address is required for DFSA CDD", Severity: "error"})
	}
	if app.Email == "" {
		violations = append(violations, Violation{RequirementID: "ae-difc-dfsa-cdd", Field: "email", Message: "Email address is required", Severity: "error"})
	}
	if app.FundingSource == "" {
		violations = append(violations, Violation{RequirementID: "ae-difc-dfsa-cdd", Field: "funding_source", Message: "Source of funds is required under DFSA AML Module s.7", Severity: "error"})
	}

	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{RequirementID: "ae-difc-dfsa-edd", Field: "is_politically_exposed", Message: "PEP status declaration is required under DFSA AML Module s.8", Severity: "error"})
	}
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{RequirementID: "ae-difc-dfsa-edd", Field: "is_politically_exposed", Message: "Enhanced due diligence required: customer is a PEP", Severity: "warning"})
	}

	return violations
}

func (d *UAEDIFC) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 500_000,
		DailyMax:             2_000_000,
		MonthlyMax:           10_000_000,
		CTRThreshold:         55_000,
		TravelRuleMin:        3_675,
		Currency:             "USD", // DIFC operates in USD
	}
}
