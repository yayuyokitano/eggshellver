-- +migrate Up
CREATE TABLE playlists (
  playlist_id TEXT NOT NULL PRIMARY KEY,
  eggs_id TEXT NOT NULL,
  FOREIGN KEY (eggs_id) REFERENCES users (eggs_id) ON DELETE CASCADE
);
CREATE INDEX playlist_eggs_id_index ON playlists (eggs_id);
-- +migrate Down
DROP TABLE playlists;