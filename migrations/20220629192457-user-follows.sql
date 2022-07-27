-- +migrate Up
CREATE TABLE user_follows (
  follower_id TEXT NOT NULL,
  followee_id TEXT NOT NULL,
  FOREIGN KEY (follower_id) REFERENCES users (eggs_id) ON DELETE CASCADE,
  FOREIGN KEY (followee_id) REFERENCES users (eggs_id) ON DELETE CASCADE
);
CREATE INDEX follows_follower_index ON user_follows (follower_id);
CREATE INDEX follows_followee_index ON user_follows (followee_id);
-- +migrate Down
DROP TABLE user_follows;