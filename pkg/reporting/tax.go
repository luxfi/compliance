package reporting

import (
	"fmt"
	"sort"
	"time"
)

// Tax1099Data consolidates 1099-B, 1099-DIV, and 1099-INT for a tax year.
type Tax1099Data struct {
	Year         int             `json:"year"`
	Recipient    Tax1099Person   `json:"recipient"`
	Payer        Tax1099Payer    `json:"payer"`
	Proceeds     []Tax1099B      `json:"proceeds,omitempty"`      // 1099-B
	Dividends    []Tax1099DIV    `json:"dividends,omitempty"`     // 1099-DIV
	Interest     []Tax1099INT    `json:"interest,omitempty"`      // 1099-INT
	Summary      Tax1099Summary  `json:"summary"`
}

// Tax1099Person is the recipient.
type Tax1099Person struct {
	Name     string `json:"name"`
	SSN      string `json:"-"`                   // never serialized
	SSNLast4 string `json:"ssn_last4,omitempty"`
	Address  string `json:"address,omitempty"`
	City     string `json:"city,omitempty"`
	State    string `json:"state,omitempty"`
	ZipCode  string `json:"zip_code,omitempty"`
}

// Tax1099Payer is the issuing institution.
type Tax1099Payer struct {
	Name    string `json:"name"`
	EIN     string `json:"-"` // never serialized
	Address string `json:"address,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

// Tax1099B is a single 1099-B proceeds entry (one per disposition).
type Tax1099B struct {
	Description   string    `json:"description"`
	Symbol        string    `json:"symbol"`
	Quantity      float64   `json:"quantity"`
	AcquiredDate  time.Time `json:"acquired_date"`
	DisposedDate  time.Time `json:"disposed_date"`
	Proceeds      float64   `json:"proceeds"`
	CostBasis     float64   `json:"cost_basis"`
	GainLoss      float64   `json:"gain_loss"`
	HoldingPeriod string    `json:"holding_period"` // short_term, long_term
	WashSale      bool      `json:"wash_sale"`
	WashAmount    float64   `json:"wash_amount,omitempty"`
	AdjustedBasis float64   `json:"adjusted_basis"` // cost_basis + wash_amount
	BasisReported bool      `json:"basis_reported"` // reported to IRS
}

// Tax1099DIV is a single 1099-DIV dividend entry.
type Tax1099DIV struct {
	Symbol             string  `json:"symbol"`
	OrdinaryDividends  float64 `json:"ordinary_dividends"`
	QualifiedDividends float64 `json:"qualified_dividends"`
	FederalTaxWH       float64 `json:"federal_tax_withheld"`
}

// Tax1099INT is a single 1099-INT interest entry.
type Tax1099INT struct {
	Source       string  `json:"source"`
	Interest     float64 `json:"interest_income"`
	FederalTaxWH float64 `json:"federal_tax_withheld"`
}

// Tax1099Summary aggregates totals.
type Tax1099Summary struct {
	TotalProceeds          float64 `json:"total_proceeds"`
	TotalCostBasis         float64 `json:"total_cost_basis"`
	ShortTermGainLoss      float64 `json:"short_term_gain_loss"`
	LongTermGainLoss       float64 `json:"long_term_gain_loss"`
	TotalWashSaleAdj       float64 `json:"total_wash_sale_adj"`
	TotalOrdinaryDividends float64 `json:"total_ordinary_dividends"`
	TotalQualifiedDividends float64 `json:"total_qualified_dividends"`
	TotalInterest          float64 `json:"total_interest"`
	TotalFederalTaxWH      float64 `json:"total_federal_tax_withheld"`
}

// Build1099 constructs 1099 tax data from trades and dividends.
// Performs wash sale detection and cost basis computation.
func Build1099(user Entity, trades []Trade, dividends []Dividend, year int) *Tax1099Data {
	td := &Tax1099Data{
		Year: year,
		Recipient: Tax1099Person{
			Name:    user.Name,
			SSNLast4: user.SSNLast4,
			Address: user.Address,
			City:    user.City,
			State:   user.State,
			ZipCode: user.ZipCode,
		},
	}

	// Sort trades by disposed date for wash sale detection.
	sorted := make([]Trade, len(trades))
	copy(sorted, trades)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].DisposedDate.Before(sorted[j].DisposedDate)
	})

	// Detect wash sales: if a substantially identical security is repurchased
	// within 30 days before or after a loss sale, the loss is disallowed and
	// added to the cost basis of the replacement.
	washAdjusted := detectWashSales(sorted)

	for _, t := range washAdjusted {
		if t.DisposedDate.Year() != year {
			continue
		}

		gainLoss := t.Proceeds - t.CostBasis
		adjustedBasis := t.CostBasis + t.WashAmount

		entry := Tax1099B{
			Description:   t.Symbol,
			Symbol:        t.Symbol,
			Quantity:      t.Qty,
			AcquiredDate:  t.AcquiredDate,
			DisposedDate:  t.DisposedDate,
			Proceeds:      t.Proceeds,
			CostBasis:     t.CostBasis,
			GainLoss:      gainLoss,
			HoldingPeriod: t.HoldingPeriod,
			WashSale:      t.WashSale,
			WashAmount:    t.WashAmount,
			AdjustedBasis: adjustedBasis,
			BasisReported: true,
		}
		td.Proceeds = append(td.Proceeds, entry)

		td.Summary.TotalProceeds += t.Proceeds
		td.Summary.TotalCostBasis += t.CostBasis

		if t.HoldingPeriod == "short" {
			td.Summary.ShortTermGainLoss += gainLoss
		} else {
			td.Summary.LongTermGainLoss += gainLoss
		}

		if t.WashSale {
			td.Summary.TotalWashSaleAdj += t.WashAmount
		}
	}

	// Aggregate dividends by symbol.
	divMap := map[string]*Tax1099DIV{}
	for _, d := range dividends {
		if d.Date.Year() != year {
			continue
		}
		entry, ok := divMap[d.Symbol]
		if !ok {
			entry = &Tax1099DIV{Symbol: d.Symbol}
			divMap[d.Symbol] = entry
		}
		entry.OrdinaryDividends += d.Amount
		if d.Qualified {
			entry.QualifiedDividends += d.Amount
		}
		entry.FederalTaxWH += d.FederalTaxWH

		td.Summary.TotalOrdinaryDividends += d.Amount
		if d.Qualified {
			td.Summary.TotalQualifiedDividends += d.Amount
		}
		td.Summary.TotalFederalTaxWH += d.FederalTaxWH
	}

	for _, v := range divMap {
		td.Dividends = append(td.Dividends, *v)
	}

	return td
}

// detectWashSales marks trades that trigger the IRS 30-day wash sale rule.
// A loss sale is a wash sale if a substantially identical security is bought
// within 30 days before or after the sale.
func detectWashSales(trades []Trade) []Trade {
	result := make([]Trade, len(trades))
	copy(result, trades)

	thirtyDays := 30 * 24 * time.Hour

	for i := range result {
		loss := result[i].Proceeds - result[i].CostBasis
		if loss >= 0 {
			continue // only loss sales can be wash sales
		}

		for j := range result {
			if i == j {
				continue
			}
			if result[i].Symbol != result[j].Symbol {
				continue
			}
			// Check if trade j is a buy within the 30-day window.
			if result[j].Side != "buy" {
				continue
			}

			diff := result[j].AcquiredDate.Sub(result[i].DisposedDate)
			if diff >= -thirtyDays && diff <= thirtyDays {
				result[i].WashSale = true
				result[i].WashAmount = -loss // disallowed loss added to basis
				break
			}
		}
	}

	return result
}

// Validate1099 returns validation errors for tax report data.
func Validate1099(td *Tax1099Data) []string {
	var errs []string

	if td.Year < 2020 || td.Year > 2100 {
		errs = append(errs, "year must be between 2020 and 2100")
	}
	if td.Recipient.Name == "" {
		errs = append(errs, "recipient name is required")
	}
	if len(td.Proceeds) == 0 && len(td.Dividends) == 0 && len(td.Interest) == 0 {
		errs = append(errs, "at least one proceeds, dividend, or interest entry required")
	}

	for i, p := range td.Proceeds {
		if p.DisposedDate.Before(p.AcquiredDate) {
			errs = append(errs, fmt.Sprintf("proceeds[%d]: disposed_date before acquired_date", i))
		}
	}

	return errs
}
