package logging

import (
	"bytes"
	"fmt"
	"log"
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
	//log.Println(time.Since(t).String(), bubbledErr.Code, bubbledErr.Err.Error(), r.Method, r.URL.Path, r.URL.Query(), string(b))
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
	//log.Println(r.Method, r.URL.Path, r.URL.Query(), string(censorKey(b, `"`)))
}

func LogRequestCompletion(w bytes.Buffer, r *http.Request, t time.Time) {
	opsRequestsCompleted.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(t).Seconds())
	//log.Println(time.Since(t).String(), "complete: ", r.Method, r.URL.Path, r.URL.Query(), w.String())
}

func LogFetch(r *http.Request) {
	opsFetches.WithLabelValues(r.Method, r.URL.Path).Inc()
	/*h := make([]byte, 0)
	for k, v := range r.Header {
		val := k + ": " + strings.Join(v, ".") + ", "
		h = append(h, []byte(val)...)
	}*/
	//log.Println(r.URL.String(), " ", r.Method, " ", string(censorKey(h, ",")))
}

func LogFetchErrored(r *http.Request, resp *http.Response) {
	opsFetchesErrored.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(resp.StatusCode)).Inc()
	log.Print(r.Body, " ", r.Method, " ", r.URL.String())
	//log.Println(r.Method, r.URL.String())
}

func LogFetchCompleted(resp *http.Response, b []byte, t time.Time) {
	opsFetchesCompleted.WithLabelValues(resp.Request.Method, resp.Request.URL.Path).Observe(time.Since(t).Seconds())
	//log.Print(time.Since(t).String(), resp.Request.Body, " ", resp.Request.Method, " ", resp.Request.URL.String(), " ", string(censorJSON(b, []string{"mail", "birthDate", "gender"})))
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
