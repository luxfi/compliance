// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package entity

import (
	"testing"
)

func TestATSInterface(t *testing.T) {
	var e RegulatedEntity = &ATS{}
	if e.Name() != "Alternative Trading System" {
		t.Fatalf("expected 'Alternative Trading System', got %q", e.Name())
	}
	if e.Type() != "ats" {
		t.Fatalf("expected 'ats', got %q", e.Type())
	}
	if e.Jurisdiction() != "US" {
		t.Fatalf("expected 'US', got %q", e.Jurisdiction())
	}
}

func TestATSLicenses(t *testing.T) {
	a := &ATS{}
	licenses := a.RequiredLicenses()
	if len(licenses) == 0 {
		t.Fatal("expected licenses for ATS")
	}
	// Must include BD registration
	found := false
	for _, l := range licenses {
		if l.Name == "Broker-Dealer Registration" {
			found = true
		}
	}
	if !found {
		t.Fatal("ATS must require Broker-Dealer registration")
	}
}

func TestATSReporting(t *testing.T) {
	a := &ATS{}
	obligations := a.ReportingObligations()
	if len(obligations) == 0 {
		t.Fatal("expected reporting obligations for ATS")
	}
	found := false
	for _, o := range obligations {
		if o.Name == "Form ATS-R Quarterly Report" {
			found = true
			if o.Frequency != "quarterly" {
				t.Fatalf("expected quarterly, got %q", o.Frequency)
			}
		}
	}
	if !found {
		t.Fatal("expected ATS-R quarterly report")
	}
}

func TestATSCapital(t *testing.T) {
	a := &ATS{}
	cap := a.CapitalRequirements()
	if cap == nil {
		t.Fatal("expected capital requirements")
	}
	if cap.MinNetCapital < 250_000 {
		t.Fatalf("expected ATS min capital >= $250k, got $%.0f", cap.MinNetCapital)
	}
	if cap.Currency != "USD" {
		t.Fatalf("expected USD, got %q", cap.Currency)
	}
}

func TestATSOperational(t *testing.T) {
	a := &ATS{}
	ops := a.OperationalRequirements()
	if len(ops) == 0 {
		t.Fatal("expected operational requirements")
	}
}

func TestBrokerDealerInterface(t *testing.T) {
	var e RegulatedEntity = &BrokerDealer{}
	if e.Name() != "Broker-Dealer" {
		t.Fatalf("expected 'Broker-Dealer', got %q", e.Name())
	}
	if e.Type() != "broker_dealer" {
		t.Fatalf("expected 'broker_dealer', got %q", e.Type())
	}
}

func TestBrokerDealerLicenses(t *testing.T) {
	bd := &BrokerDealer{}
	licenses := bd.RequiredLicenses()
	if len(licenses) < 3 {
		t.Fatalf("expected at least 3 licenses for BD, got %d", len(licenses))
	}
	// Must include SEC, FINRA, State, SIPC
	required := map[string]bool{
		"SEC Registration":   false,
		"FINRA Membership":   false,
		"State Registration": false,
		"SIPC Membership":    false,
	}
	for _, l := range licenses {
		if _, ok := required[l.Name]; ok {
			required[l.Name] = true
		}
	}
	for name, found := range required {
		if !found {
			t.Errorf("BD missing required license: %s", name)
		}
	}
}

func TestBrokerDealerReporting(t *testing.T) {
	bd := &BrokerDealer{}
	obligations := bd.ReportingObligations()
	foundFOCUS := false
	foundAudit := false
	for _, o := range obligations {
		if o.Name == "FOCUS Report" {
			foundFOCUS = true
		}
		if o.Name == "Annual Audit" {
			foundAudit = true
		}
	}
	if !foundFOCUS {
		t.Fatal("BD must file FOCUS reports")
	}
	if !foundAudit {
		t.Fatal("BD must have annual audit")
	}
}

func TestBrokerDealerCapital(t *testing.T) {
	bd := &BrokerDealer{}
	cap := bd.CapitalRequirements()
	if cap.MinNetCapital < 250_000 {
		t.Fatalf("expected BD min capital >= $250k, got $%.0f", cap.MinNetCapital)
	}
}

func TestBrokerDealerOperational(t *testing.T) {
	bd := &BrokerDealer{}
	ops := bd.OperationalRequirements()
	foundCustomer := false
	foundAML := false
	for _, o := range ops {
		if o.Name == "Customer Protection" {
			foundCustomer = true
		}
		if o.Name == "Anti-Money Laundering Program" {
			foundAML = true
		}
	}
	if !foundCustomer {
		t.Fatal("BD must have customer protection requirement")
	}
	if !foundAML {
		t.Fatal("BD must have AML program requirement")
	}
}

func TestTransferAgentInterface(t *testing.T) {
	var e RegulatedEntity = &TransferAgent{}
	if e.Name() != "Transfer Agent" {
		t.Fatalf("expected 'Transfer Agent', got %q", e.Name())
	}
	if e.Type() != "transfer_agent" {
		t.Fatalf("expected 'transfer_agent', got %q", e.Type())
	}
}

func TestTransferAgentLicenses(t *testing.T) {
	ta := &TransferAgent{}
	licenses := ta.RequiredLicenses()
	if len(licenses) == 0 {
		t.Fatal("expected licenses for TA")
	}
	if licenses[0].Regulator != "SEC" {
		t.Fatalf("expected SEC regulator, got %q", licenses[0].Regulator)
	}
}

func TestTransferAgentReporting(t *testing.T) {
	ta := &TransferAgent{}
	obligations := ta.ReportingObligations()
	if len(obligations) == 0 {
		t.Fatal("expected reporting obligations for TA")
	}
}

func TestTransferAgentOperational(t *testing.T) {
	ta := &TransferAgent{}
	ops := ta.OperationalRequirements()
	foundTurnaround := false
	for _, o := range ops {
		if o.Name == "Turnaround Performance" {
			foundTurnaround = true
		}
	}
	if !foundTurnaround {
		t.Fatal("TA must have turnaround performance requirement")
	}
}

func TestMSBInterface(t *testing.T) {
	var e RegulatedEntity = &MSB{}
	if e.Name() != "Money Service Business" {
		t.Fatalf("expected 'Money Service Business', got %q", e.Name())
	}
	if e.Type() != "msb" {
		t.Fatalf("expected 'msb', got %q", e.Type())
	}
}

func TestMSBLicenses(t *testing.T) {
	m := &MSB{}
	licenses := m.RequiredLicenses()
	if len(licenses) < 2 {
		t.Fatalf("expected at least 2 licenses for MSB, got %d", len(licenses))
	}
	foundFinCEN := false
	foundState := false
	for _, l := range licenses {
		if l.Regulator == "FinCEN" {
			foundFinCEN = true
		}
		if l.Regulator == "State Banking Regulators" {
			foundState = true
		}
	}
	if !foundFinCEN {
		t.Fatal("MSB must register with FinCEN")
	}
	if !foundState {
		t.Fatal("MSB must have state money transmitter licenses")
	}
}

func TestMSBReporting(t *testing.T) {
	m := &MSB{}
	obligations := m.ReportingObligations()
	foundCTR := false
	foundSAR := false
	for _, o := range obligations {
		if o.Name == "CTR Filing" {
			foundCTR = true
		}
		if o.Name == "SAR Filing" {
			foundSAR = true
		}
	}
	if !foundCTR {
		t.Fatal("MSB must file CTRs")
	}
	if !foundSAR {
		t.Fatal("MSB must file SARs")
	}
}

func TestMSBOperational(t *testing.T) {
	m := &MSB{}
	ops := m.OperationalRequirements()
	foundAML := false
	for _, o := range ops {
		if o.Name == "AML Compliance Program" {
			foundAML = true
		}
	}
	if !foundAML {
		t.Fatal("MSB must have AML compliance program")
	}
}

func TestGetEntity(t *testing.T) {
	cases := []struct {
		entityType string
		expected   string
	}{
		{"ats", "ats"},
		{"broker_dealer", "broker_dealer"},
		{"transfer_agent", "transfer_agent"},
		{"msb", "msb"},
	}
	for _, tc := range cases {
		e := GetEntity(tc.entityType)
		if e == nil {
			t.Fatalf("GetEntity(%q) returned nil", tc.entityType)
		}
		if e.Type() != tc.expected {
			t.Fatalf("expected type %q, got %q", tc.expected, e.Type())
		}
	}
}

func TestGetEntityUnknown(t *testing.T) {
	e := GetEntity("unknown")
	if e != nil {
		t.Fatal("expected nil for unknown entity type")
	}
}

func TestAllEntities(t *testing.T) {
	all := AllEntities()
	if len(all) != 13 {
		t.Fatalf("expected 13 entity types, got %d", len(all))
	}
	types := map[string]bool{}
	for _, e := range all {
		types[e.Type()] = true
	}
	for _, expected := range []string{
		EntityType_ATS, EntityType_BrokerDealer, EntityType_TransferAgent, EntityType_MSB,
		EntityType_SICAV, EntityType_SICAR, EntityType_RAIF, EntityType_AIFM,
		EntityType_MANCOMAN, EntityType_CRR, EntityType_ISSUER, EntityType_CUSTODIAN,
		EntityType_DLT_FACILITY,
	} {
		if !types[expected] {
			t.Errorf("missing entity type: %s", expected)
		}
	}
}
