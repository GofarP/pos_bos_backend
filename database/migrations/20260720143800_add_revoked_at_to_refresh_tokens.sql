-- +goose Up
-- +goose StatementBegin
ALTER TABLE refresh_tokens ADD COLUMN revoked_at TIMESTAMP NULL DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE refresh_tokens DROP COLUMN revoked_at;
-- +goose StatementEnd
