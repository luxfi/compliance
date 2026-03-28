# LLM.md - Lux Compliance

## Overview
Go module: github.com/luxfi/compliance

Single source of truth for all compliance domain logic: types, store interface,
identity verification, KYC/AML, payment compliance, multi-jurisdiction regulatory
frameworks, onboarding flows, RBAC, and Jube AML sidecar client. Used by
luxfi/broker (thin HTTP layer) and luxfi/bank.

## Tech Stack
- **Language**: Go 1.26.1
- **Dependencies**: standard library only (zero external deps)

## Build & Run
```bash
GOWORK=off go build ./...
GOWORK=off go test ./...
GOWORK=off go test -race ./...
```

## Structure
```
compliance/
  go.mod
  LLM.md
  migrations/
    001_initial.sql    -- PostgreSQL schema for all compliance tables
  cmd/
    complianced/       -- Standalone REST API server
  pkg/
    types/             -- All compliance domain types (single source of truth)
    store/             -- ComplianceStore interface + MemoryStore
    idv/               -- Identity verification providers (Jumio, Onfido, Plaid)
    kyc/               -- KYC orchestration service + regulatory application model
    aml/               -- AML/sanctions screening + transaction monitoring
    jube/              -- Jube AML sidecar client (HTTP, webhooks, pre-trade screen)
    rbac/              -- Role-based access control (default roles, permission check)
    onboarding/        -- 5-step investor onboarding flow logic
    regulatory/        -- Multi-jurisdiction rules (US, UK, Isle of Man)
    payments/          -- Payment compliance + stablecoin validation
    entity/            -- Regulated entity types (ATS, BD, TA, MSB)
    webhook/           -- Unified webhook handler with idempotency
```

## Package Details

### pkg/types -- Compliance Domain Types
All types shared across broker, bank, and compliance services:
- Status enums: KYCStatus, KYBStatus, SessionStatus, AMLStatus, RiskLevel,
  ApplicationStatus, EnvelopeStatus
- Core: Identity, Document, BusinessKYB
- Pipeline/Session: Pipeline, PipelineStep, Session, SessionStep
- Fund: Fund, FundInvestor
- eSign: Envelope, Signer, Template
- RBAC: Role, Permission, Module
- Users: User, Credential, Settings
- Transactions: Transaction
- Applications: Application, ApplicationStep, DocumentUpload
- Dashboard: DashboardStats, ESignStats
- Billing: Invoice, BillingInfo

### pkg/store -- ComplianceStore Interface + MemoryStore
- `ComplianceStore` interface: 40+ methods for all entity CRUD
- `MemoryStore`: thread-safe in-memory implementation with sync.RWMutex
- `GenerateID()`: crypto/rand hex ID generator
- PostgresStore lives in broker (depends on pgx driver)

### pkg/jube -- Jube AML Sidecar Client
- `Client`: HTTP client for Jube REST API (ScreenTransaction, CheckSanctions, CreateCase, GetCases, Search)
- `PreTradeScreen`: pre-trade AML screening with fail-open/fail-closed config
- Webhook: FireWebhook with HMAC-SHA256 signing and exponential backoff retries
- VerifySignature for incoming webhook validation
- Event types: aml.flagged, aml.cleared, kyc.approved, trade.executed

### pkg/rbac -- Role-Based Access Control
- `DefaultRoles()`: Owner, Admin, Manager, Developer, Agent, Reviewer
- `ComplianceModules()`: kyc, aml, applications, funds, esign, pipelines, sessions, roles
- `HasPermission(role, module, action)`: checks permission with admin-implies-all

### pkg/onboarding -- 5-Step Investor Onboarding
- `NewApplicationSteps()`: returns the 5 default onboarding steps
- `IsTerminalStatus()`: checks if application is approved/rejected/submitted
- `MarkStepCompleted()`, `MarkStepFailed()`: step lifecycle helpers

### pkg/idv -- Identity Verification Providers
- `Provider` interface: `Name()`, `InitiateVerification()`, `CheckStatus()`, `ParseWebhook()`
- Provider registry with factory pattern: `GetProvider(name, config)`
- Jumio, Onfido, Plaid implementations
- Status constants: Pending, Approved, Declined, Expired, Error

### pkg/kyc -- KYC Orchestration
- `Service`: multi-provider KYC lifecycle (initiate, webhook, status tracking)
- `Store`: in-memory regulatory application CRUD with status filtering + stats
- `Application`: regulatory application model (distinct from onboarding Application in types/)
- HMAC-SHA256 webhook signature validation per provider

### pkg/aml -- AML/Sanctions Screening & Transaction Monitoring
- `ScreeningService`: screens against OFAC SDN, EU, UK HMT, PEP, adverse media
- `MonitoringService`: real-time transaction monitoring rules engine
- Rule types: single_amount, daily_aggregate, velocity, geographic, structuring
- SAR generation from alerts

### pkg/regulatory -- Jurisdiction Framework
- `Jurisdiction` interface: requirements, validation, transaction limits
- USA: FinCEN BSA (CIP, CTR $10k, SAR), SEC/FINRA suitability/disclosures
- UK: FCA registration, 5AMLD CDD/EDD, HM Treasury sanctions
- IOM: IOMFSA Designated Business, AML/CFT Code 2019

### pkg/payments -- Payment Compliance
- `ComplianceEngine`: validates payin/payout against jurisdiction rules
- Travel Rule (FATF Rec 16), CTR threshold detection
- `StablecoinEngine`: token allowlists, address risk, mint/burn compliance

### pkg/entity -- Regulated Entity Types
- ATS, Broker-Dealer, Transfer Agent, MSB definitions with net capital rules

### pkg/webhook -- Unified Webhook Handler
- Multi-provider routing with HMAC-SHA256 signature validation
- Idempotency tracking, retry, dead letter queue

## Thread Safety
All services use sync.RWMutex for concurrent access. ID generation uses crypto/rand.

## Test Coverage
180+ tests across 10 packages. All pass with -race flag.

## Relationship to Broker
The broker (github.com/luxfi/broker) imports this library via:
```
require github.com/luxfi/compliance v0.1.0
replace github.com/luxfi/compliance => ../compliance
```
Broker's `pkg/compliance/` re-exports types as aliases and adds HTTP handlers +
PostgresStore (pgx). The library owns all domain logic; broker owns HTTP routing
and broker-specific wiring (admin auth, credentials, billing).
