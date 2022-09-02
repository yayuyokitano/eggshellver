-- +migrate Up
ALTER TABLE playlists ADD COLUMN IF NOT EXISTS last_modified TIMESTAMP(3) WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE user_likes ADD COLUMN IF NOT EXISTS added_time TIMESTAMP(3) WITH TIME ZONE NOT NULL DEFAULT NOW();
ALTER TABLE user_follows ADD COLUMN IF NOT EXISTS added_time TIMESTAMP(3) WITH TIME ZONE NOT NULL DEFAULT NOW();

-- +migrate Down
ALTER TABLE playlists DROP COLUMN last_modified;
ALTER TABLE user_likes DROP COLUMN added_time;
ALTER TABLE user_follows DROP COLUMN added_time;