// Package reporting provides regulatory report builders for FinCEN (SAR, CTR),
// SEC (Form D), FINRA (ATS-N, OATS, CAT), EU (MiFID II), and tax (1099).
// Each builder produces a typed struct; a separate filing service handles
// XML generation and e-filing submission.
package reporting

import (
	"encoding/json"
	"time"
)

// ReportType identifies a regulatory report format.
type ReportType string

const (
	ReportSAR     ReportType = "sar"      // Suspicious Activity Report (FinCEN)
	ReportCTR     ReportType = "ctr"      // Currency Transaction Report (FinCEN)
	ReportATSN    ReportType = "ats_n"    // FINRA ATS-N quarterly
	ReportOATS    ReportType = "oats"     // Order Audit Trail (FINRA)
	ReportCAT     ReportType = "cat"      // Consolidated Audit Trail
	ReportFormD   ReportType = "form_d"   // SEC Form D (Reg D notice)
	ReportBlueSky ReportType = "blue_sky" // State notice filings
	ReportMiFID   ReportType = "mifid_ii" // EU transaction report
	Report1099    ReportType = "1099"     // Tax reporting
	ReportFATCA   ReportType = "fatca"    // Foreign Account Tax Compliance
	ReportCRS     ReportType = "crs"      // Common Reporting Standard (OECD)
)

// ReportStatus tracks the lifecycle of a filing.
type ReportStatus string

const (
	StatusDraft    ReportStatus = "draft"
	StatusReview   ReportStatus = "review"
	StatusFiled    ReportStatus = "filed"
	StatusAccepted ReportStatus = "accepted"
	StatusRejected ReportStatus = "rejected"
)

// Report is the envelope for any regulatory filing.
type Report struct {
	ID           string          `json:"id"`
	Type         ReportType      `json:"type"`
	OrgID        string          `json:"org_id"`
	Jurisdiction string          `json:"jurisdiction"`
	Period       string          `json:"period"` // "2026-Q1", "2026", "2026-04-17"
	Status       ReportStatus    `json:"status"`
	Data         json.RawMessage `json:"data"`
	FiledAt      *time.Time      `json:"filed_at,omitempty"`
	FiledBy      string          `json:"filed_by,omitempty"`
	Reference    string          `json:"reference,omitempty"` // filing confirmation number
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// Entity is a normalized person or business for report population.
type Entity struct {
	Name        string `json:"name"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	DOB         string `json:"dob,omitempty"`          // YYYY-MM-DD
	SSN         string `json:"-"`                      // never serialized
	SSNLast4    string `json:"ssn_last4,omitempty"`
	TIN         string `json:"-"`                      // never serialized
	EIN         string `json:"-"`                      // never serialized
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	ZipCode     string `json:"zip_code,omitempty"`
	Country     string `json:"country,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
	Institution string `json:"institution,omitempty"` // VASP or bank name
	LEI         string `json:"lei,omitempty"`         // Legal Entity Identifier
}

// Alert is a simplified alert reference for SAR population.
type Alert struct {
	ID          string    `json:"id"`
	RuleID      string    `json:"rule_id"`
	RuleName    string    `json:"rule_name"`
	Severity    string    `json:"severity"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Transaction is a financial transaction for report population.
type Transaction struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Asset     string    `json:"asset,omitempty"`
	Side      string    `json:"side,omitempty"` // buy, sell
	Date      time.Time `json:"date"`
	AccountID string    `json:"account_id,omitempty"`
}

// Order is an exchange order for FINRA reporting.
type Order struct {
	ID            string    `json:"id"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"` // buy, sell
	Type          string    `json:"type"` // market, limit, stop
	Qty           float64   `json:"qty"`
	Price         float64   `json:"price"`
	FilledQty     float64   `json:"filled_qty"`
	FilledPrice   float64   `json:"filled_price"`
	Status        string    `json:"status"` // new, partial, filled, cancelled
	SubscriberID  string    `json:"subscriber_id,omitempty"`
	ReceivedAt    time.Time `json:"received_at"`
	ExecutedAt    time.Time `json:"executed_at,omitempty"`
}

// Offering represents a securities offering for Form D.
type Offering struct {
	IssuerName        string    `json:"issuer_name"`
	IssuerCIK         string    `json:"issuer_cik,omitempty"`
	IssuerState       string    `json:"issuer_state"`
	IssuerCountry     string    `json:"issuer_country"`
	EntityType        string    `json:"entity_type"` // corporation, llc, lp
	IndustryGroup     string    `json:"industry_group"`
	RevenueRange      string    `json:"revenue_range,omitempty"`
	ExemptionType     string    `json:"exemption_type"` // reg_d_506b, reg_d_506c, reg_a_tier1, reg_a_tier2
	SecurityType      string    `json:"security_type"`  // equity, debt, pooled_interest
	IsAmendment       bool      `json:"is_amendment"`
	FirstSaleDate     time.Time `json:"first_sale_date"`
	TotalOfferingSize float64   `json:"total_offering_size"`
	TotalSold         float64   `json:"total_sold"`
	TotalRemaining    float64   `json:"total_remaining"`
	MinInvestment     float64   `json:"min_investment"`
}

// Trade represents a completed trade for tax reporting.
type Trade struct {
	ID            string    `json:"id"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"` // buy, sell
	Qty           float64   `json:"qty"`
	Proceeds      float64   `json:"proceeds"`
	CostBasis     float64   `json:"cost_basis"`
	AcquiredDate  time.Time `json:"acquired_date"`
	DisposedDate  time.Time `json:"disposed_date"`
	HoldingPeriod string    `json:"holding_period"` // short, long
	WashSale      bool      `json:"wash_sale"`
	WashAmount    float64   `json:"wash_amount"`
}

// Dividend represents a dividend payment for tax reporting.
type Dividend struct {
	Symbol       string    `json:"symbol"`
	Amount       float64   `json:"amount"`
	Qualified    bool      `json:"qualified"`
	Date         time.Time `json:"date"`
	FederalTaxWH float64   `json:"federal_tax_withheld"`
}
