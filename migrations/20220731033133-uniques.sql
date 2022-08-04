-- +migrate Up
CREATE UNIQUE INDEX unique_follows ON user_follows (follower_id, followee_id);
CREATE UNIQUE INDEX unique_likes ON user_likes (eggs_id, track_id);

-- +migrate Down
DROP INDEX unique_follows;
DROP INDEX unique_likes;