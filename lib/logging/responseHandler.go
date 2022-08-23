package logging

import (
	"bytes"
	"log"
	"net/http"
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
	log.Println(err)
}

func HandleError(bubbledErr StatusError, r *http.Request, b []byte, t time.Time) {
	log.Println(time.Since(t).String(), bubbledErr.Code, bubbledErr.Err.Error(), r.Method, r.URL.Path, r.URL.Query(), string(b))
}

func LogRequest(r *http.Request, b []byte) {
	log.Println(r.Method, r.URL.Path, r.URL.Query(), string(b))
}

func LogRequestCompletion(w bytes.Buffer, r *http.Request, t time.Time) {
	log.Println(time.Since(t).String(), "complete: ", r.Method, r.URL.Path, r.URL.Query(), w.String())
}
