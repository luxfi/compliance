# Lux Compliance

Regulated financial compliance stack: identity verification, KYC/AML, sanctions
screening, transaction monitoring, payment compliance, and multi-jurisdiction
regulatory frameworks. Standard library only — zero external Go deps.

```bash
GOWORK=off go build -o complianced ./cmd/complianced/
COMPLIANCE_API_KEY=… ./complianced
```

Single source of truth for compliance domain logic. Consumed by
[`luxfi/broker`](https://github.com/luxfi/broker) (HTTP layer) and
[`luxfi/bank`](https://github.com/luxfi/bank) (customer apps + admin).

## Architecture

```
                    ┌──────────────────┐
                    │   complianced    │
                    │     :8091        │
                    └────────┬─────────┘
                             │
        ┌────────────────────┼────────────────────────┐
        │                    │                        │
  ┌─────┴──────┐      ┌──────┴──────┐         ┌──────┴──────┐
  │   IDV      │      │     AML     │         │  Payments   │
  │ providers  │      │  screening  │         │ compliance  │
  └──────┬─────┘      └──────┬──────┘         └──────┬──────┘
         │                   │                       │
   ┌─────┼─────┐    ┌────────┼────────┐    ┌─────────┼─────────┐
 Jumio Onfido Plaid OFAC EU UK PEP Jube  Travel CTR  Stablecoin
                                          Rule       Validation

                  ┌──────────┬──────────┬──────────┐
                  │ Regulatory framework (20 jurs) │
                  │ ── 14 frameworks ──            │
                  │ Entity types (13)              │
                  │ Onboarding (5-step)            │
                  │ Reporting + RBAC               │
                  └──────────┴──────────┴──────────┘
```

IDV providers come from [`hanzoai/idv`](https://github.com/hanzoai/idv) — the
shared identity-verification interface used across Hanzo and Lux services.
Compliance imports `idv.Provider` and registers each (`Jumio`, `Onfido`, `Plaid`).

## Packages

### `pkg/types` — Domain Types

Single source of truth for all status enums, models, and view types shared by
broker, bank, and compliance. Status enums: `KYCStatus`, `KYBStatus`,
`SessionStatus`, `AMLStatus`, `RiskLevel`, `ApplicationStatus`,
`EnvelopeStatus`. Models: `Identity`, `Document`, `BusinessKYB`, `Pipeline`,
`Session`, `Fund`, `Envelope`, `Transaction`, `Application`, `Role`,
`Permission`, `User`.

### `pkg/store` — Storage Interface

`ComplianceStore` interface with an in-memory implementation
(`MemoryStore`). PostgreSQL backing schema in `migrations/001_initial.sql`.

### `pkg/kyc` — KYC Orchestration

Multi-provider verification + application lifecycle. Uses
[`hanzoai/idv`](https://github.com/hanzoai/idv) as the provider interface;
each implementation (Jumio Netverify v4, Onfido v3.6, Plaid IDV) registers via
`RegisterProvider`. HMAC-SHA256 webhook signature validation.

```
draft → pending → pending_kyc → approved
                              → rejected

not_started → pending → verified
                      → failed
```

### `pkg/aml` — Sanctions Screening + Transaction Monitoring

Screening: OFAC SDN, EU consolidated, UK HM Treasury, PEP databases, adverse
media. Match types: exact, fuzzy (Levenshtein), partial. Risk scoring: low,
medium, high, critical.

Transaction monitoring: single-tx amount limits, daily aggregates, velocity
checks, geographic risk, structuring/smurfing detection. Alert lifecycle:
open → investigating → escalated → closed/filed.

### `pkg/jube` — Jube AML Sidecar Client

HTTP client + webhooks for the [Jube](https://github.com/jube-suite/jube)
real-time AML platform. Pre-trade screening, transaction risk scoring, model
inference. Used as a sidecar alongside the in-process screening for
high-throughput venues.

### `pkg/regulatory` — Multi-Jurisdiction Framework

**20 jurisdictions, 14 regulatory frameworks.**

| Framework | Code | Jurisdictions |
|-----------|------|---------------|
| US SEC + FINRA | `us_sec_finra` | US |
| UK FCA | `uk_fca` | GB |
| Isle of Man IOMFSA | `iom` | IM |
| Canada CIRO | `ciro` | CA |
| Brazil CVM | `cvm` | BR |
| India SEBI | `sebi` | IN |
| Singapore MAS | `mas` | SG |
| Australia ASIC | `asic` | AU |
| Switzerland FINMA | `finma` | CH |
| UAE SCA | `sca` | AE |
| Dubai DFSA | `dfsa` | AE-DIFC |
| Abu Dhabi FSRA | `fsra` | AE-ADGM |
| UAE VARA | `vara` | AE-VARA |
| EU MiCA | `mica` | LU, DE, FR, NL, IE, IT, ES |

Each `Jurisdiction` defines: regulator, framework code, application validation
rules, transaction limits, suitability tests, sanctions list selection.

### `pkg/entity` — Regulated Entity Types

**13 entity types** covering global market structure:

| Code | Type | Notes |
|------|------|-------|
| `ats` | Alternative Trading System | SEC Reg ATS, Form ATS-N, Rule 300-303 |
| `broker_dealer` | Broker-Dealer | SEC/FINRA/SIPC, Rule 15c3-1 |
| `transfer_agent` | Transfer Agent | SEC Rule 17Ad, Form TA-1/TA-2 |
| `msb` | Money Services Business | FinCEN, state MTLs, CTR/SAR |
| `sicav` | SICAV (LU) | Luxembourg open-ended investment company |
| `sicar` | SICAR (LU) | Luxembourg risk-capital investment vehicle |
| `raif` | RAIF (LU) | Reserved Alternative Investment Fund |
| `aifm` | AIFM | Alternative Investment Fund Manager |
| `mancoman` | Management Company | LU/CH manco |
| `crr` | CRR (CH) | Swiss capital-raising representative |
| `issuer` | Issuer | Tokenized-security issuance |
| `custodian` | Custodian | Qualified custodian |
| `dlt_facility` | DLT Facility | DLT Pilot / DLT Foundation regimes |

### `pkg/payments` — Payment Compliance

**Travel Rule** (FATF Recommendation 16): originator/beneficiary required
> $3,000. **CTR detection**: flags ≥ $10,000 for Currency Transaction Report.
**Stablecoin validation**: token allowlists, per-jurisdiction policies, address
risk scoring (chain-analysis integration point).

### `pkg/onboarding` — 5-Step Investor Onboarding

Pipeline-based onboarding with per-step session state. Steps: identity,
verification (KYC), suitability, agreements (eSign), funding. Resumable, with
admin review, document upload, and reg disclosures per jurisdiction.

### `pkg/rbac` — Role-Based Access Control

Default roles, permission set, module gating. Used by both the broker HTTP
layer and the bank admin app.

### `pkg/reporting` — Regulatory Reporting

CTR, SAR, suspicious-activity, transaction reporting, audit log export.
Per-framework formats.

### `pkg/webhook` — Unified Webhook Handler

Routes incoming webhooks to provider handlers. HMAC-SHA256 signature
validation, idempotency tracking (event dedup), retry with max attempts, dead
letter queue.

## API Reference

Base URL: `http://localhost:8091/v1` — `X-Api-Key` header required (skip for
`/healthz` and webhook endpoints).

### Applications

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/applications` | Create application |
| GET | `/v1/applications/{id}` | Get application |
| PATCH | `/v1/applications/{id}` | Update (draft save) |
| POST | `/v1/applications/{id}/submit` | Submit for review |
| GET | `/v1/applications` | List (`?status=`) |
| GET | `/v1/applications/stats` | Statistics |

### KYC

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/kyc/verify` | Initiate verification |
| GET | `/v1/kyc/status/{verificationId}` | Status |
| GET | `/v1/kyc/application/{applicationId}` | All for application |
| POST | `/v1/kyc/webhook/{provider}` | Provider webhook (no auth) |

### AML

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/aml/screen` | Screen identity |
| POST | `/v1/aml/monitor` | Monitor transaction |
| GET | `/v1/aml/alerts` | List (`?status=`) |
| POST | `/v1/aml/jube/webhook` | Jube callback (no auth) |

### Payments + Regulatory

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/payments/validate` | Validate payin/payout |
| GET | `/v1/regulatory/{jurisdiction}` | Jurisdiction requirements |
| GET | `/v1/regulatory/frameworks` | List frameworks |
| GET | `/v1/regulatory/frameworks/{code}` | Framework + jurisdictions |

### System

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Health (no auth) |
| GET | `/v1/providers` | List registered IDV providers |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `COMPLIANCE_LISTEN` | `:8091` | HTTP listen address |
| `COMPLIANCE_API_KEY` | — | API key for authenticated endpoints |
| `KYC_DEFAULT_PROVIDER` | first registered | Default IDV provider |
| `JUBE_BASE_URL` | — | Jube AML sidecar endpoint |
| `JUBE_API_KEY` | — | Jube auth |
| `DB_URL` | memory | PostgreSQL connection (production) |

**IDV provider credentials:**

| Provider | Variables |
|----------|-----------|
| Jumio | `JUMIO_API_TOKEN`, `JUMIO_API_SECRET`, `JUMIO_WEBHOOK_SECRET` |
| Onfido | `ONFIDO_API_TOKEN`, `ONFIDO_WEBHOOK_SECRET` |
| Plaid | `PLAID_CLIENT_ID`, `PLAID_SECRET`, `PLAID_WEBHOOK_SECRET` |

## Build & Test

```bash
make build
make test          # all packages
make test-race     # with -race
make lint          # go vet
make docker        # ghcr.io/luxfi/compliance image
```

## Docker

```bash
docker run -p 8091:8091 \
  -e COMPLIANCE_API_KEY=… \
  -e JUMIO_API_TOKEN=… \
  ghcr.io/luxfi/compliance:latest
```

Image: `ghcr.io/luxfi/compliance` — alpine, healthcheck on `/healthz`.

## Integration

### Go (library)

```go
import (
    "github.com/luxfi/compliance/pkg/kyc"
    "github.com/luxfi/compliance/pkg/aml"
    "github.com/luxfi/compliance/pkg/regulatory"
    "github.com/hanzoai/idv/provider"
)

svc := kyc.NewService()
svc.RegisterProvider(provider.NewJumio(provider.JumioConfig{…}))

screener := aml.NewScreener(aml.DefaultConfig())
result, _ := screener.Screen(ctx, identity)

jur := regulatory.For("US")
limits := jur.TransactionLimits()
```

### TypeScript (bank SDK)

```ts
import { ComplianceModule } from '@luxbank/compliance'

ComplianceModule.forRoot({
  baseUrl: process.env.COMPLIANCE_BASE_URL ?? 'http://compliance:8091',
  apiKey: process.env.COMPLIANCE_API_KEY ?? '',
})
```

`@luxbank/compliance` lives in
[luxfi/bank/pkg/compliance](https://github.com/luxfi/bank/tree/main/pkg/compliance).

## Project Structure

```
compliance/
├── cmd/complianced/   Standalone HTTP server
├── migrations/        PostgreSQL schema
└── pkg/
    ├── types/         Status enums + domain models (single source of truth)
    ├── store/         ComplianceStore interface + MemoryStore
    ├── kyc/           Service + Application
    ├── aml/           Screening + transaction monitoring
    ├── jube/          Jube AML sidecar client
    ├── regulatory/    20 jurisdictions, 14 frameworks
    ├── entity/        13 regulated entity types
    ├── payments/      Travel rule, CTR, stablecoin
    ├── onboarding/    5-step flow
    ├── rbac/          Roles + permissions
    ├── reporting/     CTR/SAR/audit export
    └── webhook/       Unified handler with idempotency
```

## Related

| Module | Purpose |
|--------|---------|
| [hanzoai/idv](https://github.com/hanzoai/idv) | Shared identity-verification provider interface (Jumio/Onfido/Plaid) |
| [luxfi/broker](https://github.com/luxfi/broker) | Multi-venue trading router, settlement |
| [luxfi/bank](https://github.com/luxfi/bank) | Customer apps, payments, admin dashboard |
| [hanzoai/iam](https://github.com/hanzoai/iam) | Identity and access management |
| [hanzoai/commerce](https://github.com/hanzoai/commerce) | Payment processors |
| [luxfi/security](https://github.com/luxfi/security) | Security stack (audits, formal verification, KMS, signing) |

## License

Copyright 2024-2026 Lux Partners Limited. All rights reserved.
