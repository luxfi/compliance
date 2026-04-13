// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// UAEADGM implements the Jurisdiction interface for Abu Dhabi Global Market (ADGM),
// regulated by the Financial Services Regulatory Authority (FSRA).
type UAEADGM struct{}

func (a *UAEADGM) Name() string              { return "UAE - ADGM" }
func (a *UAEADGM) Code() string              { return "AE-ADGM" }
func (a *UAEADGM) RegulatoryFramework() string { return "fsra" }
func (a *UAEADGM) PassportableTo() []string  { return nil }

func (a *UAEADGM) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "ae-adgm-fsra-licence",
			Category:    "licensing",
			Description: "FSRA Financial Services Permission for regulated activities",
			Mandatory:   true,
			Reference:   "FSMR 2015 (Financial Services and Markets Regulations), Part 2",
		},
		{
			ID:          "ae-adgm-fsra-virtual-assets",
			Category:    "licensing",
			Description: "FSRA Virtual Asset framework: operating a crypto exchange, providing custody, brokerage",
			Mandatory:   false,
			Reference:   "FSRA Guidance on Regulation of Virtual Asset Activities in ADGM (2020, updated 2023)",
		},
		{
			ID:          "ae-adgm-aml",
			Category:    "aml",
			Description: "AML/CFT compliance under ADGM AML Rulebook and Federal Decree-Law 20/2018",
			Mandatory:   true,
			Reference:   "ADGM AML Rulebook 2021",
		},
		{
			ID:          "ae-adgm-cdd",
			Category:    "kyc",
			Description: "Customer due diligence: identity verification, beneficial ownership, source of wealth",
			Mandatory:   true,
			Reference:   "ADGM AML Rulebook 2021, Chapter 7",
		},
		{
			ID:          "ae-adgm-edd",
			Category:    "kyc",
			Description: "Enhanced due diligence for PEPs and high-risk customers",
			Mandatory:   true,
			Reference:   "ADGM AML Rulebook 2021, Chapter 8",
		},
		{
			ID:          "ae-adgm-str",
			Category:    "reporting",
			Description: "Suspicious Activity Report to ADGM and UAE FIU",
			Mandatory:   true,
			Reference:   "ADGM AML Rulebook 2021, Chapter 12",
		},
		{
			ID:          "ae-adgm-cobs",
			Category:    "kyc",
			Description: "Conduct of Business rules: client classification, suitability assessment, best execution",
			Mandatory:   true,
			Reference:   "ADGM COBS Rulebook, Chapter 3-6",
		},
	}
}

func (a *UAEADGM) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	if app.GivenName == "" {
		violations = append(violations, Violation{RequirementID: "ae-adgm-cdd", Field: "given_name", Message: "Full name is required for FSRA CDD", Severity: "error"})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{RequirementID: "ae-adgm-cdd", Field: "family_name", Message: "Family name is required for FSRA CDD", Severity: "error"})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{RequirementID: "ae-adgm-cdd", Field: "date_of_birth", Message: "Date of birth is required for FSRA CDD", Severity: "error"})
	}
	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{RequirementID: "ae-adgm-cdd", Field: "street", Message: "Residential address is required for FSRA CDD", Severity: "error"})
	}
	if app.Email == "" {
		violations = append(violations, Violation{RequirementID: "ae-adgm-cdd", Field: "email", Message: "Email address is required", Severity: "error"})
	}
	if app.FundingSource == "" {
		violations = append(violations, Violation{RequirementID: "ae-adgm-cdd", Field: "funding_source", Message: "Source of funds/wealth is required under ADGM AML Rulebook Chapter 7", Severity: "error"})
	}

	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{RequirementID: "ae-adgm-edd", Field: "is_politically_exposed", Message: "PEP status declaration is required under ADGM AML Rulebook Chapter 8", Severity: "error"})
	}
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{RequirementID: "ae-adgm-edd", Field: "is_politically_exposed", Message: "Enhanced due diligence required: customer is a PEP", Severity: "warning"})
	}

	return violations
}

func (a *UAEADGM) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 500_000,
		DailyMax:             2_000_000,
		MonthlyMax:           10_000_000,
		CTRThreshold:         55_000,
		TravelRuleMin:        3_675,
		Currency:             "USD", // ADGM operates in USD
	}
}
