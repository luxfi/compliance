// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

import (
	"sort"
	"testing"
)

// jurisdictionTestCase defines the expected properties for each jurisdiction.
type jurisdictionTestCase struct {
	code        string
	name        string
	framework   string
	passportLen int
	minReqs     int
	currency    string
}

var jurisdictionCases = []jurisdictionTestCase{
	{"CA", "Canada", "ciro", 0, 3, "CAD"},
	{"BR", "Brazil", "cvm", 0, 3, "BRL"},
	{"IN", "India", "sebi", 0, 3, "INR"},
	{"SG", "Singapore", "mas", 0, 3, "SGD"},
	{"AU", "Australia", "asic", 0, 3, "AUD"},
	{"CH", "Switzerland", "finma", 0, 3, "CHF"},
	{"AE", "United Arab Emirates", "sca", 0, 3, "AED"},
	{"AE-DIFC", "UAE - DIFC", "dfsa", 0, 3, "USD"},
	{"AE-ADGM", "UAE - ADGM", "fsra", 0, 3, "USD"},
	{"AE-VARA", "UAE - VARA (Dubai)", "vara", 0, 3, "AED"},
	{"LU", "Luxembourg", "mica", 6, 3, "EUR"},
	{"DE", "Germany", "mica", 6, 3, "EUR"},
	{"FR", "France", "mica", 6, 3, "EUR"},
	{"NL", "Netherlands", "mica", 6, 3, "EUR"},
	{"IE", "Ireland", "mica", 6, 3, "EUR"},
	{"IT", "Italy", "mica", 6, 3, "EUR"},
	{"ES", "Spain", "mica", 6, 3, "EUR"},
}

func TestNewJurisdictions_CodeAndName(t *testing.T) {
	for _, tc := range jurisdictionCases {
		j := GetJurisdiction(tc.code)
		if j == nil {
			t.Fatalf("GetJurisdiction(%q) returned nil", tc.code)
		}
		if j.Code() != tc.code {
			t.Errorf("%s: Code() = %q, want %q", tc.code, j.Code(), tc.code)
		}
		if j.Name() == "" {
			t.Errorf("%s: Name() is empty", tc.code)
		}
		if j.Name() != tc.name {
			t.Errorf("%s: Name() = %q, want %q", tc.code, j.Name(), tc.name)
		}
	}
}

func TestNewJurisdictions_RegulatoryFramework(t *testing.T) {
	for _, tc := range jurisdictionCases {
		j := GetJurisdiction(tc.code)
		if j.RegulatoryFramework() != tc.framework {
			t.Errorf("%s: RegulatoryFramework() = %q, want %q", tc.code, j.RegulatoryFramework(), tc.framework)
		}
	}
}

func TestNewJurisdictions_PassportableTo(t *testing.T) {
	for _, tc := range jurisdictionCases {
		j := GetJurisdiction(tc.code)
		passport := j.PassportableTo()
		if len(passport) != tc.passportLen {
			t.Errorf("%s: PassportableTo() has %d entries, want %d", tc.code, len(passport), tc.passportLen)
		}
		// EU jurisdictions must not passport to themselves
		for _, p := range passport {
			if p == tc.code {
				t.Errorf("%s: PassportableTo() includes self", tc.code)
			}
		}
	}
}

func TestNewJurisdictions_Requirements(t *testing.T) {
	for _, tc := range jurisdictionCases {
		j := GetJurisdiction(tc.code)
		reqs := j.Requirements()
		if len(reqs) < tc.minReqs {
			t.Errorf("%s: Requirements() has %d items, want >= %d", tc.code, len(reqs), tc.minReqs)
		}
		// Verify each requirement has non-empty fields
		categories := map[string]bool{}
		for _, r := range reqs {
			if r.ID == "" {
				t.Errorf("%s: requirement has empty ID", tc.code)
			}
			if r.Category == "" {
				t.Errorf("%s: requirement %s has empty Category", tc.code, r.ID)
			}
			if r.Description == "" {
				t.Errorf("%s: requirement %s has empty Description", tc.code, r.ID)
			}
			if r.Reference == "" {
				t.Errorf("%s: requirement %s has empty Reference", tc.code, r.ID)
			}
			categories[r.Category] = true
		}
		// Must cover at least kyc and one of aml/reporting
		if !categories["kyc"] {
			t.Errorf("%s: Requirements() missing kyc category", tc.code)
		}
		if !categories["aml"] && !categories["reporting"] {
			t.Errorf("%s: Requirements() missing aml or reporting category", tc.code)
		}
	}
}

func TestNewJurisdictions_TransactionLimits(t *testing.T) {
	for _, tc := range jurisdictionCases {
		j := GetJurisdiction(tc.code)
		limits := j.TransactionLimits()
		if limits == nil {
			t.Fatalf("%s: TransactionLimits() returned nil", tc.code)
		}
		if limits.Currency != tc.currency {
			t.Errorf("%s: Currency = %q, want %q", tc.code, limits.Currency, tc.currency)
		}
		if limits.SingleTransactionMax <= 0 {
			t.Errorf("%s: SingleTransactionMax = %.0f, must be positive", tc.code, limits.SingleTransactionMax)
		}
		if limits.DailyMax <= limits.SingleTransactionMax {
			t.Errorf("%s: DailyMax (%.0f) should exceed SingleTransactionMax (%.0f)", tc.code, limits.DailyMax, limits.SingleTransactionMax)
		}
		if limits.MonthlyMax <= limits.DailyMax {
			t.Errorf("%s: MonthlyMax (%.0f) should exceed DailyMax (%.0f)", tc.code, limits.MonthlyMax, limits.DailyMax)
		}
		if limits.CTRThreshold <= 0 {
			t.Errorf("%s: CTRThreshold must be positive", tc.code)
		}
		if limits.TravelRuleMin <= 0 {
			t.Errorf("%s: TravelRuleMin must be positive", tc.code)
		}
		if limits.TravelRuleMin >= limits.CTRThreshold {
			t.Errorf("%s: TravelRuleMin (%.0f) should be less than CTRThreshold (%.0f)", tc.code, limits.TravelRuleMin, limits.CTRThreshold)
		}
	}
}

func TestNewJurisdictions_ValidateRejectsEmpty(t *testing.T) {
	for _, tc := range jurisdictionCases {
		j := GetJurisdiction(tc.code)
		violations := j.ValidateApplication(&ApplicationData{})
		if len(violations) == 0 {
			t.Errorf("%s: ValidateApplication(empty) returned no violations", tc.code)
		}
		errorCount := 0
		for _, v := range violations {
			if v.Severity == "error" {
				errorCount++
			}
		}
		if errorCount < 3 {
			t.Errorf("%s: ValidateApplication(empty) returned only %d errors, want >= 3", tc.code, errorCount)
		}
	}
}

func TestNewJurisdictions_ValidateRejectsMissingTaxID(t *testing.T) {
	// Only test jurisdictions that check tax ID for residents
	taxIDJurisdictions := map[string]string{
		"CA": "CA", "BR": "BR", "IN": "IN", "SG": "SG", "AU": "AU", "CH": "CH", "AE": "AE",
	}
	for code, country := range taxIDJurisdictions {
		j := GetJurisdiction(code)
		app := &ApplicationData{
			Country:    country,
			CountryOfTax: country,
			// No TaxID
		}
		violations := j.ValidateApplication(app)
		foundTaxID := false
		for _, v := range violations {
			if v.Field == "tax_id" && v.Severity == "error" {
				foundTaxID = true
			}
		}
		if !foundTaxID {
			t.Errorf("%s: ValidateApplication should require tax_id for %s residents", code, country)
		}
	}
}

func TestNewJurisdictions_ValidatePEPWarning(t *testing.T) {
	for _, tc := range jurisdictionCases {
		j := GetJurisdiction(tc.code)
		boolTrue := true
		boolFalse := false
		app := &ApplicationData{
			GivenName:              "Test",
			FamilyName:             "PEP",
			DateOfBirth:            "1970-01-01",
			Email:                  "test@example.com",
			Street:                 []string{"1 Main St"},
			City:                   "TestCity",
			PostalCode:             "12345",
			Country:                "XX",
			FundingSource:          "savings",
			AnnualIncome:           "100k-200k",
			InvestmentObjective:    "growth",
			EmploymentStatus:       "employed",
			IsPoliticallyExposed:   &boolTrue,
			ImmediateFamilyExposed: &boolFalse,
		}
		violations := j.ValidateApplication(app)
		foundPEPWarning := false
		for _, v := range violations {
			if v.Severity == "warning" && v.Field == "is_politically_exposed" {
				foundPEPWarning = true
			}
		}
		if !foundPEPWarning {
			t.Errorf("%s: ValidateApplication should emit PEP warning", tc.code)
		}
	}
}

// --- Existing jurisdictions: new interface methods ---

func TestUSA_NewMethods(t *testing.T) {
	j := &USA{}
	if j.RegulatoryFramework() != "us_sec_finra" {
		t.Errorf("USA.RegulatoryFramework() = %q, want us_sec_finra", j.RegulatoryFramework())
	}
	if len(j.PassportableTo()) != 0 {
		t.Error("USA should not passport to any jurisdiction")
	}
}

func TestUK_NewMethods(t *testing.T) {
	j := &UK{}
	if j.RegulatoryFramework() != "uk_fca" {
		t.Errorf("UK.RegulatoryFramework() = %q, want uk_fca", j.RegulatoryFramework())
	}
	if len(j.PassportableTo()) != 0 {
		t.Error("UK should not passport (post-Brexit)")
	}
}

func TestIOM_NewMethods(t *testing.T) {
	j := &IOM{}
	if j.RegulatoryFramework() != "iom" {
		t.Errorf("IOM.RegulatoryFramework() = %q, want iom", j.RegulatoryFramework())
	}
}

// --- Framework tests ---

func TestFrameworks(t *testing.T) {
	micaCodes := JurisdictionsByFramework(Framework_MICA)
	if len(micaCodes) != 7 {
		t.Fatalf("MiCA should have 7 jurisdictions, got %d", len(micaCodes))
	}
	expected := []string{"LU", "DE", "FR", "NL", "IE", "IT", "ES"}
	sort.Strings(expected)
	sort.Strings(micaCodes)
	for i, c := range expected {
		if micaCodes[i] != c {
			t.Errorf("MiCA[%d] = %q, want %q", i, micaCodes[i], c)
		}
	}
}

func TestFrameworksUnknown(t *testing.T) {
	if codes := JurisdictionsByFramework("nonexistent"); codes != nil {
		t.Errorf("unknown framework should return nil, got %v", codes)
	}
}

func TestAllFrameworks(t *testing.T) {
	all := AllFrameworks()
	if len(all) != 14 {
		t.Fatalf("expected 14 frameworks, got %d", len(all))
	}
}

func TestFrameworkCoverage(t *testing.T) {
	// Every jurisdiction's RegulatoryFramework() must map to a known Framework constant
	for _, j := range AllJurisdictions() {
		fw := j.RegulatoryFramework()
		codes := JurisdictionsByFramework(Framework(fw))
		if codes == nil {
			t.Errorf("%s: RegulatoryFramework() returns %q which has no JurisdictionsByFramework mapping", j.Code(), fw)
		}
		// The jurisdiction's code must appear in its own framework
		found := false
		for _, c := range codes {
			if c == j.Code() {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: code not found in JurisdictionsByFramework(%q)", j.Code(), fw)
		}
	}
}

// --- EU passporting symmetry ---

func TestEUPassportingSymmetry(t *testing.T) {
	euCodes := []string{"LU", "DE", "FR", "NL", "IE", "IT", "ES"}
	for _, code := range euCodes {
		j := GetJurisdiction(code)
		passport := j.PassportableTo()
		if len(passport) != 6 {
			t.Errorf("%s: PassportableTo() should have 6 entries, got %d", code, len(passport))
			continue
		}
		// Each EU code should be passportable to all other EU codes
		for _, other := range euCodes {
			if other == code {
				continue
			}
			found := false
			for _, p := range passport {
				if p == other {
					found = true
				}
			}
			if !found {
				t.Errorf("%s: PassportableTo() missing %s", code, other)
			}
		}
	}
}

// --- GetJurisdiction extended ---

func TestGetJurisdictionAllCodes(t *testing.T) {
	codes := []string{
		"US", "GB", "IM", "CA", "BR", "IN", "SG", "AU", "CH",
		"AE", "AE-DIFC", "AE-ADGM", "AE-VARA",
		"LU", "DE", "FR", "NL", "IE", "IT", "ES",
	}
	for _, code := range codes {
		j := GetJurisdiction(code)
		if j == nil {
			t.Errorf("GetJurisdiction(%q) returned nil", code)
			continue
		}
		if j.Code() != code {
			t.Errorf("GetJurisdiction(%q).Code() = %q", code, j.Code())
		}
	}
}

func TestGetJurisdictionUnknownReturnsNil(t *testing.T) {
	for _, code := range []string{"XX", "ZZ", "US-NY", ""} {
		if j := GetJurisdiction(code); j != nil {
			t.Errorf("GetJurisdiction(%q) should return nil", code)
		}
	}
}

// --- Specific jurisdiction validation tests ---

func TestCanada_ValidateComplete(t *testing.T) {
	j := GetJurisdiction("CA")
	boolFalse := false
	app := &ApplicationData{
		GivenName: "Jean", FamilyName: "Tremblay", DateOfBirth: "1985-03-15",
		Email: "jean@example.ca", Street: []string{"100 Wellington St"}, City: "Ottawa",
		PostalCode: "K1A 0A2", Country: "CA", TaxID: "123-456-789", TaxIDType: "sin",
		CountryOfTax: "CA", InvestmentObjective: "growth", AnnualIncome: "75k-100k",
		IsPoliticallyExposed: &boolFalse,
	}
	for _, v := range j.ValidateApplication(app) {
		if v.Severity == "error" {
			t.Errorf("CA complete app: unexpected error: %s - %s", v.Field, v.Message)
		}
	}
}

func TestBrazil_ValidateComplete(t *testing.T) {
	j := GetJurisdiction("BR")
	boolFalse := false
	app := &ApplicationData{
		GivenName: "Maria", FamilyName: "Silva", DateOfBirth: "1990-07-22",
		Email: "maria@example.com.br", Street: []string{"Av Paulista 1000"}, City: "Sao Paulo",
		PostalCode: "01311-100", Country: "BR", TaxID: "123.456.789-00", TaxIDType: "cpf",
		CountryOfTax: "BR", InvestmentObjective: "income", AnnualIncome: "50k-100k",
		IsPoliticallyExposed: &boolFalse,
	}
	for _, v := range j.ValidateApplication(app) {
		if v.Severity == "error" {
			t.Errorf("BR complete app: unexpected error: %s - %s", v.Field, v.Message)
		}
	}
}

func TestIndia_ValidateComplete(t *testing.T) {
	j := GetJurisdiction("IN")
	boolFalse := false
	app := &ApplicationData{
		GivenName: "Rahul", FamilyName: "Sharma", DateOfBirth: "1988-11-05",
		Email: "rahul@example.in", Street: []string{"42 MG Road"}, City: "Mumbai",
		PostalCode: "400001", Country: "IN", TaxID: "ABCDE1234F", TaxIDType: "pan",
		CountryOfTax: "IN", AnnualIncome: "10L-25L",
		IsPoliticallyExposed: &boolFalse,
	}
	for _, v := range j.ValidateApplication(app) {
		if v.Severity == "error" {
			t.Errorf("IN complete app: unexpected error: %s - %s", v.Field, v.Message)
		}
	}
}

func TestSingapore_ValidateComplete(t *testing.T) {
	j := GetJurisdiction("SG")
	boolFalse := false
	app := &ApplicationData{
		GivenName: "Wei", FamilyName: "Tan", DateOfBirth: "1992-02-28",
		Email: "wei@example.sg", Street: []string{"1 Raffles Place"}, City: "Singapore",
		PostalCode: "048616", Country: "SG", TaxID: "S9200001A", TaxIDType: "nric",
		CountryOfTax: "SG", FundingSource: "employment_income",
		IsPoliticallyExposed: &boolFalse,
	}
	for _, v := range j.ValidateApplication(app) {
		if v.Severity == "error" {
			t.Errorf("SG complete app: unexpected error: %s - %s", v.Field, v.Message)
		}
	}
}

func TestAustralia_ValidateComplete(t *testing.T) {
	j := GetJurisdiction("AU")
	boolFalse := false
	app := &ApplicationData{
		GivenName: "Liam", FamilyName: "Mitchell", DateOfBirth: "1987-09-14",
		Email: "liam@example.com.au", Street: []string{"1 George St"}, City: "Sydney",
		PostalCode: "2000", Country: "AU", TaxID: "123456789", TaxIDType: "tfn",
		CountryOfTax: "AU",
		IsPoliticallyExposed: &boolFalse,
	}
	for _, v := range j.ValidateApplication(app) {
		if v.Severity == "error" {
			t.Errorf("AU complete app: unexpected error: %s - %s", v.Field, v.Message)
		}
	}
}

func TestSwitzerland_ValidateComplete(t *testing.T) {
	j := GetJurisdiction("CH")
	boolFalse := false
	app := &ApplicationData{
		GivenName: "Hans", FamilyName: "Mueller", DateOfBirth: "1975-06-01",
		Email: "hans@example.ch", Street: []string{"Bahnhofstrasse 1"}, City: "Zurich",
		PostalCode: "8001", Country: "CH", TaxID: "756.1234.5678.97", TaxIDType: "ahv",
		CountryOfTax: "CH", FundingSource: "employment_income",
		IsPoliticallyExposed: &boolFalse,
	}
	for _, v := range j.ValidateApplication(app) {
		if v.Severity == "error" {
			t.Errorf("CH complete app: unexpected error: %s - %s", v.Field, v.Message)
		}
	}
}

func TestSwitzerland_DLTCustodianWarning(t *testing.T) {
	j := GetJurisdiction("CH")
	boolFalse := false
	app := &ApplicationData{
		GivenName: "Corp", FamilyName: "AG", DateOfBirth: "2020-01-01",
		Email: "corp@example.ch", Street: []string{"1 Finanzplatz"}, City: "Zug",
		PostalCode: "6300", Country: "CH", TaxID: "CHE-123.456.789", TaxIDType: "uid",
		CountryOfTax: "CH", FundingSource: "corporate_funds",
		IsPoliticallyExposed: &boolFalse,
		AccountType: "entity",
	}
	violations := j.ValidateApplication(app)
	found := false
	for _, v := range violations {
		if v.RequirementID == "ch-dlt-custodian" && v.Severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("CH entity account should trigger DLT custodian warning")
	}
}

func TestUAE_ValidateComplete(t *testing.T) {
	j := GetJurisdiction("AE")
	boolFalse := false
	app := &ApplicationData{
		GivenName: "Ahmed", FamilyName: "Al Maktoum", DateOfBirth: "1980-03-25",
		Email: "ahmed@example.ae", Street: []string{"1 Sheikh Zayed Road"}, City: "Dubai",
		Country: "AE", TaxID: "784-1234-5678901-1", CountryOfTax: "AE",
		FundingSource:        "business_income",
		IsPoliticallyExposed: &boolFalse,
	}
	for _, v := range j.ValidateApplication(app) {
		if v.Severity == "error" {
			t.Errorf("AE complete app: unexpected error: %s - %s", v.Field, v.Message)
		}
	}
}

func TestLuxembourg_PassportableTo(t *testing.T) {
	j := GetJurisdiction("LU")
	passport := j.PassportableTo()
	expected := map[string]bool{"DE": true, "FR": true, "NL": true, "IE": true, "IT": true, "ES": true}
	for _, code := range passport {
		if !expected[code] {
			t.Errorf("LU passports to unexpected %q", code)
		}
		delete(expected, code)
	}
	for code := range expected {
		t.Errorf("LU missing passport to %q", code)
	}
}

func TestEU_ValidateComplete(t *testing.T) {
	boolFalse := false
	app := &ApplicationData{
		GivenName: "Pierre", FamilyName: "Dupont", DateOfBirth: "1982-04-12",
		Email: "pierre@example.fr", Street: []string{"1 Rue de Rivoli"}, City: "Paris",
		PostalCode: "75001", Country: "FR", FundingSource: "employment_income",
		IsPoliticallyExposed: &boolFalse,
	}
	for _, code := range []string{"LU", "DE", "FR", "NL", "IE", "IT", "ES"} {
		j := GetJurisdiction(code)
		violations := j.ValidateApplication(app)
		for _, v := range violations {
			if v.Severity == "error" {
				t.Errorf("%s complete EU app: unexpected error: %s - %s", code, v.Field, v.Message)
			}
		}
	}
}
