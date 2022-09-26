package wsendpoint

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/router"
	"github.com/yayuyokitano/eggshellver/lib/services"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "https://eggs.mu"
	},
}

func Establish(w http.ResponseWriter, r *http.Request) *logging.StatusError {
	userSplit := strings.Split(r.URL.Path, "/")
	user := userSplit[len(userSplit)-1]
	if user == "" {
		return logging.SE(http.StatusBadRequest, errors.New("please specify room to join"))
	}

	hub := services.GetHub(user)
	if hub == nil {
		return logging.SE(http.StatusBadRequest, errors.New("room does not exist"))
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}

	client := &services.Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
	return nil
}

func Create(w http.ResponseWriter, r *http.Request) *logging.StatusError {
	userSplit := strings.Split(r.URL.Path, "/")
	user := userSplit[len(userSplit)-1]
	if user == "" {
		return logging.SE(http.StatusBadRequest, errors.New("please specify room to create"))
	}

	err := router.AuthenticateSpecificUser(r, user)
	if err != nil {
		return err
	}

	services.AttachHub(user)
	return Establish(w, r)
}
