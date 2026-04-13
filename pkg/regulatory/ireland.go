// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Ireland implements the Jurisdiction interface for Central Bank of Ireland (CBI) regulation.
// Covers MiCA, CBI UCITS authorisation, and SI 110/2010 (Criminal Justice Act) AML.
type Ireland struct{}

func (i *Ireland) Name() string              { return "Ireland" }
func (i *Ireland) Code() string              { return "IE" }
func (i *Ireland) RegulatoryFramework() string { return "mica" }
func (i *Ireland) PassportableTo() []string  { return []string{"LU", "DE", "FR", "NL", "IT", "ES"} }

func (i *Ireland) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "ie-cbi-licence",
			Category:    "licensing",
			Description: "CBI authorisation for investment firm, fund manager, or payment institution",
			Mandatory:   true,
			Reference:   "Investment Intermediaries Act 1995; MiFID II (as transposed by SI 375/2017)",
		},
		{
			ID:          "ie-mica-casp",
			Category:    "licensing",
			Description: "MiCA CASP authorisation via CBI for crypto-asset services",
			Mandatory:   false,
			Reference:   "Regulation (EU) 2023/1114 (MiCA), Title V",
		},
		{
			ID:          "ie-ucits-authorisation",
			Category:    "licensing",
			Description: "CBI UCITS fund authorisation for Irish-domiciled UCITS",
			Mandatory:   false,
			Reference:   "UCITS Regulations (SI 352/2011); CBI UCITS Q&A",
		},
		{
			ID:          "ie-aml-kyc",
			Category:    "kyc",
			Description: "CDD under Criminal Justice (Money Laundering and Terrorist Financing) Act 2010",
			Mandatory:   true,
			Reference:   "CJA 2010, Part 4, s.33",
		},
		{
			ID:          "ie-aml-edd",
			Category:    "kyc",
			Description: "Enhanced due diligence for PEPs and high-risk relationships",
			Mandatory:   true,
			Reference:   "CJA 2010, s.37-39",
		},
		{
			ID:          "ie-fiu-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Report to Garda FIU and Revenue Commissioners",
			Mandatory:   true,
			Reference:   "CJA 2010, s.42",
		},
		{
			ID:          "ie-si692-crypto",
			Category:    "licensing",
			Description: "Registration under SI 692/2020 for virtual asset service providers (pre-MiCA)",
			Mandatory:   false,
			Reference:   "SI 692/2020 (Criminal Justice Act 2010, as amended)",
		},
		{
			ID:          "ie-fatca-crs",
			Category:    "reporting",
			Description: "FATCA/CRS reporting to Revenue Commissioners",
			Mandatory:   true,
			Reference:   "Taxes Consolidation Act 1997, s.891F; SI 583/2015 (CRS)",
		},
	}
}

func (i *Ireland) ValidateApplication(app *ApplicationData) []Violation {
	return validateEUApplication(app, "ie-aml-kyc", "ie-aml-edd", "ie-fiu-str")
}

func (i *Ireland) TransactionLimits() *Limits {
	return euTransactionLimits()
}
