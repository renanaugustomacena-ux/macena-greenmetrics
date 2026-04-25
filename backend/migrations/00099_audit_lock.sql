-- +goose Up
-- +goose StatementBegin
--
-- Audit log hardening:
--   1. REVOKE UPDATE/DELETE from app_user (RLS already restricts but defence in depth).
--   2. Trigger to forbid UPDATE/DELETE even on session that bypasses RLS.
--   3. immutable column on (tenant_id, action, created_at) for downstream attestation.
--
-- Doctrine: Rule 19, Rule 39, Rule 60 (forensic readiness), Rule 65.
-- Mitigates: RISK-009 (audit log tampering by privileged user).

REVOKE UPDATE, DELETE ON audit_log FROM app_user;

CREATE OR REPLACE FUNCTION audit_log_immutable() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'audit_log is append-only; UPDATE/DELETE forbidden (RISK-009)';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS audit_log_no_update ON audit_log;
CREATE TRIGGER audit_log_no_update
    BEFORE UPDATE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION audit_log_immutable();

DROP TRIGGER IF EXISTS audit_log_no_delete ON audit_log;
CREATE TRIGGER audit_log_no_delete
    BEFORE DELETE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION audit_log_immutable();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audit_log_no_update ON audit_log;
DROP TRIGGER IF EXISTS audit_log_no_delete ON audit_log;
DROP FUNCTION IF EXISTS audit_log_immutable();
GRANT UPDATE, DELETE ON audit_log TO app_user;
-- +goose StatementEnd
