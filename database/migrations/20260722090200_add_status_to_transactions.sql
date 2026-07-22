-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions ADD COLUMN status VARCHAR(20) DEFAULT 'COMPLETED' AFTER idempotency_key;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions DROP COLUMN status;
-- +goose StatementEnd
