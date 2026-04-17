package reporting

import "time"

// FormDData is a SEC Form D filing for Regulation D exempt offerings.
// Maps to SEC EDGAR XML schema.
type FormDData struct {
	IsAmendment    bool           `json:"is_amendment"`
	Issuer         FormDIssuer    `json:"issuer"`
	Offering       FormDOffering  `json:"offering"`
	SalesComp      []FormDSalesPerson `json:"sales_compensation,omitempty"`
	InvestorCounts map[string]int `json:"investor_counts_by_state,omitempty"` // state code -> count
	SignedBy       string         `json:"signed_by"`
	SignedAt       time.Time      `json:"signed_at"`
}

// FormDIssuer describes the entity making the offering.
type FormDIssuer struct {
	Name               string `json:"name"`
	CIK                string `json:"cik,omitempty"`
	EntityType         string `json:"entity_type"` // corporation, llc, lp, trust
	StateOfIncorporation string `json:"state_of_incorporation"`
	CountryOfIncorporation string `json:"country_of_incorporation"`
	YearOfIncorporation string `json:"year_of_incorporation,omitempty"`
	Address            string `json:"address"`
	City               string `json:"city"`
	State              string `json:"state"`
	ZipCode            string `json:"zip_code"`
	Phone              string `json:"phone"`
	IndustryGroup      string `json:"industry_group"`
	IssuerSize         string `json:"issuer_size,omitempty"` // revenue range
	EIN                string `json:"-"` // never serialized
}

// FormDOffering describes the securities being offered.
type FormDOffering struct {
	SecurityType      string    `json:"security_type"`    // equity, debt, pooled_interest, other
	ExemptionClaimed  []string  `json:"exemption_claimed"` // ["rule_506_b"], ["rule_506_c"], etc.
	TotalOfferingSize float64   `json:"total_offering_size"`
	TotalSold         float64   `json:"total_sold"`
	TotalRemaining    float64   `json:"total_remaining"`
	MinInvestment     float64   `json:"min_investment,omitempty"`
	HasNonAccredited  bool      `json:"has_non_accredited"`
	NonAccreditedCount int      `json:"non_accredited_count,omitempty"`
	TotalInvestors    int       `json:"total_investors"`
	FirstSaleDate     time.Time `json:"first_sale_date"`
	DurationOrIndef   string    `json:"duration_or_indef,omitempty"` // "indefinite" or months
}

// FormDSalesPerson is a sales compensation recipient.
type FormDSalesPerson struct {
	Name       string  `json:"name"`
	CRDNumber  string  `json:"crd_number,omitempty"`
	BrokerDealer string `json:"broker_dealer,omitempty"`
	State      string  `json:"state"`
	Commission float64 `json:"commission"`
	FinderFee  float64 `json:"finder_fee"`
}

// BuildFormD constructs a Form D from an offering, issuer entity, and per-state sales.
func BuildFormD(offering Offering, issuer Entity, salesByState map[string]float64) *FormDData {
	fd := &FormDData{
		IsAmendment: offering.IsAmendment,
		Issuer: FormDIssuer{
			Name:                    offering.IssuerName,
			CIK:                     offering.IssuerCIK,
			EntityType:              offering.EntityType,
			StateOfIncorporation:    offering.IssuerState,
			CountryOfIncorporation:  offering.IssuerCountry,
			IndustryGroup:           offering.IndustryGroup,
			IssuerSize:              offering.RevenueRange,
		},
		Offering: FormDOffering{
			SecurityType:      offering.SecurityType,
			TotalOfferingSize: offering.TotalOfferingSize,
			TotalSold:         offering.TotalSold,
			TotalRemaining:    offering.TotalRemaining,
			MinInvestment:     offering.MinInvestment,
			FirstSaleDate:     offering.FirstSaleDate,
		},
	}

	// Map exemption type to SEC rule references.
	switch offering.ExemptionType {
	case "reg_d_506b":
		fd.Offering.ExemptionClaimed = []string{"rule_506_b"}
	case "reg_d_506c":
		fd.Offering.ExemptionClaimed = []string{"rule_506_c"}
	case "reg_a_tier1":
		fd.Offering.ExemptionClaimed = []string{"regulation_a_tier_1"}
	case "reg_a_tier2":
		fd.Offering.ExemptionClaimed = []string{"regulation_a_tier_2"}
	default:
		fd.Offering.ExemptionClaimed = []string{offering.ExemptionType}
	}

	if len(salesByState) > 0 {
		fd.InvestorCounts = make(map[string]int, len(salesByState))
		for state := range salesByState {
			fd.InvestorCounts[state] = 1 // default 1 per state; caller refines
		}
	}

	return fd
}

// ValidateFormD returns validation errors for a Form D before filing.
func ValidateFormD(fd *FormDData) []string {
	var errs []string

	if fd.Issuer.Name == "" {
		errs = append(errs, "issuer name is required")
	}
	if fd.Issuer.EntityType == "" {
		errs = append(errs, "issuer entity_type is required")
	}
	if fd.Issuer.StateOfIncorporation == "" {
		errs = append(errs, "issuer state_of_incorporation is required")
	}
	if len(fd.Offering.ExemptionClaimed) == 0 {
		errs = append(errs, "at least one exemption_claimed is required")
	}
	if fd.Offering.SecurityType == "" {
		errs = append(errs, "offering security_type is required")
	}
	if fd.Offering.TotalOfferingSize <= 0 {
		errs = append(errs, "offering total_offering_size must be positive")
	}
	if fd.Offering.FirstSaleDate.IsZero() {
		errs = append(errs, "offering first_sale_date is required")
	}

	// 506(b) allows up to 35 non-accredited investors.
	for _, ex := range fd.Offering.ExemptionClaimed {
		if ex == "rule_506_b" && fd.Offering.NonAccreditedCount > 35 {
			errs = append(errs, "Rule 506(b) allows at most 35 non-accredited investors")
		}
		if ex == "rule_506_c" && fd.Offering.HasNonAccredited {
			errs = append(errs, "Rule 506(c) requires all investors to be accredited")
		}
	}

	return errs
}
