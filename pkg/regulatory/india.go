// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// India implements the Jurisdiction interface for Indian securities regulation.
// Covers SEBI (Securities and Exchange Board of India), RBI FEMA for NRI,
// PMLA 2002 AML, and KRA (KYC Registration Agencies).
type India struct{}

func (i *India) Name() string              { return "India" }
func (i *India) Code() string              { return "IN" }
func (i *India) RegulatoryFramework() string { return "sebi" }
func (i *India) PassportableTo() []string  { return nil }

func (i *India) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "in-sebi-registration",
			Category:    "licensing",
			Description: "SEBI registration as stockbroker, merchant banker, or investment adviser",
			Mandatory:   true,
			Reference:   "SEBI (Stock Brokers) Regulations 1992; SEBI Act 1992, s.12",
		},
		{
			ID:          "in-sebi-nism",
			Category:    "licensing",
			Description: "NISM certification for registered intermediary personnel",
			Mandatory:   true,
			Reference:   "SEBI Circular SEBI/HO/MIRSD/MIRSD-PoD-1/P/CIR/2023/10",
		},
		{
			ID:          "in-kra-kyc",
			Category:    "kyc",
			Description: "KYC via SEBI-registered KRA: PAN, Aadhaar, address proof, photograph, in-person verification",
			Mandatory:   true,
			Reference:   "SEBI KYC Circular CIR/MIRSD/16/2011; SEBI (KYC Registration Agency) Regulations 2011",
		},
		{
			ID:          "in-pmla-cdd",
			Category:    "kyc",
			Description: "Customer Due Diligence under Prevention of Money Laundering Act 2002",
			Mandatory:   true,
			Reference:   "PMLA 2002, s.12; PML Rules 2005, Rule 9",
		},
		{
			ID:          "in-pmla-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Report to Financial Intelligence Unit-India (FIU-IND)",
			Mandatory:   true,
			Reference:   "PMLA 2002, s.12(1)(b); PML Rules 2005, Rule 7",
		},
		{
			ID:          "in-pmla-ctr",
			Category:    "reporting",
			Description: "Cash Transaction Report for transactions over INR 10 lakh to FIU-IND",
			Mandatory:   true,
			Reference:   "PML Rules 2005, Rule 3",
		},
		{
			ID:          "in-rbi-fema",
			Category:    "licensing",
			Description: "RBI FEMA compliance for NRI/foreign investment in Indian securities",
			Mandatory:   false,
			Reference:   "FEMA 1999; RBI Master Direction on Foreign Investment (updated 2024)",
		},
		{
			ID:          "in-sebi-investor-protection",
			Category:    "kyc",
			Description: "Risk profiling and suitability for retail investors per SEBI guidelines",
			Mandatory:   true,
			Reference:   "SEBI (Investment Advisers) Regulations 2013, Reg 17",
		},
		{
			ID:          "in-fatca-crs",
			Category:    "reporting",
			Description: "FATCA/CRS reporting via Form 61B to CBDT",
			Mandatory:   true,
			Reference:   "Income Tax Act 1961, s.285BA; IT Rules 114F-H",
		},
		{
			ID:          "in-sebi-margin",
			Category:    "capital",
			Description: "SEBI peak margin and upfront margin requirements",
			Mandatory:   true,
			Reference:   "SEBI Circular SEBI/HO/MRD/MRD-PoD-3/P/CIR/2024/125",
		},
	}
}

func (i *India) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	if app.GivenName == "" {
		violations = append(violations, Violation{
			RequirementID: "in-kra-kyc",
			Field:         "given_name",
			Message:       "Full name is required for KRA KYC registration",
			Severity:      "error",
		})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{
			RequirementID: "in-kra-kyc",
			Field:         "family_name",
			Message:       "Family name is required for KRA KYC registration",
			Severity:      "error",
		})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{
			RequirementID: "in-kra-kyc",
			Field:         "date_of_birth",
			Message:       "Date of birth is required for KRA KYC registration",
			Severity:      "error",
		})
	}

	// PAN is mandatory for all securities market participants in India
	if (app.Country == "IN" || app.CountryOfTax == "IN") && app.TaxID == "" {
		violations = append(violations, Violation{
			RequirementID: "in-kra-kyc",
			Field:         "tax_id",
			Message:       "PAN (Permanent Account Number) is mandatory for Indian securities market participation",
			Severity:      "error",
		})
	}

	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{
			RequirementID: "in-kra-kyc",
			Field:         "street",
			Message:       "Residential address with proof is required for KRA KYC",
			Severity:      "error",
		})
	}
	if app.City == "" {
		violations = append(violations, Violation{
			RequirementID: "in-kra-kyc",
			Field:         "city",
			Message:       "City is required",
			Severity:      "error",
		})
	}
	if app.PostalCode == "" {
		violations = append(violations, Violation{
			RequirementID: "in-kra-kyc",
			Field:         "postal_code",
			Message:       "PIN code is required",
			Severity:      "error",
		})
	}

	if app.Email == "" {
		violations = append(violations, Violation{
			RequirementID: "in-kra-kyc",
			Field:         "email",
			Message:       "Email address is required",
			Severity:      "error",
		})
	}

	if app.AnnualIncome == "" {
		violations = append(violations, Violation{
			RequirementID: "in-sebi-investor-protection",
			Field:         "annual_income",
			Message:       "Annual income is required for SEBI risk profiling",
			Severity:      "error",
		})
	}

	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "in-pmla-cdd",
			Field:         "is_politically_exposed",
			Message:       "PEP status declaration is required under PMLA 2002",
			Severity:      "error",
		})
	}
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{
			RequirementID: "in-pmla-cdd",
			Field:         "is_politically_exposed",
			Message:       "Enhanced Due Diligence required: customer is a PEP under PMLA 2002",
			Severity:      "warning",
		})
	}

	return violations
}

func (i *India) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 10_000_000, // INR 1 crore
		DailyMax:             50_000_000, // INR 5 crore
		MonthlyMax:           200_000_000,
		CTRThreshold:         1_000_000,  // INR 10 lakh cash reporting to FIU-IND
		TravelRuleMin:        500_000,    // INR 5 lakh FATF travel rule
		Currency:             "INR",
	}
}
