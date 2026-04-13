// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Netherlands implements the Jurisdiction interface for AFM/DNB regulation.
// Covers MiCA, Wet op het financieel toezicht (Wft), and DNB fit-and-proper requirements.
type Netherlands struct{}

func (n *Netherlands) Name() string              { return "Netherlands" }
func (n *Netherlands) Code() string              { return "NL" }
func (n *Netherlands) RegulatoryFramework() string { return "mica" }
func (n *Netherlands) PassportableTo() []string  { return []string{"LU", "DE", "FR", "IE", "IT", "ES"} }

func (n *Netherlands) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "nl-afm-licence",
			Category:    "licensing",
			Description: "AFM licence for investment firm or financial service provider under Wft",
			Mandatory:   true,
			Reference:   "Wet op het financieel toezicht (Wft), s.2:11",
		},
		{
			ID:          "nl-mica-casp",
			Category:    "licensing",
			Description: "MiCA CASP authorisation via AFM/DNB for crypto-asset services",
			Mandatory:   false,
			Reference:   "Regulation (EU) 2023/1114 (MiCA), Title V",
		},
		{
			ID:          "nl-dnb-fit-proper",
			Category:    "licensing",
			Description: "DNB fit-and-proper assessment for directors and senior management",
			Mandatory:   true,
			Reference:   "Wft, s.3:8-3:9; DNB Beleidsregel Geschiktheid",
		},
		{
			ID:          "nl-wwft-kyc",
			Category:    "kyc",
			Description: "CDD under Wet ter voorkoming van witwassen en financieren van terrorisme (Wwft)",
			Mandatory:   true,
			Reference:   "Wwft, art. 3",
		},
		{
			ID:          "nl-wwft-edd",
			Category:    "kyc",
			Description: "Enhanced due diligence for PEPs and high-risk third countries",
			Mandatory:   true,
			Reference:   "Wwft, art. 8",
		},
		{
			ID:          "nl-fiu-str",
			Category:    "reporting",
			Description: "Unusual Transaction Report to FIU-Nederland",
			Mandatory:   true,
			Reference:   "Wwft, art. 16",
		},
		{
			ID:          "nl-fatca-crs",
			Category:    "reporting",
			Description: "FATCA/CRS reporting to Belastingdienst (Dutch Tax Authority)",
			Mandatory:   true,
			Reference:   "Wet uitvoering Common Reporting Standard; FATCA Implementation Act",
		},
	}
}

func (n *Netherlands) ValidateApplication(app *ApplicationData) []Violation {
	return validateEUApplication(app, "nl-wwft-kyc", "nl-wwft-edd", "nl-fiu-str")
}

func (n *Netherlands) TransactionLimits() *Limits {
	return euTransactionLimits()
}
