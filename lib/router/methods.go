package router

import (
	"fmt"
	"net/http"
)

type Methods struct {
	GET  func(http.ResponseWriter, *http.Request)
	POST func(http.ResponseWriter, *http.Request)
}

func Handle(endpoint string, m Methods) {
	http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			m.GET(w, r)
		case "POST":
			m.POST(w, r)
		default:
			ReturnMethodNotAllowed(w, r)
		}
	})
}

func ReturnMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("Method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
}
