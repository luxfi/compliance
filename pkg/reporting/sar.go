package reporting

import (
	"fmt"
	"strings"
	"time"
)

// SARData is a Suspicious Activity Report per FinCEN BSA E-Filing (31 CFR 1020.320).
type SARData struct {
	FilingType          string         `json:"filing_type"` // initial, correct, supplement
	Subject             SARSubject     `json:"subject"`
	Institution         SARInstitution `json:"financial_institution"`
	SuspiciousActivity  SARActivity    `json:"suspicious_activity"`
	Narrative           string         `json:"narrative"`
	FilingInstitutionID string         `json:"filing_institution_id,omitempty"`
}

// SARSubject is the person or entity under suspicion.
type SARSubject struct {
	Name        string `json:"name"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	SSN         string `json:"-"`                      // never serialized
	SSNLast4    string `json:"ssn_last4,omitempty"`
	DOB         string `json:"dob,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	ZipCode     string `json:"zip_code,omitempty"`
	Country     string `json:"country,omitempty"`
	Phone       string `json:"phone,omitempty"`
	AccountIDs  []string `json:"account_ids,omitempty"`
	Occupation  string `json:"occupation,omitempty"`
	EntityType  string `json:"entity_type,omitempty"` // individual, business
	Relationship string `json:"relationship,omitempty"` // customer, former_customer, employee
}

// SARInstitution identifies the filing financial institution.
type SARInstitution struct {
	Name     string `json:"name"`
	RSSD     string `json:"rssd,omitempty"`     // RSSD number
	EIN      string `json:"-"`                  // never serialized
	Address  string `json:"address,omitempty"`
	City     string `json:"city,omitempty"`
	State    string `json:"state,omitempty"`
	ZipCode  string `json:"zip_code,omitempty"`
}

// SARActivity describes the suspicious conduct.
type SARActivity struct {
	DateStart       time.Time         `json:"date_start"`
	DateEnd         time.Time         `json:"date_end"`
	TotalAmount     float64           `json:"total_amount"`
	Instruments     []string          `json:"instruments"`      // cash, wire, check, crypto, etc.
	Categories      []SARCategory     `json:"categories"`
	TransactionIDs  []string          `json:"transaction_ids,omitempty"`
}

// SARCategory is a FinCEN suspicious activity category.
type SARCategory struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// Standard SAR activity categories per FinCEN.
var SARCategories = map[string]string{
	"a": "Structuring/smurfing",
	"b": "Terrorist financing",
	"c": "Fraud against financial institution",
	"d": "Mortgage fraud",
	"e": "Identity theft",
	"f": "Money laundering",
	"g": "Bribery/corruption",
	"h": "Computer intrusion",
	"i": "Counterfeit instruments",
	"j": "Embezzlement",
	"k": "Insider trading",
	"l": "Market manipulation",
	"m": "Suspicious use of accounts",
	"n": "Wire transfer fraud",
	"o": "Unknown/other",
}

// BuildSAR constructs a SAR from a subject entity, triggering alerts, and an
// analyst-written narrative. The narrative is critical — FinCEN requires it
// to explain the why, not just the what.
func BuildSAR(subject Entity, alerts []Alert, narrative string) *SARData {
	sar := &SARData{
		FilingType: "initial",
		Subject: SARSubject{
			Name:      subject.Name,
			FirstName: subject.FirstName,
			LastName:  subject.LastName,
			SSNLast4:  subject.SSNLast4,
			DOB:       subject.DOB,
			Address:   subject.Address,
			City:      subject.City,
			State:     subject.State,
			ZipCode:   subject.ZipCode,
			Country:   subject.Country,
			Phone:     subject.Phone,
		},
		Narrative: narrative,
	}

	if subject.AccountID != "" {
		sar.Subject.AccountIDs = []string{subject.AccountID}
	}

	// Build activity from alerts.
	var totalAmount float64
	var earliest, latest time.Time
	instruments := map[string]bool{}
	txIDs := map[string]bool{}

	for _, a := range alerts {
		totalAmount += a.Amount
		if earliest.IsZero() || a.CreatedAt.Before(earliest) {
			earliest = a.CreatedAt
		}
		if a.CreatedAt.After(latest) {
			latest = a.CreatedAt
		}
	}

	// Default to 90-day window if only one alert.
	if earliest.Equal(latest) {
		earliest = latest.Add(-90 * 24 * time.Hour)
	}

	instrSlice := make([]string, 0, len(instruments))
	for k := range instruments {
		instrSlice = append(instrSlice, k)
	}

	txSlice := make([]string, 0, len(txIDs))
	for k := range txIDs {
		txSlice = append(txSlice, k)
	}

	sar.SuspiciousActivity = SARActivity{
		DateStart:      earliest,
		DateEnd:        latest,
		TotalAmount:    totalAmount,
		Instruments:    instrSlice,
		TransactionIDs: txSlice,
	}

	return sar
}

// ValidateSAR returns validation errors for a SAR before filing.
// An empty slice means the SAR is ready for submission.
func ValidateSAR(sar *SARData) []string {
	var errs []string

	if sar.FilingType == "" {
		errs = append(errs, "filing_type is required")
	}

	// Subject validation — FinCEN requires at minimum a name.
	if sar.Subject.Name == "" && (sar.Subject.FirstName == "" || sar.Subject.LastName == "") {
		errs = append(errs, "subject name is required (full name or first + last)")
	}

	// Activity must have a date range.
	if sar.SuspiciousActivity.DateStart.IsZero() {
		errs = append(errs, "suspicious_activity.date_start is required")
	}
	if sar.SuspiciousActivity.DateEnd.IsZero() {
		errs = append(errs, "suspicious_activity.date_end is required")
	}
	if !sar.SuspiciousActivity.DateStart.IsZero() && !sar.SuspiciousActivity.DateEnd.IsZero() {
		if sar.SuspiciousActivity.DateEnd.Before(sar.SuspiciousActivity.DateStart) {
			errs = append(errs, "suspicious_activity.date_end must be after date_start")
		}
	}

	// FinCEN requires SAR filing for suspicious activity >= $5,000.
	if sar.SuspiciousActivity.TotalAmount < 5000 {
		errs = append(errs, fmt.Sprintf("total_amount $%.2f below $5,000 SAR threshold", sar.SuspiciousActivity.TotalAmount))
	}

	// Narrative is mandatory and must be substantive.
	narrative := strings.TrimSpace(sar.Narrative)
	if narrative == "" {
		errs = append(errs, "narrative is required")
	} else if len(narrative) < 50 {
		errs = append(errs, "narrative must be at least 50 characters — explain the suspicious activity")
	}

	return errs
}
