-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audit_log_no_update ON audit_log;
DROP TRIGGER IF EXISTS audit_log_no_delete ON audit_log;
DROP FUNCTION IF EXISTS audit_log_immutable();
GRANT UPDATE, DELETE ON audit_log TO app_user;
-- +goose StatementEnd
