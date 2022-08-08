-- +migrate Up
CREATE TABLE user_likes (
  eggs_id TEXT NOT NULL,
  target_id TEXT NOT NULL,
  target_type TEXT NOT NULL,
  FOREIGN KEY (eggs_id) REFERENCES users (eggs_id) ON DELETE CASCADE
);
CREATE INDEX likes_eggs_id_index ON user_likes (eggs_id);
CREATE INDEX likes_target_index ON user_likes (target_id);
-- +migrate Down
DROP TABLE user_likes;