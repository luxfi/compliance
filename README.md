# Lux Compliance

Regulated financial compliance stack: identity verification, KYC/AML, sanctions screening, transaction monitoring, payment compliance, and multi-jurisdiction regulatory frameworks.

Zero external dependencies. Standard library only.

```
go build -o complianced ./cmd/complianced/
JUMIO_API_TOKEN=... COMPLIANCE_API_KEY=... ./complianced
```

## Architecture

```
                    ┌──────────────────┐
                    │   complianced    │
                    │     :8091        │
                    └────────┬─────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
  ┌─────┴─────┐       ┌─────┴─────┐       ┌─────┴─────┐
  │    IDV    │       │    AML    │       │ Payments  │
  │ Providers │       │ Screening │       │Compliance │
  └─────┬─────┘       └─────┬─────┘       └─────┬─────┘
        │                    │                    │
   ┌────┼────┐          ┌───┼───┐           ┌───┼───┐
   │    │    │          │   │   │           │   │   │
 Jumio Onfido Plaid   OFAC EU  PEP     Travel CTR  Stablecoin
                       SDN  HMT           Rule      Validation
```

Extends the [hanzoai/iam](https://github.com/hanzoai/iam) `idv/` provider pattern. Consumed by [luxfi/broker](https://github.com/luxfi/broker) and [luxfi/bank](https://github.com/luxfi/bank).

## Packages

### `pkg/idv` — Identity Verification Providers

Provider interface with factory pattern. Each provider implements initiate, status check, and webhook parsing.

| Provider | API | Features |
|----------|-----|----------|
| Jumio | Netverify v4 | ID + selfie, liveness, document verification |
| Onfido | v3.6 | Applicant, check, SDK token, watchlist screening |
| Plaid | Identity Verification | Session-based IDV, bank-linked identity |

All providers support HMAC-SHA256 webhook signature validation.

### `pkg/kyc` — KYC Orchestration

Full application lifecycle with multi-provider KYC verification.

**Application status flow:**
```
draft → pending → pending_kyc → approved
                              → rejected
```

**KYC status flow:**
```
not_started → pending → verified
                      → failed
```

Application model includes: identity, address, tax info, regulatory disclosures, employment, financial profile, account preferences, documents, and admin review fields.

### `pkg/aml` — AML/Sanctions Screening & Transaction Monitoring

**Screening** checks applicants against:
- OFAC SDN list (US Treasury)
- EU consolidated sanctions
- UK HM Treasury sanctions
- PEP (Politically Exposed Persons) databases
- Adverse media

Match types: exact, fuzzy (Levenshtein distance), partial. Risk scoring: low, medium, high, critical.

**Transaction monitoring** rules engine:
- Single transaction amount limits
- Daily aggregate limits
- Velocity checks (too many transactions in time window)
- Geographic risk (high-risk country detection)
- Structuring/smurfing pattern detection

Alert lifecycle: open → investigating → escalated → closed/filed.

### `pkg/regulatory` — Multi-Jurisdiction Framework

| Jurisdiction | Regulator | Key Requirements |
|-------------|-----------|------------------|
| USA | FinCEN, SEC, FINRA | BSA (CIP, CTR $10k+, SAR), suitability, accredited investor |
| UK | FCA | Registration, 5AMLD CDD/EDD, HM Treasury sanctions |
| Isle of Man | IOMFSA | Designated Business, AML/CFT Code 2019, source of wealth/funds |

Each jurisdiction defines: requirements, application validation rules, and transaction limits.

### `pkg/payments` — Payment Compliance

**Travel Rule** (FATF Recommendation 16): originator and beneficiary information required for transfers over $3,000.

**CTR detection**: flags transactions at or above $10,000 for Currency Transaction Report filing.

**Stablecoin validation**: token allowlists, per-jurisdiction policies, address risk scoring (chain analysis integration point).

### `pkg/entity` — Regulated Entity Types

| Entity | Registration | Net Capital | Key Rules |
|--------|-------------|-------------|-----------|
| ATS | SEC Reg ATS, Form ATS-N | $250,000 | Rule 300-303 |
| Broker-Dealer | SEC/FINRA/SIPC | $250,000 | Rule 15c3-1 |
| Transfer Agent | SEC Rule 17Ad | $25,000 | Form TA-1/TA-2 |
| MSB | FinCEN, state MTLs | Varies | CTR/SAR filing |

### `pkg/webhook` — Unified Webhook Handler

Routes incoming webhooks to the correct provider handler. Features:
- HMAC-SHA256 signature validation per provider
- Idempotency tracking (event deduplication)
- Configurable retry with max attempts
- Dead letter queue for failed webhooks

## API Reference

Base URL: `http://localhost:8091/v1`

Auth: `X-Api-Key` header (skip for `/healthz` and webhook endpoints).

### Applications

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/applications` | Create application |
| GET | `/v1/applications/{id}` | Get application |
| PATCH | `/v1/applications/{id}` | Update (draft save) |
| POST | `/v1/applications/{id}/submit` | Submit for review |
| GET | `/v1/applications` | List applications (`?status=`) |
| GET | `/v1/applications/stats` | Application statistics |

### KYC Verification

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/kyc/verify` | Initiate KYC verification |
| GET | `/v1/kyc/status/{verificationId}` | Check verification status |
| GET | `/v1/kyc/application/{applicationId}` | Verifications for application |
| POST | `/v1/kyc/webhook/{provider}` | Receive provider webhooks (no auth) |

### AML Screening

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/aml/screen` | Screen individual against sanctions/PEP |
| POST | `/v1/aml/monitor` | Monitor transaction |
| GET | `/v1/aml/alerts` | List alerts (`?status=`) |

### Payments & Regulatory

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/payments/validate` | Validate payin/payout compliance |
| GET | `/v1/regulatory/{jurisdiction}` | Get jurisdiction requirements |

### System

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Health check (no auth) |
| GET | `/v1/providers` | List registered IDV providers |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `COMPLIANCE_LISTEN` | `:8091` | HTTP listen address |
| `COMPLIANCE_API_KEY` | — | API key for authenticated endpoints |
| `KYC_DEFAULT_PROVIDER` | first registered | Default IDV provider |

**IDV Provider credentials** — set for each provider:

| Provider | Variables |
|----------|-----------|
| Jumio | `JUMIO_API_TOKEN`, `JUMIO_API_SECRET`, `JUMIO_WEBHOOK_SECRET` |
| Onfido | `ONFIDO_API_TOKEN`, `ONFIDO_WEBHOOK_SECRET` |
| Plaid | `PLAID_CLIENT_ID`, `PLAID_SECRET`, `PLAID_WEBHOOK_SECRET` |

## Build & Test

```bash
make build          # Build binary
make test           # Run tests (173 tests across 7 packages)
make test-race      # Run with race detector (0 data races)
make lint           # go vet
make docker         # Build Docker image
make docker-push    # Push to ghcr.io
```

## Docker

```bash
docker build --platform linux/amd64 -t ghcr.io/luxfi/compliance:latest .
docker run -p 8091:8091 \
  -e COMPLIANCE_API_KEY=your-key \
  -e JUMIO_API_TOKEN=... \
  -e JUMIO_API_SECRET=... \
  ghcr.io/luxfi/compliance:latest
```

Image: `ghcr.io/luxfi/compliance` — 6.8 MB, alpine-based, healthcheck on `/healthz`.

## Integration

### Go (import as library)

```go
import (
    "github.com/luxfi/compliance/pkg/kyc"
    "github.com/luxfi/compliance/pkg/idv"
    "github.com/luxfi/compliance/pkg/aml"
)

svc := kyc.NewService()
svc.RegisterProvider(idv.NewJumio(idv.JumioConfig{...}))
```

### TypeScript (bank SDK)

```typescript
import { ComplianceModule } from '@luxbank/compliance'

// In NestJS app.module.ts
ComplianceModule.forRoot({
  baseUrl: process.env.COMPLIANCE_BASE_URL || 'http://compliance:8091',
  apiKey: process.env.COMPLIANCE_API_KEY || '',
})

// Inject anywhere
constructor(private readonly compliance: ComplianceService) {}
await compliance.initiateKYC(applicationId, 'jumio')
```

The `@luxbank/compliance` TypeScript SDK lives in [luxfi/bank/pkg/compliance](https://github.com/luxfi/bank/tree/main/pkg/compliance).

## Project Structure

```
compliance/
├── cmd/complianced/        Standalone HTTP server
├── pkg/
│   ├── idv/                Identity verification providers
│   │   ├── provider.go     Provider interface + registry
│   │   ├── jumio.go        Jumio Netverify v4
│   │   ├── onfido.go       Onfido v3.6
│   │   └── plaid.go        Plaid Identity Verification
│   ├── kyc/                KYC orchestration
│   │   ├── kyc.go          Service (multi-provider, webhooks)
│   │   └── application.go  Application model + store
│   ├── aml/                AML compliance
│   │   ├── screening.go    Sanctions/PEP screening
│   │   └── monitoring.go   Transaction monitoring rules
│   ├── regulatory/         Jurisdiction framework
│   │   ├── jurisdiction.go Interface + factory
│   │   ├── usa.go          FinCEN/SEC/FINRA rules
│   │   ├── uk.go           FCA/5AMLD rules
│   │   └── iom.go          IOMFSA rules
│   ├── payments/           Payment compliance
│   │   ├── compliance.go   Travel rule, CTR, validation
│   │   └── stablecoin.go   Token policies, address risk
│   ├── entity/             Regulated entity types
│   │   └── entity.go       ATS, BD, TA, MSB definitions
│   └── webhook/            Unified webhook handler
│       └── handler.go      Routing, sig validation, idempotency
├── Dockerfile
├── Makefile
└── go.mod                  Zero external dependencies
```

## Related Projects

| Module | Purpose |
|--------|---------|
| [luxfi/broker](https://github.com/luxfi/broker) | Multi-venue trading router, settlement engine |
| [luxfi/bank](https://github.com/luxfi/bank) | Customer apps, payments, admin dashboard |
| [hanzoai/iam](https://github.com/hanzoai/iam) | Identity and access management (base `idv/` pattern) |
| [hanzoai/commerce](https://github.com/hanzoai/commerce) | Payment processors (Plaid Link, Braintree, Stripe) |

## License

Copyright 2024-2026 Lux Partners Limited. All rights reserved.
