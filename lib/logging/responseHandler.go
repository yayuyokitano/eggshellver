package logging

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
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

func LogError(err error) {
	//log.Println(err)
}

func HandleError(bubbledErr StatusError, r *http.Request, b []byte, t time.Time) {
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
	//log.Println(r.Method, r.URL.Path, r.URL.Query(), string(censorKey(b, `"`)))
}

func LogRequestCompletion(w bytes.Buffer, r *http.Request, t time.Time) {
	//log.Println(time.Since(t).String(), "complete: ", r.Method, r.URL.Path, r.URL.Query(), w.String())
}

func LogFetch(r *http.Request) {
	h := make([]byte, 0)
	for k, v := range r.Header {
		val := k + ": " + strings.Join(v, ".") + ", "
		h = append(h, []byte(val)...)
	}
	//log.Println(r.URL.String(), " ", r.Method, " ", string(censorKey(h, ",")))
}

func LogFetchCompleted(resp *http.Response, b []byte, t time.Time) {
	//log.Print(time.Since(t).String(), " ", resp.Request.Method, " ", resp.Request.URL.String(), " ", string(censorJSON(b, []string{"mail", "birthDate", "gender"})))
}
