package reporting

import (
	"testing"
	"time"
)

// --- SAR Tests ---

func TestBuildSAR(t *testing.T) {
	subject := Entity{
		Name:      "Jane Smith",
		FirstName: "Jane",
		LastName:  "Smith",
		DOB:       "1985-03-15",
		SSNLast4:  "4567",
		AccountID: "acct-001",
		Address:   "123 Main St",
		City:      "New York",
		State:     "NY",
		ZipCode:   "10001",
		Country:   "US",
	}

	alerts := []Alert{
		{
			ID:        "alert-1",
			RuleName:  "Structuring",
			Severity:  "medium",
			Amount:    9500,
			CreatedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:        "alert-2",
			RuleName:  "Structuring",
			Severity:  "medium",
			Amount:    9800,
			CreatedAt: time.Date(2026, 1, 20, 14, 0, 0, 0, time.UTC),
		},
	}

	sar := BuildSAR(subject, alerts, "Subject made multiple transactions just below the $10,000 CTR threshold over a 5-day period, consistent with structuring to avoid reporting.")
	if sar == nil {
		t.Fatal("BuildSAR returned nil")
	}
	if sar.Subject.Name != "Jane Smith" {
		t.Errorf("subject name = %q, want %q", sar.Subject.Name, "Jane Smith")
	}
	if sar.SuspiciousActivity.TotalAmount != 19300 {
		t.Errorf("total amount = %.2f, want %.2f", sar.SuspiciousActivity.TotalAmount, 19300.0)
	}
	if sar.FilingType != "initial" {
		t.Errorf("filing_type = %q, want %q", sar.FilingType, "initial")
	}
}

func TestValidateSAR_Valid(t *testing.T) {
	sar := &SARData{
		FilingType: "initial",
		Subject:    SARSubject{Name: "John Doe"},
		SuspiciousActivity: SARActivity{
			DateStart:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			DateEnd:     time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
			TotalAmount: 25000,
		},
		Narrative: "Subject transferred $25,000 across multiple accounts in a pattern consistent with layering, using accounts opened within days of each other.",
	}

	errs := ValidateSAR(sar)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateSAR_MissingFields(t *testing.T) {
	sar := &SARData{}
	errs := ValidateSAR(sar)
	if len(errs) == 0 {
		t.Error("expected validation errors for empty SAR")
	}

	// Must flag: filing_type, subject name, dates, amount, narrative.
	want := 6
	if len(errs) < want {
		t.Errorf("got %d errors, want at least %d", len(errs), want)
	}
}

func TestValidateSAR_ShortNarrative(t *testing.T) {
	sar := &SARData{
		FilingType: "initial",
		Subject:    SARSubject{Name: "X"},
		SuspiciousActivity: SARActivity{
			DateStart:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			DateEnd:     time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
			TotalAmount: 10000,
		},
		Narrative: "Suspicious.",
	}

	errs := ValidateSAR(sar)
	found := false
	for _, e := range errs {
		if e == "narrative must be at least 50 characters — explain the suspicious activity" {
			found = true
		}
	}
	if !found {
		t.Error("expected short narrative error")
	}
}

func TestValidateSAR_BelowThreshold(t *testing.T) {
	sar := &SARData{
		FilingType: "initial",
		Subject:    SARSubject{Name: "X"},
		SuspiciousActivity: SARActivity{
			DateStart:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			DateEnd:     time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
			TotalAmount: 3000,
		},
		Narrative: "Subject transferred money around, and it looked kind of suspicious but was below the threshold amount.",
	}

	errs := ValidateSAR(sar)
	found := false
	for _, e := range errs {
		if e == "total_amount $3000.00 below $5,000 SAR threshold" {
			found = true
		}
	}
	if !found {
		t.Error("expected below-threshold error")
	}
}

// --- CTR Tests ---

func TestBuildCTR(t *testing.T) {
	user := Entity{Name: "Alice Example", SSNLast4: "1234"}
	date := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	txs := []Transaction{
		{ID: "tx-1", Type: "deposit", Amount: 7000, Currency: "USD", Date: date},
		{ID: "tx-2", Type: "deposit", Amount: 5000, Currency: "USD", Date: date},
	}

	ctr := BuildCTR(user, txs, date)
	if ctr.TotalCashIn != 12000 {
		t.Errorf("total_cash_in = %.2f, want %.2f", ctr.TotalCashIn, 12000.0)
	}
	if len(ctr.Transactions) != 2 {
		t.Errorf("got %d transactions, want 2", len(ctr.Transactions))
	}
}

func TestBuildCTR_FiltersOtherDays(t *testing.T) {
	user := Entity{Name: "Bob"}
	date := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	txs := []Transaction{
		{ID: "tx-1", Type: "deposit", Amount: 11000, Currency: "USD", Date: date},
		{ID: "tx-2", Type: "deposit", Amount: 5000, Currency: "USD", Date: date.AddDate(0, 0, 1)},
	}

	ctr := BuildCTR(user, txs, date)
	if len(ctr.Transactions) != 1 {
		t.Errorf("got %d transactions, want 1 (other day filtered)", len(ctr.Transactions))
	}
}

func TestValidateCTR_BelowThreshold(t *testing.T) {
	ctr := &CTRData{
		Person:       CTRPerson{Name: "X"},
		Date:         time.Now(),
		Transactions: []CTRTxDetail{{Amount: 5000, Direction: "cash_in"}},
		TotalCashIn:  5000,
	}
	errs := ValidateCTR(ctr)
	found := false
	for _, e := range errs {
		if e == "aggregate total below $10,000 CTR threshold" {
			found = true
		}
	}
	if !found {
		t.Error("expected below-threshold error")
	}
}

func TestRequiresCTR(t *testing.T) {
	if !RequiresCTR(10001) {
		t.Error("10001 should require CTR")
	}
	if RequiresCTR(9999) {
		t.Error("9999 should not require CTR")
	}
	if RequiresCTR(10000) {
		t.Error("exactly 10000 should not require CTR (threshold is >)")
	}
}

// --- Form D Tests ---

func TestBuildFormD(t *testing.T) {
	offering := Offering{
		IssuerName:        "Acme Fund LP",
		IssuerState:       "DE",
		IssuerCountry:     "US",
		EntityType:        "lp",
		ExemptionType:     "reg_d_506b",
		SecurityType:      "pooled_interest",
		TotalOfferingSize: 10000000,
		TotalSold:         3000000,
		FirstSaleDate:     time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	issuer := Entity{Name: "Acme Fund LP"}
	sales := map[string]float64{"NY": 1000000, "CA": 2000000}

	fd := BuildFormD(offering, issuer, sales)
	if fd.Issuer.Name != "Acme Fund LP" {
		t.Errorf("issuer name = %q, want %q", fd.Issuer.Name, "Acme Fund LP")
	}
	if fd.Offering.ExemptionClaimed[0] != "rule_506_b" {
		t.Errorf("exemption = %q, want %q", fd.Offering.ExemptionClaimed[0], "rule_506_b")
	}
}

func TestValidateFormD_506cNonAccredited(t *testing.T) {
	fd := &FormDData{
		Issuer: FormDIssuer{
			Name:                 "X",
			EntityType:           "corporation",
			StateOfIncorporation: "DE",
		},
		Offering: FormDOffering{
			SecurityType:      "equity",
			ExemptionClaimed:  []string{"rule_506_c"},
			TotalOfferingSize: 1000000,
			HasNonAccredited:  true,
			FirstSaleDate:     time.Now(),
		},
	}

	errs := ValidateFormD(fd)
	found := false
	for _, e := range errs {
		if e == "Rule 506(c) requires all investors to be accredited" {
			found = true
		}
	}
	if !found {
		t.Error("expected 506(c) non-accredited error")
	}
}

// --- ATS-N Tests ---

func TestBuildATSN(t *testing.T) {
	orders := []Order{
		{ID: "o-1", Side: "buy", Type: "limit", Qty: 100, Price: 50, FilledQty: 100, FilledPrice: 50, Status: "filled", SubscriberID: "sub-1"},
		{ID: "o-2", Side: "sell", Type: "market", Qty: 200, Price: 25, FilledQty: 0, Status: "cancelled", SubscriberID: "sub-2"},
		{ID: "o-3", Side: "buy", Type: "limit", Qty: 50, Price: 100, FilledQty: 50, FilledPrice: 99, Status: "filled", SubscriberID: "sub-1"},
	}

	atsn := BuildATSN(orders, "2026-Q1")
	if atsn.OrderSummary.TotalOrders != 3 {
		t.Errorf("total_orders = %d, want 3", atsn.OrderSummary.TotalOrders)
	}
	if atsn.OrderSummary.BuyOrders != 2 {
		t.Errorf("buy_orders = %d, want 2", atsn.OrderSummary.BuyOrders)
	}
	if atsn.OrderSummary.CancelledOrders != 1 {
		t.Errorf("cancelled = %d, want 1", atsn.OrderSummary.CancelledOrders)
	}
	if atsn.ExecutionSummary.TotalExecutions != 2 {
		t.Errorf("executions = %d, want 2", atsn.ExecutionSummary.TotalExecutions)
	}
	if atsn.Subscribers.Active != 2 {
		t.Errorf("active_subscribers = %d, want 2", atsn.Subscribers.Active)
	}
}

// --- MiFID Tests ---

func TestBuildMiFID(t *testing.T) {
	tx := Transaction{
		ID:       "tx-001",
		Amount:   15000,
		Currency: "EUR",
		Side:     "buy",
		Asset:    "DE000BASF111",
		Date:     time.Now(),
	}
	entity := Entity{
		LEI:       "5299001Y2H7A4QBQ2N56",
		FirstName: "Hans",
		LastName:  "Mueller",
		Country:   "DE",
	}

	m := BuildMiFID(tx, entity)
	if m.Buyer.LEI != entity.LEI {
		t.Errorf("buyer LEI = %q, want %q", m.Buyer.LEI, entity.LEI)
	}
	if m.Execution.Side != "buy" {
		t.Errorf("side = %q, want %q", m.Execution.Side, "buy")
	}
}

func TestBuildMiFID_Sell(t *testing.T) {
	tx := Transaction{ID: "tx-002", Amount: 5000, Currency: "EUR", Side: "sell", Date: time.Now()}
	entity := Entity{LEI: "ABC123", LastName: "Test"}

	m := BuildMiFID(tx, entity)
	if m.Seller.LEI != "ABC123" {
		t.Error("seller not populated for sell side")
	}
}

func TestValidateMiFID_MissingLEI(t *testing.T) {
	m := &MiFIDData{
		Execution:       MiFIDExecution{Price: 100, Side: "buy"},
		Instrument:      MiFIDInstrument{Currency: "EUR"},
		TradingDateTime: time.Now(),
	}
	errs := ValidateMiFID(m)
	if len(errs) == 0 {
		t.Error("expected validation errors for missing LEI")
	}
}

// --- Tax 1099 Tests ---

func TestBuild1099(t *testing.T) {
	user := Entity{Name: "Tax Payer", SSNLast4: "9999"}
	trades := []Trade{
		{
			Symbol:       "AAPL",
			Side:         "sell",
			Qty:          10,
			Proceeds:     1500,
			CostBasis:    1000,
			AcquiredDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			DisposedDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			HoldingPeriod: "short",
		},
		{
			Symbol:       "MSFT",
			Side:         "sell",
			Qty:          5,
			Proceeds:     5000,
			CostBasis:    3000,
			AcquiredDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			DisposedDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			HoldingPeriod: "long",
		},
	}
	dividends := []Dividend{
		{Symbol: "AAPL", Amount: 50, Qualified: true, Date: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)},
	}

	td := Build1099(user, trades, dividends, 2026)
	if td.Year != 2026 {
		t.Errorf("year = %d, want 2026", td.Year)
	}
	if len(td.Proceeds) != 2 {
		t.Errorf("proceeds count = %d, want 2", len(td.Proceeds))
	}
	if td.Summary.ShortTermGainLoss != 500 {
		t.Errorf("short_term = %.2f, want 500", td.Summary.ShortTermGainLoss)
	}
	if td.Summary.LongTermGainLoss != 2000 {
		t.Errorf("long_term = %.2f, want 2000", td.Summary.LongTermGainLoss)
	}
	if td.Summary.TotalQualifiedDividends != 50 {
		t.Errorf("qualified_dividends = %.2f, want 50", td.Summary.TotalQualifiedDividends)
	}
}

func TestWashSaleDetection(t *testing.T) {
	trades := []Trade{
		{
			Symbol:       "AAPL",
			Side:         "sell",
			Qty:          10,
			Proceeds:     900,
			CostBasis:    1000,
			AcquiredDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			DisposedDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			HoldingPeriod: "short",
		},
		{
			Symbol:       "AAPL",
			Side:         "buy",
			Qty:          10,
			Proceeds:     950,
			CostBasis:    950,
			AcquiredDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			DisposedDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			HoldingPeriod: "short",
		},
	}

	result := detectWashSales(trades)
	if !result[0].WashSale {
		t.Error("first trade should be flagged as wash sale")
	}
	if result[0].WashAmount != 100 {
		t.Errorf("wash amount = %.2f, want 100", result[0].WashAmount)
	}
}

// --- Travel Rule Tests ---

func TestRequiresTravelRule_USThreshold(t *testing.T) {
	if !RequiresTravelRule(5000, "USD", "US") {
		t.Error("$5,000 USD in US should require travel rule")
	}
	if RequiresTravelRule(2000, "USD", "US") {
		t.Error("$2,000 USD in US should not require travel rule")
	}
}

func TestRequiresTravelRule_UnknownJurisdiction(t *testing.T) {
	// Unknown jurisdiction uses conservative $1,000 threshold.
	if !RequiresTravelRule(1500, "USD", "XX") {
		t.Error("$1,500 in unknown jurisdiction should require travel rule")
	}
}

func TestBuildTravelRuleMessage(t *testing.T) {
	data := TravelRuleData{
		Originator: TravelRuleParty{
			Name:        "Alice Corp",
			AccountID:   "alice-001",
			Institution: "VASP Alpha",
		},
		Beneficiary: TravelRuleParty{
			Name:        "Bob Ltd",
			AccountID:   "bob-002",
			Institution: "VASP Beta",
		},
		Amount:    50000,
		Currency:  "USD",
		Timestamp: time.Now(),
	}

	msg := BuildTravelRuleMessage(data)
	if msg.Version != "IVMS101" {
		t.Errorf("version = %q, want %q", msg.Version, "IVMS101")
	}
	if msg.Originator.Name != "Alice Corp" {
		t.Errorf("originator name = %q, want %q", msg.Originator.Name, "Alice Corp")
	}
}

func TestValidateTravelRule(t *testing.T) {
	// Missing everything.
	errs := ValidateTravelRule(TravelRuleData{})
	if len(errs) < 4 {
		t.Errorf("expected at least 4 errors, got %d", len(errs))
	}

	// Valid.
	errs = ValidateTravelRule(TravelRuleData{
		Originator:  TravelRuleParty{Name: "A", AccountID: "a1"},
		Beneficiary: TravelRuleParty{Name: "B", AccountID: "b1"},
		Amount:      5000,
		Currency:    "USD",
	})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}
