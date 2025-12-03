// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package aml

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMonitorSingleAmountAlert(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID:              "r1",
		Type:            RuleSingleAmount,
		ThresholdAmount: 10_000,
		Enabled:         true,
		Severity:        SeverityHigh,
	})

	tx := &Transaction{
		ID:        "tx-1",
		AccountID: "acct-1",
		Amount:    15_000,
		Timestamp: time.Now(),
	}

	result, err := svc.Monitor(context.Background(), tx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Alerts) == 0 {
		t.Fatal("expected alert for amount over threshold")
	}
	if result.Alerts[0].RuleType != RuleSingleAmount {
		t.Fatalf("expected single_amount rule, got %q", result.Alerts[0].RuleType)
	}
	if result.Alerts[0].Severity != SeverityHigh {
		t.Fatalf("expected high severity, got %q", result.Alerts[0].Severity)
	}
}

func TestMonitorSingleAmountBelowThreshold(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID:              "r1",
		Type:            RuleSingleAmount,
		ThresholdAmount: 10_000,
		Enabled:         true,
		Severity:        SeverityMedium,
	})

	tx := &Transaction{
		ID:        "tx-1",
		AccountID: "acct-1",
		Amount:    5_000,
		Timestamp: time.Now(),
	}

	result, err := svc.Monitor(context.Background(), tx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(result.Alerts))
	}
	if !result.Allowed {
		t.Fatal("expected transaction to be allowed")
	}
}

func TestMonitorDailyAggregate(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID:              "r2",
		Type:            RuleDailyAggregate,
		ThresholdAmount: 20_000,
		Enabled:         true,
		Severity:        SeverityHigh,
	})

	now := time.Now()
	// First transaction
	svc.Monitor(context.Background(), &Transaction{
		ID: "tx-1", AccountID: "acct-1", Amount: 8_000, Timestamp: now,
	})
	// Second transaction
	svc.Monitor(context.Background(), &Transaction{
		ID: "tx-2", AccountID: "acct-1", Amount: 7_000, Timestamp: now,
	})
	// Third transaction pushes over daily limit
	result, err := svc.Monitor(context.Background(), &Transaction{
		ID: "tx-3", AccountID: "acct-1", Amount: 6_000, Timestamp: now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Alerts) == 0 {
		t.Fatal("expected daily aggregate alert")
	}
}

func TestMonitorVelocity(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID:       "r3",
		Type:     RuleVelocity,
		MaxCount: 3,
		Window:   time.Hour,
		Enabled:  true,
		Severity: SeverityMedium,
	})

	now := time.Now()
	for i := 0; i < 3; i++ {
		svc.Monitor(context.Background(), &Transaction{
			ID:        fmt.Sprintf("tx-v-%d", i),
			AccountID: "acct-1",
			Amount:    100,
			Timestamp: now,
		})
	}

	// 4th transaction should trigger velocity
	result, err := svc.Monitor(context.Background(), &Transaction{
		ID: "tx-v-3", AccountID: "acct-1", Amount: 100, Timestamp: now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, a := range result.Alerts {
		if a.RuleType == RuleVelocity {
			found = true
		}
	}
	if !found {
		t.Fatal("expected velocity alert")
	}
}

func TestMonitorGeographic(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID:                "r4",
		Type:              RuleGeographic,
		HighRiskCountries: []string{"IR", "KP", "SY"},
		Enabled:           true,
		Severity:          SeverityCritical,
	})

	result, err := svc.Monitor(context.Background(), &Transaction{
		ID: "tx-g-1", AccountID: "acct-1", Amount: 1000, Country: "IR", Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Alerts) == 0 {
		t.Fatal("expected geographic alert for Iran")
	}
	if result.Alerts[0].Severity != SeverityCritical {
		t.Fatalf("expected critical severity, got %q", result.Alerts[0].Severity)
	}
	if result.Allowed {
		t.Fatal("expected transaction blocked for critical alert")
	}
}

func TestMonitorGeographicSafeCountry(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID:                "r4",
		Type:              RuleGeographic,
		HighRiskCountries: []string{"IR", "KP"},
		Enabled:           true,
		Severity:          SeverityCritical,
	})

	result, err := svc.Monitor(context.Background(), &Transaction{
		ID: "tx-g-2", AccountID: "acct-1", Amount: 1000, Country: "US", Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Alerts) != 0 {
		t.Fatalf("expected no alerts for US, got %d", len(result.Alerts))
	}
}

func TestMonitorStructuring(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID:                   "r5",
		Type:                 RuleStructuring,
		StructuringThreshold: 10_000,
		StructuringMargin:    1_000,
		StructuringWindow:    24 * time.Hour,
		StructuringMinCount:  3,
		Enabled:              true,
		Severity:             SeverityHigh,
	})

	now := time.Now()
	// Multiple transactions just below $10k
	for i := 0; i < 3; i++ {
		svc.Monitor(context.Background(), &Transaction{
			ID:        fmt.Sprintf("tx-s-%d", i),
			AccountID: "acct-1",
			Amount:    9_500,
			Timestamp: now,
		})
	}

	// 4th should trigger structuring detection
	result, err := svc.Monitor(context.Background(), &Transaction{
		ID: "tx-s-3", AccountID: "acct-1", Amount: 9_800, Timestamp: now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, a := range result.Alerts {
		if a.RuleType == RuleStructuring {
			found = true
		}
	}
	if !found {
		t.Fatal("expected structuring alert")
	}
}

func TestMonitorDisabledRule(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID:              "r-disabled",
		Type:            RuleSingleAmount,
		ThresholdAmount: 100,
		Enabled:         false,
		Severity:        SeverityHigh,
	})

	result, err := svc.Monitor(context.Background(), &Transaction{
		ID: "tx-d", AccountID: "acct-1", Amount: 1000, Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Alerts) != 0 {
		t.Fatal("expected no alerts for disabled rule")
	}
}

func TestMonitorNoID(t *testing.T) {
	svc := NewMonitoringService()
	_, err := svc.Monitor(context.Background(), &Transaction{AccountID: "acct-1"})
	if err == nil {
		t.Fatal("expected error for missing transaction ID")
	}
}

func TestGetAlerts(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID:              "r1",
		Type:            RuleSingleAmount,
		ThresholdAmount: 100,
		Enabled:         true,
		Severity:        SeverityMedium,
	})

	svc.Monitor(context.Background(), &Transaction{
		ID: "tx-1", AccountID: "acct-1", Amount: 200, Timestamp: time.Now(),
	})

	alerts := svc.GetAlerts("")
	if len(alerts) == 0 {
		t.Fatal("expected at least one alert")
	}
	if alerts[0].Status != AlertOpen {
		t.Fatalf("expected open status, got %q", alerts[0].Status)
	}

	// Filter by status
	open := svc.GetAlerts("open")
	if len(open) != len(alerts) {
		t.Fatal("expected same count for open filter")
	}
	closed := svc.GetAlerts("closed")
	if len(closed) != 0 {
		t.Fatalf("expected 0 closed alerts, got %d", len(closed))
	}
}

func TestUpdateAlertStatus(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID: "r1", Type: RuleSingleAmount, ThresholdAmount: 100, Enabled: true, Severity: SeverityMedium,
	})
	svc.Monitor(context.Background(), &Transaction{
		ID: "tx-1", AccountID: "acct-1", Amount: 200, Timestamp: time.Now(),
	})

	alerts := svc.GetAlerts("")
	err := svc.UpdateAlertStatus(alerts[0].ID, AlertInvestigating)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a, _ := svc.GetAlert(alerts[0].ID)
	if a.Status != AlertInvestigating {
		t.Fatalf("expected investigating, got %q", a.Status)
	}
}

func TestUpdateAlertStatusNotFound(t *testing.T) {
	svc := NewMonitoringService()
	err := svc.UpdateAlertStatus("nonexistent", AlertClosed)
	if err == nil {
		t.Fatal("expected error for nonexistent alert")
	}
}

func TestCreateSAR(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID: "r1", Type: RuleSingleAmount, ThresholdAmount: 100, Enabled: true, Severity: SeverityMedium,
	})
	svc.Monitor(context.Background(), &Transaction{
		ID: "tx-1", AccountID: "acct-1", Amount: 200, Timestamp: time.Now(),
	})

	alerts := svc.GetAlerts("")
	sar, err := svc.CreateSAR([]string{alerts[0].ID}, "acct-1", "Suspicious large transaction")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sar.Status != "draft" {
		t.Fatalf("expected draft status, got %q", sar.Status)
	}

	// Alert should be marked as filed
	a, _ := svc.GetAlert(alerts[0].ID)
	if a.Status != AlertFiled {
		t.Fatalf("expected filed status, got %q", a.Status)
	}
}

func TestCreateSARInvalidAlert(t *testing.T) {
	svc := NewMonitoringService()
	_, err := svc.CreateSAR([]string{"nonexistent"}, "acct-1", "test")
	if err == nil {
		t.Fatal("expected error for nonexistent alert")
	}
}

func TestGetRules(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{ID: "r1", Type: RuleSingleAmount, Enabled: true})
	svc.AddRule(Rule{ID: "r2", Type: RuleVelocity, Enabled: true})

	rules := svc.GetRules()
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}

func TestCriticalAlertBlocksTransaction(t *testing.T) {
	svc := NewMonitoringService()
	svc.AddRule(Rule{
		ID:                "r-crit",
		Type:              RuleGeographic,
		HighRiskCountries: []string{"KP"},
		Enabled:           true,
		Severity:          SeverityCritical,
	})

	result, err := svc.Monitor(context.Background(), &Transaction{
		ID: "tx-kp", AccountID: "acct-1", Amount: 100, Country: "KP", Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Allowed {
		t.Fatal("expected transaction to be blocked for critical alert")
	}
	if !result.ReviewRequired {
		t.Fatal("expected review required")
	}
}

func TestAlertConstants(t *testing.T) {
	if AlertOpen != "open" {
		t.Fatalf("AlertOpen = %q", AlertOpen)
	}
	if AlertFiled != "filed" {
		t.Fatalf("AlertFiled = %q", AlertFiled)
	}
}
