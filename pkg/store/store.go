// Package store defines the ComplianceStore interface and its implementations.
package store

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/luxfi/compliance/pkg/types"
)

// ComplianceStore is the interface for compliance data persistence.
// MemoryStore and PostgresStore both implement this interface.
type ComplianceStore interface {
	// Identity
	SaveIdentity(id *types.Identity) error
	GetIdentity(id string) (*types.Identity, error)
	ListIdentitiesByUser(userID string) []*types.Identity

	// AML Screening
	SaveAMLScreening(s *types.AMLScreening) error
	GetAMLScreening(id string) (*types.AMLScreening, error)
	ListAMLScreeningsByAccount(accountID string) []*types.AMLScreening
	ListAMLScreeningsByStatus(status types.AMLStatus) []*types.AMLScreening

	// Application (onboarding)
	SaveApplication(app *types.Application) error
	GetApplication(id string) (*types.Application, error)
	GetApplicationByUser(userID string) (*types.Application, error)
	ListApplications() []*types.Application
	ListApplicationsByStatus(status types.ApplicationStatus) []*types.Application

	// Document Upload
	SaveDocumentUpload(doc *types.DocumentUpload) error
	GetDocumentUpload(id string) (*types.DocumentUpload, error)
	ListDocumentUploads(applicationID string) []*types.DocumentUpload

	// Pipeline
	SavePipeline(p *types.Pipeline) error
	GetPipeline(id string) (*types.Pipeline, error)
	ListPipelines() []*types.Pipeline
	DeletePipeline(id string) error

	// Session
	SaveSession(sess *types.Session) error
	GetSession(id string) (*types.Session, error)
	ListSessions() []*types.Session

	// Fund
	SaveFund(f *types.Fund) error
	GetFund(id string) (*types.Fund, error)
	ListFunds() []*types.Fund
	DeleteFund(id string) error
	AddFundInvestor(inv *types.FundInvestor) error
	ListFundInvestors(fundID string) []*types.FundInvestor

	// Envelope
	SaveEnvelope(env *types.Envelope) error
	GetEnvelope(id string) (*types.Envelope, error)
	ListEnvelopes() []*types.Envelope
	ListEnvelopesByDirection(direction string) []*types.Envelope

	// Template
	SaveTemplate(t *types.Template) error
	GetTemplate(id string) (*types.Template, error)
	ListTemplates() []*types.Template

	// Role
	SaveRole(role *types.Role) error
	GetRole(id string) (*types.Role, error)
	GetRoleByName(name string) (*types.Role, error)
	ListRoles() []*types.Role
	DeleteRole(id string) error

	// User
	SaveUser(u *types.User) error
	GetUser(id string) (*types.User, error)
	ListUsers() []*types.User

	// Transaction
	SaveTransaction(tx *types.Transaction) error
	ListTransactions() []*types.Transaction

	// Credential
	SaveCredential(c *types.Credential) error
	ListCredentials() []*types.Credential
	DeleteCredential(id string) error

	// Settings
	GetSettings() *types.Settings
	SaveSettings(settings *types.Settings)

	// Dashboard
	ComputeDashboard() *types.DashboardStats
	ComputeESignStats() *types.ESignStats
}

// GenerateID returns a random hex ID. Panics if the system CSPRNG fails,
// which indicates a catastrophic OS-level problem.
func GenerateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return hex.EncodeToString(b)
}
