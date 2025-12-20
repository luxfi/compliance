# LLM.md - Lux Compliance

## Overview
Go module: github.com/luxfi/compliance

Regulated financial compliance stack for identity verification, KYC/AML,
payment compliance, and multi-jurisdiction regulatory frameworks. Extends
the hanzoai/iam IDV pattern. Used by luxfi/broker and luxfi/bank.

## Tech Stack
- **Language**: Go 1.26.1
- **Dependencies**: standard library only (zero external deps)

## Build & Run
```bash
go build ./...
go test ./...
go test -race ./...
```

## Structure
```
compliance/
  go.mod
  LLM.md
  pkg/
    idv/           -- Identity verification providers (Jumio, Onfido, Plaid)
    kyc/           -- KYC orchestration service + application model
    aml/           -- AML/sanctions screening + transaction monitoring
    regulatory/    -- Multi-jurisdiction rules (US, UK, Isle of Man)
    payments/      -- Payment compliance + stablecoin validation
    entity/        -- Regulated entity types (ATS, BD, TA, MSB)
    webhook/       -- Unified webhook handler with idempotency
```

## Package Details

### pkg/idv — Identity Verification Providers
- `Provider` interface: `Name()`, `InitiateVerification()`, `CheckStatus()`, `ParseWebhook()`
- Provider registry with factory pattern: `GetProvider(name, config)`
- Jumio: API v4, initiate/status/webhook, HMAC-SHA256 sig validation
- Onfido: API v3.6, applicant+check+SDK token, webhook parsing
- Plaid: Identity Verification sessions, verification/get, webhook parsing
- Status constants: Pending, Approved, Declined, Expired, Error

### pkg/kyc — KYC Orchestration
- `Service`: multi-provider KYC lifecycle (initiate, webhook, status tracking)
- `Store`: in-memory application CRUD with status filtering + stats
- `Application`: full model with identity, address, tax, disclosures, employment,
  financial, account prefs, KYC state, documents, admin notes
- Status lifecycle: draft -> pending -> pending_kyc -> approved/rejected
- KYC status: not_started -> pending -> verified/failed
- HMAC-SHA256 webhook signature validation per provider

### pkg/aml — AML/Sanctions Screening & Transaction Monitoring
- `ScreeningService`: screens against OFAC SDN, EU, UK HMT, PEP, adverse media
- Match types: exact, fuzzy (Levenshtein), partial
- Risk scoring: low, medium, high, critical
- `MonitoringService`: real-time transaction monitoring rules engine
- Rule types: single_amount, daily_aggregate, velocity, geographic, structuring
- SAR generation from alerts
- Alert lifecycle: open -> investigating -> escalated -> closed/filed

### pkg/regulatory — Jurisdiction Framework
- `Jurisdiction` interface: requirements, validation, transaction limits
- USA: FinCEN BSA (CIP, CTR $10k, SAR), SEC/FINRA suitability/disclosures
- UK: FCA registration, 5AMLD CDD/EDD, HM Treasury sanctions
- IOM: IOMFSA Designated Business, AML/CFT Code 2019, source of wealth/funds

### pkg/payments — Payment Compliance
- `ComplianceEngine`: validates payin/payout against jurisdiction rules
- Travel Rule (FATF Rec 16): originator/beneficiary info for transfers >$3k
- CTR threshold detection, sanctions screening on counterparties
- `StablecoinEngine`: token allowlists, address risk (chain analysis integration),
  mint/burn compliance, per-jurisdiction stablecoin policies

### pkg/entity — Regulated Entity Types
- ATS: SEC Reg ATS, Form ATS-N, $250k net capital
- Broker-Dealer: SEC/FINRA/SIPC, $250k net capital, Rule 15c3-1
- Transfer Agent: SEC Rule 17Ad, Form TA-1/TA-2
- MSB: FinCEN registration, state MTLs, CTR/SAR filing

### pkg/webhook — Unified Webhook Handler
- Multi-provider routing with HMAC-SHA256 signature validation
- Idempotency tracking (event deduplication)
- Retry with configurable max attempts
- Dead letter queue for failed webhooks

## Thread Safety
All services use sync.RWMutex for concurrent access. ID generation uses crypto/rand.

## Test Coverage
168 tests across 7 packages. All pass with -race flag.
