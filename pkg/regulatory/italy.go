// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Italy implements the Jurisdiction interface for CONSOB/Bank of Italy regulation.
// Covers MiCA, Testo Unico della Finanza (TUF), and DPR 148/2024 for DLT.
type Italy struct{}

func (it *Italy) Name() string              { return "Italy" }
func (it *Italy) Code() string              { return "IT" }
func (it *Italy) RegulatoryFramework() string { return "mica" }
func (it *Italy) PassportableTo() []string  { return []string{"LU", "DE", "FR", "NL", "IE", "ES"} }

func (it *Italy) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "it-consob-licence",
			Category:    "licensing",
			Description: "CONSOB authorisation for investment firm (SIM) or market operator",
			Mandatory:   true,
			Reference:   "Testo Unico della Finanza (TUF - D.Lgs. 58/1998), art. 18-19",
		},
		{
			ID:          "it-mica-casp",
			Category:    "licensing",
			Description: "MiCA CASP authorisation via CONSOB/Bank of Italy for crypto-asset services",
			Mandatory:   false,
			Reference:   "Regulation (EU) 2023/1114 (MiCA), Title V",
		},
		{
			ID:          "it-dlt-pilot",
			Category:    "licensing",
			Description: "EU DLT Pilot Regime participation (DPR 148/2024 transposition) for tokenized securities",
			Mandatory:   false,
			Reference:   "Regulation (EU) 2022/858 (DLT Pilot Regime); DPR 148/2024",
		},
		{
			ID:          "it-aml-kyc",
			Category:    "kyc",
			Description: "CDD under D.Lgs. 231/2007 (Italian AML decree, transposing 4AMLD/5AMLD)",
			Mandatory:   true,
			Reference:   "D.Lgs. 231/2007, art. 17-19",
		},
		{
			ID:          "it-aml-edd",
			Category:    "kyc",
			Description: "Enhanced due diligence for PEPs and high-risk relationships",
			Mandatory:   true,
			Reference:   "D.Lgs. 231/2007, art. 24-25",
		},
		{
			ID:          "it-uif-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Report to UIF (Unita di Informazione Finanziaria per l'Italia)",
			Mandatory:   true,
			Reference:   "D.Lgs. 231/2007, art. 35",
		},
		{
			ID:          "it-fatca-crs",
			Category:    "reporting",
			Description: "FATCA/CRS reporting to Agenzia delle Entrate",
			Mandatory:   true,
			Reference:   "Law 95/2015 (FATCA); DM 28 December 2015 (CRS)",
		},
	}
}

func (it *Italy) ValidateApplication(app *ApplicationData) []Violation {
	return validateEUApplication(app, "it-aml-kyc", "it-aml-edd", "it-uif-str")
}

func (it *Italy) TransactionLimits() *Limits {
	return euTransactionLimits()
}
