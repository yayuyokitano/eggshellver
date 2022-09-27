package wsendpoint

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/yayuyokitano/eggshellver/lib/hub"
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
	room := userSplit[len(userSplit)-2]
	if room == "" {
		return logging.SE(http.StatusBadRequest, errors.New("please specify room to join"))
	}

	user, se := router.EggsIDFromToken(r)
	if se != nil {
		return se
	}

	targetHub := services.GetHub(room)
	if targetHub == nil {
		return logging.SE(http.StatusBadRequest, errors.New("room does not exist"))
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}

	client := &hub.Client{Hub: targetHub, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump(user)
	return nil
}

func Create(w http.ResponseWriter, r *http.Request) *logging.StatusError {
	userSplit := strings.Split(r.URL.Path, "/")
	user := userSplit[len(userSplit)-2]
	if user == "" {
		return logging.SE(http.StatusBadRequest, errors.New("please specify room to create"))
	}

	err := router.AuthenticateSpecificUser(r)
	if err != nil {
		return err
	}

	services.AttachHub(user)
	return Establish(w, r)
}
