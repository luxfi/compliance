// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Spain implements the Jurisdiction interface for CNMV/Bank of Spain regulation.
// Covers MiCA, Ley del Mercado de Valores (LMV), and Real Decreto 217/2008.
type Spain struct{}

func (s *Spain) Name() string              { return "Spain" }
func (s *Spain) Code() string              { return "ES" }
func (s *Spain) RegulatoryFramework() string { return "mica" }
func (s *Spain) PassportableTo() []string  { return []string{"LU", "DE", "FR", "NL", "IE", "IT"} }

func (s *Spain) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "es-cnmv-licence",
			Category:    "licensing",
			Description: "CNMV authorisation for investment firm (empresa de servicios de inversion)",
			Mandatory:   true,
			Reference:   "Ley del Mercado de Valores (LMV - RDL 4/2015), art. 143",
		},
		{
			ID:          "es-mica-casp",
			Category:    "licensing",
			Description: "MiCA CASP authorisation via CNMV for crypto-asset services",
			Mandatory:   false,
			Reference:   "Regulation (EU) 2023/1114 (MiCA), Title V",
		},
		{
			ID:          "es-rd217-conduct",
			Category:    "kyc",
			Description: "Conduct of business rules per Real Decreto 217/2008: client classification, suitability, appropriateness",
			Mandatory:   true,
			Reference:   "Real Decreto 217/2008, Titulo IV",
		},
		{
			ID:          "es-aml-kyc",
			Category:    "kyc",
			Description: "CDD under Ley 10/2010 de prevencion del blanqueo de capitales",
			Mandatory:   true,
			Reference:   "Ley 10/2010, art. 3-6",
		},
		{
			ID:          "es-aml-edd",
			Category:    "kyc",
			Description: "Enhanced due diligence for PEPs and high-risk relationships",
			Mandatory:   true,
			Reference:   "Ley 10/2010, art. 11-12",
		},
		{
			ID:          "es-sepblac-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Report to SEPBLAC (Servicio Ejecutivo de la Comision de Prevencion del Blanqueo de Capitales e Infracciones Monetarias)",
			Mandatory:   true,
			Reference:   "Ley 10/2010, art. 18",
		},
		{
			ID:          "es-bde-prudential",
			Category:    "capital",
			Description: "Bank of Spain prudential supervision for credit institutions and payment entities",
			Mandatory:   false,
			Reference:   "Ley 10/2014 de ordenacion, supervision y solvencia de entidades de credito",
		},
		{
			ID:          "es-fatca-crs",
			Category:    "reporting",
			Description: "FATCA/CRS reporting to Agencia Estatal de Administracion Tributaria (AEAT)",
			Mandatory:   true,
			Reference:   "Real Decreto 1021/2015 (CRS); Orden HAP/1136/2014 (FATCA)",
		},
	}
}

func (s *Spain) ValidateApplication(app *ApplicationData) []Violation {
	return validateEUApplication(app, "es-aml-kyc", "es-aml-edd", "es-sepblac-str")
}

func (s *Spain) TransactionLimits() *Limits {
	return euTransactionLimits()
}
