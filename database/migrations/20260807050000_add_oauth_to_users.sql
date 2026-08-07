-- +goose Up
-- +goose StatementBegin
ALTER TABLE users 
ADD COLUMN auth_provider VARCHAR(50) DEFAULT 'local',
ADD COLUMN provider_id VARCHAR(255) NULL,
MODIFY COLUMN password VARCHAR(255) NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users 
DROP COLUMN auth_provider,
DROP COLUMN provider_id,
MODIFY COLUMN password VARCHAR(255) NOT NULL;
-- +goose StatementEnd
