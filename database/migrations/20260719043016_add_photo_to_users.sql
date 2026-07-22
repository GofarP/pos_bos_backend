-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN photo VARCHAR(255) NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN photo;
-- +goose StatementEnd
