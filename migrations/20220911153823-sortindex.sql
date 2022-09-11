-- +migrate Up
CREATE INDEX playlists_last_modified ON playlists (last_modified desc);
CREATE INDEX likes_added_time ON user_likes (added_time desc);
CREATE INDEX follows_added_time ON user_follows (added_time desc);
CREATE INDEX songs_release_date ON songs (release_date desc);

-- +migrate Down
DROP INDEX playlists_last_modified;
DROP INDEX likes_added_time;
DROP INDEX follows_added_time;
DROP INDEX songs_release_date;