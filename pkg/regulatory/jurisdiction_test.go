// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

import (
	"testing"
)

// --- USA ---

func TestUSAName(t *testing.T) {
	j := &USA{}
	if j.Name() != "United States" {
		t.Fatalf("expected 'United States', got %q", j.Name())
	}
	if j.Code() != "US" {
		t.Fatalf("expected 'US', got %q", j.Code())
	}
}

func TestUSARequirements(t *testing.T) {
	j := &USA{}
	reqs := j.Requirements()
	if len(reqs) == 0 {
		t.Fatal("expected requirements")
	}
	// Check a known requirement exists
	found := false
	for _, r := range reqs {
		if r.ID == "us-bsa-cip" {
			found = true
			if !r.Mandatory {
				t.Fatal("CIP should be mandatory")
			}
		}
	}
	if !found {
		t.Fatal("expected us-bsa-cip requirement")
	}
}

func TestUSATransactionLimits(t *testing.T) {
	j := &USA{}
	limits := j.TransactionLimits()
	if limits.CTRThreshold != 10_000 {
		t.Fatalf("expected CTR threshold $10k, got $%.0f", limits.CTRThreshold)
	}
	if limits.TravelRuleMin != 3_000 {
		t.Fatalf("expected travel rule $3k, got $%.0f", limits.TravelRuleMin)
	}
	if limits.Currency != "USD" {
		t.Fatalf("expected USD, got %q", limits.Currency)
	}
}

func TestUSAValidateComplete(t *testing.T) {
	j := &USA{}
	boolTrue := true
	boolFalse := false
	app := &ApplicationData{
		GivenName:           "John",
		FamilyName:          "Doe",
		DateOfBirth:         "1990-01-15",
		Email:               "john@example.com",
		Street:              []string{"123 Main St"},
		City:                "New York",
		State:               "NY",
		PostalCode:          "10001",
		Country:             "US",
		TaxID:               "123-45-6789",
		TaxIDType:           "ssn",
		CountryOfTax:        "US",
		IsControlPerson:     &boolFalse,
		IsAffiliatedExchange: &boolFalse,
		IsPoliticallyExposed: &boolFalse,
		ImmediateFamilyExposed: &boolTrue,
		EmploymentStatus:    "employed",
		EmployerName:        "Acme Corp",
		AnnualIncome:        "100k-200k",
		FundingSource:       "employment_income",
		InvestmentObjective: "growth",
	}
	violations := j.ValidateApplication(app)
	// Should have no errors (but may have warnings)
	errors := 0
	for _, v := range violations {
		if v.Severity == "error" {
			t.Errorf("unexpected error: %s - %s", v.Field, v.Message)
			errors++
		}
	}
	if errors > 0 {
		t.Fatalf("expected 0 errors for complete application, got %d", errors)
	}
}

func TestUSAValidateIncomplete(t *testing.T) {
	j := &USA{}
	app := &ApplicationData{
		Country:    "US",
		CountryOfTax: "US",
	}
	violations := j.ValidateApplication(app)
	if len(violations) == 0 {
		t.Fatal("expected violations for empty application")
	}

	// Check specific violations
	fieldErrors := map[string]bool{}
	for _, v := range violations {
		if v.Severity == "error" {
			fieldErrors[v.Field] = true
		}
	}
	required := []string{"given_name", "family_name", "date_of_birth", "tax_id", "tax_id_type",
		"street", "city", "state", "postal_code", "investment_objective", "annual_income",
		"is_control_person", "is_affiliated_exchange_or_finra", "is_politically_exposed",
		"immediate_family_exposed", "funding_source", "email", "employment_status"}
	for _, f := range required {
		if !fieldErrors[f] {
			t.Errorf("expected error for field %q", f)
		}
	}
}

func TestUSAValidatePEPWarning(t *testing.T) {
	j := &USA{}
	boolTrue := true
	boolFalse := false
	app := &ApplicationData{
		GivenName:           "Jane",
		FamilyName:          "Senator",
		DateOfBirth:         "1970-06-15",
		Email:               "jane@gov.example.com",
		Street:              []string{"456 Capitol Hill"},
		City:                "Washington",
		State:               "DC",
		PostalCode:          "20001",
		Country:             "US",
		TaxID:               "999-88-7777",
		TaxIDType:           "ssn",
		CountryOfTax:        "US",
		IsControlPerson:     &boolFalse,
		IsAffiliatedExchange: &boolFalse,
		IsPoliticallyExposed: &boolTrue,
		ImmediateFamilyExposed: &boolFalse,
		EmploymentStatus:    "employed",
		EmployerName:        "US Senate",
		AnnualIncome:        "100k-200k",
		FundingSource:       "employment_income",
		InvestmentObjective: "growth",
	}
	violations := j.ValidateApplication(app)
	foundPEPWarning := false
	for _, v := range violations {
		if v.Severity == "warning" && v.Field == "is_politically_exposed" {
			foundPEPWarning = true
		}
	}
	if !foundPEPWarning {
		t.Fatal("expected PEP warning for politically exposed person")
	}
}

func TestUSAValidateEmployedRequiresEmployer(t *testing.T) {
	j := &USA{}
	boolFalse := false
	app := &ApplicationData{
		GivenName:           "Bob",
		FamilyName:          "Worker",
		DateOfBirth:         "1985-03-20",
		Email:               "bob@example.com",
		Street:              []string{"789 Oak St"},
		City:                "Chicago",
		State:               "IL",
		PostalCode:          "60601",
		Country:             "US",
		TaxID:               "111-22-3333",
		TaxIDType:           "ssn",
		CountryOfTax:        "US",
		IsControlPerson:     &boolFalse,
		IsAffiliatedExchange: &boolFalse,
		IsPoliticallyExposed: &boolFalse,
		ImmediateFamilyExposed: &boolFalse,
		EmploymentStatus:    "employed",
		// Missing EmployerName
		AnnualIncome:        "50k-100k",
		FundingSource:       "employment_income",
		InvestmentObjective: "income",
	}
	violations := j.ValidateApplication(app)
	found := false
	for _, v := range violations {
		if v.Field == "employer_name" && v.Severity == "error" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected employer_name error when employed but no employer")
	}
}

func TestUSAStateFormat(t *testing.T) {
	j := &USA{}
	boolFalse := false
	app := &ApplicationData{
		GivenName: "Test", FamilyName: "User", DateOfBirth: "1990-01-01",
		Email: "test@example.com", Street: []string{"1 St"}, City: "LA",
		State: "california", PostalCode: "90001", Country: "US",
		TaxID: "000-00-0000", TaxIDType: "ssn", CountryOfTax: "US",
		IsControlPerson: &boolFalse, IsAffiliatedExchange: &boolFalse,
		IsPoliticallyExposed: &boolFalse, ImmediateFamilyExposed: &boolFalse,
		EmploymentStatus: "retired", AnnualIncome: "25k-50k",
		FundingSource: "savings", InvestmentObjective: "preservation",
	}
	violations := j.ValidateApplication(app)
	found := false
	for _, v := range violations {
		if v.Field == "state" && v.Severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected state format warning for 'california'")
	}
}

// --- UK ---

func TestUKName(t *testing.T) {
	j := &UK{}
	if j.Name() != "United Kingdom" {
		t.Fatalf("expected 'United Kingdom', got %q", j.Name())
	}
	if j.Code() != "GB" {
		t.Fatalf("expected 'GB', got %q", j.Code())
	}
}

func TestUKRequirements(t *testing.T) {
	j := &UK{}
	reqs := j.Requirements()
	if len(reqs) == 0 {
		t.Fatal("expected requirements")
	}
	found := false
	for _, r := range reqs {
		if r.ID == "uk-5amld-cdd" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected uk-5amld-cdd requirement")
	}
}

func TestUKTransactionLimits(t *testing.T) {
	j := &UK{}
	limits := j.TransactionLimits()
	if limits.Currency != "GBP" {
		t.Fatalf("expected GBP, got %q", limits.Currency)
	}
	if limits.CTRThreshold != 15_000 {
		t.Fatalf("expected CTR threshold 15k, got %.0f", limits.CTRThreshold)
	}
}

func TestUKValidateComplete(t *testing.T) {
	j := &UK{}
	boolFalse := false
	app := &ApplicationData{
		GivenName:              "Alice",
		FamilyName:             "Smith",
		DateOfBirth:            "1988-07-22",
		Email:                  "alice@example.co.uk",
		Street:                 []string{"10 Downing St"},
		City:                   "London",
		PostalCode:             "SW1A 2AA",
		Country:                "GB",
		FundingSource:          "employment_income",
		IsPoliticallyExposed:   &boolFalse,
		ImmediateFamilyExposed: &boolFalse,
	}
	violations := j.ValidateApplication(app)
	errors := 0
	for _, v := range violations {
		if v.Severity == "error" {
			t.Errorf("unexpected error: %s - %s", v.Field, v.Message)
			errors++
		}
	}
	if errors > 0 {
		t.Fatalf("expected 0 errors for complete UK application, got %d", errors)
	}
}

func TestUKValidateIncomplete(t *testing.T) {
	j := &UK{}
	violations := j.ValidateApplication(&ApplicationData{})
	if len(violations) == 0 {
		t.Fatal("expected violations for empty UK application")
	}
	errorCount := 0
	for _, v := range violations {
		if v.Severity == "error" {
			errorCount++
		}
	}
	if errorCount < 5 {
		t.Fatalf("expected at least 5 errors for empty UK app, got %d", errorCount)
	}
}

func TestUKPEPWarning(t *testing.T) {
	j := &UK{}
	boolTrue := true
	boolFalse := false
	app := &ApplicationData{
		GivenName: "PM", FamilyName: "Test", DateOfBirth: "1970-01-01",
		Email: "pm@gov.uk", Street: []string{"1 St"}, City: "London",
		PostalCode: "SW1", Country: "GB", FundingSource: "savings",
		IsPoliticallyExposed: &boolTrue, ImmediateFamilyExposed: &boolFalse,
	}
	violations := j.ValidateApplication(app)
	found := false
	for _, v := range violations {
		if v.Severity == "warning" && v.RequirementID == "uk-5amld-edd" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected EDD warning for PEP in UK")
	}
}

// --- Isle of Man ---

func TestIOMName(t *testing.T) {
	j := &IOM{}
	if j.Name() != "Isle of Man" {
		t.Fatalf("expected 'Isle of Man', got %q", j.Name())
	}
	if j.Code() != "IM" {
		t.Fatalf("expected 'IM', got %q", j.Code())
	}
}

func TestIOMRequirements(t *testing.T) {
	j := &IOM{}
	reqs := j.Requirements()
	if len(reqs) == 0 {
		t.Fatal("expected requirements")
	}
	found := false
	for _, r := range reqs {
		if r.ID == "im-amlcft-source-wealth" {
			found = true
			if !r.Mandatory {
				t.Fatal("source of wealth should be mandatory for IOM")
			}
		}
	}
	if !found {
		t.Fatal("expected im-amlcft-source-wealth requirement")
	}
}

func TestIOMTransactionLimits(t *testing.T) {
	j := &IOM{}
	limits := j.TransactionLimits()
	if limits.Currency != "GBP" {
		t.Fatalf("expected GBP, got %q", limits.Currency)
	}
}

func TestIOMValidateComplete(t *testing.T) {
	j := &IOM{}
	boolFalse := false
	app := &ApplicationData{
		GivenName:              "Bob",
		FamilyName:             "Douglas",
		DateOfBirth:            "1975-11-30",
		Email:                  "bob@example.im",
		Street:                 []string{"1 Loch Promenade"},
		City:                   "Douglas",
		PostalCode:             "IM1 1AA",
		Country:                "IM",
		NetWorth:               "500k-1M",
		AnnualIncome:           "100k-200k",
		FundingSource:          "employment_income",
		EmploymentStatus:       "employed",
		IsPoliticallyExposed:   &boolFalse,
		ImmediateFamilyExposed: &boolFalse,
	}
	violations := j.ValidateApplication(app)
	errors := 0
	for _, v := range violations {
		if v.Severity == "error" {
			t.Errorf("unexpected error: %s - %s", v.Field, v.Message)
			errors++
		}
	}
	if errors > 0 {
		t.Fatalf("expected 0 errors for complete IOM application, got %d", errors)
	}
}

func TestIOMValidateIncomplete(t *testing.T) {
	j := &IOM{}
	violations := j.ValidateApplication(&ApplicationData{})
	if len(violations) == 0 {
		t.Fatal("expected violations for empty IOM application")
	}
}

func TestIOMSourceOfWealth(t *testing.T) {
	j := &IOM{}
	boolFalse := false
	app := &ApplicationData{
		GivenName: "Test", FamilyName: "User", DateOfBirth: "1990-01-01",
		Email: "test@example.im", Street: []string{"1 St"}, City: "Douglas",
		PostalCode: "IM1", Country: "IM", FundingSource: "savings",
		EmploymentStatus: "employed",
		IsPoliticallyExposed: &boolFalse, ImmediateFamilyExposed: &boolFalse,
		// Missing NetWorth and AnnualIncome
	}
	violations := j.ValidateApplication(app)
	foundWealth := false
	for _, v := range violations {
		if v.RequirementID == "im-amlcft-source-wealth" && v.Severity == "error" {
			foundWealth = true
		}
	}
	if !foundWealth {
		t.Fatal("expected source of wealth error for missing net_worth/annual_income")
	}
}

// --- GetJurisdiction ---

func TestGetJurisdiction(t *testing.T) {
	us := GetJurisdiction("US")
	if us == nil || us.Code() != "US" {
		t.Fatal("expected US jurisdiction")
	}
	gb := GetJurisdiction("GB")
	if gb == nil || gb.Code() != "GB" {
		t.Fatal("expected GB jurisdiction")
	}
	im := GetJurisdiction("IM")
	if im == nil || im.Code() != "IM" {
		t.Fatal("expected IM jurisdiction")
	}
	unknown := GetJurisdiction("XX")
	if unknown != nil {
		t.Fatal("expected nil for unknown jurisdiction")
	}
}

func TestAllJurisdictions(t *testing.T) {
	all := AllJurisdictions()
	if len(all) != 3 {
		t.Fatalf("expected 3 jurisdictions, got %d", len(all))
	}
}
