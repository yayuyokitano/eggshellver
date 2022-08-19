package router

import (
	"fmt"
	"net/http"
)

type Methods struct {
	GET    func(http.ResponseWriter, *http.Request)
	POST   func(http.ResponseWriter, *http.Request)
	DELETE func(http.ResponseWriter, *http.Request)
}

func Handle(endpoint string, m Methods) {
	http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		handleCors(w)

		switch r.Method {
		case "GET":
			m.GET(w, r)
		case "POST":
			m.POST(w, r)
		case "DELETE":
			m.DELETE(w, r)
		case "OPTIONS": // CORS preflight request
			w.WriteHeader(http.StatusOK)
		default:
			ReturnMethodNotAllowed(w, r)
		}
	})
}

func ReturnMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("Method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
}

func handleCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "https://eggs.mu")
}
