package logging

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yayuyokitano/eggshellver/lib/queries"
)

var (
	opsRequestsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "eggshellver_requests_received",
		Help: "The total number of received requests, by method and path",
	}, []string{"method", "path"})
	opsRequestsCompleted = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "eggshellver_requests_completed",
		Help: "The total number of completed requests, by method and path, with latency",
	}, []string{"method", "path"})
	opsRequestsErrored = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "eggshellver_requests_errored",
		Help: "The total number of errored requests, by method and path",
	}, []string{"method", "path", "code"})
	opsFetches = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "eggshellver_fetches",
		Help: "The total number of fetches to external APIs, by method and path",
	}, []string{"method", "path"})
	opsFetchesCompleted = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "eggshellver_fetches_completed",
		Help: "The total number of completed fetches to external APIs, by method and path, with latency",
	}, []string{"method", "path"})
	opsFetchesErrored = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "eggshellver_fetches_errored",
		Help: "The total number of errored fetches to external APIs, by method and path",
	}, []string{"method", "path", "code"})
	authenticatedUserCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "eggshellver_authenticated_user_count",
		Help: "The number of authenticated users",
	})
	cachedUserCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "eggshellver_cached_user_count",
		Help: "The number of cached users",
	})
	followCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "eggshellver_follow_count",
		Help: "The number of follows",
	})
	playlistLikeCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "eggshellver_playlist_like_count",
		Help: "The number of playlist likes",
	})
	trackLikeCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "eggshellver_track_like_count",
		Help: "The number of track likes",
	})
	playlistCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "eggshellver_playlist_count",
		Help: "The number of playlists",
	})
	songCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "eggshellver_song_count",
		Help: "The number of songs",
	})
	opPartialCacheSucceeded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "eggshellver_partial_cache_succeeded",
		Help: "The number of partial caches succeeded",
	})
	opPartialCacheErrored = promauto.NewCounter(prometheus.CounterOpts{
		Name: "eggshellver_partial_cache_errored",
		Help: "The number of partial caches that errored",
	})
)

func setUserCounts() {
	authedCount, err := queries.GetAuthenticatedUserCount(context.Background())
	if err != nil {
		metricError("authenticateduser", err)
		return
	}
	authenticatedUserCount.Set(float64(authedCount))

	cachedCount, err := queries.GetCachedUserCount(context.Background())
	if err != nil {
		metricError("cacheduser", err)
		return
	}
	cachedUserCount.Set(float64(cachedCount))

	follow, err := queries.GetFollowCount(context.Background())
	if err != nil {
		metricError("follow", err)
		return
	}
	followCount.Set(float64(follow))

	playlistLike, err := queries.GetLikeCount(context.Background(), "playlist")
	if err != nil {
		metricError("playlistlike", err)
		return
	}
	playlistLikeCount.Set(float64(playlistLike))

	trackLike, err := queries.GetLikeCount(context.Background(), "track")
	if err != nil {
		metricError("tracklike", err)
		return
	}
	trackLikeCount.Set(float64(trackLike))

	playlists, err := queries.GetPlaylistCount(context.Background())
	if err != nil {
		metricError("playlist", err)
		return
	}
	playlistCount.Set(float64(playlists))

	songs, err := queries.GetSongCount(context.Background())
	if err != nil {
		metricError("song", err)
		return
	}
	songCount.Set(float64(songs))
}

func ServeLogs() {
	fmt.Println("serving metrics on port 2112")
	setUserCounts()
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
