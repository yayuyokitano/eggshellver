package queries

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
)

type ArtistData struct {
	ArtistID       int    `json:"artistId"`
	ArtistName     string `json:"artistName"`
	DisplayName    string `json:"displayName"`
	PrefectureCode int    `json:"prefectureCode"`
	ImageDataPath  string `json:"imageDataPath"`
	Profile        string `json:"profile"`
}

type SongData struct {
	MusicID     string     `json:"musicId"`
	ReleaseDate time.Time  `json:"releaseDate"`
	ArtistData  ArtistData `json:"artistData"`
}

type SearchSongResp struct {
	Data       []SongData `json:"data"`
	TotalCount int        `json:"totalCount"`
}

func PostSongs(ctx context.Context, songData []SongData) (n int64, err error) {
	songs := make([][]interface{}, 0)
	for _, song := range songData {
		songs = append(songs, []interface{}{song.ArtistData.ArtistName, song.MusicID, song.ReleaseDate})
	}
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(ctx, "DROP TABLE IF EXISTS _temp_upsert_songs")
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(ctx, "CREATE TEMPORARY TABLE _temp_upsert_songs (LIKE songs INCLUDING ALL) ON COMMIT DROP")
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"_temp_upsert_songs"},
		[]string{"eggs_id", "music_id", "release_date"},
		pgx.CopyFromRows(songs),
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	cmd, err := tx.Exec(ctx, "INSERT INTO songs SELECT * FROM _temp_upsert_songs ON CONFLICT DO NOTHING")
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	n = cmd.RowsAffected()
	err = commitTransaction(tx, "_temp_upsert_songs")
	return
}

func SongExists(ctx context.Context, musicID string) (exists bool, err error) {
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = tx.QueryRow(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM songs WHERE music_id = $1)",
		musicID,
	).Scan(&exists)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = commitTransaction(tx)
	return
}

func GetSongCount(ctx context.Context) (n int64, err error) {
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = tx.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM songs",
	).Scan(&n)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = commitTransaction(tx)
	return
}
