package queries

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type rawFollow struct {
	EggsID1         string    `db:"eggs_id1"`
	DisplayName1    string    `db:"display_name1"`
	IsArtist1       bool      `db:"is_artist1"`
	ImageDataPath1  string    `db:"image_data_path1"`
	PrefectureCode1 int       `db:"prefecture_code1"`
	ProfileText1    string    `db:"profile_text1"`
	EggsID2         string    `db:"eggs_id2"`
	DisplayName2    string    `db:"display_name2"`
	IsArtist2       bool      `db:"is_artist2"`
	ImageDataPath2  string    `db:"image_data_path2"`
	PrefectureCode2 int       `db:"prefecture_code2"`
	ProfileText2    string    `db:"profile_text2"`
	AddedTime       time.Time `db:"added_time"`
}
type rawFollows []rawFollow

type StructuredFollow struct {
	Follower  UserStub  `json:"follower"`
	Followee  UserStub  `json:"followee"`
	Timestamp time.Time `json:"timestamp"`
}
type StructuredFollows struct {
	Follows []StructuredFollow `json:"follows"`
	Total   int64              `json:"total"`
}

func (r rawFollow) ToFollow() StructuredFollow {
	return StructuredFollow{
		Follower: UserStub{
			EggsID:         r.EggsID1,
			DisplayName:    r.DisplayName1,
			IsArtist:       r.IsArtist1,
			ImageDataPath:  r.ImageDataPath1,
			PrefectureCode: r.PrefectureCode1,
			ProfileText:    r.ProfileText1,
		},
		Followee: UserStub{
			EggsID:         r.EggsID2,
			DisplayName:    r.DisplayName2,
			IsArtist:       r.IsArtist2,
			ImageDataPath:  r.ImageDataPath2,
			PrefectureCode: r.PrefectureCode2,
			ProfileText:    r.ProfileText2,
		},
		Timestamp: r.AddedTime,
	}
}

func (arr StructuredFollows) ContainsFollower(b UserStub) bool {
	for _, a := range arr.Follows {
		if a.Follower.EggsID == b.EggsID && a.Follower.DisplayName == b.DisplayName && a.Follower.IsArtist == b.IsArtist && a.Follower.ImageDataPath == b.ImageDataPath && a.Follower.PrefectureCode == b.PrefectureCode && a.Follower.ProfileText == b.ProfileText {
			return true
		}
	}
	return false
}

func (arr StructuredFollows) ContainsFollowee(b UserStub) bool {
	for _, a := range arr.Follows {
		if a.Followee.EggsID == b.EggsID && a.Followee.DisplayName == b.DisplayName && a.Followee.IsArtist == b.IsArtist && a.Followee.ImageDataPath == b.ImageDataPath && a.Followee.PrefectureCode == b.PrefectureCode && a.Followee.ProfileText == b.ProfileText {
			return true
		}
	}
	return false
}

func (arr StructuredFollows) ContainsFollowerID(b string) bool {
	for _, a := range arr.Follows {
		if a.Follower.EggsID == b {
			return true
		}
	}
	return false
}

func (arr StructuredFollows) ContainsFolloweeID(b string) bool {
	for _, a := range arr.Follows {
		if a.Followee.EggsID == b {
			return true
		}
	}
	return false
}

func (arr rawFollows) ToFollows(total int64) (follows StructuredFollows) {
	followSlice := make([]StructuredFollow, 0)
	for _, r := range arr {
		followSlice = append(followSlice, r.ToFollow())
	}
	follows = StructuredFollows{
		Follows: followSlice,
		Total:   total,
	}
	return
}

type Follow struct {
	FollowerID string    `json:"followerID" db:"follower_id"`
	FolloweeID string    `json:"followeeID" db:"followee_id"`
	Timestamp  time.Time `json:"timestamp" db:"added_time"`
}
type Follows []Follow

type PartialFollow struct {
	FollowerID string `json:"followerID" db:"follower_id"`
	FolloweeID string `json:"followeeID" db:"followee_id"`
}

func (arr Follows) Contains(b PartialFollow) bool {
	for _, a := range arr {
		if a.FollowerID == b.FollowerID && a.FolloweeID == b.FolloweeID {
			return true
		}
	}
	return false
}

func GetFollows(ctx context.Context, followerIDs []string, followeeIDs []string, paginator Paginator) (follows StructuredFollows, err error) {
	if len(followerIDs) == 0 && len(followeeIDs) == 0 {
		err = errors.New("no users specified")
		return
	}
	var query string
	var query2 string
	prefix := "SELECT u1.eggs_id AS eggs_id1, u1.display_name AS display_name1, u1.is_artist AS is_artist1, u1.image_data_path AS image_data_path1, u1.prefecture_code AS prefecture_code1, u1.profile_text AS profile_text1, u2.eggs_id AS eggs_id2, u2.display_name AS display_name2, u2.is_artist AS is_artist2, u2.image_data_path AS image_data_path2, u2.prefecture_code AS prefecture_code2, u2.profile_text AS profile_text2, uf.added_time FROM user_follows uf "
	prefix2 := "SELECT COUNT(*) FROM user_follows "
	args := make([]interface{}, 0)
	if len(followerIDs) == 0 {
		query = prefix + "INNER JOIN users u1 ON uf.follower_id = u1.eggs_id INNER JOIN users u2 ON uf.followee_id = u2.eggs_id AND uf.followee_id = ANY($1) ORDER BY added_time DESC LIMIT $2 OFFSET $3"
		query2 = prefix2 + "WHERE followee_id = ANY($1)"
		args = append(args, followeeIDs)
	} else if len(followeeIDs) == 0 {
		query = prefix + "INNER JOIN users u1 ON uf.follower_id = u1.eggs_id AND uf.follower_id = ANY($1) INNER JOIN users u2 ON uf.followee_id = u2.eggs_id ORDER BY added_time DESC LIMIT $2 OFFSET $3"
		query2 = prefix2 + "WHERE follower_id = ANY($1)"
		args = append(args, followerIDs)
	} else {
		query = prefix + "INNER JOIN users u1 ON uf.follower_id = u1.eggs_id AND uf.follower_id = ANY($1) INNER JOIN users u2 ON uf.followee_id = u2.eggs_id AND uf.followee_id = ANY($2) ORDER BY added_time DESC LIMIT $3 OFFSET $4"
		query2 = prefix2 + "WHERE follower_id = ANY($1) AND followee_id = ANY($2)"
		args = append(args, followerIDs, followeeIDs)
	}
	args = append(args, paginator.Limit, paginator.Offset)
	rawFollows := make(rawFollows, 0)

	tx, err := fetchTransaction()
	if err != nil {
		return
	}
	err = pgxscan.Select(
		ctx,
		tx,
		&rawFollows,
		query,
		args...,
	)
	if err != nil {
		return
	}
	var total int64
	err = tx.QueryRow(ctx, query2, args[:len(args)-2]...).Scan(&total)
	err = commitTransaction(tx)
	follows = rawFollows.ToFollows(total)
	return
}

func SubmitFollows(ctx context.Context, followerID string, followeeIDs []string) (n int64, err error) {
	timestamp := time.Now().UnixMilli()
	follows := make([][]interface{}, 0)
	for i, followeeID := range followeeIDs {
		follows = append(follows, []interface{}{followerID, followeeID, time.UnixMilli(timestamp - int64(i))})
	}
	tx, err := fetchTransaction()
	if err != nil {
		return
	}
	n, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"user_follows"},
		[]string{"follower_id", "followee_id", "added_time"},
		pgx.CopyFromRows(follows),
	)
	if err != nil {
		return
	}
	err = commitTransaction(tx)
	return
}

func DeleteFollows(ctx context.Context, followerID string, followeeIDs []string) (err error) {
	tx, err := fetchTransaction()
	if err != nil {
		return
	}
	_, err = tx.Exec(
		ctx,
		"DELETE FROM user_follows WHERE follower_id = $1 AND followee_id = ANY($2)",
		followerID,
		followeeIDs,
	)
	if err != nil {
		return
	}
	err = commitTransaction(tx)
	return
}
