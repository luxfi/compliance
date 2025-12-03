// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package aml

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// AlertSeverity is the severity of a monitoring alert.
type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "low"
	SeverityMedium   AlertSeverity = "medium"
	SeverityHigh     AlertSeverity = "high"
	SeverityCritical AlertSeverity = "critical"
)

// AlertStatus is the lifecycle of an alert.
type AlertStatus string

const (
	AlertOpen       AlertStatus = "open"
	AlertInvestigating AlertStatus = "investigating"
	AlertEscalated  AlertStatus = "escalated"
	AlertClosed     AlertStatus = "closed"
	AlertFiled      AlertStatus = "filed" // SAR filed
)

// RuleType describes the type of monitoring rule.
type RuleType string

const (
	RuleSingleAmount   RuleType = "single_amount"    // single tx exceeds threshold
	RuleDailyAggregate RuleType = "daily_aggregate"   // daily total exceeds threshold
	RuleVelocity       RuleType = "velocity"          // too many transactions in window
	RuleGeographic     RuleType = "geographic"         // high-risk country
	RuleStructuring    RuleType = "structuring"        // pattern detection (just below threshold)
)

// Transaction is a financial transaction to monitor.
type Transaction struct {
	ID            string    `json:"id"`
	AccountID     string    `json:"account_id"`
	Type          string    `json:"type"`           // wire, ach, crypto, card
	Direction     string    `json:"direction"`       // inbound, outbound
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Country       string    `json:"country,omitempty"`
	CounterpartyID string  `json:"counterparty_id,omitempty"`
	CounterpartyName string `json:"counterparty_name,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	Description   string    `json:"description,omitempty"`
	Meta          map[string]string `json:"meta,omitempty"`
}

// MonitoringResult is the outcome of monitoring a transaction.
type MonitoringResult struct {
	TransactionID string  `json:"transaction_id"`
	Allowed       bool    `json:"allowed"`
	Alerts        []Alert `json:"alerts,omitempty"`
	ReviewRequired bool   `json:"review_required"`
}

// Alert is a triggered monitoring alert.
type Alert struct {
	ID            string        `json:"id"`
	TransactionID string        `json:"transaction_id"`
	AccountID     string        `json:"account_id"`
	RuleID        string        `json:"rule_id"`
	RuleType      RuleType      `json:"rule_type"`
	Severity      AlertSeverity `json:"severity"`
	Status        AlertStatus   `json:"status"`
	Description   string        `json:"description"`
	Details       map[string]interface{} `json:"details,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// Rule is a configurable monitoring rule.
type Rule struct {
	ID          string   `json:"id"`
	Type        RuleType `json:"rule_type"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`

	// For amount-based rules
	ThresholdAmount float64 `json:"threshold_amount,omitempty"`
	Currency        string  `json:"currency,omitempty"`

	// For velocity rules
	MaxCount int           `json:"max_count,omitempty"`
	Window   time.Duration `json:"window,omitempty"`

	// For geographic rules
	HighRiskCountries []string `json:"high_risk_countries,omitempty"`

	// For structuring rules
	StructuringThreshold float64 `json:"structuring_threshold,omitempty"` // e.g. 10000 for CTR
	StructuringMargin    float64 `json:"structuring_margin,omitempty"`    // e.g. 1000 (flag txns 9000-9999)
	StructuringWindow    time.Duration `json:"structuring_window,omitempty"`
	StructuringMinCount  int     `json:"structuring_min_count,omitempty"` // min txns in window to flag

	// Severity when triggered
	Severity AlertSeverity `json:"severity"`

	// Jurisdiction restriction (nil = all)
	Jurisdictions []string `json:"jurisdictions,omitempty"`
}

// SARReport is a Suspicious Activity Report reference.
type SARReport struct {
	ID          string    `json:"id"`
	AlertIDs    []string  `json:"alert_ids"`
	AccountID   string    `json:"account_id"`
	FilingDate  time.Time `json:"filing_date"`
	Narrative   string    `json:"narrative"`
	Status      string    `json:"status"` // draft, submitted, acknowledged
}

// MonitoringService provides real-time transaction monitoring.
type MonitoringService struct {
	mu     sync.RWMutex
	rules  map[string]*Rule           // ruleID -> Rule
	alerts map[string]*Alert          // alertID -> Alert
	txLog  map[string][]Transaction   // accountID -> recent transactions
	sars   map[string]*SARReport      // sarID -> SARReport
}

// NewMonitoringService creates a transaction monitoring service.
func NewMonitoringService() *MonitoringService {
	return &MonitoringService{
		rules:  make(map[string]*Rule),
		alerts: make(map[string]*Alert),
		txLog:  make(map[string][]Transaction),
		sars:   make(map[string]*SARReport),
	}
}

// AddRule adds a monitoring rule.
func (m *MonitoringService) AddRule(rule Rule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rule.ID == "" {
		rule.ID = newAlertID("rule")
	}
	m.rules[rule.ID] = &rule
}

// Monitor evaluates a transaction against all active rules.
func (m *MonitoringService) Monitor(ctx context.Context, tx *Transaction) (*MonitoringResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tx.ID == "" {
		return nil, fmt.Errorf("transaction ID is required")
	}

	// Record transaction
	m.txLog[tx.AccountID] = append(m.txLog[tx.AccountID], *tx)

	result := &MonitoringResult{
		TransactionID: tx.ID,
		Allowed:       true,
		Alerts:        []Alert{},
	}

	for _, rule := range m.rules {
		if !rule.Enabled {
			continue
		}

		alerts := m.evaluateRule(rule, tx)
		for i := range alerts {
			m.alerts[alerts[i].ID] = &alerts[i]
		}
		result.Alerts = append(result.Alerts, alerts...)
	}

	if len(result.Alerts) > 0 {
		result.ReviewRequired = true
		// Block transaction if any critical alerts
		for _, a := range result.Alerts {
			if a.Severity == SeverityCritical {
				result.Allowed = false
				break
			}
		}
	}

	return result, nil
}

func (m *MonitoringService) evaluateRule(rule *Rule, tx *Transaction) []Alert {
	switch rule.Type {
	case RuleSingleAmount:
		return m.checkSingleAmount(rule, tx)
	case RuleDailyAggregate:
		return m.checkDailyAggregate(rule, tx)
	case RuleVelocity:
		return m.checkVelocity(rule, tx)
	case RuleGeographic:
		return m.checkGeographic(rule, tx)
	case RuleStructuring:
		return m.checkStructuring(rule, tx)
	default:
		return nil
	}
}

func (m *MonitoringService) checkSingleAmount(rule *Rule, tx *Transaction) []Alert {
	if tx.Amount >= rule.ThresholdAmount {
		return []Alert{m.createAlert(rule, tx, fmt.Sprintf(
			"Single transaction $%.2f exceeds threshold $%.2f",
			tx.Amount, rule.ThresholdAmount,
		), map[string]interface{}{
			"amount":    tx.Amount,
			"threshold": rule.ThresholdAmount,
		})}
	}
	return nil
}

func (m *MonitoringService) checkDailyAggregate(rule *Rule, tx *Transaction) []Alert {
	window := 24 * time.Hour
	cutoff := tx.Timestamp.Add(-window)
	total := tx.Amount
	for _, past := range m.txLog[tx.AccountID] {
		if past.ID != tx.ID && past.Timestamp.After(cutoff) {
			total += past.Amount
		}
	}
	if total >= rule.ThresholdAmount {
		return []Alert{m.createAlert(rule, tx, fmt.Sprintf(
			"Daily aggregate $%.2f exceeds threshold $%.2f",
			total, rule.ThresholdAmount,
		), map[string]interface{}{
			"daily_total": total,
			"threshold":   rule.ThresholdAmount,
		})}
	}
	return nil
}

func (m *MonitoringService) checkVelocity(rule *Rule, tx *Transaction) []Alert {
	window := rule.Window
	if window == 0 {
		window = time.Hour
	}
	cutoff := tx.Timestamp.Add(-window)
	count := 0
	for _, past := range m.txLog[tx.AccountID] {
		if past.Timestamp.After(cutoff) {
			count++
		}
	}
	if count >= rule.MaxCount {
		return []Alert{m.createAlert(rule, tx, fmt.Sprintf(
			"Velocity: %d transactions in %v (limit %d)",
			count, window, rule.MaxCount,
		), map[string]interface{}{
			"count":     count,
			"window":    window.String(),
			"max_count": rule.MaxCount,
		})}
	}
	return nil
}

func (m *MonitoringService) checkGeographic(rule *Rule, tx *Transaction) []Alert {
	if tx.Country == "" {
		return nil
	}
	for _, c := range rule.HighRiskCountries {
		if tx.Country == c {
			return []Alert{m.createAlert(rule, tx, fmt.Sprintf(
				"Transaction involves high-risk country: %s", tx.Country,
			), map[string]interface{}{
				"country": tx.Country,
			})}
		}
	}
	return nil
}

func (m *MonitoringService) checkStructuring(rule *Rule, tx *Transaction) []Alert {
	threshold := rule.StructuringThreshold
	margin := rule.StructuringMargin
	if threshold == 0 || margin == 0 {
		return nil
	}

	// Check if this transaction is just below the threshold
	lower := threshold - margin
	if tx.Amount >= lower && tx.Amount < threshold {
		// Check if there are multiple such transactions in the window
		window := rule.StructuringWindow
		if window == 0 {
			window = 24 * time.Hour
		}
		minCount := rule.StructuringMinCount
		if minCount == 0 {
			minCount = 3
		}
		cutoff := tx.Timestamp.Add(-window)
		belowCount := 0
		for _, past := range m.txLog[tx.AccountID] {
			if past.Timestamp.After(cutoff) && past.Amount >= lower && past.Amount < threshold {
				belowCount++
			}
		}
		if belowCount >= minCount {
			return []Alert{m.createAlert(rule, tx, fmt.Sprintf(
				"Potential structuring: %d transactions between $%.2f-$%.2f in %v",
				belowCount, lower, threshold, window,
			), map[string]interface{}{
				"below_threshold_count": belowCount,
				"threshold":             threshold,
				"margin":                margin,
				"window":                window.String(),
			})}
		}
	}
	return nil
}

func (m *MonitoringService) createAlert(rule *Rule, tx *Transaction, desc string, details map[string]interface{}) Alert {
	now := time.Now()
	return Alert{
		ID:            newAlertID("alert"),
		TransactionID: tx.ID,
		AccountID:     tx.AccountID,
		RuleID:        rule.ID,
		RuleType:      rule.Type,
		Severity:      rule.Severity,
		Status:        AlertOpen,
		Description:   desc,
		Details:       details,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// GetAlerts returns alerts, optionally filtered by status.
func (m *MonitoringService) GetAlerts(status string) []Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Alert, 0, len(m.alerts))
	for _, a := range m.alerts {
		if status == "" || string(a.Status) == status {
			result = append(result, *a)
		}
	}
	return result
}

// GetAlert returns a single alert by ID.
func (m *MonitoringService) GetAlert(id string) (*Alert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.alerts[id]
	if !ok {
		return nil, fmt.Errorf("alert %s not found", id)
	}
	return a, nil
}

// UpdateAlertStatus updates the status of an alert.
func (m *MonitoringService) UpdateAlertStatus(id string, status AlertStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alerts[id]
	if !ok {
		return fmt.Errorf("alert %s not found", id)
	}
	a.Status = status
	a.UpdatedAt = time.Now()
	return nil
}

// CreateSAR creates a Suspicious Activity Report from alerts.
func (m *MonitoringService) CreateSAR(alertIDs []string, accountID, narrative string) (*SARReport, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate alert IDs
	for _, id := range alertIDs {
		if _, ok := m.alerts[id]; !ok {
			return nil, fmt.Errorf("alert %s not found", id)
		}
	}

	sar := &SARReport{
		ID:         newAlertID("sar"),
		AlertIDs:   alertIDs,
		AccountID:  accountID,
		FilingDate: time.Now(),
		Narrative:  narrative,
		Status:     "draft",
	}

	m.sars[sar.ID] = sar

	// Mark alerts as filed
	for _, id := range alertIDs {
		if a, ok := m.alerts[id]; ok {
			a.Status = AlertFiled
			a.UpdatedAt = time.Now()
		}
	}

	return sar, nil
}

// GetRules returns all configured rules.
func (m *MonitoringService) GetRules() []Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Rule, 0, len(m.rules))
	for _, r := range m.rules {
		result = append(result, *r)
	}
	return result
}

func newAlertID(prefix string) string {
	b := make([]byte, 12)
	rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}
