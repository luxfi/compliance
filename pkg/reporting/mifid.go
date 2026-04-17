package reporting

import "time"

// MiFIDData is a MiFID II transaction report per RTS 25 / ESMA XML schema.
// Required for all EU-regulated transactions within 24 hours.
type MiFIDData struct {
	ReportID          string    `json:"report_id"`
	ExecutingEntity   MiFIDEntity `json:"executing_entity"`
	SubmittingEntity  MiFIDEntity `json:"submitting_entity"`
	Buyer             MiFIDParty  `json:"buyer"`
	Seller            MiFIDParty  `json:"seller"`
	Instrument        MiFIDInstrument `json:"instrument"`
	Execution         MiFIDExecution  `json:"execution"`
	TransactionRefNum string    `json:"transaction_ref_num"`
	Venue             string    `json:"venue"` // MIC code
	TradingDateTime   time.Time `json:"trading_date_time"`
}

// MiFIDEntity is a firm involved in the report.
type MiFIDEntity struct {
	LEI     string `json:"lei"`
	Name    string `json:"name,omitempty"`
	Country string `json:"country,omitempty"` // ISO 3166-1 alpha-2
}

// MiFIDParty is a buyer or seller.
type MiFIDParty struct {
	LEI            string `json:"lei,omitempty"`
	NationalID     string `json:"-"`                       // never serialized
	NationalIDType string `json:"national_id_type,omitempty"` // passport, concat, etc.
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	DOB            string `json:"dob,omitempty"`
	Country        string `json:"country,omitempty"`
	DecisionMaker  string `json:"decision_maker,omitempty"` // algorithm ID or person
}

// MiFIDInstrument is the traded instrument.
type MiFIDInstrument struct {
	ISIN           string `json:"isin"`
	FullName       string `json:"full_name,omitempty"`
	Classification string `json:"classification,omitempty"` // CFI code
	Currency       string `json:"currency"`
	Underlying     string `json:"underlying,omitempty"` // underlying ISIN for derivatives
}

// MiFIDExecution contains price, quantity, and execution details.
type MiFIDExecution struct {
	Price          float64 `json:"price"`
	PriceCurrency  string  `json:"price_currency"`
	Quantity       float64 `json:"quantity"`
	QuantityType   string  `json:"quantity_type"` // units, nominal, monetary_value
	Side           string  `json:"side"`          // buy, sell
	WaiveFlag      string  `json:"waive_flag,omitempty"`
	ShortSelling   string  `json:"short_selling,omitempty"` // no_short, short_no_exempt, short_exempt
}

// BuildMiFID constructs a MiFID II transaction report from a transaction and entity.
func BuildMiFID(tx Transaction, entity Entity) *MiFIDData {
	m := &MiFIDData{
		TradingDateTime:   tx.Date,
		TransactionRefNum: tx.ID,
		Instrument: MiFIDInstrument{
			Currency: tx.Currency,
		},
		Execution: MiFIDExecution{
			Price:         tx.Amount,
			PriceCurrency: tx.Currency,
			Quantity:      1, // default; caller overrides for multi-unit
			QuantityType:  "units",
		},
	}

	// Assign buyer/seller based on side.
	party := MiFIDParty{
		LEI:       entity.LEI,
		FirstName: entity.FirstName,
		LastName:  entity.LastName,
		DOB:       entity.DOB,
		Country:   entity.Country,
	}

	if tx.Side == "buy" {
		m.Buyer = party
		m.Execution.Side = "buy"
	} else {
		m.Seller = party
		m.Execution.Side = "sell"
	}

	// Populate asset identifier if available.
	if tx.Asset != "" {
		m.Instrument.FullName = tx.Asset
	}

	return m
}

// ValidateMiFID returns validation errors for a MiFID II report.
func ValidateMiFID(m *MiFIDData) []string {
	var errs []string

	if m.ExecutingEntity.LEI == "" {
		errs = append(errs, "executing_entity LEI is required")
	}
	if m.Instrument.Currency == "" {
		errs = append(errs, "instrument currency is required")
	}
	if m.TradingDateTime.IsZero() {
		errs = append(errs, "trading_date_time is required")
	}
	if m.Execution.Price <= 0 {
		errs = append(errs, "execution price must be positive")
	}
	if m.Execution.Side != "buy" && m.Execution.Side != "sell" {
		errs = append(errs, "execution side must be 'buy' or 'sell'")
	}

	// MiFID II requires LEI for legal entities.
	if m.Buyer.LEI == "" && m.Buyer.FirstName == "" {
		errs = append(errs, "buyer requires either LEI (entity) or name (natural person)")
	}
	if m.Seller.LEI == "" && m.Seller.FirstName == "" {
		errs = append(errs, "seller requires either LEI (entity) or name (natural person)")
	}

	return errs
}
