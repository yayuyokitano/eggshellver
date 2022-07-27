-- +migrate Up
ALTER TABLE users ADD COLUMN last_modified TIMESTAMP(3) WITH TIME ZONE NOT NULL DEFAULT NOW();
CREATE INDEX user_last_modified_index ON users (last_modified);

-- +migrate Down
ALTER TABLE users DROP COLUMN last_modified;