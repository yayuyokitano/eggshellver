package wsendpoint

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/yayuyokitano/eggshellver/lib/hub"
	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/queries"
	"github.com/yayuyokitano/eggshellver/lib/router"
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

	targetHub := hub.GetHub(room)
	if targetHub == nil {
		return logging.SE(http.StatusBadRequest, errors.New("room does not exist"))
	}

	users, err := queries.GetUsers(context.Background(), []string{user}, []int{})
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	if len(users) == 0 {
		return logging.SE(http.StatusBadRequest, errors.New("user does not exist"))
	}

	userStub := users[0]
	if userStub.EggsID != user {
		return logging.SE(http.StatusBadRequest, errors.New("user does not exist"))
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}

	client := &hub.Client{Hub: targetHub, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump(userStub)
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

	hub.AttachHub(user)
	return Establish(w, r)
}

func GetHubs(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	output := hub.GetHubs()

	b, err := json.Marshal(output)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}

	w.Write(b)
	return nil
}
