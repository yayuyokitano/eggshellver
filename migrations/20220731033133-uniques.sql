-- +migrate Up
CREATE UNIQUE INDEX unique_follows ON user_follows (follower_id, followee_id);
CREATE UNIQUE INDEX unique_likes ON user_likes (eggs_id, track_id);

-- +migrate Down
ALTER TABLE users DROP INDEX unique_follows;
ALTER TABLE users DROP INDEX unique_likes;