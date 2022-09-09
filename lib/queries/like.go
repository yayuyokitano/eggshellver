package queries

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type rawLike struct {
	ID             string    `db:"target_id"`
	Type           string    `db:"target_type"`
	UserID         int       `db:"user_id"`
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
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	User      UserStub  `json:"user"`
	Timestamp time.Time `json:"timestamp"`
}
type StructuredLikes struct {
	Likes []StructuredLike `json:"likes"`
	Total int64            `json:"total"`
}

func (r rawLike) ToLike() StructuredLike {
	return StructuredLike{
		ID:   r.ID,
		Type: r.Type,
		User: UserStub{
			UserID:         r.UserID,
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
		if a.ID == b.TargetID && a.User.EggsID == b.EggsID {
			return true
		}
	}
	return false
}

type Like struct {
	TargetID  string    `json:"targetID" db:"target_id"`
	EggsID    string    `json:"eggsID" db:"eggs_id"`
	Timestamp time.Time `json:"timestamp" db:"added_time"`
}
type Likes []Like

type PartialLike struct {
	TargetID string `json:"targetID" db:"target_id"`
	EggsID   string `json:"eggsID" db:"eggs_id"`
}

type LikeTarget struct {
	ID   string `json:"id" db:"target_id"`
	Type string `json:"type" db:"target_type"`
}

type LikeTargets []LikeTarget

type LikeTargetsFixed struct {
	Targets []LikeTarget `json:"targets"`
	Type    string       `json:"type"`
}

func (arr LikeTargets) IDs() []string {
	ids := make([]string, 0)
	for _, a := range arr {
		ids = append(ids, a.ID)
	}
	return ids
}

func (arr LikeTargetsFixed) IDs() []string {
	ids := make([]string, 0)
	for _, a := range arr.Targets {
		ids = append(ids, a.ID)
	}
	return ids
}

func (a LikeTarget) IsValid() bool {
	if a.ID == "" || a.Type != "track" && a.Type != "playlist" {
		return false
	}
	return true
}

func (arr LikeTargets) IsValid() bool {
	for _, a := range arr {
		if !a.IsValid() {
			return false
		}
	}
	return true
}

func (arr LikeTargetsFixed) IsValid() bool {
	arrType := arr.Type
	if arrType != "track" && arrType != "playlist" {
		return false
	}
	for _, a := range arr.Targets {
		if a.ID == "" || a.Type != arrType {
			return false
		}
	}
	return true
}

func GetLikedObjects(ctx context.Context, eggsIDs []string, targetIDs []string, targetType string, paginator Paginator) (likes StructuredLikes, err error) {
	if len(targetIDs) == 0 && len(eggsIDs) == 0 {
		err = errors.New("no target IDs or eggs IDs")
		return
	}

	query := "SELECT ul.target_id, ul.target_type, u.user_id, u.eggs_id, u.display_name, u.is_artist, u.image_data_path, u.prefecture_code, u.profile_text, ul.added_time FROM user_likes ul INNER JOIN users u ON ul.eggs_id = u.eggs_id AND "
	query2 := "SELECT COUNT(*) FROM user_likes WHERE "
	args := make([]interface{}, 0)

	if targetType == "track" {
		query += "ul.target_type = 'track' AND "
		query2 += "target_type = 'track' AND "
	} else if targetType == "playlist" {
		query += "ul.target_type = 'playlist' AND "
		query2 += "target_type = 'playlist' AND "
	}

	if len(targetIDs) == 0 {
		query += "ul.eggs_id = ANY($1) ORDER BY added_time DESC LIMIT $2 OFFSET $3"
		query2 += "eggs_id = ANY($1)"
		args = append(args, eggsIDs)
	} else if len(eggsIDs) == 0 {
		query += "ul.target_id = ANY($1) ORDER BY added_time DESC LIMIT $2 OFFSET $3"
		query2 += "target_id = ANY($1)"
		args = append(args, targetIDs)
	} else {
		query += "ul.target_id = ANY($1) AND ul.eggs_id = ANY($2) ORDER BY added_time DESC LIMIT $3 OFFSET $4"
		query2 += "target_id = ANY($1) AND eggs_id = ANY($2)"
		args = append(args, targetIDs, eggsIDs)
	}
	args = append(args, paginator.Limit, paginator.Offset)
	rawLikes := make(rawLikes, 0)
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
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
		RollbackTransaction(tx)
		return
	}
	var total int64
	err = tx.QueryRow(ctx, query2, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = commitTransaction(tx)
	likes = rawLikes.ToLikes(total)
	return
}

func LikeObjects(ctx context.Context, eggsID string, targets LikeTargetsFixed) (n int64, err error) {
	timestamp := time.Now().UnixMilli()
	likes := make([][]interface{}, 0)
	for i, target := range targets.Targets {
		likes = append(likes, []interface{}{eggsID, target.ID, target.Type, time.UnixMilli(timestamp - int64(i))})
	}
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(ctx, "DROP TABLE IF EXISTS _temp_upsert_likes")
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(ctx, "CREATE TEMPORARY TABLE _temp_upsert_likes (LIKE user_likes INCLUDING ALL) ON COMMIT DROP")
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"_temp_upsert_likes"},
		[]string{"eggs_id", "target_id", "target_type", "added_time"},
		pgx.CopyFromRows(likes),
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	cmd, err := tx.Exec(ctx, "INSERT INTO user_likes SELECT * FROM _temp_upsert_likes ON CONFLICT DO NOTHING")
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	n = cmd.RowsAffected()
	err = commitTransaction(tx, "_temp_upsert_likes")
	return
}

func PutLikes(ctx context.Context, eggsID string, targets LikeTargetsFixed) (delta int64, total int64, err error) {
	timestamp := time.Now().UnixMilli()
	likes := make([][]interface{}, 0)
	for i, target := range targets.Targets {
		likes = append(likes, []interface{}{eggsID, target.ID, target.Type, time.UnixMilli(timestamp - int64(i))})
	}

	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(ctx, "DROP TABLE IF EXISTS _temp_upsert_likes")
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(ctx, "CREATE TEMPORARY TABLE _temp_upsert_likes (LIKE user_likes INCLUDING ALL) ON COMMIT DROP")
	if err != nil {
		RollbackTransaction(tx)
		return
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"_temp_upsert_likes"},
		[]string{"eggs_id", "target_id", "target_type", "added_time"},
		pgx.CopyFromRows(likes),
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}

	cmd, err := tx.Exec(ctx, "INSERT INTO user_likes SELECT * FROM _temp_upsert_likes ON CONFLICT DO NOTHING")
	if err != nil {
		RollbackTransaction(tx)
		return
	}

	delta = cmd.RowsAffected()

	err = tx.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM user_likes WHERE eggs_id = $1 AND target_type = $2",
		eggsID,
		targets.Type,
	).Scan(&total)
	if err != nil {
		RollbackTransaction(tx)
		return
	}

	cmd, err = tx.Exec(ctx, "DELETE FROM user_likes WHERE eggs_id = $1 AND target_id != ALL($2) AND target_type = $3", eggsID, targets.IDs(), targets.Type)
	if err != nil {
		RollbackTransaction(tx)
		return
	}

	total -= cmd.RowsAffected()
	delta -= cmd.RowsAffected()
	err = commitTransaction(tx, "_temp_upsert_likes")
	return
}

func ToggleLike(ctx context.Context, eggsID string, target LikeTarget) (isFollowing bool, err error) {
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}

	rows, err := tx.Query(
		ctx,
		"SELECT 1 FROM user_likes WHERE eggs_id = $1 AND target_id = $2 AND target_type = $3",
		eggsID,
		target.ID,
		target.Type,
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	if rows.Next() {
		rows.Close()
		_, err = tx.Exec(
			ctx,
			"DELETE FROM user_likes WHERE eggs_id = $1 AND target_id = $2 AND target_type = $3",
			eggsID,
			target.ID,
			target.Type,
		)
		if err != nil {
			RollbackTransaction(tx)
			return
		}
		isFollowing = false
		err = commitTransaction(tx)
		return
	}
	rows.Close()
	_, err = tx.Exec(
		ctx,
		"INSERT INTO user_likes (eggs_id, target_id, target_type) VALUES ($1, $2, $3)",
		eggsID,
		target.ID,
		target.Type,
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	isFollowing = true
	err = commitTransaction(tx)
	return
}

func GetLikeCount(ctx context.Context, target string) (n int64, err error) {
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	if target != "track" && target != "playlist" {
		RollbackTransaction(tx)
		err = errors.New("invalid target")
		return
	}
	err = tx.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM user_likes WHERE target_type = $1",
		target,
	).Scan(&n)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = commitTransaction(tx)
	return
}
