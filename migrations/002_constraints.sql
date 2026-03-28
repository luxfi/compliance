-- Additional constraints and indexes for compliance tables.
-- Idempotent: uses IF NOT EXISTS and DO NOTHING patterns.

-- Status CHECK constraints prevent invalid state values at the database level,
-- providing defense-in-depth beyond application-level validation.

-- identities: valid KYC statuses
ALTER TABLE identities DROP CONSTRAINT IF EXISTS chk_identities_status;
ALTER TABLE identities ADD CONSTRAINT chk_identities_status
    CHECK (status IN ('pending', 'verified', 'failed', 'expired'));

-- sessions: valid session and KYC statuses
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS chk_sessions_status;
ALTER TABLE sessions ADD CONSTRAINT chk_sessions_status
    CHECK (status IN ('pending', 'in_progress', 'completed', 'failed', 'archived'));

ALTER TABLE sessions DROP CONSTRAINT IF EXISTS chk_sessions_kyc_status;
ALTER TABLE sessions ADD CONSTRAINT chk_sessions_kyc_status
    CHECK (kyc_status IN ('pending', 'verified', 'failed', 'expired'));

-- funds: valid statuses and non-negative amounts
ALTER TABLE funds DROP CONSTRAINT IF EXISTS chk_funds_status;
ALTER TABLE funds ADD CONSTRAINT chk_funds_status
    CHECK (status IN ('open', 'closed', 'raising'));

ALTER TABLE funds DROP CONSTRAINT IF EXISTS chk_funds_min_investment;
ALTER TABLE funds ADD CONSTRAINT chk_funds_min_investment
    CHECK (min_investment >= 0);

ALTER TABLE funds DROP CONSTRAINT IF EXISTS chk_funds_total_raised;
ALTER TABLE funds ADD CONSTRAINT chk_funds_total_raised
    CHECK (total_raised >= 0);

-- fund_investors: non-negative amount
ALTER TABLE fund_investors DROP CONSTRAINT IF EXISTS chk_fund_investors_amount;
ALTER TABLE fund_investors ADD CONSTRAINT chk_fund_investors_amount
    CHECK (amount >= 0);

-- envelopes: valid statuses
ALTER TABLE envelopes DROP CONSTRAINT IF EXISTS chk_envelopes_status;
ALTER TABLE envelopes ADD CONSTRAINT chk_envelopes_status
    CHECK (status IN ('pending', 'sent', 'viewed', 'signed', 'completed', 'declined', 'voided'));

-- aml_screenings: valid statuses, risk levels, non-negative score
ALTER TABLE aml_screenings DROP CONSTRAINT IF EXISTS chk_aml_status;
ALTER TABLE aml_screenings ADD CONSTRAINT chk_aml_status
    CHECK (status IN ('pending', 'cleared', 'flagged', 'blocked', 'expired'));

ALTER TABLE aml_screenings DROP CONSTRAINT IF EXISTS chk_aml_risk_level;
ALTER TABLE aml_screenings ADD CONSTRAINT chk_aml_risk_level
    CHECK (risk_level IN ('low', 'medium', 'high', 'critical'));

ALTER TABLE aml_screenings DROP CONSTRAINT IF EXISTS chk_aml_risk_score;
ALTER TABLE aml_screenings ADD CONSTRAINT chk_aml_risk_score
    CHECK (risk_score >= 0);

-- applications: valid statuses, step range
ALTER TABLE applications DROP CONSTRAINT IF EXISTS chk_applications_status;
ALTER TABLE applications ADD CONSTRAINT chk_applications_status
    CHECK (status IN ('draft', 'in_progress', 'submitted', 'under_review', 'approved', 'rejected'));

ALTER TABLE applications DROP CONSTRAINT IF EXISTS chk_applications_current_step;
ALTER TABLE applications ADD CONSTRAINT chk_applications_current_step
    CHECK (current_step >= 1 AND current_step <= 5);

ALTER TABLE applications DROP CONSTRAINT IF EXISTS chk_applications_kyc_status;
ALTER TABLE applications ADD CONSTRAINT chk_applications_kyc_status
    CHECK (kyc_status IN ('pending', 'verified', 'failed', 'expired'));

ALTER TABLE applications DROP CONSTRAINT IF EXISTS chk_applications_aml_status;
ALTER TABLE applications ADD CONSTRAINT chk_applications_aml_status
    CHECK (aml_status IN ('pending', 'cleared', 'flagged', 'blocked', 'expired'));

-- transactions: non-negative amounts
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS chk_transactions_amount;
ALTER TABLE transactions ADD CONSTRAINT chk_transactions_amount
    CHECK (amount >= 0);

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS chk_transactions_fee;
ALTER TABLE transactions ADD CONSTRAINT chk_transactions_fee
    CHECK (fee >= 0);

-- document_uploads: non-negative size, valid statuses
ALTER TABLE document_uploads DROP CONSTRAINT IF EXISTS chk_uploads_size;
ALTER TABLE document_uploads ADD CONSTRAINT chk_uploads_size
    CHECK (size >= 0);

ALTER TABLE document_uploads DROP CONSTRAINT IF EXISTS chk_uploads_status;
ALTER TABLE document_uploads ADD CONSTRAINT chk_uploads_status
    CHECK (status IN ('pending', 'accepted', 'rejected'));

-- settings: singleton constraint
ALTER TABLE settings DROP CONSTRAINT IF EXISTS chk_settings_singleton;
ALTER TABLE settings ADD CONSTRAINT chk_settings_singleton
    CHECK (id = 1);

-- Missing indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_pipeline ON sessions(pipeline_id);
CREATE INDEX IF NOT EXISTS idx_envelopes_status ON envelopes(status);
CREATE INDEX IF NOT EXISTS idx_compliance_users_email ON compliance_users(email);
CREATE INDEX IF NOT EXISTS idx_applications_email ON applications(email);
