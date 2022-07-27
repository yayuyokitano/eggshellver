-- +migrate Up
CREATE TABLE user_likes (
  eggs_id TEXT NOT NULL,
  track_id TEXT NOT NULL,
  FOREIGN KEY (eggs_id) REFERENCES users (eggs_id) ON DELETE CASCADE
);
CREATE INDEX likes_eggs_id_index ON user_likes (eggs_id);
CREATE INDEX likes_track_index ON user_likes (track_id);
-- +migrate Down
DROP TABLE user_likes;