package aml

// DefaultMonitoringRules returns the standard AML monitoring rules mandated
// by FinCEN regulations. Any ATS, BD, or TA should install these as a baseline.
//
//   - CTR threshold: Flag transactions over $10,000 (31 CFR 1010.311)
//   - Structuring: Detect potential structuring at $9,000-$9,999 (31 USC 5324)
//   - Velocity: Block if 24h volume exceeds $50,000 without enhanced KYC
func DefaultMonitoringRules() []Rule {
	return []Rule{
		{
			ID:              "ctr-threshold",
			Type:            RuleSingleAmount,
			ThresholdAmount: 10000,
			Currency:        "USD",
			Enabled:         true,
			Description:     "Flag transactions over $10,000 (CTR threshold per 31 CFR 1010.311)",
		},
		{
			ID:              "structuring",
			Type:            RuleStructuring,
			ThresholdAmount: 9000,
			Currency:        "USD",
			Enabled:         true,
			Description:     "Detect potential structuring ($9,000-$9,999 per 31 USC 5324)",
		},
		{
			ID:              "velocity-24h",
			Type:            RuleVelocity,
			ThresholdAmount: 50000,
			Currency:        "USD",
			MaxCount:        100,
			Window:          24 * 60 * 60 * 1e9, // 24 hours in nanoseconds
			Enabled:         true,
			Description:     "Block if 24h volume exceeds $50,000 without enhanced KYC",
		},
	}
}

// InstallDefaultRules adds all DefaultMonitoringRules to the given service.
func InstallDefaultRules(svc *MonitoringService) {
	for _, r := range DefaultMonitoringRules() {
		svc.AddRule(r)
	}
}
