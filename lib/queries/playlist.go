package queries

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type rawPlaylist struct {
	PlaylistID     string    `db:"playlist_id"`
	EggsID         string    `db:"eggs_id"`
	DisplayName    string    `db:"display_name"`
	IsArtist       bool      `db:"is_artist"`
	ImageDataPath  string    `db:"image_data_path"`
	PrefectureCode int       `db:"prefecture_code"`
	ProfileText    string    `db:"profile_text"`
	LastModified   time.Time `db:"last_modified"`
}
type rawPlaylists []rawPlaylist

type StructuredPlaylist struct {
	PlaylistID string    `json:"playlistID"`
	User       UserStub  `json:"user"`
	Timestamp  time.Time `json:"timestamp"`
}
type StructuredPlaylists struct {
	Playlists []StructuredPlaylist `json:"playlists"`
	Total     int64                `json:"total"`
}

func (r rawPlaylist) ToPlaylist() StructuredPlaylist {
	return StructuredPlaylist{
		PlaylistID: r.PlaylistID,
		User: UserStub{
			EggsID:         r.EggsID,
			DisplayName:    r.DisplayName,
			IsArtist:       r.IsArtist,
			ImageDataPath:  r.ImageDataPath,
			PrefectureCode: r.PrefectureCode,
			ProfileText:    r.ProfileText,
		},
		Timestamp: r.LastModified,
	}
}

func (arr rawPlaylists) ToPlaylists(total int64) (playlists StructuredPlaylists) {
	playlistSlice := make([]StructuredPlaylist, 0)
	for _, r := range arr {
		playlistSlice = append(playlistSlice, r.ToPlaylist())
	}
	playlists = StructuredPlaylists{
		Playlists: playlistSlice,
		Total:     total,
	}
	return
}

func (arr StructuredPlaylists) Contains(b PartialPlaylist) bool {
	for _, a := range arr.Playlists {
		if a.PlaylistID == b.PlaylistID && a.User.EggsID == b.EggsID {
			return true
		}
	}
	return false
}

type PartialPlaylist struct {
	PlaylistID string `json:"playlistID" db:"playlist_id"`
	EggsID     string `json:"eggsID" db:"eggs_id"`
}

func GetPlaylists(ctx context.Context, eggsIDs []string, playlistIDs []string, paginator Paginator) (playlists StructuredPlaylists, err error) {
	if len(playlistIDs) == 0 && len(eggsIDs) == 0 {
		err = errors.New("no playlist IDs or eggs IDs")
		return
	}

	var query string
	var query2 string
	prefix := "SELECT ul.playlist_id, u.eggs_id, u.display_name, u.is_artist, u.image_data_path, u.prefecture_code, u.profile_text, ul.last_modified FROM playlists ul INNER JOIN users u ON ul.eggs_id = u.eggs_id AND "
	prefix2 := "SELECT COUNT(*) FROM playlists WHERE "
	args := make([]interface{}, 0)
	if len(playlistIDs) == 0 {
		query = prefix + "ul.eggs_id = ANY($1) ORDER BY last_modified DESC LIMIT $2 OFFSET $3"
		query2 = prefix2 + "eggs_id = ANY($1)"
		args = append(args, eggsIDs)
	} else if len(eggsIDs) == 0 {
		query = prefix + "ul.playlist_id = ANY($1) ORDER BY last_modified DESC LIMIT $2 OFFSET $3"
		query2 = prefix2 + "playlist_id = ANY($1)"
		args = append(args, playlistIDs)
	} else {
		query = prefix + "ul.playlist_id = ANY($1) AND ul.eggs_id = ANY($2) ORDER BY last_modified DESC LIMIT $3 OFFSET $4"
		query2 = prefix2 + "playlist_id = ANY($1) AND eggs_id = ANY($2)"
		args = append(args, playlistIDs, eggsIDs)
	}
	args = append(args, paginator.Limit, paginator.Offset)
	rawPlaylists := make(rawPlaylists, 0)
	tx, err := fetchTransaction()
	if err != nil {
		return
	}
	err = pgxscan.Select(
		ctx,
		tx,
		&rawPlaylists,
		query,
		args...,
	)
	if err != nil {
		return
	}
	var total int64
	err = tx.QueryRow(ctx, query2, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return
	}
	err = commitTransaction(tx)
	playlists = rawPlaylists.ToPlaylists(total)
	return
}

func PostPlaylists(ctx context.Context, eggsID string, playlistIDs []string) (n int64, err error) {
	timestamp := time.Now().UnixMilli()
	playlists := make([][]interface{}, 0)
	for i, playlistID := range playlistIDs {
		playlists = append(playlists, []interface{}{eggsID, playlistID, time.UnixMilli(timestamp - int64(i))})
	}
	tx, err := fetchTransaction()
	if err != nil {
		return
	}
	n, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"playlists"},
		[]string{"eggs_id", "playlist_id", "last_modified"},
		pgx.CopyFromRows(playlists),
	)
	if err != nil {
		return
	}
	err = commitTransaction(tx)
	return
}

func DeletePlaylists(ctx context.Context, eggsID string, playlistIDs []string) (err error) {
	tx, err := fetchTransaction()
	if err != nil {
		return
	}
	_, err = tx.Exec(
		ctx,
		"DELETE FROM playlists WHERE eggs_id = $1 AND playlist_id = ANY($2)",
		eggsID,
		playlistIDs,
	)
	if err != nil {
		return
	}
	err = commitTransaction(tx)
	return
}
