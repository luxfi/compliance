-- Compliance module database tables.
-- All tables use upsert (ON CONFLICT DO UPDATE) so this migration is idempotent.

CREATE TABLE IF NOT EXISTS identities (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    provider    TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'pending',
    data        JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_identities_user ON identities(user_id);

CREATE TABLE IF NOT EXISTS pipelines (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    business_id TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'draft',
    steps       JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    id              TEXT PRIMARY KEY,
    pipeline_id     TEXT NOT NULL DEFAULT '',
    investor_email  TEXT NOT NULL DEFAULT '',
    investor_name   TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'pending',
    kyc_status      TEXT NOT NULL DEFAULT 'pending',
    steps           JSONB NOT NULL DEFAULT '[]',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS funds (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    business_id     TEXT NOT NULL DEFAULT '',
    type            TEXT NOT NULL DEFAULT 'equity',
    min_investment  DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_raised    DOUBLE PRECISION NOT NULL DEFAULT 0,
    investor_count  INTEGER NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'raising',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fund_investors (
    id          TEXT PRIMARY KEY,
    fund_id     TEXT NOT NULL REFERENCES funds(id) ON DELETE CASCADE,
    investor_id TEXT NOT NULL DEFAULT '',
    name        TEXT NOT NULL DEFAULT '',
    email       TEXT NOT NULL DEFAULT '',
    amount      DOUBLE PRECISION NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'committed',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fund_investors_fund ON fund_investors(fund_id);

CREATE TABLE IF NOT EXISTS envelopes (
    id          TEXT PRIMARY KEY,
    template_id TEXT NOT NULL DEFAULT '',
    subject     TEXT NOT NULL DEFAULT '',
    message     TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'pending',
    signers     JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS templates (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    content     TEXT NOT NULL DEFAULT '',
    roles       JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS roles (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS compliance_users (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL DEFAULT '',
    email       TEXT NOT NULL DEFAULT '',
    role        TEXT NOT NULL DEFAULT 'agent',
    status      TEXT NOT NULL DEFAULT 'active',
    last_login  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    id      TEXT PRIMARY KEY,
    type    TEXT NOT NULL DEFAULT '',
    asset   TEXT NOT NULL DEFAULT '',
    amount  DOUBLE PRECISION NOT NULL DEFAULT 0,
    fee     DOUBLE PRECISION NOT NULL DEFAULT 0,
    status  TEXT NOT NULL DEFAULT 'pending',
    date    TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS credentials (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL DEFAULT '',
    key_prefix  TEXT NOT NULL DEFAULT '',
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS settings (
    id                  INTEGER PRIMARY KEY DEFAULT 1,
    business_name       TEXT NOT NULL DEFAULT 'Your Company',
    timezone            TEXT NOT NULL DEFAULT 'America/New_York',
    currency            TEXT NOT NULL DEFAULT 'USD',
    notification_email  TEXT NOT NULL DEFAULT 'compliance@example.com'
);

CREATE TABLE IF NOT EXISTS aml_screenings (
    id             TEXT PRIMARY KEY,
    account_id     TEXT NOT NULL,
    user_id        TEXT NOT NULL DEFAULT '',
    type           TEXT NOT NULL DEFAULT 'sanctions',
    status         TEXT NOT NULL DEFAULT 'pending',
    risk_level     TEXT NOT NULL DEFAULT 'low',
    risk_score     DOUBLE PRECISION NOT NULL DEFAULT 0,
    provider       TEXT NOT NULL DEFAULT 'manual',
    details        TEXT NOT NULL DEFAULT '',
    sanctions_hit  BOOLEAN NOT NULL DEFAULT FALSE,
    pep_match      BOOLEAN NOT NULL DEFAULT FALSE,
    reviewed_by    TEXT NOT NULL DEFAULT '',
    reviewed_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_aml_screenings_account ON aml_screenings(account_id);
CREATE INDEX IF NOT EXISTS idx_aml_screenings_status ON aml_screenings(status);

CREATE TABLE IF NOT EXISTS applications (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL,
    email          TEXT NOT NULL DEFAULT '',
    first_name     TEXT NOT NULL DEFAULT '',
    last_name      TEXT NOT NULL DEFAULT '',
    phone          TEXT NOT NULL DEFAULT '',
    date_of_birth  TEXT NOT NULL DEFAULT '',
    ssn_hash       TEXT NOT NULL DEFAULT '',
    ssn_last4      TEXT NOT NULL DEFAULT '',
    address_line1  TEXT NOT NULL DEFAULT '',
    address_line2  TEXT NOT NULL DEFAULT '',
    city           TEXT NOT NULL DEFAULT '',
    state          TEXT NOT NULL DEFAULT '',
    zip_code       TEXT NOT NULL DEFAULT '',
    country        TEXT NOT NULL DEFAULT '',
    status         TEXT NOT NULL DEFAULT 'draft',
    current_step   INTEGER NOT NULL DEFAULT 1,
    kyc_status     TEXT NOT NULL DEFAULT 'pending',
    aml_status     TEXT NOT NULL DEFAULT 'pending',
    steps          JSONB NOT NULL DEFAULT '[]',
    documents      JSONB NOT NULL DEFAULT '[]',
    risk_level     TEXT NOT NULL DEFAULT '',
    risk_score     DOUBLE PRECISION NOT NULL DEFAULT 0,
    submitted_at   TIMESTAMPTZ,
    reviewed_by    TEXT NOT NULL DEFAULT '',
    reviewed_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_applications_user ON applications(user_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);

CREATE TABLE IF NOT EXISTS document_uploads (
    id              TEXT PRIMARY KEY,
    application_id  TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    user_id         TEXT NOT NULL DEFAULT '',
    type            TEXT NOT NULL,
    name            TEXT NOT NULL,
    mime_type       TEXT NOT NULL DEFAULT '',
    size            BIGINT NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'pending',
    review_note     TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_document_uploads_application ON document_uploads(application_id);
