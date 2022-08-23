package router

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yayuyokitano/eggshellver/lib/logging"
)

type HTTPImplementer = func(io.Writer, *http.Request, []byte) *logging.StatusError

type Methods struct {
	GET    HTTPImplementer
	POST   HTTPImplementer
	PUT    HTTPImplementer
	DELETE HTTPImplementer
}

func Handle(endpoint string, m Methods) {
	http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		var method HTTPImplementer

		switch r.Method {
		case "GET":
			method = m.GET
		case "POST":
			method = m.POST
		case "PUT":
			method = m.PUT
		case "DELETE":
			method = m.DELETE
		case "OPTIONS": // CORS preflight request
			method = HandleCORSPreflight
		default:
			method = ReturnMethodNotAllowed
		}

		HandleMethod(method, w, r)
	})
}

func HandleMethod(m HTTPImplementer, w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		logging.HandleError(*logging.SE(http.StatusInternalServerError, err), r, b, t)
		return
	}
	logging.LogRequest(r, b)
	handleCors(w)

	var log bytes.Buffer
	mw := io.MultiWriter(w, &log)
	se := m(mw, r, b)
	if se != nil {
		logging.HandleError(*se, r, b, t)
		http.Error(w, se.Err.Error(), se.Code)
		return
	}
	logging.LogRequestCompletion(log, r, t)
}

func HandleCORSPreflight(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	return nil
}

func ReturnMethodNotAllowed(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	return logging.SE(http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
}

func handleCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "https://eggs.mu")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization")
}
