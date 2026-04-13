// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// France implements the Jurisdiction interface for AMF/ACPR regulation.
// Covers MiCA, Sapin II anti-corruption, PACTE law for ICOs/DASPs,
// and Code monetaire et financier.
type France struct{}

func (f *France) Name() string              { return "France" }
func (f *France) Code() string              { return "FR" }
func (f *France) RegulatoryFramework() string { return "mica" }
func (f *France) PassportableTo() []string  { return []string{"LU", "DE", "NL", "IE", "IT", "ES"} }

func (f *France) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "fr-amf-licence",
			Category:    "licensing",
			Description: "AMF authorisation for investment services provider (prestataire de services d'investissement)",
			Mandatory:   true,
			Reference:   "Code monetaire et financier (CMF), art. L.532-1",
		},
		{
			ID:          "fr-mica-casp",
			Category:    "licensing",
			Description: "MiCA CASP authorisation via AMF (replaces PACTE DASP regime effective 30 Dec 2024)",
			Mandatory:   false,
			Reference:   "Regulation (EU) 2023/1114 (MiCA), Title V; Loi PACTE art. 86 (transitional)",
		},
		{
			ID:          "fr-acpr-aml",
			Category:    "aml",
			Description: "AML/CFT compliance supervised by ACPR (Autorite de controle prudentiel et de resolution)",
			Mandatory:   true,
			Reference:   "CMF, art. L.561-1 et seq.",
		},
		{
			ID:          "fr-aml-kyc",
			Category:    "kyc",
			Description: "CDD: verify identity, beneficial ownership, purpose of business relationship",
			Mandatory:   true,
			Reference:   "CMF, art. L.561-5 to L.561-5-1",
		},
		{
			ID:          "fr-aml-edd",
			Category:    "kyc",
			Description: "Enhanced due diligence for PEPs and high-risk relationships",
			Mandatory:   true,
			Reference:   "CMF, art. L.561-10",
		},
		{
			ID:          "fr-tracfin-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Report to Tracfin (Traitement du renseignement et action contre les circuits financiers clandestins)",
			Mandatory:   true,
			Reference:   "CMF, art. L.561-15",
		},
		{
			ID:          "fr-sapin-ii",
			Category:    "aml",
			Description: "Sapin II anti-corruption compliance for entities with 500+ employees or EUR 100M+ revenue",
			Mandatory:   false,
			Reference:   "Loi Sapin II (Law 2016-1691), art. 17",
		},
		{
			ID:          "fr-fatca-crs",
			Category:    "reporting",
			Description: "FATCA/CRS reporting to Direction generale des finances publiques (DGFiP)",
			Mandatory:   true,
			Reference:   "Code general des impots, art. 1649 AC",
		},
	}
}

func (f *France) ValidateApplication(app *ApplicationData) []Violation {
	return validateEUApplication(app, "fr-aml-kyc", "fr-aml-edd", "fr-tracfin-str")
}

func (f *France) TransactionLimits() *Limits {
	return euTransactionLimits()
}
