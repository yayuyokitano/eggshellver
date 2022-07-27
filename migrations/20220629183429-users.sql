-- +migrate Up
CREATE TABLE users (
  eggs_id TEXT NOT NULL PRIMARY KEY,
  display_name TEXT NOT NULL,
  is_artist BOOLEAN NOT NULL,
  image_data_path TEXT NOT NULL DEFAULT '',
  prefecture_code INTEGER NOT NULL DEFAULT 0,
  profile_text TEXT NOT NULL DEFAULT '',
  token TEXT NOT NULL DEFAULT ''
);
CREATE INDEX user_prefecture_index ON users (prefecture_code);
CREATE INDEX user_display_name_index ON users (display_name);
CREATE INDEX user_token_index ON users (token);
-- +migrate Down
DROP TABLE users;