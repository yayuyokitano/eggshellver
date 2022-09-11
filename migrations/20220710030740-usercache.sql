-- +migrate Up
ALTER TABLE users ADD COLUMN last_modified TIMESTAMP(3) WITH TIME ZONE NOT NULL DEFAULT NOW();

-- +migrate Down
ALTER TABLE users DROP COLUMN last_modified;