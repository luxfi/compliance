// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// UAE implements the Jurisdiction interface for UAE federal (SCA) regulation.
// Covers Securities and Commodities Authority rules applying outside the
// free zones (DIFC, ADGM) and outside VARA's Dubai crypto scope.
type UAE struct{}

func (u *UAE) Name() string              { return "United Arab Emirates" }
func (u *UAE) Code() string              { return "AE" }
func (u *UAE) RegulatoryFramework() string { return "sca" }
func (u *UAE) PassportableTo() []string  { return nil }

func (u *UAE) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "ae-sca-licence",
			Category:    "licensing",
			Description: "SCA licence for securities brokerage, dealing, or advisory in the UAE (outside DIFC/ADGM)",
			Mandatory:   true,
			Reference:   "Federal Decree-Law No. 32/2021 on Commercial Companies; SCA Board Resolution 3/Chairman/2000",
		},
		{
			ID:          "ae-sca-aml",
			Category:    "aml",
			Description: "AML/CFT compliance under Federal Decree-Law No. 20/2018 and Cabinet Resolution No. 10/2019",
			Mandatory:   true,
			Reference:   "Federal Decree-Law No. 20/2018; Cabinet Resolution No. 10/2019",
		},
		{
			ID:          "ae-sca-kyc",
			Category:    "kyc",
			Description: "Customer identification: Emirates ID, passport, proof of address, source of funds",
			Mandatory:   true,
			Reference:   "Cabinet Resolution No. 10/2019, art. 4-8",
		},
		{
			ID:          "ae-sca-edd",
			Category:    "kyc",
			Description: "Enhanced due diligence for PEPs and high-risk relationships",
			Mandatory:   true,
			Reference:   "Cabinet Resolution No. 10/2019, art. 15-17",
		},
		{
			ID:          "ae-sca-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Report to UAE Financial Intelligence Unit (FIU)",
			Mandatory:   true,
			Reference:   "Federal Decree-Law No. 20/2018, art. 15",
		},
		{
			ID:          "ae-sca-sanctions",
			Category:    "aml",
			Description: "Screening against UAE, UN, and OFAC sanctions lists",
			Mandatory:   true,
			Reference:   "Federal Law No. 7/2014 on Combating Terrorism; Cabinet Resolution No. 74/2020",
		},
		{
			ID:          "ae-fatca-crs",
			Category:    "reporting",
			Description: "FATCA/CRS reporting to Federal Tax Authority (FTA)",
			Mandatory:   true,
			Reference:   "Cabinet Resolution No. 9/2016 (CRS); MoF Guidance Note 2017 (FATCA)",
		},
	}
}

func (u *UAE) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	if app.GivenName == "" {
		violations = append(violations, Violation{RequirementID: "ae-sca-kyc", Field: "given_name", Message: "Full name is required for SCA KYC", Severity: "error"})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{RequirementID: "ae-sca-kyc", Field: "family_name", Message: "Family name is required for SCA KYC", Severity: "error"})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{RequirementID: "ae-sca-kyc", Field: "date_of_birth", Message: "Date of birth is required for SCA KYC", Severity: "error"})
	}

	// Emirates ID for UAE residents
	if (app.Country == "AE" || app.CountryOfTax == "AE") && app.TaxID == "" {
		violations = append(violations, Violation{RequirementID: "ae-sca-kyc", Field: "tax_id", Message: "Emirates ID number is required for UAE residents", Severity: "error"})
	}

	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{RequirementID: "ae-sca-kyc", Field: "street", Message: "Residential address is required", Severity: "error"})
	}
	if app.Email == "" {
		violations = append(violations, Violation{RequirementID: "ae-sca-kyc", Field: "email", Message: "Email address is required", Severity: "error"})
	}
	if app.FundingSource == "" {
		violations = append(violations, Violation{RequirementID: "ae-sca-kyc", Field: "funding_source", Message: "Source of funds is required", Severity: "error"})
	}

	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{RequirementID: "ae-sca-edd", Field: "is_politically_exposed", Message: "PEP status declaration is required", Severity: "error"})
	}
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{RequirementID: "ae-sca-edd", Field: "is_politically_exposed", Message: "Enhanced due diligence required: customer is a PEP", Severity: "warning"})
	}

	return violations
}

func (u *UAE) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 500_000,   // AED 500k
		DailyMax:             2_000_000, // AED 2M
		MonthlyMax:           10_000_000,
		CTRThreshold:         55_000,    // AED 55k (~USD 15k equivalent, Cabinet Resolution 10/2019)
		TravelRuleMin:        3_675,     // AED 3,675 (~USD 1k FATF travel rule)
		Currency:             "AED",
	}
}
