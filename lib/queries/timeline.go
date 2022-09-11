package queries

import (
	"context"
	"time"

	"github.com/georgysavva/scany/pgxscan"
)

type TimelineItem struct {
	ID        string    `json:"id" db:"id"`
	Type      string    `json:"type" db:"type"`
	Target    string    `json:"target" db:"target"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

func GetTimeline(ctx context.Context, eggsID string, offset int, limit int) (timeline []TimelineItem, err error) {
	tx, err := fetchTransaction()
	if err != nil {
		RollbackTransaction(tx)
		return
	}

	query := `
		WITH followed_users AS (
			SELECT followee_id AS id FROM user_follows WHERE follower_id = $1
		)
		SELECT eggs_id AS id, 'music' AS type, music_id AS target, release_date AS timestamp FROM songs WHERE eggs_id = ANY(SELECT id FROM followed_users)
		UNION ALL
		SELECT eggs_id AS id, 'musiclike' AS type, target_id AS target, added_time AS timestamp FROM user_likes WHERE target_type = 'track' AND eggs_id = ANY(SELECT id FROM followed_users)
		UNION ALL
		SELECT eggs_id AS id, 'playlist' AS type, playlist_id AS target, last_modified AS timestamp FROM playlists WHERE eggs_id = ANY(SELECT id FROM followed_users)
		UNION ALL
		SELECT eggs_id AS id, 'playlistlike' AS type, target_id AS target, added_time AS timestamp FROM user_likes WHERE target_type = 'playlist' AND eggs_id = ANY(SELECT id FROM followed_users)
		UNION ALL
		SELECT follower_id AS id, 'follow' AS type, followee_id AS target, added_time AS timestamp FROM user_follows WHERE follower_id = ANY(SELECT id FROM followed_users)
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	err = pgxscan.Select(
		ctx,
		tx,
		&timeline,
		query,
		eggsID,
		limit,
		offset,
	)
	if err != nil {
		RollbackTransaction(tx)
		return
	}

	err = commitTransaction(tx)
	return
}
