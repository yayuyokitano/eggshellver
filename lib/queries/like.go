package queries

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type rawLike struct {
	TrackID        string    `db:"track_id"`
	EggsID         string    `db:"eggs_id"`
	DisplayName    string    `db:"display_name"`
	IsArtist       bool      `db:"is_artist"`
	ImageDataPath  string    `db:"image_data_path"`
	PrefectureCode int       `db:"prefecture_code"`
	ProfileText    string    `db:"profile_text"`
	AddedTime      time.Time `db:"added_time"`
}
type rawLikes []rawLike

type StructuredLike struct {
	TrackID   string    `json:"trackID"`
	User      UserStub  `json:"user"`
	Timestamp time.Time `json:"timestamp"`
}
type StructuredLikes struct {
	Likes []StructuredLike `json:"likes"`
	Total int64            `json:"total"`
}

func (r rawLike) ToLike() StructuredLike {
	return StructuredLike{
		TrackID: r.TrackID,
		User: UserStub{
			EggsID:         r.EggsID,
			DisplayName:    r.DisplayName,
			IsArtist:       r.IsArtist,
			ImageDataPath:  r.ImageDataPath,
			PrefectureCode: r.PrefectureCode,
			ProfileText:    r.ProfileText,
		},
		Timestamp: r.AddedTime,
	}
}

func (arr rawLikes) ToLikes(total int64) (likes StructuredLikes) {
	likeSlice := make([]StructuredLike, 0)
	for _, r := range arr {
		likeSlice = append(likeSlice, r.ToLike())
	}
	likes = StructuredLikes{
		Likes: likeSlice,
		Total: total,
	}
	return
}

func (arr StructuredLikes) Contains(b PartialLike) bool {
	for _, a := range arr.Likes {
		if a.TrackID == b.TrackID && a.User.EggsID == b.EggsID {
			return true
		}
	}
	return false
}

type Like struct {
	TrackID   string    `json:"trackID" db:"track_id"`
	EggsID    string    `json:"eggsID" db:"eggs_id"`
	Timestamp time.Time `json:"timestamp" db:"added_time"`
}
type Likes []Like

type PartialLike struct {
	TrackID string `json:"trackID" db:"track_id"`
	EggsID  string `json:"eggsID" db:"eggs_id"`
}

func GetLikedTracks(ctx context.Context, eggsIDs []string, trackIDs []string, paginator Paginator) (likes StructuredLikes, err error) {
	if len(trackIDs) == 0 && len(eggsIDs) == 0 {
		err = errors.New("no track IDs or eggs IDs")
		return
	}
	var query string
	var query2 string
	prefix := "SELECT ul.track_id, u.eggs_id, u.display_name, u.is_artist, u.image_data_path, u.prefecture_code, u.profile_text, ul.added_time FROM user_likes ul INNER JOIN users u ON ul.eggs_id = u.eggs_id AND "
	prefix2 := "SELECT COUNT(*) FROM user_likes WHERE "
	args := make([]interface{}, 0)
	if len(trackIDs) == 0 {
		query = prefix + "ul.eggs_id = ANY($1) ORDER BY added_time DESC LIMIT $2 OFFSET $3"
		query2 = prefix2 + "eggs_id = ANY($1)"
		args = append(args, eggsIDs)
	} else if len(eggsIDs) == 0 {
		query = prefix + "ul.track_id = ANY($1) ORDER BY added_time DESC LIMIT $2 OFFSET $3"
		query2 = prefix2 + "track_id = ANY($1)"
		args = append(args, trackIDs)
	} else {
		query = prefix + "ul.track_id = ANY($1) AND ul.eggs_id = ANY($2) ORDER BY added_time DESC LIMIT $3 OFFSET $4"
		query2 = prefix2 + "track_id = ANY($1) AND eggs_id = ANY($2)"
		args = append(args, trackIDs, eggsIDs)
	}
	args = append(args, paginator.Limit, paginator.Offset)
	rawLikes := make(rawLikes, 0)
	tx, err := fetchTransaction()
	if err != nil {
		return
	}
	err = pgxscan.Select(
		ctx,
		tx,
		&rawLikes,
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
	likes = rawLikes.ToLikes(total)
	return
}

func LikeTracks(ctx context.Context, eggsID string, trackIDs []string) (n int64, err error) {
	timestamp := time.Now().UnixMilli()
	likes := make([][]interface{}, 0)
	for i, trackID := range trackIDs {
		likes = append(likes, []interface{}{eggsID, trackID, time.UnixMilli(timestamp - int64(i))})
	}
	tx, err := fetchTransaction()
	if err != nil {
		return
	}
	n, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"user_likes"},
		[]string{"eggs_id", "track_id", "added_time"},
		pgx.CopyFromRows(likes),
	)
	if err != nil {
		return
	}
	err = commitTransaction(tx)
	return
}

func DeleteLikes(ctx context.Context, eggsID string, trackIDs []string) (err error) {
	tx, err := fetchTransaction()
	if err != nil {
		return
	}
	_, err = tx.Exec(
		ctx,
		"DELETE FROM user_likes WHERE eggs_id = $1 AND track_id = ANY($2)",
		eggsID,
		trackIDs,
	)
	if err != nil {
		return
	}
	err = commitTransaction(tx)
	return
}
