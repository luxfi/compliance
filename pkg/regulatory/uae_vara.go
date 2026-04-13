// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// UAEVARA implements the Jurisdiction interface for Dubai's Virtual Assets
// Regulatory Authority (VARA). VARA regulates virtual asset activities
// specifically within Dubai (outside DIFC), established by Law No. 4 of 2022.
type UAEVARA struct{}

func (v *UAEVARA) Name() string              { return "UAE - VARA (Dubai)" }
func (v *UAEVARA) Code() string              { return "AE-VARA" }
func (v *UAEVARA) RegulatoryFramework() string { return "vara" }
func (v *UAEVARA) PassportableTo() []string  { return nil }

func (v *UAEVARA) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "ae-vara-licence",
			Category:    "licensing",
			Description: "VARA licence for virtual asset service provision (exchange, broker-dealer, custody, lending, payments, advisory, transfer/settlement)",
			Mandatory:   true,
			Reference:   "Law No. 4 of 2022 (Virtual Assets and Related Activities); VARA Rulebook 2023",
		},
		{
			ID:          "ae-vara-aml",
			Category:    "aml",
			Description: "AML/CFT compliance per VARA Compliance and Risk Management Rulebook",
			Mandatory:   true,
			Reference:   "VARA Compliance and Risk Management Rulebook 2023; Federal Decree-Law 20/2018",
		},
		{
			ID:          "ae-vara-kyc",
			Category:    "kyc",
			Description: "Customer identification: Emirates ID/passport, proof of address, source of funds for virtual asset transactions",
			Mandatory:   true,
			Reference:   "VARA Compliance and Risk Management Rulebook 2023, Chapter 4",
		},
		{
			ID:          "ae-vara-travel-rule",
			Category:    "aml",
			Description: "FATF Travel Rule compliance for virtual asset transfers (originator/beneficiary info)",
			Mandatory:   true,
			Reference:   "VARA Compliance and Risk Management Rulebook 2023; FATF Recommendation 16",
		},
		{
			ID:          "ae-vara-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Report to UAE FIU and VARA",
			Mandatory:   true,
			Reference:   "VARA Compliance and Risk Management Rulebook 2023, Chapter 5",
		},
		{
			ID:          "ae-vara-technology",
			Category:    "licensing",
			Description: "Technology and information governance requirements for VASP platforms",
			Mandatory:   true,
			Reference:   "VARA Technology and Information Governance Rulebook 2023",
		},
		{
			ID:          "ae-vara-market-conduct",
			Category:    "reporting",
			Description: "Market conduct rules: prohibition on market manipulation and insider trading for virtual assets",
			Mandatory:   true,
			Reference:   "VARA Market Conduct Rulebook 2023",
		},
	}
}

func (v *UAEVARA) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	if app.GivenName == "" {
		violations = append(violations, Violation{RequirementID: "ae-vara-kyc", Field: "given_name", Message: "Full name is required for VARA KYC", Severity: "error"})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{RequirementID: "ae-vara-kyc", Field: "family_name", Message: "Family name is required for VARA KYC", Severity: "error"})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{RequirementID: "ae-vara-kyc", Field: "date_of_birth", Message: "Date of birth is required for VARA KYC", Severity: "error"})
	}
	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{RequirementID: "ae-vara-kyc", Field: "street", Message: "Residential address is required", Severity: "error"})
	}
	if app.Email == "" {
		violations = append(violations, Violation{RequirementID: "ae-vara-kyc", Field: "email", Message: "Email address is required", Severity: "error"})
	}
	if app.FundingSource == "" {
		violations = append(violations, Violation{RequirementID: "ae-vara-kyc", Field: "funding_source", Message: "Source of funds is required under VARA Compliance Rulebook Chapter 4", Severity: "error"})
	}

	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{RequirementID: "ae-vara-aml", Field: "is_politically_exposed", Message: "PEP status declaration is required under VARA AML rules", Severity: "error"})
	}
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{RequirementID: "ae-vara-aml", Field: "is_politically_exposed", Message: "Enhanced due diligence required: customer is a PEP", Severity: "warning"})
	}

	return violations
}

func (v *UAEVARA) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 500_000,
		DailyMax:             2_000_000,
		MonthlyMax:           10_000_000,
		CTRThreshold:         55_000,
		TravelRuleMin:        3_675, // AED 3,675 (~USD 1k FATF travel rule, applicable to virtual assets)
		Currency:             "AED",
	}
}
