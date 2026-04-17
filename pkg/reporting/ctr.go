package reporting

import "time"

// CTRData is a Currency Transaction Report per FinCEN (31 CFR 1010.311).
// Auto-generated when a person's daily aggregate cash transactions exceed $10,000.
type CTRData struct {
	Person      CTRPerson      `json:"person"`
	Transactions []CTRTxDetail `json:"transactions"`
	Institution CTRInstitution `json:"financial_institution"`
	Date        time.Time      `json:"date"`
	TotalCashIn  float64       `json:"total_cash_in"`
	TotalCashOut float64       `json:"total_cash_out"`
}

// CTRPerson is the individual conducting or on whose behalf the transactions were made.
type CTRPerson struct {
	Name      string `json:"name"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	SSN       string `json:"-"`                   // never serialized
	SSNLast4  string `json:"ssn_last4,omitempty"`
	DOB       string `json:"dob,omitempty"`
	Address   string `json:"address,omitempty"`
	City      string `json:"city,omitempty"`
	State     string `json:"state,omitempty"`
	ZipCode   string `json:"zip_code,omitempty"`
	Country   string `json:"country,omitempty"`
	IDType    string `json:"id_type,omitempty"`     // drivers_license, passport
	IDNumber  string `json:"-"`                     // never serialized
	IDState   string `json:"id_state,omitempty"`
	AccountID string `json:"account_id,omitempty"`
	Occupation string `json:"occupation,omitempty"`
}

// CTRTxDetail is a single cash transaction within the CTR.
type CTRTxDetail struct {
	Amount    float64   `json:"amount"`
	Direction string    `json:"direction"` // cash_in, cash_out
	Type      string    `json:"type"`      // deposit, withdrawal, exchange, payment
	Timestamp time.Time `json:"timestamp"`
	TxID      string    `json:"tx_id,omitempty"`
}

// CTRInstitution identifies the filing financial institution.
type CTRInstitution struct {
	Name    string `json:"name"`
	RSSD    string `json:"rssd,omitempty"`
	EIN     string `json:"-"` // never serialized
	Address string `json:"address,omitempty"`
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
	ZipCode string `json:"zip_code,omitempty"`
}

// CTRThreshold is the FinCEN threshold for mandatory CTR filing.
const CTRThreshold = 10000.0

// BuildCTR constructs a CTR from a user and their transactions for a given date.
// Only call this when daily aggregate exceeds CTRThreshold.
func BuildCTR(user Entity, transactions []Transaction, date time.Time) *CTRData {
	ctr := &CTRData{
		Person: CTRPerson{
			Name:      user.Name,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			SSNLast4:  user.SSNLast4,
			DOB:       user.DOB,
			Address:   user.Address,
			City:      user.City,
			State:     user.State,
			ZipCode:   user.ZipCode,
			Country:   user.Country,
			AccountID: user.AccountID,
		},
		Date: date,
	}

	for _, tx := range transactions {
		if tx.Date.Year() != date.Year() || tx.Date.YearDay() != date.YearDay() {
			continue
		}

		direction := "cash_in"
		if tx.Type == "withdrawal" || tx.Type == "payout" {
			direction = "cash_out"
		}

		detail := CTRTxDetail{
			Amount:    tx.Amount,
			Direction: direction,
			Type:      tx.Type,
			Timestamp: tx.Date,
			TxID:      tx.ID,
		}
		ctr.Transactions = append(ctr.Transactions, detail)

		if direction == "cash_in" {
			ctr.TotalCashIn += tx.Amount
		} else {
			ctr.TotalCashOut += tx.Amount
		}
	}

	return ctr
}

// ValidateCTR returns validation errors for a CTR before filing.
func ValidateCTR(ctr *CTRData) []string {
	var errs []string

	if ctr.Person.Name == "" && (ctr.Person.FirstName == "" || ctr.Person.LastName == "") {
		errs = append(errs, "person name is required")
	}
	if ctr.Date.IsZero() {
		errs = append(errs, "date is required")
	}
	if len(ctr.Transactions) == 0 {
		errs = append(errs, "at least one transaction is required")
	}

	total := ctr.TotalCashIn + ctr.TotalCashOut
	if total < CTRThreshold {
		errs = append(errs, "aggregate total below $10,000 CTR threshold")
	}

	return errs
}

// RequiresCTR returns true if the daily aggregate for an account exceeds the threshold.
func RequiresCTR(dailyAggregate float64) bool {
	return dailyAggregate > CTRThreshold
}
