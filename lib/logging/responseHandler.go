package logging

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type StatusError struct {
	Code int
	Err  error
}

func SE(code int, err error) *StatusError {
	return &StatusError{
		Code: code,
		Err:  err,
	}
}

func HandleError(bubbledErr StatusError, r *http.Request, b []byte, t time.Time) {
	opsRequestsErrored.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(bubbledErr.Code)).Inc()
	logger.Error().Err(bubbledErr.Err).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("query", r.URL.Query().Encode()).
		Str("body", string(censorKey(b, `"`))).
		Msg("requesterror")
}

func censorKey(b []byte, endChar string) []byte {
	re := regexp.MustCompile(fmt.Sprintf(`(Bearer ).*?(%s)`, endChar))
	return re.ReplaceAll(b, []byte(fmt.Sprintf("$1%s$2", strings.Repeat("*", 10))))
}

func censorJSON(b []byte, keys []string) []byte {
	re := regexp.MustCompile(fmt.Sprintf(`("(?:%s)":").*?("(?:,"|}}))`, strings.Join(keys, "|")))
	return re.ReplaceAll(b, []byte(fmt.Sprintf("$1%s$2", strings.Repeat("*", 10))))
}

func LogRequest(r *http.Request, b []byte) {
	opsRequestsReceived.WithLabelValues(r.Method, r.URL.Path).Inc()
	logger.Debug().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("query", r.URL.Query().Encode()).
		Msg("request")
}

func LogRequestCompletion(w bytes.Buffer, r *http.Request, t time.Time) {
	opsRequestsCompleted.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(t).Seconds())
	logger.Debug().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("query", r.URL.Query().Encode()).
		Msg("requestcomplete")
}

func LogFetch(r *http.Request) {
	opsFetches.WithLabelValues(r.Method, r.URL.Path).Inc()
	h := make([]byte, 0)
	for k, v := range r.Header {
		val := k + ": " + strings.Join(v, ".") + ", "
		h = append(h, []byte(val)...)
	}
	logger.Debug().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("query", r.URL.Query().Encode()).
		Str("headers", string(h)).
		Msg("fetch")
}

func LogFetchErrored(r *http.Request, resp *http.Response) {
	if resp != nil {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error().Err(errors.New("no response body")).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("query", r.URL.Query().Encode()).
				Int("status", resp.StatusCode).
				Msg("fetcherror")
		}
		opsFetchesErrored.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(resp.StatusCode)).Inc()
		logger.Error().Err(errors.New(string(b))).
			Int("status", resp.StatusCode).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("query", r.URL.Query().Encode()).
			Msg("fetcherror")
	} else {
		opsFetchesErrored.WithLabelValues(r.Method, r.URL.Path, "0").Inc()
		logger.Error().Err(errors.New("no response")).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("query", r.URL.Query().Encode()).
			Msg("fetcherror")
	}
}

func LogFetchCompleted(resp *http.Response, b []byte, t time.Time) {
	opsFetchesCompleted.WithLabelValues(resp.Request.Method, resp.Request.URL.Path).Observe(time.Since(t).Seconds())
	logger.Debug().
		Str("method", resp.Request.Method).
		Str("path", resp.Request.URL.Path).
		Str("query", resp.Request.URL.Query().Encode()).
		Msg("fetchcomplete")
}

func WebsocketError(err error) {
	logger.Error().Err(err).Msg("websocketerror")
}

func WebsocketMessage(msgType string, sender string, content string) {
	logger.Debug().Str("type", msgType).Str("sender", sender).Str("content", content).Msg("websocketmessage")
}

func metricError(metricType string, err error) {
	logger.Error().Err(err).Str("type", metricType).Msg("metricerror")
}

func AddCachedUsers(count int) {
	cachedUserCount.Add(float64(count))
}

func AddAuthenticatedUsers(count int) {
	authenticatedUserCount.Add(float64(count))
}

func AddFollows(count int) {
	followCount.Add(float64(count))
}

func AddLikes(count int, targetType string) {
	if targetType == "playlist" {
		playlistLikeCount.Add(float64(count))
	} else if targetType == "track" {
		trackLikeCount.Add(float64(count))
	}
}

func AddPlaylists(count int) {
	playlistCount.Add(float64(count))
}

func AddSongs(count int) {
	songCount.Add(float64(count))
}

func CompleteCache() {
	opPartialCacheSucceeded.Inc()
}

func FailCache(err error) {
	logger.Error().Err(err).Msg("cacheerror")
	opPartialCacheErrored.Inc()
}
