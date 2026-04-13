// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package entity

import "testing"

// entityTestCase defines expected properties for each new entity type.
type entityTestCase struct {
	entityType   string
	name         string
	jurisdiction string
	minLicenses  int
	minReporting int
	minCapital   float64
	currency     string
}

var newEntityCases = []entityTestCase{
	{EntityType_SICAV, "SICAV", "LU", 1, 2, 1_250_000, "EUR"},
	{EntityType_SICAR, "SICAR", "LU", 1, 1, 1_000_000, "EUR"},
	{EntityType_RAIF, "RAIF", "LU", 1, 1, 1_250_000, "EUR"},
	{EntityType_AIFM, "Alternative Investment Fund Manager", "LU", 1, 1, 125_000, "EUR"},
	{EntityType_MANCOMAN, "Management Company", "LU", 1, 1, 125_000, "EUR"},
	{EntityType_CRR, "Credit Institution (CRR)", "LU", 1, 2, 5_000_000, "EUR"},
	{EntityType_ISSUER, "Securities Issuer (SPV)", "LU", 1, 1, 0, "EUR"},
	{EntityType_CUSTODIAN, "Qualified Custodian", "LU", 1, 1, 5_000_000, "EUR"},
	{EntityType_DLT_FACILITY, "DLT Trading Facility", "CH", 1, 1, 1_000_000, "CHF"},
}

func TestNewEntities_Interface(t *testing.T) {
	for _, tc := range newEntityCases {
		e := GetEntity(tc.entityType)
		if e == nil {
			t.Fatalf("GetEntity(%q) returned nil", tc.entityType)
		}

		var _ RegulatedEntity = e // compile-time interface check

		if e.Name() != tc.name {
			t.Errorf("%s: Name() = %q, want %q", tc.entityType, e.Name(), tc.name)
		}
		if e.Type() != tc.entityType {
			t.Errorf("%s: Type() = %q, want %q", tc.entityType, e.Type(), tc.entityType)
		}
		if e.Jurisdiction() != tc.jurisdiction {
			t.Errorf("%s: Jurisdiction() = %q, want %q", tc.entityType, e.Jurisdiction(), tc.jurisdiction)
		}
	}
}

func TestNewEntities_Licenses(t *testing.T) {
	for _, tc := range newEntityCases {
		e := GetEntity(tc.entityType)
		licenses := e.RequiredLicenses()
		if len(licenses) < tc.minLicenses {
			t.Errorf("%s: RequiredLicenses() has %d, want >= %d", tc.entityType, len(licenses), tc.minLicenses)
		}
		for _, l := range licenses {
			if l.Name == "" {
				t.Errorf("%s: license has empty Name", tc.entityType)
			}
			if l.Regulator == "" {
				t.Errorf("%s: license %q has empty Regulator", tc.entityType, l.Name)
			}
			if l.Reference == "" {
				t.Errorf("%s: license %q has empty Reference", tc.entityType, l.Name)
			}
		}
	}
}

func TestNewEntities_Reporting(t *testing.T) {
	for _, tc := range newEntityCases {
		e := GetEntity(tc.entityType)
		obligations := e.ReportingObligations()
		if len(obligations) < tc.minReporting {
			t.Errorf("%s: ReportingObligations() has %d, want >= %d", tc.entityType, len(obligations), tc.minReporting)
		}
		for _, o := range obligations {
			if o.Name == "" {
				t.Errorf("%s: reporting obligation has empty Name", tc.entityType)
			}
			if o.Frequency == "" {
				t.Errorf("%s: obligation %q has empty Frequency", tc.entityType, o.Name)
			}
		}
	}
}

func TestNewEntities_Capital(t *testing.T) {
	for _, tc := range newEntityCases {
		e := GetEntity(tc.entityType)
		cap := e.CapitalRequirements()
		if cap == nil {
			t.Fatalf("%s: CapitalRequirements() returned nil", tc.entityType)
		}
		if cap.MinNetCapital < tc.minCapital {
			t.Errorf("%s: MinNetCapital = %.0f, want >= %.0f", tc.entityType, cap.MinNetCapital, tc.minCapital)
		}
		if cap.Currency != tc.currency {
			t.Errorf("%s: Currency = %q, want %q", tc.entityType, cap.Currency, tc.currency)
		}
		if cap.CalculationRule == "" {
			t.Errorf("%s: CalculationRule is empty", tc.entityType)
		}
		if cap.Reference == "" {
			t.Errorf("%s: Reference is empty", tc.entityType)
		}
	}
}

func TestNewEntities_Operational(t *testing.T) {
	for _, tc := range newEntityCases {
		e := GetEntity(tc.entityType)
		ops := e.OperationalRequirements()
		if len(ops) == 0 {
			t.Errorf("%s: OperationalRequirements() is empty", tc.entityType)
		}
		for _, o := range ops {
			if o.Name == "" {
				t.Errorf("%s: operational requirement has empty Name", tc.entityType)
			}
			if o.Category == "" {
				t.Errorf("%s: requirement %q has empty Category", tc.entityType, o.Name)
			}
			if o.Description == "" {
				t.Errorf("%s: requirement %q has empty Description", tc.entityType, o.Name)
			}
		}
	}
}

func TestEntityTypeConstants(t *testing.T) {
	// Verify all constant values are unique
	types := map[string]bool{}
	constants := []string{
		EntityType_ATS, EntityType_BrokerDealer, EntityType_TransferAgent, EntityType_MSB,
		EntityType_SICAV, EntityType_SICAR, EntityType_RAIF, EntityType_AIFM,
		EntityType_MANCOMAN, EntityType_CRR, EntityType_ISSUER, EntityType_CUSTODIAN,
		EntityType_DLT_FACILITY,
	}
	for _, c := range constants {
		if types[c] {
			t.Errorf("duplicate EntityType constant: %q", c)
		}
		types[c] = true
	}
	if len(constants) != 13 {
		t.Errorf("expected 13 entity type constants, got %d", len(constants))
	}
}

func TestGetEntityAllTypes(t *testing.T) {
	types := []string{
		EntityType_ATS, EntityType_BrokerDealer, EntityType_TransferAgent, EntityType_MSB,
		EntityType_SICAV, EntityType_SICAR, EntityType_RAIF, EntityType_AIFM,
		EntityType_MANCOMAN, EntityType_CRR, EntityType_ISSUER, EntityType_CUSTODIAN,
		EntityType_DLT_FACILITY,
	}
	for _, typ := range types {
		e := GetEntity(typ)
		if e == nil {
			t.Errorf("GetEntity(%q) returned nil", typ)
		}
	}
}

// Specific entity tests

func TestSICAV_Depositary(t *testing.T) {
	s := &SICAV{}
	ops := s.OperationalRequirements()
	found := false
	for _, o := range ops {
		if o.Name == "Depositary Appointment" {
			found = true
		}
	}
	if !found {
		t.Error("SICAV must require depositary appointment")
	}
}

func TestSICAR_WellInformedInvestors(t *testing.T) {
	s := &SICAR{}
	ops := s.OperationalRequirements()
	found := false
	for _, o := range ops {
		if o.Name == "Well-Informed Investors Only" {
			found = true
		}
	}
	if !found {
		t.Error("SICAR must restrict to well-informed investors")
	}
}

func TestAIFM_Capital(t *testing.T) {
	a := &AIFM{}
	cap := a.CapitalRequirements()
	if cap.MinNetCapital < 125_000 {
		t.Errorf("AIFM min capital should be >= EUR 125k, got %.0f", cap.MinNetCapital)
	}
}

func TestCRR_CapitalRatios(t *testing.T) {
	c := &CRR{}
	cap := c.CapitalRequirements()
	if cap.MinNetCapital < 5_000_000 {
		t.Errorf("CRR min capital should be >= EUR 5M, got %.0f", cap.MinNetCapital)
	}
}

func TestDLTFacility_FINMACustodian(t *testing.T) {
	d := &DLTFacility{}
	ops := d.OperationalRequirements()
	found := false
	for _, o := range ops {
		if o.Name == "FINMA-Registered Custodian" {
			found = true
		}
	}
	if !found {
		t.Error("DLT Facility must require FINMA-registered custodian")
	}
}
