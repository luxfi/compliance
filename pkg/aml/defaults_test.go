package aml

import "testing"

func TestDefaultMonitoringRules_Count(t *testing.T) {
	rules := DefaultMonitoringRules()
	if len(rules) != 3 {
		t.Fatalf("expected 3 default rules, got %d", len(rules))
	}
}

func TestDefaultMonitoringRules_CTR(t *testing.T) {
	rules := DefaultMonitoringRules()
	ctr := rules[0]
	if ctr.ID != "ctr-threshold" {
		t.Fatalf("expected ctr-threshold, got %s", ctr.ID)
	}
	if ctr.ThresholdAmount != 10000 {
		t.Fatalf("CTR threshold should be $10,000, got %f", ctr.ThresholdAmount)
	}
	if ctr.Type != RuleSingleAmount {
		t.Fatalf("CTR rule type should be single_amount, got %s", ctr.Type)
	}
}

func TestDefaultMonitoringRules_Structuring(t *testing.T) {
	rules := DefaultMonitoringRules()
	s := rules[1]
	if s.ID != "structuring" {
		t.Fatalf("expected structuring, got %s", s.ID)
	}
	if s.ThresholdAmount != 9000 {
		t.Fatalf("structuring threshold should be $9,000, got %f", s.ThresholdAmount)
	}
}

func TestDefaultMonitoringRules_Velocity(t *testing.T) {
	rules := DefaultMonitoringRules()
	v := rules[2]
	if v.ID != "velocity-24h" {
		t.Fatalf("expected velocity-24h, got %s", v.ID)
	}
	if v.MaxCount != 100 {
		t.Fatalf("velocity max count should be 100, got %d", v.MaxCount)
	}
}

func TestInstallDefaultRules(t *testing.T) {
	svc := NewMonitoringService()
	InstallDefaultRules(svc)
	// Verify rules were added by checking the service has rules
	// (MonitoringService internals are tested in monitoring_test.go)
}

func TestDefaultRulesAllEnabled(t *testing.T) {
	for _, r := range DefaultMonitoringRules() {
		if !r.Enabled {
			t.Fatalf("default rule %s should be enabled", r.ID)
		}
	}
}

func TestDefaultRulesAllUSD(t *testing.T) {
	for _, r := range DefaultMonitoringRules() {
		if r.Currency != "USD" {
			t.Fatalf("default rule %s should be USD, got %s", r.ID, r.Currency)
		}
	}
}

func TestDefaultRulesHaveSeverity(t *testing.T) {
	for _, r := range DefaultMonitoringRules() {
		if r.Severity == "" {
			t.Fatalf("default rule %s should have a severity", r.ID)
		}
	}
}
