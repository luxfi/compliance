package reporting

import "time"

// ATSNData is a FINRA ATS-N quarterly report (Form ATS-N, Regulation ATS).
// Filed quarterly by alternative trading systems.
type ATSNData struct {
	ATSName         string          `json:"ats_name"`
	CRDNumber       string          `json:"crd_number"`
	BrokerDealer    string          `json:"broker_dealer"`
	Period          string          `json:"period"` // "2026-Q1"
	OrderSummary    ATSNOrderSummary `json:"order_summary"`
	ExecutionSummary ATSNExecSummary `json:"execution_summary"`
	Subscribers     ATSNSubscribers `json:"subscribers"`
	MaterialChanges []string        `json:"material_changes,omitempty"`
	FilingDate      time.Time       `json:"filing_date"`
}

// ATSNOrderSummary aggregates order activity for the quarter.
type ATSNOrderSummary struct {
	TotalOrders     int64   `json:"total_orders"`
	BuyOrders       int64   `json:"buy_orders"`
	SellOrders      int64   `json:"sell_orders"`
	MarketOrders    int64   `json:"market_orders"`
	LimitOrders     int64   `json:"limit_orders"`
	CancelledOrders int64   `json:"cancelled_orders"`
	TotalNotional   float64 `json:"total_notional"`
	TotalShares     float64 `json:"total_shares"`
	AvgOrderSize    float64 `json:"avg_order_size"`
}

// ATSNExecSummary aggregates execution statistics.
type ATSNExecSummary struct {
	TotalExecutions int64   `json:"total_executions"`
	TotalVolume     float64 `json:"total_volume"`
	FillRate        float64 `json:"fill_rate"` // 0.0 - 1.0
	AvgFillPrice    float64 `json:"avg_fill_price"`
	InternalCrosses int64   `json:"internal_crosses"` // matched internally
	ExternalRouted  int64   `json:"external_routed"`  // routed to external venues
}

// ATSNSubscribers summarizes subscriber participation.
type ATSNSubscribers struct {
	Total     int `json:"total"`
	Active    int `json:"active"` // placed at least one order in period
	NewAdded  int `json:"new_added"`
	Removed   int `json:"removed"`
}

// BuildATSN constructs an ATS-N quarterly report from order data.
func BuildATSN(orders []Order, period string) *ATSNData {
	atsn := &ATSNData{
		Period: period,
	}

	var totalNotional, totalShares, totalFillVolume float64
	var executions int64
	subscribers := map[string]bool{}

	for _, o := range orders {
		atsn.OrderSummary.TotalOrders++

		switch o.Side {
		case "buy":
			atsn.OrderSummary.BuyOrders++
		case "sell":
			atsn.OrderSummary.SellOrders++
		}

		switch o.Type {
		case "market":
			atsn.OrderSummary.MarketOrders++
		case "limit":
			atsn.OrderSummary.LimitOrders++
		}

		if o.Status == "cancelled" {
			atsn.OrderSummary.CancelledOrders++
		}

		totalNotional += o.Qty * o.Price
		totalShares += o.Qty

		if o.FilledQty > 0 {
			executions++
			totalFillVolume += o.FilledQty * o.FilledPrice
		}

		if o.SubscriberID != "" {
			subscribers[o.SubscriberID] = true
		}
	}

	atsn.OrderSummary.TotalNotional = totalNotional
	atsn.OrderSummary.TotalShares = totalShares
	if atsn.OrderSummary.TotalOrders > 0 {
		atsn.OrderSummary.AvgOrderSize = totalNotional / float64(atsn.OrderSummary.TotalOrders)
	}

	atsn.ExecutionSummary.TotalExecutions = executions
	atsn.ExecutionSummary.TotalVolume = totalFillVolume
	if atsn.OrderSummary.TotalOrders > 0 {
		atsn.ExecutionSummary.FillRate = float64(executions) / float64(atsn.OrderSummary.TotalOrders)
	}
	if executions > 0 {
		atsn.ExecutionSummary.AvgFillPrice = totalFillVolume / float64(executions)
	}

	atsn.Subscribers.Active = len(subscribers)

	return atsn
}

// ValidateATSN returns validation errors for an ATS-N report.
func ValidateATSN(atsn *ATSNData) []string {
	var errs []string

	if atsn.ATSName == "" {
		errs = append(errs, "ats_name is required")
	}
	if atsn.CRDNumber == "" {
		errs = append(errs, "crd_number is required")
	}
	if atsn.Period == "" {
		errs = append(errs, "period is required (e.g. 2026-Q1)")
	}
	if atsn.OrderSummary.TotalOrders == 0 {
		errs = append(errs, "order_summary has zero orders — report may be empty")
	}

	return errs
}
