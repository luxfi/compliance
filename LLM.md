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
    regulatory/        -- Multi-jurisdiction rules (20 jurisdictions, 14 frameworks)
    payments/          -- Payment compliance + stablecoin validation
    entity/            -- Regulated entity types (13 types: ATS, BD, TA, MSB + EU/CH vehicles)
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

### pkg/regulatory -- Multi-Jurisdiction Framework
- `Jurisdiction` interface: Code, Name, RegulatoryFramework, PassportableTo, Requirements, ValidateApplication, TransactionLimits
- 20 jurisdictions: US, GB, IM, CA, BR, IN, SG, AU, CH, AE, AE-DIFC, AE-ADGM, AE-VARA, LU, DE, FR, NL, IE, IT, ES
- 14 frameworks: us_sec_finra, uk_fca, iom, ciro, cvm, sebi, mas, asic, finma, sca, dfsa, fsra, vara, mica
- `Framework` type + `JurisdictionsByFramework()` + `AllFrameworks()`
- EU passporting: all 7 MiCA jurisdictions (LU/DE/FR/NL/IE/IT/ES) passport to each other
- `GetJurisdiction(code)` returns nil for unknown codes; `AllJurisdictions()` returns all 20
- Shared `validateEUApplication()` and `euTransactionLimits()` for EU jurisdictions

### pkg/payments -- Payment Compliance
- `ComplianceEngine`: validates payin/payout against jurisdiction rules
- Travel Rule (FATF Rec 16), CTR threshold detection
- `StablecoinEngine`: token allowlists, address risk, mint/burn compliance

### pkg/entity -- Regulated Entity Types
- 13 types via EntityType_* constants: ats, broker_dealer, transfer_agent, msb,
  sicav, sicar, raif, aifm, mancoman, crr, issuer, custodian, dlt_facility
- `GetEntity(type)` and `AllEntities()` for registry access
- Each implements RegulatedEntity: licenses, reporting, capital, operational requirements

### pkg/types -- Asset Classes
- `AssetClass` type + 38 constants: us_equity, us_crypto, eu_aif, ch_dlt_security, etc.
- `AllAssetClasses()` returns all defined classes

### pkg/webhook -- Unified Webhook Handler
- Multi-provider routing with HMAC-SHA256 signature validation
- Idempotency tracking, retry, dead letter queue

## Thread Safety
All services use sync.RWMutex for concurrent access. ID generation uses crypto/rand.

## Test Coverage
250 tests across 11 packages. All pass with -race flag.

## Relationship to Broker
The broker (github.com/luxfi/broker) imports this library via:
```
require github.com/luxfi/compliance v0.1.0
replace github.com/luxfi/compliance => ../compliance
```
Broker's `pkg/compliance/` re-exports types as aliases and adds HTTP handlers +
PostgresStore (pgx). The library owns all domain logic; broker owns HTTP routing
and broker-specific wiring (admin auth, credentials, billing).
