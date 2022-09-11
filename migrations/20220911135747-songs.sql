-- +migrate Up
CREATE TABLE songs (
  eggs_id TEXT NOT NULL,
  music_id TEXT NOT NULL,
	release_date TIMESTAMP(3) WITH TIME ZONE NOT NULL,
  FOREIGN KEY (eggs_id) REFERENCES users (eggs_id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX unique_songs ON songs (eggs_id, music_id);
-- +migrate Down
DROP TABLE songs;