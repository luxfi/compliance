// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Brazil implements the Jurisdiction interface for Brazilian securities regulation.
// Covers CVM (Comissao de Valores Mobiliarios), B3 exchange rules, COAF AML,
// and Receita Federal tax obligations.
type Brazil struct{}

func (b *Brazil) Name() string              { return "Brazil" }
func (b *Brazil) Code() string              { return "BR" }
func (b *Brazil) RegulatoryFramework() string { return "cvm" }
func (b *Brazil) PassportableTo() []string  { return nil }

func (b *Brazil) Requirements() []Requirement {
	return []Requirement{
		{
			ID:          "br-cvm-registration",
			Category:    "licensing",
			Description: "CVM registration as securities intermediary (corretora or distribuidora)",
			Mandatory:   true,
			Reference:   "Lei 6.385/76, art. 16; CVM Resolucao 178/2025",
		},
		{
			ID:          "br-cvm-crowdfunding",
			Category:    "licensing",
			Description: "CVM crowdfunding platform registration under Instrucao CVM 88/2022 (replaced ICVM 588)",
			Mandatory:   false,
			Reference:   "Instrucao CVM 88/2022",
		},
		{
			ID:          "br-coaf-registration",
			Category:    "aml",
			Description: "COAF (Conselho de Controle de Atividades Financeiras) registration for AML obligations",
			Mandatory:   true,
			Reference:   "Lei 9.613/98, art. 9",
		},
		{
			ID:          "br-coaf-kyc",
			Category:    "kyc",
			Description: "Customer identification: CPF/CNPJ, full name, address, occupation, income",
			Mandatory:   true,
			Reference:   "BCB Circular 3.978/2020, art. 8",
		},
		{
			ID:          "br-coaf-pep",
			Category:    "kyc",
			Description: "PEP screening: identify pessoas politicamente expostas per COAF Resolucao 40/2021",
			Mandatory:   true,
			Reference:   "COAF Resolucao 40/2021",
		},
		{
			ID:          "br-coaf-str",
			Category:    "reporting",
			Description: "Suspicious Transaction Report (Comunicacao de Operacao Suspeita) to COAF",
			Mandatory:   true,
			Reference:   "Lei 9.613/98, art. 11",
		},
		{
			ID:          "br-receita-fatca",
			Category:    "reporting",
			Description: "FATCA reporting to Receita Federal via e-Financeira",
			Mandatory:   true,
			Reference:   "IN RFB 1.571/2015; IGA Model 1 (Brazil-US)",
		},
		{
			ID:          "br-crypto-marco",
			Category:    "licensing",
			Description: "Virtual Asset Service Provider registration under Lei 14.478/2022 (Marco Legal das Criptomoedas)",
			Mandatory:   false,
			Reference:   "Lei 14.478/2022; BCB Resolucao 338/2023",
		},
		{
			ID:          "br-b3-rules",
			Category:    "licensing",
			Description: "B3 exchange participant rules for listed securities trading",
			Mandatory:   false,
			Reference:   "B3 Regulamento de Operacoes",
		},
		{
			ID:          "br-cvm-suitability",
			Category:    "kyc",
			Description: "Suitability assessment: risk profile, investment objectives, financial capacity",
			Mandatory:   true,
			Reference:   "CVM Resolucao 30/2021 (replaced ICVM 539)",
		},
	}
}

func (b *Brazil) ValidateApplication(app *ApplicationData) []Violation {
	var violations []Violation

	if app.GivenName == "" {
		violations = append(violations, Violation{
			RequirementID: "br-coaf-kyc",
			Field:         "given_name",
			Message:       "Nome completo is required for Brazilian KYC",
			Severity:      "error",
		})
	}
	if app.FamilyName == "" {
		violations = append(violations, Violation{
			RequirementID: "br-coaf-kyc",
			Field:         "family_name",
			Message:       "Sobrenome is required for Brazilian KYC",
			Severity:      "error",
		})
	}
	if app.DateOfBirth == "" {
		violations = append(violations, Violation{
			RequirementID: "br-coaf-kyc",
			Field:         "date_of_birth",
			Message:       "Data de nascimento is required for Brazilian KYC",
			Severity:      "error",
		})
	}

	// CPF (11 digits) or CNPJ (14 digits) required for Brazilian residents
	if (app.Country == "BR" || app.CountryOfTax == "BR") && app.TaxID == "" {
		violations = append(violations, Violation{
			RequirementID: "br-coaf-kyc",
			Field:         "tax_id",
			Message:       "CPF (pessoa fisica) or CNPJ (pessoa juridica) is required for Brazilian tax residents",
			Severity:      "error",
		})
	}

	if len(app.Street) == 0 || app.Street[0] == "" {
		violations = append(violations, Violation{
			RequirementID: "br-coaf-kyc",
			Field:         "street",
			Message:       "Residential address is required",
			Severity:      "error",
		})
	}
	if app.City == "" {
		violations = append(violations, Violation{
			RequirementID: "br-coaf-kyc",
			Field:         "city",
			Message:       "City is required",
			Severity:      "error",
		})
	}

	if app.Email == "" {
		violations = append(violations, Violation{
			RequirementID: "br-coaf-kyc",
			Field:         "email",
			Message:       "Email address is required",
			Severity:      "error",
		})
	}

	// CVM suitability
	if app.InvestmentObjective == "" {
		violations = append(violations, Violation{
			RequirementID: "br-cvm-suitability",
			Field:         "investment_objective",
			Message:       "Investment objective is required for CVM suitability (Resolucao 30/2021)",
			Severity:      "error",
		})
	}
	if app.AnnualIncome == "" {
		violations = append(violations, Violation{
			RequirementID: "br-cvm-suitability",
			Field:         "annual_income",
			Message:       "Annual income is required for CVM suitability assessment",
			Severity:      "error",
		})
	}

	// PEP screening (COAF requirement)
	if app.IsPoliticallyExposed == nil {
		violations = append(violations, Violation{
			RequirementID: "br-coaf-pep",
			Field:         "is_politically_exposed",
			Message:       "PEP status declaration is required under COAF Resolucao 40/2021",
			Severity:      "error",
		})
	}
	if app.IsPoliticallyExposed != nil && *app.IsPoliticallyExposed {
		violations = append(violations, Violation{
			RequirementID: "br-coaf-pep",
			Field:         "is_politically_exposed",
			Message:       "Enhanced Due Diligence required: customer is a pessoa politicamente exposta",
			Severity:      "warning",
		})
	}

	return violations
}

func (b *Brazil) TransactionLimits() *Limits {
	return &Limits{
		SingleTransactionMax: 500_000,   // BRL 500k
		DailyMax:             2_000_000, // BRL 2M
		MonthlyMax:           10_000_000,
		CTRThreshold:         50_000,    // COAF automatic reporting at BRL 50k (BCB Circular 3.978/2020)
		TravelRuleMin:        5_000,     // BRL 5k for FATF travel rule
		Currency:             "BRL",
	}
}
