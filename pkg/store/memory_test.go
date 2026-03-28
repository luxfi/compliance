package store

import (
	"testing"

	"github.com/luxfi/compliance/pkg/types"
)

func TestMemoryStoreIdentity(t *testing.T) {
	s := NewMemoryStore()

	ident := &types.Identity{UserID: "user-1", Provider: "onfido", Status: types.KYCPending}
	if err := s.SaveIdentity(ident); err != nil {
		t.Fatalf("SaveIdentity: %v", err)
	}
	if ident.ID == "" {
		t.Fatal("expected ID to be generated")
	}

	got, err := s.GetIdentity(ident.ID)
	if err != nil {
		t.Fatalf("GetIdentity: %v", err)
	}
	if got.UserID != "user-1" {
		t.Fatalf("UserID = %q, want user-1", got.UserID)
	}

	list := s.ListIdentitiesByUser("user-1")
	if len(list) != 1 {
		t.Fatalf("ListIdentitiesByUser len = %d, want 1", len(list))
	}

	list = s.ListIdentitiesByUser("nonexistent")
	if len(list) != 0 {
		t.Fatalf("ListIdentitiesByUser for nonexistent = %d, want 0", len(list))
	}
}

func TestMemoryStoreAMLScreening(t *testing.T) {
	s := NewMemoryStore()

	sc := &types.AMLScreening{AccountID: "acct-1", Status: types.AMLPending, RiskLevel: types.RiskLow}
	if err := s.SaveAMLScreening(sc); err != nil {
		t.Fatalf("SaveAMLScreening: %v", err)
	}

	got, err := s.GetAMLScreening(sc.ID)
	if err != nil {
		t.Fatalf("GetAMLScreening: %v", err)
	}
	if got.AccountID != "acct-1" {
		t.Fatalf("AccountID = %q, want acct-1", got.AccountID)
	}

	byAcct := s.ListAMLScreeningsByAccount("acct-1")
	if len(byAcct) != 1 {
		t.Fatalf("ListAMLScreeningsByAccount len = %d, want 1", len(byAcct))
	}

	byStatus := s.ListAMLScreeningsByStatus(types.AMLPending)
	if len(byStatus) != 1 {
		t.Fatalf("ListAMLScreeningsByStatus len = %d, want 1", len(byStatus))
	}
}

func TestMemoryStoreApplication(t *testing.T) {
	s := NewMemoryStore()

	app := &types.Application{UserID: "user-1", Email: "test@example.com", Status: types.AppDraft}
	if err := s.SaveApplication(app); err != nil {
		t.Fatalf("SaveApplication: %v", err)
	}

	got, err := s.GetApplication(app.ID)
	if err != nil {
		t.Fatalf("GetApplication: %v", err)
	}
	if got.Email != "test@example.com" {
		t.Fatalf("Email = %q, want test@example.com", got.Email)
	}

	byUser, err := s.GetApplicationByUser("user-1")
	if err != nil {
		t.Fatalf("GetApplicationByUser: %v", err)
	}
	if byUser.ID != app.ID {
		t.Fatalf("GetApplicationByUser ID mismatch")
	}

	all := s.ListApplications()
	if len(all) != 1 {
		t.Fatalf("ListApplications len = %d, want 1", len(all))
	}

	byStatus := s.ListApplicationsByStatus(types.AppDraft)
	if len(byStatus) != 1 {
		t.Fatalf("ListApplicationsByStatus len = %d, want 1", len(byStatus))
	}
}

func TestMemoryStorePipeline(t *testing.T) {
	s := NewMemoryStore()

	p := &types.Pipeline{Name: "Test", Status: "active"}
	if err := s.SavePipeline(p); err != nil {
		t.Fatalf("SavePipeline: %v", err)
	}

	got, err := s.GetPipeline(p.ID)
	if err != nil {
		t.Fatalf("GetPipeline: %v", err)
	}
	if got.Name != "Test" {
		t.Fatalf("Name = %q, want Test", got.Name)
	}

	list := s.ListPipelines()
	if len(list) != 1 {
		t.Fatalf("ListPipelines len = %d, want 1", len(list))
	}

	if err := s.DeletePipeline(p.ID); err != nil {
		t.Fatalf("DeletePipeline: %v", err)
	}
	if len(s.ListPipelines()) != 0 {
		t.Fatal("expected 0 pipelines after delete")
	}
}

func TestMemoryStoreRole(t *testing.T) {
	s := NewMemoryStore()

	role := &types.Role{Name: "Admin", Permissions: []types.Permission{{Module: "kyc", Action: "read"}}}
	if err := s.SaveRole(role); err != nil {
		t.Fatalf("SaveRole: %v", err)
	}

	byName, err := s.GetRoleByName("Admin")
	if err != nil {
		t.Fatalf("GetRoleByName: %v", err)
	}
	if byName.ID != role.ID {
		t.Fatal("GetRoleByName ID mismatch")
	}
}

func TestMemoryStoreFund(t *testing.T) {
	s := NewMemoryStore()

	f := &types.Fund{Name: "Test Fund", Status: "open"}
	if err := s.SaveFund(f); err != nil {
		t.Fatalf("SaveFund: %v", err)
	}

	inv := &types.FundInvestor{FundID: f.ID, Name: "Alice", Amount: 50000, Status: "funded"}
	if err := s.AddFundInvestor(inv); err != nil {
		t.Fatalf("AddFundInvestor: %v", err)
	}

	investors := s.ListFundInvestors(f.ID)
	if len(investors) != 1 {
		t.Fatalf("ListFundInvestors len = %d, want 1", len(investors))
	}

	got, _ := s.GetFund(f.ID)
	if got.InvestorCount != 1 {
		t.Fatalf("InvestorCount = %d, want 1", got.InvestorCount)
	}
	if got.TotalRaised != 50000 {
		t.Fatalf("TotalRaised = %f, want 50000", got.TotalRaised)
	}
}

func TestMemoryStoreSettings(t *testing.T) {
	s := NewMemoryStore()

	settings := s.GetSettings()
	if settings.Currency != "USD" {
		t.Fatalf("default Currency = %q, want USD", settings.Currency)
	}

	s.SaveSettings(&types.Settings{BusinessName: "Test", Timezone: "UTC", Currency: "EUR", NotificationEmail: "a@b.com"})
	got := s.GetSettings()
	if got.Currency != "EUR" {
		t.Fatalf("Currency = %q, want EUR", got.Currency)
	}
}

func TestMemoryStoreDashboard(t *testing.T) {
	s := NewMemoryStore()

	s.SaveSession(&types.Session{Status: types.SessionPending, KYCStatus: types.KYCPending})
	s.SaveFund(&types.Fund{Name: "F1", Status: "open"})
	s.SaveTransaction(&types.Transaction{Type: "deposit", Amount: 100})

	stats := s.ComputeDashboard()
	if stats.ActiveSessions != 1 {
		t.Fatalf("ActiveSessions = %d, want 1", stats.ActiveSessions)
	}
	if stats.PendingKYC != 1 {
		t.Fatalf("PendingKYC = %d, want 1", stats.PendingKYC)
	}
	if stats.TotalFunds != 1 {
		t.Fatalf("TotalFunds = %d, want 1", stats.TotalFunds)
	}
}
