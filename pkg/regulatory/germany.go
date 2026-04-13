// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Germany implements the Jurisdiction interface for BaFin regulation.
// Covers KWG (Banking Act), KAGB (Investment Code), MiCA, and WpHG (Securities Trading Act).
type Germany struct{}

func (g *Germany) Name() string              { return "Germany" }
func (g *Germany) Code() string              { return "DE" }
func (g *Germany) RegulatoryFramework() string { return "mica" }
func (g *Germany) PassportableTo() []string  { return []string{"LU", "FR", "NL", "IE", "IT", "ES"} }

func (g *Germany) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "de-bafin-licence",
			Category:    "licensing",
			Description: "BaFin licence for banking, financial services, or investment firm under KWG",
			Mandatory:   true,
			Reference:   "Kreditwesengesetz (KWG), s.32",
		},
		{
			ID:          "de-mica-casp",
			Category:    "licensing",
			Description: "MiCA CASP authorisation via BaFin for crypto-asset services",
			Mandatory:   false,
			Reference:   "Regulation (EU) 2023/1114 (MiCA), Title V; BaFin guidance on MiCA implementation",
		},
		{
			ID:          "de-kagb",
			Category:    "licensing",
			Description: "KAGB (Kapitalanlagegesetzbuch) compliance for investment fund management (AIF, UCITS)",
			Mandatory:   false,
			Reference:   "KAGB, s.20-22",
		},
		{
			ID:          "de-gwg-kyc",
			Category:    "kyc",
			Description: "CDD under Geldwaeschegesetz (GwG): verify identity, beneficial owner, purpose",
			Mandatory:   true,
			Reference:   "GwG (Money Laundering Act), s.10-11",
		},
		{
			ID:          "de-gwg-edd",
			Category:    "kyc",
			Description: "Enhanced due diligence for PEPs and high-risk relationships",
			Mandatory:   true,
			Reference:   "GwG, s.15",
		},
		{
			ID:          "de-gwg-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Report to FIU (Zentralstelle fuer Finanztransaktionsuntersuchungen)",
			Mandatory:   true,
			Reference:   "GwG, s.43",
		},
		{
			ID:          "de-wphg-conduct",
			Category:    "kyc",
			Description: "WpHG conduct of business rules: investor classification, suitability, product governance",
			Mandatory:   true,
			Reference:   "Wertpapierhandelsgesetz (WpHG), s.63-83",
		},
		{
			ID:          "de-fatca-crs",
			Category:    "reporting",
			Description: "FATCA/CRS reporting to BZSt (Bundeszentralamt fuer Steuern)",
			Mandatory:   true,
			Reference:   "Finanzkonten-Informationsaustauschgesetz (FKAustG)",
		},
	}
}

func (g *Germany) ValidateApplication(app *ApplicationData) []Violation {
	return validateEUApplication(app, "de-gwg-kyc", "de-gwg-edd", "de-gwg-str")
}

func (g *Germany) TransactionLimits() *Limits {
	return euTransactionLimits()
}
