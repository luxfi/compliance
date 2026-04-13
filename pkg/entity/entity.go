// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package entity defines regulated financial entity types (ATS, BD, TransferAgent, MSB)
// with their required licenses, reporting obligations, and capital requirements.
package entity

// RegulatedEntity is the interface for a regulated financial entity.
type RegulatedEntity interface {
	Name() string
	Type() string
	Jurisdiction() string
	RequiredLicenses() []License
	ReportingObligations() []ReportingObligation
	CapitalRequirements() *CapitalRequirement
	OperationalRequirements() []OperationalRequirement
}

// License describes a required regulatory license.
type License struct {
	Name        string `json:"name"`
	Regulator   string `json:"regulator"`
	Reference   string `json:"reference"`
	Description string `json:"description"`
}

// ReportingObligation describes a required regulatory report.
type ReportingObligation struct {
	Name       string `json:"name"`
	Frequency  string `json:"frequency"` // daily, quarterly, annual, event-driven
	Regulator  string `json:"regulator"`
	Reference  string `json:"reference"`
	Description string `json:"description"`
}

// CapitalRequirement describes minimum capital requirements.
type CapitalRequirement struct {
	MinNetCapital   float64 `json:"min_net_capital"`
	Currency        string  `json:"currency"`
	CalculationRule string  `json:"calculation_rule"`
	Reference       string  `json:"reference"`
}

// OperationalRequirement describes an operational compliance requirement.
type OperationalRequirement struct {
	Name        string `json:"name"`
	Category    string `json:"category"` // technology, personnel, books_and_records, supervisory
	Description string `json:"description"`
	Reference   string `json:"reference"`
}

// Entity type constants. Use these instead of string literals.
const (
	EntityType_ATS          = "ats"
	EntityType_BrokerDealer = "broker_dealer"
	EntityType_TransferAgent = "transfer_agent"
	EntityType_MSB          = "msb"
	EntityType_SICAV        = "sicav"
	EntityType_SICAR        = "sicar"
	EntityType_RAIF         = "raif"
	EntityType_AIFM         = "aifm"
	EntityType_MANCOMAN     = "mancoman"
	EntityType_CRR          = "crr"
	EntityType_ISSUER       = "issuer"
	EntityType_CUSTODIAN    = "custodian"
	EntityType_DLT_FACILITY = "dlt_facility"
)

// ATS is an Alternative Trading System registered under SEC Regulation ATS.
type ATS struct{}

func (a *ATS) Name() string         { return "Alternative Trading System" }
func (a *ATS) Type() string         { return EntityType_ATS }
func (a *ATS) Jurisdiction() string { return "US" }

func (a *ATS) RequiredLicenses() []License {
	return []License{
		{
			Name:        "Broker-Dealer Registration",
			Regulator:   "SEC/FINRA",
			Reference:   "Securities Exchange Act Section 15(a)",
			Description: "ATS must register as a broker-dealer before operating",
		},
		{
			Name:        "ATS Registration (Form ATS/ATS-N)",
			Regulator:   "SEC",
			Reference:   "SEC Rule 301(b)(2), Regulation ATS",
			Description: "File Form ATS or ATS-N with SEC and operate under Regulation ATS",
		},
		{
			Name:        "FINRA Membership",
			Regulator:   "FINRA",
			Reference:   "FINRA By-Laws, Article III",
			Description: "FINRA member firm membership required for BD operating ATS",
		},
	}
}

func (a *ATS) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{
			Name:        "Form ATS-R Quarterly Report",
			Frequency:   "quarterly",
			Regulator:   "SEC",
			Reference:   "SEC Rule 301(b)(9), Regulation ATS",
			Description: "Quarterly transaction volume and subscriber reporting",
		},
		{
			Name:        "Form ATS-N Amendment",
			Frequency:   "event-driven",
			Regulator:   "SEC",
			Reference:   "SEC Rule 304, Regulation ATS",
			Description: "Material changes to ATS operations must be reported",
		},
		{
			Name:        "FOCUS Report",
			Frequency:   "quarterly",
			Regulator:   "FINRA/SEC",
			Reference:   "SEC Rule 17a-5",
			Description: "Financial and Operational Combined Uniform Single Report",
		},
		{
			Name:        "SAR Filing",
			Frequency:   "event-driven",
			Regulator:   "FinCEN",
			Reference:   "31 CFR 1023.320",
			Description: "Suspicious Activity Reports for BD-operated ATS",
		},
	}
}

func (a *ATS) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{
		MinNetCapital:   250_000,
		Currency:        "USD",
		CalculationRule: "SEC Rule 15c3-1 (net capital rule)",
		Reference:       "17 CFR 240.15c3-1",
	}
}

func (a *ATS) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{
			Name:        "Fair Access",
			Category:    "supervisory",
			Description: "ATS exceeding 5% volume in NMS stock must provide fair access",
			Reference:   "SEC Rule 301(b)(5), Regulation ATS",
		},
		{
			Name:        "Order Display",
			Category:    "technology",
			Description: "Display best-priced orders if exceeding 5% volume threshold",
			Reference:   "SEC Rule 301(b)(3), Regulation ATS",
		},
		{
			Name:        "System Capacity and Integrity",
			Category:    "technology",
			Description: "Adequate system capacity, security, and business continuity",
			Reference:   "SEC Rule 301(b)(6), Regulation ATS",
		},
		{
			Name:        "Books and Records",
			Category:    "books_and_records",
			Description: "Maintain subscriber records and order/trade records per SEC rules",
			Reference:   "SEC Rule 301(b)(8), Regulation ATS",
		},
	}
}

// BrokerDealer is a Broker-Dealer registered with SEC and FINRA.
type BrokerDealer struct{}

func (bd *BrokerDealer) Name() string         { return "Broker-Dealer" }
func (bd *BrokerDealer) Type() string         { return EntityType_BrokerDealer }
func (bd *BrokerDealer) Jurisdiction() string { return "US" }

func (bd *BrokerDealer) RequiredLicenses() []License {
	return []License{
		{
			Name:        "SEC Registration",
			Regulator:   "SEC",
			Reference:   "Securities Exchange Act Section 15(a)",
			Description: "Register with SEC on Form BD",
		},
		{
			Name:        "FINRA Membership",
			Regulator:   "FINRA",
			Reference:   "Securities Exchange Act Section 15(b)",
			Description: "Membership in FINRA as self-regulatory organization",
		},
		{
			Name:        "State Registration",
			Regulator:   "State Securities Regulators",
			Reference:   "Uniform Securities Act",
			Description: "Register in each state where business is conducted",
		},
		{
			Name:        "SIPC Membership",
			Regulator:   "SIPC",
			Reference:   "Securities Investor Protection Act",
			Description: "Securities Investor Protection Corporation membership",
		},
	}
}

func (bd *BrokerDealer) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{
			Name:        "FOCUS Report",
			Frequency:   "quarterly",
			Regulator:   "FINRA/SEC",
			Reference:   "SEC Rule 17a-5",
			Description: "Financial and Operational Combined Uniform Single Report",
		},
		{
			Name:        "Annual Audit",
			Frequency:   "annual",
			Regulator:   "SEC",
			Reference:   "SEC Rule 17a-5(d)",
			Description: "Annual audited financial statements by independent public accountant",
		},
		{
			Name:        "Customer Complaints",
			Frequency:   "quarterly",
			Regulator:   "FINRA",
			Reference:   "FINRA Rule 4530",
			Description: "Report customer complaints and regulatory actions",
		},
		{
			Name:        "SAR Filing",
			Frequency:   "event-driven",
			Regulator:   "FinCEN",
			Reference:   "31 CFR 1023.320",
			Description: "Suspicious Activity Reports",
		},
		{
			Name:        "CTR Filing",
			Frequency:   "event-driven",
			Regulator:   "FinCEN",
			Reference:   "31 CFR 1010.311",
			Description: "Currency Transaction Reports for cash transactions over $10,000",
		},
	}
}

func (bd *BrokerDealer) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{
		MinNetCapital:   250_000,
		Currency:        "USD",
		CalculationRule: "SEC Rule 15c3-1 (aggregate indebtedness method or alternative)",
		Reference:       "17 CFR 240.15c3-1",
	}
}

func (bd *BrokerDealer) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{
			Name:        "Customer Protection",
			Category:    "supervisory",
			Description: "Segregate customer funds and securities per Rule 15c3-3",
			Reference:   "17 CFR 240.15c3-3",
		},
		{
			Name:        "Supervisory System",
			Category:    "supervisory",
			Description: "Written supervisory procedures and designated supervisory personnel",
			Reference:   "FINRA Rule 3110",
		},
		{
			Name:        "Anti-Money Laundering Program",
			Category:    "supervisory",
			Description: "AML compliance program with CIP, ongoing monitoring, SAR filing",
			Reference:   "31 CFR 1023.210",
		},
		{
			Name:        "Books and Records",
			Category:    "books_and_records",
			Description: "Maintain books and records per SEC Rules 17a-3 and 17a-4",
			Reference:   "17 CFR 240.17a-3, 17a-4",
		},
		{
			Name:        "Business Continuity Plan",
			Category:    "technology",
			Description: "Written BCP addressing data backup, mission critical systems, communications",
			Reference:   "FINRA Rule 4370",
		},
	}
}

// TransferAgent is a Transfer Agent registered with SEC.
type TransferAgent struct{}

func (ta *TransferAgent) Name() string         { return "Transfer Agent" }
func (ta *TransferAgent) Type() string         { return EntityType_TransferAgent }
func (ta *TransferAgent) Jurisdiction() string { return "US" }

func (ta *TransferAgent) RequiredLicenses() []License {
	return []License{
		{
			Name:        "SEC Registration",
			Regulator:   "SEC",
			Reference:   "Securities Exchange Act Section 17A",
			Description: "Register with SEC on Form TA-1",
		},
	}
}

func (ta *TransferAgent) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{
			Name:        "Form TA-2 Annual Report",
			Frequency:   "annual",
			Regulator:   "SEC",
			Reference:   "SEC Rule 17Ad-17(b)",
			Description: "Annual report of transfer agent activities and financial condition",
		},
		{
			Name:        "Lost Securityholder Searches",
			Frequency:   "annual",
			Regulator:   "SEC",
			Reference:   "SEC Rule 17Ad-17",
			Description: "Annual search for lost securityholders using database",
		},
		{
			Name:        "Form TA-W Withdrawal",
			Frequency:   "event-driven",
			Regulator:   "SEC",
			Reference:   "SEC Rule 17Ac2-2",
			Description: "Withdrawal of registration when ceasing transfer agent activities",
		},
	}
}

func (ta *TransferAgent) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{
		MinNetCapital:   0, // No specific minimum, but must demonstrate operational capacity
		Currency:        "USD",
		CalculationRule: "No minimum net capital rule; operational surety bond may be required",
		Reference:       "SEC Rule 17Ad",
	}
}

func (ta *TransferAgent) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{
			Name:        "Turnaround Performance",
			Category:    "supervisory",
			Description: "Process 90% of routine items within 3 business days",
			Reference:   "SEC Rule 17Ad-2",
		},
		{
			Name:        "Record Keeping",
			Category:    "books_and_records",
			Description: "Maintain records of transfers, cancellations, and issuances",
			Reference:   "SEC Rule 17Ad-6, 17Ad-7",
		},
		{
			Name:        "Safeguarding",
			Category:    "technology",
			Description: "Safeguard funds and securities, maintain adequate insurance",
			Reference:   "SEC Rule 17Ad-12",
		},
		{
			Name:        "Written Procedures",
			Category:    "supervisory",
			Description: "Written procedures for transfer operations and error correction",
			Reference:   "SEC Rule 17Ad-15",
		},
	}
}

// MSB is a Money Service Business registered with FinCEN.
type MSB struct{}

func (m *MSB) Name() string         { return "Money Service Business" }
func (m *MSB) Type() string         { return EntityType_MSB }
func (m *MSB) Jurisdiction() string { return "US" }

func (m *MSB) RequiredLicenses() []License {
	return []License{
		{
			Name:        "FinCEN Registration",
			Regulator:   "FinCEN",
			Reference:   "31 CFR 1022.380",
			Description: "Register with FinCEN as an MSB",
		},
		{
			Name:        "State Money Transmitter Licenses",
			Regulator:   "State Banking Regulators",
			Reference:   "State Money Transmission Laws",
			Description: "Obtain money transmitter license in each operating state",
		},
	}
}

func (m *MSB) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{
			Name:        "CTR Filing",
			Frequency:   "event-driven",
			Regulator:   "FinCEN",
			Reference:   "31 CFR 1010.311",
			Description: "Currency Transaction Report for cash transactions over $10,000",
		},
		{
			Name:        "SAR Filing",
			Frequency:   "event-driven",
			Regulator:   "FinCEN",
			Reference:   "31 CFR 1022.320",
			Description: "Suspicious Activity Report for transactions $2,000+ that appear suspicious",
		},
		{
			Name:        "FinCEN Re-registration",
			Frequency:   "biennial",
			Regulator:   "FinCEN",
			Reference:   "31 CFR 1022.380(b)(4)",
			Description: "Renew FinCEN registration every 2 years",
		},
		{
			Name:        "FBAR Filing",
			Frequency:   "annual",
			Regulator:   "FinCEN",
			Reference:   "31 CFR 1010.350",
			Description: "Report of Foreign Bank and Financial Accounts if aggregate >$10,000",
		},
	}
}

func (m *MSB) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{
		MinNetCapital:   0, // Varies by state; FinCEN has no federal minimum
		Currency:        "USD",
		CalculationRule: "State-specific: typically surety bond ($10k-$2M) plus net worth requirements",
		Reference:       "State Money Transmission Laws",
	}
}

func (m *MSB) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{
			Name:        "AML Compliance Program",
			Category:    "supervisory",
			Description: "Written AML program: internal policies, designated compliance officer, training, independent audit",
			Reference:   "31 CFR 1022.210",
		},
		{
			Name:        "Record Keeping",
			Category:    "books_and_records",
			Description: "Maintain records of money transfers $3,000+ for 5 years",
			Reference:   "31 CFR 1010.410(e)",
		},
		{
			Name:        "Customer Identification",
			Category:    "supervisory",
			Description: "Verify identity for money transfers $3,000+",
			Reference:   "31 CFR 1010.410(e)",
		},
		{
			Name:        "Agent Oversight",
			Category:    "supervisory",
			Description: "Maintain list of agents and ensure agent compliance with BSA",
			Reference:   "31 CFR 1022.380(d)",
		},
	}
}

// GetEntity returns a regulated entity by type string.
func GetEntity(entityType string) RegulatedEntity {
	switch entityType {
	case EntityType_ATS:
		return &ATS{}
	case EntityType_BrokerDealer:
		return &BrokerDealer{}
	case EntityType_TransferAgent:
		return &TransferAgent{}
	case EntityType_MSB:
		return &MSB{}
	case EntityType_SICAV:
		return &SICAV{}
	case EntityType_SICAR:
		return &SICAR{}
	case EntityType_RAIF:
		return &RAIF{}
	case EntityType_AIFM:
		return &AIFM{}
	case EntityType_MANCOMAN:
		return &ManCo{}
	case EntityType_CRR:
		return &CRR{}
	case EntityType_ISSUER:
		return &Issuer{}
	case EntityType_CUSTODIAN:
		return &Custodian{}
	case EntityType_DLT_FACILITY:
		return &DLTFacility{}
	default:
		return nil
	}
}

// AllEntities returns all regulated entity types.
func AllEntities() []RegulatedEntity {
	return []RegulatedEntity{
		&ATS{}, &BrokerDealer{}, &TransferAgent{}, &MSB{},
		&SICAV{}, &SICAR{}, &RAIF{}, &AIFM{}, &ManCo{},
		&CRR{}, &Issuer{}, &Custodian{}, &DLTFacility{},
	}
}

// --- Luxembourg investment vehicles ---

// SICAV is a Luxembourg variable-capital investment company (societe d'investissement a capital variable).
type SICAV struct{}

func (s *SICAV) Name() string         { return "SICAV" }
func (s *SICAV) Type() string         { return EntityType_SICAV }
func (s *SICAV) Jurisdiction() string { return "LU" }

func (s *SICAV) RequiredLicenses() []License {
	return []License{
		{Name: "CSSF Authorisation", Regulator: "CSSF", Reference: "Law of 17 December 2010 (UCITS) or Law of 13 February 2007 (SIF)", Description: "CSSF authorisation as UCITS Part I or SIF fund"},
		{Name: "Management Company Appointment", Regulator: "CSSF", Reference: "Law of 17 December 2010, Chapter 15", Description: "Appointment of authorised UCITS management company or self-managed fund"},
	}
}

func (s *SICAV) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{Name: "Annual Financial Statements", Frequency: "annual", Regulator: "CSSF", Reference: "Law of 17 December 2010, art. 146", Description: "Audited annual accounts and annual report"},
		{Name: "Semi-Annual Report", Frequency: "semi-annual", Regulator: "CSSF", Reference: "Law of 17 December 2010, art. 146", Description: "Semi-annual financial report"},
		{Name: "CSSF Reporting (S-Reporting)", Frequency: "quarterly", Regulator: "CSSF", Reference: "CSSF Circular 15/633", Description: "Quarterly statistical reporting to CSSF"},
	}
}

func (s *SICAV) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{MinNetCapital: 1_250_000, Currency: "EUR", CalculationRule: "Minimum share capital EUR 1,250,000 within 6 months of authorisation", Reference: "Law of 17 December 2010, art. 27"}
}

func (s *SICAV) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{Name: "Depositary Appointment", Category: "supervisory", Description: "Appoint a Luxembourg-established credit institution as depositary", Reference: "Law of 17 December 2010, art. 17"},
		{Name: "Risk Management", Category: "supervisory", Description: "Risk management process covering market, credit, liquidity, and operational risk", Reference: "CSSF Circular 11/512"},
		{Name: "NAV Calculation", Category: "technology", Description: "Regular NAV calculation and publication per prospectus frequency", Reference: "Law of 17 December 2010, art. 37"},
	}
}

// SICAR is a Luxembourg risk capital investment company (societe d'investissement en capital a risque).
type SICAR struct{}

func (s *SICAR) Name() string         { return "SICAR" }
func (s *SICAR) Type() string         { return EntityType_SICAR }
func (s *SICAR) Jurisdiction() string { return "LU" }

func (s *SICAR) RequiredLicenses() []License {
	return []License{
		{Name: "CSSF Authorisation", Regulator: "CSSF", Reference: "Law of 15 June 2004 on SICAR, art. 1-2", Description: "CSSF authorisation as SICAR for risk capital investments"},
	}
}

func (s *SICAR) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{Name: "Annual Report", Frequency: "annual", Regulator: "CSSF", Reference: "Law of 15 June 2004, art. 26", Description: "Audited annual accounts filed with CSSF"},
		{Name: "CSSF Reporting", Frequency: "quarterly", Regulator: "CSSF", Reference: "CSSF Circular 15/633", Description: "Statistical reporting to CSSF"},
	}
}

func (s *SICAR) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{MinNetCapital: 1_000_000, Currency: "EUR", CalculationRule: "Minimum share capital EUR 1,000,000 within 12 months of authorisation", Reference: "Law of 15 June 2004, art. 5"}
}

func (s *SICAR) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{Name: "Well-Informed Investors Only", Category: "supervisory", Description: "Restricted to institutional, professional, and well-informed investors (EUR 125k min or written declaration)", Reference: "Law of 15 June 2004, art. 2"},
		{Name: "Depositary Appointment", Category: "supervisory", Description: "Appoint a Luxembourg-established depositary", Reference: "Law of 15 June 2004, art. 12"},
	}
}

// RAIF is a Luxembourg Reserved Alternative Investment Fund (fonds d'investissement alternatif reserve).
type RAIF struct{}

func (r *RAIF) Name() string         { return "RAIF" }
func (r *RAIF) Type() string         { return EntityType_RAIF }
func (r *RAIF) Jurisdiction() string { return "LU" }

func (r *RAIF) RequiredLicenses() []License {
	return []License{
		{Name: "AIFM Appointment", Regulator: "CSSF", Reference: "Law of 23 July 2016, art. 2", Description: "Must be managed by an authorised AIFM (no product-level CSSF approval required)"},
	}
}

func (r *RAIF) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{Name: "Annual Report", Frequency: "annual", Regulator: "CSSF", Reference: "Law of 23 July 2016, art. 45", Description: "Audited annual report (AIFM reporting obligations apply)"},
		{Name: "AIFMD Annex IV Reporting", Frequency: "quarterly", Regulator: "CSSF", Reference: "AIFMD art. 24; CSSF FAQ on AIFMD reporting", Description: "AIFM-level reporting under AIFMD Annex IV"},
	}
}

func (r *RAIF) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{MinNetCapital: 1_250_000, Currency: "EUR", CalculationRule: "EUR 1,250,000 net assets within 12 months of establishment", Reference: "Law of 23 July 2016, art. 27"}
}

func (r *RAIF) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{Name: "Well-Informed Investors Only", Category: "supervisory", Description: "Restricted to well-informed investors (same definition as SICAR/SIF)", Reference: "Law of 23 July 2016, art. 3"},
		{Name: "Depositary Appointment", Category: "supervisory", Description: "Appoint a Luxembourg-established depositary", Reference: "Law of 23 July 2016, art. 15"},
	}
}

// --- EU-wide entity types ---

// AIFM is an Alternative Investment Fund Manager authorised under AIFMD.
type AIFM struct{}

func (a *AIFM) Name() string         { return "Alternative Investment Fund Manager" }
func (a *AIFM) Type() string         { return EntityType_AIFM }
func (a *AIFM) Jurisdiction() string { return "LU" } // Most common domicile; overridden per-entity in practice

func (a *AIFM) RequiredLicenses() []License {
	return []License{
		{Name: "AIFMD Authorisation", Regulator: "CSSF", Reference: "AIFMD (Directive 2011/61/EU), art. 6; Law of 12 July 2013, art. 4", Description: "Full AIFM authorisation under AIFMD"},
		{Name: "MiFID Top-Up", Regulator: "CSSF", Reference: "AIFMD art. 6(4); MiFID II art. 4", Description: "Optional MiFID II top-up licence for portfolio management and non-core services"},
	}
}

func (a *AIFM) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{Name: "AIFMD Annex IV Reporting", Frequency: "quarterly", Regulator: "CSSF", Reference: "AIFMD art. 24; Level 2 Regulation (EU) 231/2013", Description: "AIFMD Annex IV transparency reporting for each managed AIF"},
		{Name: "Annual Report per AIF", Frequency: "annual", Regulator: "CSSF", Reference: "AIFMD art. 22", Description: "Audited annual report for each managed AIF"},
	}
}

func (a *AIFM) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{MinNetCapital: 125_000, Currency: "EUR", CalculationRule: "EUR 125,000 initial capital + 0.02% of AUM above EUR 250M (capped at EUR 10M); OR EUR 300,000 for internally-managed AIF", Reference: "AIFMD art. 9"}
}

func (a *AIFM) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{Name: "Depositary Appointment", Category: "supervisory", Description: "Appoint a single depositary for each managed AIF", Reference: "AIFMD art. 21"},
		{Name: "Risk Management", Category: "supervisory", Description: "Functionally and hierarchically separate risk management from portfolio management", Reference: "AIFMD art. 15"},
		{Name: "Leverage Monitoring", Category: "reporting", Description: "Report leverage levels to NCA; comply with leverage limits if imposed", Reference: "AIFMD art. 25"},
		{Name: "Remuneration Policy", Category: "supervisory", Description: "Remuneration policies consistent with sound risk management", Reference: "AIFMD art. 13; Annex II"},
	}
}

// ManCo is a Luxembourg management company (societe de gestion) for UCITS/AIF.
type ManCo struct{}

func (m *ManCo) Name() string         { return "Management Company" }
func (m *ManCo) Type() string         { return EntityType_MANCOMAN }
func (m *ManCo) Jurisdiction() string { return "LU" }

func (m *ManCo) RequiredLicenses() []License {
	return []License{
		{Name: "Chapter 15 Authorisation", Regulator: "CSSF", Reference: "Law of 17 December 2010, Chapter 15", Description: "CSSF authorisation as UCITS management company"},
		{Name: "Chapter 16 Authorisation (Optional)", Regulator: "CSSF", Reference: "Law of 17 December 2010, Chapter 16", Description: "Optional: combined UCITS ManCo and AIFM authorisation"},
	}
}

func (m *ManCo) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{Name: "Annual Financial Statements", Frequency: "annual", Regulator: "CSSF", Reference: "Law of 17 December 2010, art. 105", Description: "Audited annual accounts of the ManCo itself"},
		{Name: "CSSF Reporting", Frequency: "quarterly", Regulator: "CSSF", Reference: "CSSF Circular 15/633", Description: "Quarterly reporting to CSSF on managed funds"},
	}
}

func (m *ManCo) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{MinNetCapital: 125_000, Currency: "EUR", CalculationRule: "EUR 125,000 initial capital + additional own funds equal to 0.02% of AUM above EUR 250M", Reference: "Law of 17 December 2010, art. 102"}
}

func (m *ManCo) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{Name: "Substance Requirements", Category: "personnel", Description: "Central administration and registered office in Luxembourg; at least 2 conducting officers resident in Luxembourg", Reference: "CSSF Circular 12/546"},
		{Name: "Delegation Oversight", Category: "supervisory", Description: "ManCo remains responsible for delegated functions; no letter-box entity", Reference: "Law of 17 December 2010, art. 110"},
	}
}

// CRR is a credit institution regulated under the EU Capital Requirements Regulation.
type CRR struct{}

func (c *CRR) Name() string         { return "Credit Institution (CRR)" }
func (c *CRR) Type() string         { return EntityType_CRR }
func (c *CRR) Jurisdiction() string { return "LU" }

func (c *CRR) RequiredLicenses() []License {
	return []License{
		{Name: "Banking Licence", Regulator: "ECB/CSSF", Reference: "CRD IV (Directive 2013/36/EU), art. 8; CRR (Regulation (EU) 575/2013)", Description: "ECB/NCA banking licence for deposit-taking and lending"},
	}
}

func (c *CRR) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{Name: "COREP/FINREP Reporting", Frequency: "quarterly", Regulator: "ECB/CSSF", Reference: "CRR art. 99; EBA ITS on supervisory reporting", Description: "Common Reporting (COREP) and Financial Reporting (FINREP)"},
		{Name: "Pillar 3 Disclosure", Frequency: "annual", Regulator: "ECB", Reference: "CRR Part Eight", Description: "Public disclosure of risk, capital, and remuneration"},
		{Name: "Resolution Reporting", Frequency: "annual", Regulator: "SRB", Reference: "BRRD (Directive 2014/59/EU)", Description: "Data submission for resolution planning (SRB)"},
	}
}

func (c *CRR) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{MinNetCapital: 5_000_000, Currency: "EUR", CalculationRule: "EUR 5M initial capital; CET1 ratio >= 4.5%, Tier 1 >= 6%, Total capital >= 8% of RWA + buffers (CCoB, SyRB, O-SII as applicable)", Reference: "CRR art. 92; CRD IV art. 129-133"}
}

func (c *CRR) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{Name: "Liquidity Coverage Ratio", Category: "supervisory", Description: "LCR >= 100%: sufficient high-quality liquid assets to cover 30-day net outflows", Reference: "CRR art. 412; Delegated Regulation (EU) 2015/61"},
		{Name: "Net Stable Funding Ratio", Category: "supervisory", Description: "NSFR >= 100%: stable funding to cover assets over 1 year", Reference: "CRR art. 428a-428az"},
		{Name: "Internal Governance", Category: "supervisory", Description: "Management body with sufficient knowledge, skills, and experience; separation of executive and supervisory functions", Reference: "CRD IV art. 88-91; EBA Guidelines on Internal Governance"},
	}
}

// Issuer is an SPV issuer of securities (e.g. securitisation vehicle, tokenised security issuer).
type Issuer struct{}

func (i *Issuer) Name() string         { return "Securities Issuer (SPV)" }
func (i *Issuer) Type() string         { return EntityType_ISSUER }
func (i *Issuer) Jurisdiction() string { return "LU" }

func (i *Issuer) RequiredLicenses() []License {
	return []License{
		{Name: "Prospectus Approval", Regulator: "CSSF", Reference: "Prospectus Regulation (EU) 2017/1129, art. 20", Description: "CSSF-approved prospectus for public offering or admission to trading"},
		{Name: "Securitisation Vehicle Registration", Regulator: "CSSF", Reference: "Law of 22 March 2004 on Securitisation", Description: "Registration as securitisation vehicle for structured issuance"},
	}
}

func (i *Issuer) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{Name: "Annual Financial Statements", Frequency: "annual", Regulator: "CSSF", Reference: "Transparency Directive (2004/109/EC), art. 4", Description: "Audited annual accounts for issuers with securities admitted to trading"},
		{Name: "Semi-Annual Report", Frequency: "semi-annual", Regulator: "CSSF", Reference: "Transparency Directive, art. 5", Description: "Half-yearly financial report"},
		{Name: "Major Holdings Notifications", Frequency: "event-driven", Regulator: "CSSF", Reference: "Transparency Directive, art. 9-16", Description: "Notifications of acquisition or disposal of major holdings"},
	}
}

func (i *Issuer) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{MinNetCapital: 0, Currency: "EUR", CalculationRule: "No minimum net capital for SPVs; adequacy depends on issuance structure and collateralisation", Reference: "Law of 22 March 2004 on Securitisation"}
}

func (i *Issuer) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{Name: "Paying Agent Appointment", Category: "supervisory", Description: "Appoint a paying agent for distributions to security holders", Reference: "Prospectus Regulation (EU) 2017/1129; LuxSE listing rules"},
		{Name: "Ongoing Disclosure", Category: "reporting", Description: "Inside information disclosure per MAR (Regulation (EU) 596/2014) if listed", Reference: "MAR art. 17"},
	}
}

// Custodian is a qualified custodian holding client assets.
type Custodian struct{}

func (c *Custodian) Name() string         { return "Qualified Custodian" }
func (c *Custodian) Type() string         { return EntityType_CUSTODIAN }
func (c *Custodian) Jurisdiction() string { return "LU" }

func (c *Custodian) RequiredLicenses() []License {
	return []License{
		{Name: "Banking Licence (Depositary)", Regulator: "CSSF", Reference: "Law of 17 December 2010, art. 17 (UCITS); AIFMD art. 21(3)", Description: "Credit institution licence required for UCITS/AIF depositary function"},
		{Name: "PSF Licence (Alternative)", Regulator: "CSSF", Reference: "Law of 5 April 1993, art. 28-1", Description: "Professional of the Financial Sector licence if not a credit institution"},
	}
}

func (c *Custodian) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{Name: "Depositary Report", Frequency: "annual", Regulator: "CSSF", Reference: "UCITS Directive art. 73; AIFMD art. 21(15)", Description: "Annual report to NCA on depositary duties and any irregularities"},
		{Name: "COREP/FINREP (if bank)", Frequency: "quarterly", Regulator: "ECB/CSSF", Reference: "CRR art. 99", Description: "Supervisory reporting if custodian is a credit institution"},
	}
}

func (c *Custodian) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{MinNetCapital: 5_000_000, Currency: "EUR", CalculationRule: "EUR 5M if credit institution; EUR 730,000 if PSF; plus client asset segregation", Reference: "CRR art. 12; Law of 5 April 1993, art. 23"}
}

func (c *Custodian) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{Name: "Asset Segregation", Category: "technology", Description: "Client assets held in segregated accounts; no commingling with own assets", Reference: "UCITS Directive art. 22; AIFMD art. 21(8)"},
		{Name: "Cash Flow Monitoring", Category: "supervisory", Description: "Monitor fund cash flows and reconcile with records", Reference: "UCITS Directive art. 22(3); AIFMD Delegated Regulation art. 86"},
		{Name: "Safekeeping Duties", Category: "books_and_records", Description: "Maintain records of all assets held in custody; regular reconciliation", Reference: "AIFMD Delegated Regulation art. 89-90"},
	}
}

// DLTFacility is a Swiss DLT Trading Facility licensed under the DLT Act 2021.
type DLTFacility struct{}

func (d *DLTFacility) Name() string         { return "DLT Trading Facility" }
func (d *DLTFacility) Type() string         { return EntityType_DLT_FACILITY }
func (d *DLTFacility) Jurisdiction() string { return "CH" }

func (d *DLTFacility) RequiredLicenses() []License {
	return []License{
		{Name: "FINMA DLT Trading Facility Licence", Regulator: "FINMA", Reference: "FMIA (FinfraG), art. 73a-73i (DLT Act 2021 amendments)", Description: "FINMA licence to operate a DLT trading facility for ledger-based securities"},
	}
}

func (d *DLTFacility) ReportingObligations() []ReportingObligation {
	return []ReportingObligation{
		{Name: "Annual FINMA Report", Frequency: "annual", Regulator: "FINMA", Reference: "FMIA art. 73i; FINMASA art. 29", Description: "Annual report to FINMA on operations, technology, and risk management"},
		{Name: "Incident Reporting", Frequency: "event-driven", Regulator: "FINMA", Reference: "FINMASA art. 29; FINMA Circular 2023/1", Description: "Immediate notification of material incidents, cyberattacks, or system failures"},
	}
}

func (d *DLTFacility) CapitalRequirements() *CapitalRequirement {
	return &CapitalRequirement{MinNetCapital: 1_000_000, Currency: "CHF", CalculationRule: "CHF 1M minimum capital; additional capital based on DLT-specific risk assessment", Reference: "FMIA art. 73d; FINMA Circular on DLT facilities"}
}

func (d *DLTFacility) OperationalRequirements() []OperationalRequirement {
	return []OperationalRequirement{
		{Name: "FINMA-Registered Custodian", Category: "supervisory", Description: "Participant assets must be held with a FINMA-registered custodian per CO art. 973d-i", Reference: "CO art. 973d-973i; FMIA art. 73f"},
		{Name: "Participant Protection", Category: "supervisory", Description: "Rules ensuring participant protection: admission criteria, fair access, conflict management", Reference: "FMIA art. 73e"},
		{Name: "Technology Governance", Category: "technology", Description: "Robust IT governance, cyber risk management, and operational resilience for DLT infrastructure", Reference: "FINMA Circular 2023/1 on Operational Risks"},
	}
}
