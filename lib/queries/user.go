package queries

import (
	"context"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type User struct {
	Data struct {
		EggsID         string `json:"userName"`
		DisplayName    string `json:"displayName"`
		ArtistID       int    `json:"artistId"`
		UserID         int    `json:"userId"`
		ImageDataPath  string `json:"imageDataPath"`
		PrefectureCode int    `json:"prefectureCode"`
		ProfileText    string `json:"profile"`
	} `json:"data"`
}

type UserStub struct {
	EggsID         string `json:"userName" db:"eggs_id"`
	DisplayName    string `json:"displayName" db:"display_name"`
	IsArtist       bool   `json:"isArtist" db:"is_artist"`
	ImageDataPath  string `json:"imageDataPath" db:"image_data_path"`
	PrefectureCode int    `json:"prefectureCode" db:"prefecture_code"`
	ProfileText    string `json:"profile" db:"profile_text"`
}
type UserStubs []UserStub

func (arr UserStubs) Contains(b UserStub) bool {
	for _, a := range arr {
		if a.EggsID != b.EggsID {
			continue
		}
		if a.DisplayName != b.DisplayName {
			continue
		}
		if a.IsArtist != b.IsArtist {
			continue
		}
		if a.ImageDataPath != b.ImageDataPath {
			continue
		}
		if a.PrefectureCode != b.PrefectureCode {
			continue
		}
		if a.ProfileText != b.ProfileText {
			continue
		}
		return true
	}
	return false
}

func (arr UserStubs) ContainsID(b string) bool {
	for _, a := range arr {
		if a.EggsID == b {
			return true
		}
	}
	return false
}

func (arr UserStubs) StringSlice() (output []string) {
	output = make([]string, 0)
	for _, user := range arr {
		output = append(output, user.EggsID)
	}
	return
}

func (u User) IsArtist() bool {
	return u.Data.ArtistID != 0
}

func PostUserStubs(ctx context.Context, users []UserStub) (n int64, err error) {
	userStubs := make([][]interface{}, 0)
	for _, user := range users {
		userStubs = append(userStubs, []interface{}{
			user.EggsID,
			user.DisplayName,
			user.IsArtist,
			user.ImageDataPath,
			user.PrefectureCode,
			user.ProfileText,
		})
	}
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(ctx, "DROP TABLE IF EXISTS _temp_upsert_users")
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(ctx, "CREATE TEMPORARY TABLE _temp_upsert_users (LIKE users INCLUDING ALL) ON COMMIT DROP")
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"_temp_upsert_users"},
		[]string{"eggs_id", "display_name", "is_artist", "image_data_path", "prefecture_code", "profile_text"},
		pgx.CopyFromRows(userStubs),
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	cmd, err := tx.Exec(ctx, "INSERT INTO users SELECT * FROM _temp_upsert_users ON CONFLICT (eggs_id) DO UPDATE SET display_name = EXCLUDED.display_name, is_artist = EXCLUDED.is_artist, image_data_path = EXCLUDED.image_data_path, prefecture_code = EXCLUDED.prefecture_code, profile_text = EXCLUDED.profile_text")
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	n = cmd.RowsAffected()
	err = commitTransaction(tx, "_temp_upsert_users")
	return
}

func InsertUser(ctx context.Context, user User, token string) (err error) {
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(
		ctx,
		"INSERT INTO users (eggs_id, display_name, is_artist, image_data_path, prefecture_code, profile_text, token) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		user.Data.EggsID,
		user.Data.DisplayName,
		user.IsArtist(),
		user.Data.ImageDataPath,
		user.Data.PrefectureCode,
		user.Data.ProfileText,
		token,
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = commitTransaction(tx)
	return
}

func UpdateUserToken(ctx context.Context, user User, token string) (err error) {
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(
		ctx,
		"UPDATE users SET token = $1 WHERE eggs_id = $2",
		token,
		user.Data.EggsID,
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = commitTransaction(tx)
	return
}

func UpdateUserDetails(ctx context.Context, user User) (err error) {
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(
		ctx,
		"UPDATE users SET display_name = $1, is_artist = $2, image_data_path = $3, prefecture_code = $4, profile_text = $5 WHERE eggs_id = $6",
		user.Data.DisplayName,
		user.IsArtist(),
		user.Data.ImageDataPath,
		user.Data.PrefectureCode,
		user.Data.ProfileText,
		user.Data.EggsID,
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = commitTransaction(tx)
	return
}

func GetUsers(ctx context.Context, users []string) (output []UserStub, err error) {
	output = make([]UserStub, 0)
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = pgxscan.Select(
		ctx,
		tx,
		&output,
		"SELECT eggs_id, display_name, is_artist, image_data_path, prefecture_code, profile_text FROM users WHERE eggs_id = ANY($1)",
		users,
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = commitTransaction(tx)
	return
}

func GetEggsIDByToken(ctx context.Context, token string) (eggsID string, err error) {
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = tx.QueryRow(
		ctx,
		"SELECT eggs_id FROM users WHERE token = $1",
		token,
	).Scan(&eggsID)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = commitTransaction(tx)
	return
}

func GetUserCredentials(ctx context.Context, user User) (eggsID string, token string, err error) {
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = tx.QueryRow(
		ctx,
		"SELECT eggs_id, token FROM users WHERE eggs_id = $1",
		user.Data.EggsID,
	).Scan(&eggsID, &token)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = commitTransaction(tx)
	return
}

func UNSAFEDeleteUser(ctx context.Context, eggsID string) (err error) {
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	_, err = tx.Exec(
		ctx,
		"DELETE FROM users WHERE eggs_id = $1",
		eggsID,
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}
	err = commitTransaction(tx)
	return
}
