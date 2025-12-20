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

// ATS is an Alternative Trading System registered under SEC Regulation ATS.
type ATS struct{}

func (a *ATS) Name() string         { return "Alternative Trading System" }
func (a *ATS) Type() string         { return "ats" }
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
func (bd *BrokerDealer) Type() string         { return "broker_dealer" }
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
func (ta *TransferAgent) Type() string         { return "transfer_agent" }
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
func (m *MSB) Type() string         { return "msb" }
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
	case "ats":
		return &ATS{}
	case "broker_dealer":
		return &BrokerDealer{}
	case "transfer_agent":
		return &TransferAgent{}
	case "msb":
		return &MSB{}
	default:
		return nil
	}
}

// AllEntities returns all regulated entity types.
func AllEntities() []RegulatedEntity {
	return []RegulatedEntity{
		&ATS{},
		&BrokerDealer{},
		&TransferAgent{},
		&MSB{},
	}
}
