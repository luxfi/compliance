// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Luxembourg implements the Jurisdiction interface for CSSF regulation.
// Covers MiCA (Markets in Crypto-Assets Regulation), AIFMD (RAIF/SICAR/SICAV),
// UCITS V, and Luxembourg AML law.
type Luxembourg struct{}

func (l *Luxembourg) Name() string              { return "Luxembourg" }
func (l *Luxembourg) Code() string              { return "LU" }
func (l *Luxembourg) RegulatoryFramework() string { return "mica" }
func (l *Luxembourg) PassportableTo() []string  { return []string{"DE", "FR", "NL", "IE", "IT", "ES"} }

func (l *Luxembourg) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "lu-cssf-licence",
			Category:    "licensing",
			Description: "CSSF authorisation for investment firm, credit institution, or AIFM",
			Mandatory:   true,
			Reference:   "Law of 5 April 1993 on the Financial Sector (LSF), Part I",
		},
		{
			ID:          "lu-mica-casp",
			Category:    "licensing",
			Description: "MiCA Crypto-Asset Service Provider (CASP) authorisation via CSSF",
			Mandatory:   false,
			Reference:   "Regulation (EU) 2023/1114 (MiCA), Title V",
		},
		{
			ID:          "lu-aifmd-raif",
			Category:    "licensing",
			Description: "Reserved Alternative Investment Fund (RAIF) regime: no CSSF product approval needed, AIFM-managed",
			Mandatory:   false,
			Reference:   "Law of 23 July 2016 on RAIF",
		},
		{
			ID:          "lu-aml-kyc",
			Category:    "kyc",
			Description: "CDD under Luxembourg AML Law: verify identity, beneficial owner, source of wealth",
			Mandatory:   true,
			Reference:   "Law of 12 November 2004 (as amended), art. 3",
		},
		{
			ID:          "lu-aml-edd",
			Category:    "kyc",
			Description: "Enhanced due diligence for PEPs and high-risk third countries",
			Mandatory:   true,
			Reference:   "Law of 12 November 2004 (as amended), art. 3-2",
		},
		{
			ID:          "lu-aml-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Report to Cellule de Renseignement Financier (CRF)",
			Mandatory:   true,
			Reference:   "Law of 12 November 2004, art. 5",
		},
		{
			ID:          "lu-ucits-passport",
			Category:    "licensing",
			Description: "UCITS V fund passport for distribution across EU member states",
			Mandatory:   false,
			Reference:   "Directive 2009/65/EC (UCITS V), as transposed by Law of 17 December 2010, Part I",
		},
		{
			ID:          "lu-fatca-crs",
			Category:    "reporting",
			Description: "FATCA/CRS reporting to Administration des Contributions Directes (ACD)",
			Mandatory:   true,
			Reference:   "Law of 24 July 2015 (CRS); Law of 24 July 2015 (FATCA IGA)",
		},
		{
			ID:          "lu-cssf-circular",
			Category:    "capital",
			Description: "CSSF prudential requirements: own funds, liquidity, and capital adequacy",
			Mandatory:   true,
			Reference:   "CSSF Circular 12/552 (as amended)",
		},
	}
}

func (l *Luxembourg) ValidateApplication(app *ApplicationData) []Violation {
	return validateEUApplication(app, "lu-aml-kyc", "lu-aml-edd", "lu-aml-str")
}

func (l *Luxembourg) TransactionLimits() *Limits {
	return euTransactionLimits()
}

// validateEUApplication provides common validation for EU MiCA jurisdictions.
// The kycReqID, eddReqID, and strReqID are jurisdiction-specific requirement IDs.
func validateEUApplication(app *ApplicationData, kycReqID, eddReqID, _ string) []Violation {
	var violations []Violation

	if app.GivenName == "" {
		violations = append(violations, Violation{RequirementID: kycReqID, Field: "given_name", Message: "Full name is required for EU CDD", Severity: "error"})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{RequirementID: kycReqID, Field: "family_name", Message: "Family name is required for EU CDD", Severity: "error"})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{RequirementID: kycReqID, Field: "date_of_birth", Message: "Date of birth is required for EU CDD", Severity: "error"})
	}
	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{RequirementID: kycReqID, Field: "street", Message: "Residential address is required for EU CDD", Severity: "error"})
	}
	if app.City == "" {
		violations = append(violations, Violation{RequirementID: kycReqID, Field: "city", Message: "City is required", Severity: "error"})
	}
	if app.PostalCode == "" {
		violations = append(violations, Violation{RequirementID: kycReqID, Field: "postal_code", Message: "Postal code is required", Severity: "error"})
	}
	if app.Email == "" {
		violations = append(violations, Violation{RequirementID: kycReqID, Field: "email", Message: "Email address is required", Severity: "error"})
	}
	if app.FundingSource == "" {
		violations = append(violations, Violation{RequirementID: kycReqID, Field: "funding_source", Message: "Source of funds is required under EU AMLD", Severity: "error"})
	}

	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{RequirementID: eddReqID, Field: "is_politically_exposed", Message: "PEP status declaration is required under EU AMLD", Severity: "error"})
	}
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{RequirementID: eddReqID, Field: "is_politically_exposed", Message: "Enhanced due diligence required: customer is a PEP under EU AMLD", Severity: "warning"})
	}

	return violations
}

// euTransactionLimits returns the standard EU transaction limits.
func euTransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 200_000,   // EUR 200k
		DailyMax:             1_000_000, // EUR 1M
		MonthlyMax:           5_000_000,
		CTRThreshold:         15_000,    // EUR 15k (5AMLD/6AMLD threshold)
		TravelRuleMin:        1_000,     // EUR 1k (EU Transfer of Funds Regulation)
		Currency:             "EUR",
	}
}
