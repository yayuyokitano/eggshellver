package wsendpoint

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/yayuyokitano/eggshellver/lib/hub"
	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/router"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true //firefox doesn't allow me to do a proper check, as it sends a null origin
	},
}

func Establish(w http.ResponseWriter, r *http.Request) *logging.StatusError {
	userSplit := strings.Split(r.URL.Path, "/")
	room := userSplit[len(userSplit)-2]
	if room == "" {
		return logging.SE(http.StatusBadRequest, errors.New("please specify room to join"))
	}

	userStub, se := router.UserStubFromToken(r)
	if se != nil {
		return se
	}

	targetHub := hub.GetHub(room)
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
	go client.ReadPump(userStub)
	return nil
}

func Create(w http.ResponseWriter, r *http.Request) *logging.StatusError {
	userStub, err := router.AuthenticateSpecificUser(r)
	if err != nil {
		return err
	}

	hub.AttachHub(userStub)
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
