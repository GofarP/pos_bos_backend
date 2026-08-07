-- +goose Up
ALTER TABLE roles ADD COLUMN description TEXT;

-- +goose Down
ALTER TABLE roles DROP COLUMN description;
