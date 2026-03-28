package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/luxfi/compliance/pkg/types"
)

// MemoryStore is an in-memory compliance data store.
// Use PostgresStore in production.
type MemoryStore struct {
	mu              sync.RWMutex
	identities      map[string]*types.Identity
	businesses      map[string]*types.BusinessKYB
	pipelines       map[string]*types.Pipeline
	sessions        map[string]*types.Session
	funds           map[string]*types.Fund
	investors       map[string][]*types.FundInvestor
	envelopes       map[string]*types.Envelope
	templates       map[string]*types.Template
	roles           map[string]*types.Role
	users           map[string]*types.User
	transactions    map[string]*types.Transaction
	credentials     map[string]*types.Credential
	amlScreenings   map[string]*types.AMLScreening
	applications    map[string]*types.Application
	documentUploads map[string]*types.DocumentUpload
	settings        *types.Settings
}

// NewMemoryStore creates an empty in-memory compliance store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		identities:      make(map[string]*types.Identity),
		businesses:      make(map[string]*types.BusinessKYB),
		pipelines:       make(map[string]*types.Pipeline),
		sessions:        make(map[string]*types.Session),
		funds:           make(map[string]*types.Fund),
		investors:       make(map[string][]*types.FundInvestor),
		envelopes:       make(map[string]*types.Envelope),
		templates:       make(map[string]*types.Template),
		roles:           make(map[string]*types.Role),
		users:           make(map[string]*types.User),
		transactions:    make(map[string]*types.Transaction),
		credentials:     make(map[string]*types.Credential),
		amlScreenings:   make(map[string]*types.AMLScreening),
		applications:    make(map[string]*types.Application),
		documentUploads: make(map[string]*types.DocumentUpload),
		settings: &types.Settings{
			BusinessName:      "Your Company",
			Timezone:          "America/New_York",
			Currency:          "USD",
			NotificationEmail: "compliance@example.com",
		},
	}
}

// --- Identity ---

func (s *MemoryStore) SaveIdentity(id *types.Identity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id.ID == "" {
		id.ID = GenerateID()
	}
	if id.CreatedAt.IsZero() {
		id.CreatedAt = time.Now().UTC()
	}
	id.UpdatedAt = time.Now().UTC()
	s.identities[id.ID] = id
	return nil
}

func (s *MemoryStore) GetIdentity(id string) (*types.Identity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ident, ok := s.identities[id]
	if !ok {
		return nil, fmt.Errorf("identity not found")
	}
	return ident, nil
}

func (s *MemoryStore) ListIdentitiesByUser(userID string) []*types.Identity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*types.Identity
	for _, ident := range s.identities {
		if ident.UserID == userID {
			out = append(out, ident)
		}
	}
	if out == nil {
		out = make([]*types.Identity, 0)
	}
	return out
}

// --- Pipeline ---

func (s *MemoryStore) SavePipeline(p *types.Pipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		p.ID = GenerateID()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	p.UpdatedAt = time.Now().UTC()
	s.pipelines[p.ID] = p
	return nil
}

func (s *MemoryStore) GetPipeline(id string) (*types.Pipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.pipelines[id]
	if !ok {
		return nil, fmt.Errorf("pipeline not found")
	}
	return p, nil
}

func (s *MemoryStore) ListPipelines() []*types.Pipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*types.Pipeline, 0, len(s.pipelines))
	for _, p := range s.pipelines {
		out = append(out, p)
	}
	return out
}

func (s *MemoryStore) DeletePipeline(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.pipelines[id]; !ok {
		return fmt.Errorf("pipeline not found")
	}
	delete(s.pipelines, id)
	return nil
}

// --- Session ---

func (s *MemoryStore) SaveSession(sess *types.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess.ID == "" {
		sess.ID = GenerateID()
	}
	if sess.CreatedAt.IsZero() {
		sess.CreatedAt = time.Now().UTC()
	}
	s.sessions[sess.ID] = sess
	return nil
}

func (s *MemoryStore) GetSession(id string) (*types.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return sess, nil
}

func (s *MemoryStore) ListSessions() []*types.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*types.Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		out = append(out, sess)
	}
	return out
}

// --- Fund ---

func (s *MemoryStore) SaveFund(f *types.Fund) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if f.ID == "" {
		f.ID = GenerateID()
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}
	f.UpdatedAt = time.Now().UTC()
	s.funds[f.ID] = f
	return nil
}

func (s *MemoryStore) GetFund(id string) (*types.Fund, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.funds[id]
	if !ok {
		return nil, fmt.Errorf("fund not found")
	}
	return f, nil
}

func (s *MemoryStore) ListFunds() []*types.Fund {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*types.Fund, 0, len(s.funds))
	for _, f := range s.funds {
		out = append(out, f)
	}
	return out
}

func (s *MemoryStore) DeleteFund(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.funds[id]; !ok {
		return fmt.Errorf("fund not found")
	}
	delete(s.funds, id)
	delete(s.investors, id)
	return nil
}

func (s *MemoryStore) AddFundInvestor(inv *types.FundInvestor) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if inv.ID == "" {
		inv.ID = GenerateID()
	}
	if inv.CreatedAt.IsZero() {
		inv.CreatedAt = time.Now().UTC()
	}
	f, ok := s.funds[inv.FundID]
	if !ok {
		return fmt.Errorf("fund not found")
	}
	s.investors[inv.FundID] = append(s.investors[inv.FundID], inv)
	f.InvestorCount = len(s.investors[inv.FundID])
	f.TotalRaised += inv.Amount
	return nil
}

func (s *MemoryStore) ListFundInvestors(fundID string) []*types.FundInvestor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.investors[fundID]
}

// --- Envelope ---

func (s *MemoryStore) SaveEnvelope(env *types.Envelope) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if env.ID == "" {
		env.ID = GenerateID()
	}
	if env.CreatedAt.IsZero() {
		env.CreatedAt = time.Now().UTC()
	}
	env.UpdatedAt = time.Now().UTC()
	s.envelopes[env.ID] = env
	return nil
}

func (s *MemoryStore) GetEnvelope(id string) (*types.Envelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	env, ok := s.envelopes[id]
	if !ok {
		return nil, fmt.Errorf("envelope not found")
	}
	return env, nil
}

func (s *MemoryStore) ListEnvelopes() []*types.Envelope {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*types.Envelope, 0, len(s.envelopes))
	for _, env := range s.envelopes {
		out = append(out, env)
	}
	return out
}

func (s *MemoryStore) ListEnvelopesByDirection(direction string) []*types.Envelope {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*types.Envelope, 0)
	for _, env := range s.envelopes {
		switch direction {
		case "inbox":
			if env.Status == types.EnvelopeSent || env.Status == types.EnvelopePending || env.Status == types.EnvelopeViewed {
				out = append(out, env)
			}
		case "sent":
			if env.Status == types.EnvelopeSigned || env.Status == types.EnvelopeCompleted || env.Status == types.EnvelopeDeclined || env.Status == types.EnvelopeVoided {
				out = append(out, env)
			}
		}
	}
	return out
}

// --- Template ---

func (s *MemoryStore) SaveTemplate(t *types.Template) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.ID == "" {
		t.ID = GenerateID()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	s.templates[t.ID] = t
	return nil
}

func (s *MemoryStore) GetTemplate(id string) (*types.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.templates[id]
	if !ok {
		return nil, fmt.Errorf("template not found")
	}
	return t, nil
}

func (s *MemoryStore) ListTemplates() []*types.Template {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*types.Template, 0, len(s.templates))
	for _, t := range s.templates {
		out = append(out, t)
	}
	return out
}

// --- Role ---

func (s *MemoryStore) SaveRole(role *types.Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if role.ID == "" {
		role.ID = GenerateID()
	}
	if role.CreatedAt.IsZero() {
		role.CreatedAt = time.Now().UTC()
	}
	role.UpdatedAt = time.Now().UTC()
	s.roles[role.ID] = role
	return nil
}

func (s *MemoryStore) GetRole(id string) (*types.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	role, ok := s.roles[id]
	if !ok {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

func (s *MemoryStore) GetRoleByName(name string) (*types.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, role := range s.roles {
		if role.Name == name {
			return role, nil
		}
	}
	return nil, fmt.Errorf("role not found: %s", name)
}

func (s *MemoryStore) ListRoles() []*types.Role {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*types.Role, 0, len(s.roles))
	for _, role := range s.roles {
		out = append(out, role)
	}
	return out
}

func (s *MemoryStore) DeleteRole(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.roles[id]; !ok {
		return fmt.Errorf("role not found")
	}
	delete(s.roles, id)
	return nil
}

// --- User ---

func (s *MemoryStore) SaveUser(u *types.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u.ID == "" {
		u.ID = GenerateID()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	s.users[u.ID] = u
	return nil
}

func (s *MemoryStore) GetUser(id string) (*types.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (s *MemoryStore) ListUsers() []*types.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*types.User, 0, len(s.users))
	for _, u := range s.users {
		out = append(out, u)
	}
	return out
}

// --- Transaction ---

func (s *MemoryStore) SaveTransaction(tx *types.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if tx.ID == "" {
		tx.ID = GenerateID()
	}
	s.transactions[tx.ID] = tx
	return nil
}

func (s *MemoryStore) ListTransactions() []*types.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*types.Transaction, 0, len(s.transactions))
	for _, tx := range s.transactions {
		out = append(out, tx)
	}
	return out
}

// --- Credential ---

func (s *MemoryStore) SaveCredential(c *types.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.ID == "" {
		c.ID = GenerateID()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	s.credentials[c.ID] = c
	return nil
}

func (s *MemoryStore) ListCredentials() []*types.Credential {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*types.Credential, 0, len(s.credentials))
	for _, c := range s.credentials {
		out = append(out, c)
	}
	return out
}

func (s *MemoryStore) DeleteCredential(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.credentials[id]; !ok {
		return fmt.Errorf("credential not found")
	}
	delete(s.credentials, id)
	return nil
}

// --- Settings ---

func (s *MemoryStore) GetSettings() *types.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := *s.settings
	return &cp
}

func (s *MemoryStore) SaveSettings(settings *types.Settings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = settings
}

// --- AML Screening ---

func (s *MemoryStore) SaveAMLScreening(sc *types.AMLScreening) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sc.ID == "" {
		sc.ID = GenerateID()
	}
	if sc.CreatedAt.IsZero() {
		sc.CreatedAt = time.Now().UTC()
	}
	sc.UpdatedAt = time.Now().UTC()
	s.amlScreenings[sc.ID] = sc
	return nil
}

func (s *MemoryStore) GetAMLScreening(id string) (*types.AMLScreening, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok := s.amlScreenings[id]
	if !ok {
		return nil, fmt.Errorf("aml screening not found")
	}
	return sc, nil
}

func (s *MemoryStore) ListAMLScreeningsByAccount(accountID string) []*types.AMLScreening {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*types.AMLScreening
	for _, sc := range s.amlScreenings {
		if sc.AccountID == accountID {
			out = append(out, sc)
		}
	}
	if out == nil {
		out = make([]*types.AMLScreening, 0)
	}
	return out
}

func (s *MemoryStore) ListAMLScreeningsByStatus(status types.AMLStatus) []*types.AMLScreening {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*types.AMLScreening
	for _, sc := range s.amlScreenings {
		if sc.Status == status {
			out = append(out, sc)
		}
	}
	if out == nil {
		out = make([]*types.AMLScreening, 0)
	}
	return out
}

// --- Application ---

func (s *MemoryStore) SaveApplication(app *types.Application) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if app.ID == "" {
		app.ID = GenerateID()
	}
	if app.CreatedAt.IsZero() {
		app.CreatedAt = time.Now().UTC()
	}
	app.UpdatedAt = time.Now().UTC()
	s.applications[app.ID] = app
	return nil
}

func (s *MemoryStore) GetApplication(id string) (*types.Application, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	app, ok := s.applications[id]
	if !ok {
		return nil, fmt.Errorf("application not found")
	}
	return app, nil
}

func (s *MemoryStore) GetApplicationByUser(userID string) (*types.Application, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, app := range s.applications {
		if app.UserID == userID {
			return app, nil
		}
	}
	return nil, fmt.Errorf("application not found")
}

func (s *MemoryStore) ListApplications() []*types.Application {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*types.Application, 0, len(s.applications))
	for _, app := range s.applications {
		out = append(out, app)
	}
	return out
}

func (s *MemoryStore) ListApplicationsByStatus(status types.ApplicationStatus) []*types.Application {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*types.Application
	for _, app := range s.applications {
		if app.Status == status {
			out = append(out, app)
		}
	}
	if out == nil {
		out = make([]*types.Application, 0)
	}
	return out
}

// --- Document Upload ---

func (s *MemoryStore) SaveDocumentUpload(doc *types.DocumentUpload) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if doc.ID == "" {
		doc.ID = GenerateID()
	}
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now().UTC()
	}
	doc.UpdatedAt = time.Now().UTC()
	s.documentUploads[doc.ID] = doc
	return nil
}

func (s *MemoryStore) GetDocumentUpload(id string) (*types.DocumentUpload, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	doc, ok := s.documentUploads[id]
	if !ok {
		return nil, fmt.Errorf("document not found")
	}
	return doc, nil
}

func (s *MemoryStore) ListDocumentUploads(applicationID string) []*types.DocumentUpload {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*types.DocumentUpload
	for _, doc := range s.documentUploads {
		if doc.ApplicationID == applicationID {
			out = append(out, doc)
		}
	}
	if out == nil {
		out = make([]*types.DocumentUpload, 0)
	}
	return out
}

// --- Dashboard ---

func (s *MemoryStore) ComputeDashboard() *types.DashboardStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &types.DashboardStats{
		TotalFunds:          len(s.funds),
		MonthlyTransactions: len(s.transactions),
	}

	for _, sess := range s.sessions {
		if sess.Status == types.SessionPending || sess.Status == types.SessionInProgress {
			stats.ActiveSessions++
		}
		if sess.KYCStatus == types.KYCPending {
			stats.PendingKYC++
		}
	}

	stats.RecentSessions = make([]*types.Session, 0, 5)
	count := 0
	for _, sess := range s.sessions {
		if count >= 5 {
			break
		}
		stats.RecentSessions = append(stats.RecentSessions, sess)
		count++
	}

	stats.RecentTransactions = make([]*types.Transaction, 0, 5)
	count = 0
	for _, tx := range s.transactions {
		if count >= 5 {
			break
		}
		stats.RecentTransactions = append(stats.RecentTransactions, tx)
		count++
	}

	return stats
}

func (s *MemoryStore) ComputeESignStats() *types.ESignStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &types.ESignStats{
		Templates: len(s.templates),
	}

	for _, env := range s.envelopes {
		switch env.Status {
		case types.EnvelopePending, types.EnvelopeSent, types.EnvelopeViewed:
			stats.Pending++
		case types.EnvelopeCompleted, types.EnvelopeSigned:
			stats.Completed++
		}
	}
	stats.Draft = stats.Pending

	stats.Recent = make([]*types.Envelope, 0, 5)
	count := 0
	for _, env := range s.envelopes {
		if count >= 5 {
			break
		}
		stats.Recent = append(stats.Recent, env)
		count++
	}

	return stats
}
