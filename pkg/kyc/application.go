// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package kyc

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Status represents the lifecycle state of an application.
type Status string

const (
	StatusDraft      Status = "draft"
	StatusPending    Status = "pending"
	StatusPendingKYC Status = "pending_kyc"
	StatusApproved   Status = "approved"
	StatusRejected   Status = "rejected"
)

// KYCStatus represents the identity verification state.
type KYCStatus string

const (
	KYCNotStarted KYCStatus = "not_started"
	KYCPending    KYCStatus = "pending"
	KYCVerified   KYCStatus = "verified"
	KYCFailed     KYCStatus = "failed"
)

// Application is a brokerage/bank account application with full KYC lifecycle.
type Application struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id,omitempty"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`

	// Applicant identity
	GivenName   string `json:"given_name"`
	FamilyName  string `json:"family_name"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Email       string `json:"email"`
	Phone       string `json:"phone,omitempty"`

	// Address
	Street     []string `json:"street,omitempty"`
	City       string   `json:"city,omitempty"`
	State      string   `json:"state,omitempty"`
	PostalCode string   `json:"postal_code,omitempty"`
	Country    string   `json:"country,omitempty"`

	// Tax & regulatory
	TaxID        string `json:"tax_id,omitempty"`
	TaxIDType    string `json:"tax_id_type,omitempty"`    // ssn, itin, ein, nino, utr
	CountryOfTax string `json:"country_of_tax_residence,omitempty"`

	// Disclosures
	IsControlPerson        *bool `json:"is_control_person,omitempty"`
	IsAffiliatedExchange   *bool `json:"is_affiliated_exchange_or_finra,omitempty"`
	IsPoliticallyExposed   *bool `json:"is_politically_exposed,omitempty"`
	ImmediateFamilyExposed *bool `json:"immediate_family_exposed,omitempty"`

	// Employment
	EmploymentStatus string `json:"employment_status,omitempty"` // employed, unemployed, retired, student
	EmployerName     string `json:"employer_name,omitempty"`
	EmployerAddress  string `json:"employer_address,omitempty"`
	JobTitle         string `json:"job_title,omitempty"`

	// Financial
	AnnualIncome        string `json:"annual_income,omitempty"`        // range: <25k, 25k-50k, 50k-100k, 100k-200k, 200k-500k, >500k
	NetWorth            string `json:"net_worth,omitempty"`
	LiquidNetWorth      string `json:"liquid_net_worth,omitempty"`
	FundingSource       string `json:"funding_source,omitempty"`       // employment_income, investments, inheritance, savings, other
	InvestmentObjective string `json:"investment_objective,omitempty"` // growth, income, speculation, preservation, other

	// Account preferences
	AccountType   string   `json:"account_type,omitempty"` // individual, joint, ira, entity
	EnabledAssets []string `json:"enabled_assets,omitempty"`
	Provider      string   `json:"provider,omitempty"` // target broker/bank provider

	// KYC state
	KYCStatus     KYCStatus  `json:"kyc_status"`
	KYCProvider   string     `json:"kyc_provider,omitempty"`
	KYCReference  string     `json:"kyc_reference,omitempty"`
	KYCResult     string     `json:"kyc_result,omitempty"`
	KYCVerifiedAt *time.Time `json:"kyc_verified_at,omitempty"`

	// Documents
	Documents []Document `json:"documents,omitempty"`

	// Admin notes
	ReviewNotes string `json:"review_notes,omitempty"`
	ReviewedBy  string `json:"reviewed_by,omitempty"`

	// Source tracking
	Source    string            `json:"source,omitempty"` // web, api, import
	IPAddress string           `json:"ip_address,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// Document is a document attached to an application.
type Document struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`          // identity, proof_of_address, tax_form, source_of_funds
	SubType    string    `json:"sub_type,omitempty"` // passport, drivers_license, w8ben, utility_bill
	FileName   string    `json:"file_name,omitempty"`
	MimeType   string    `json:"mime_type,omitempty"`
	Status     string    `json:"status,omitempty"` // uploaded, verified, rejected
	UploadedAt time.Time `json:"uploaded_at"`
}

// Store is the application persistence layer.
type Store struct {
	mu   sync.RWMutex
	apps map[string]*Application
}

// NewStore creates an in-memory application store.
func NewStore() *Store {
	return &Store{apps: make(map[string]*Application)}
}

// Create persists a new application and assigns an ID.
func (s *Store) Create(app *Application) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if app.ID == "" {
		app.ID = newID()
	}
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now
	if app.Status == "" {
		app.Status = StatusDraft
	}
	if app.KYCStatus == "" {
		app.KYCStatus = KYCNotStarted
	}
	s.apps[app.ID] = app
	return nil
}

// Get returns a single application by ID.
func (s *Store) Get(id string) (*Application, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	app, ok := s.apps[id]
	if !ok {
		return nil, fmt.Errorf("application %s not found", id)
	}
	return app, nil
}

// Update modifies an existing application.
func (s *Store) Update(app *Application) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.apps[app.ID]; !ok {
		return fmt.Errorf("application %s not found", app.ID)
	}
	app.UpdatedAt = time.Now()
	s.apps[app.ID] = app
	return nil
}

// List returns applications, optionally filtered by status.
func (s *Store) List(status string) []*Application {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Application, 0, len(s.apps))
	for _, app := range s.apps {
		if status == "" || string(app.Status) == status {
			result = append(result, app)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result
}

// Stats returns aggregate statistics about applications.
func (s *Store) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := map[Status]int{}
	kycCounts := map[KYCStatus]int{}
	var total int

	for _, app := range s.apps {
		total++
		counts[app.Status]++
		kycCounts[app.KYCStatus]++
	}

	return map[string]interface{}{
		"total":           total,
		"draft":           counts[StatusDraft],
		"pending":         counts[StatusPending],
		"pending_kyc":     counts[StatusPendingKYC],
		"approved":        counts[StatusApproved],
		"rejected":        counts[StatusRejected],
		"kyc_not_started": kycCounts[KYCNotStarted],
		"kyc_pending":     kycCounts[KYCPending],
		"kyc_verified":    kycCounts[KYCVerified],
		"kyc_failed":      kycCounts[KYCFailed],
	}
}

// Count returns the total number of applications.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.apps)
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
