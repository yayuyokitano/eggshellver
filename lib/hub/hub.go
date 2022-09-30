package hub

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yayuyokitano/eggshellver/lib/queries"
)

var hubs map[string]*AuthedHub

func Init() {
	hubs = make(map[string]*AuthedHub)
}

func websocketError(err error) {
	log.Println(err)
}

type AuthedHub struct {
	Hub       *Hub
	Owner     queries.UserStub
	Blocklist map[string]bool
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 65536
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub *AuthedHub

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte
}

type AuthedMessage struct {
	Privileged bool             `json:"privileged"`
	Blocked    bool             `json:"blocked"`
	Sender     queries.UserStub `json:"sender"`
	Message    string           `json:"message"`
}

type RawMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump(user queries.UserStub) {
	defer func() {
		c.Hub.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, rawMessage, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				websocketError(err)
			}
			break
		}

		var message RawMessage
		err = json.Unmarshal(rawMessage, &message)
		if err != nil {
			websocketError(err)
			break
		}
		if message.Type == "start" {
			var songStub RawSongStub
			err = json.Unmarshal([]byte(message.Message), &songStub)
			if err != nil {
				websocketError(err)
				break
			}
			c.Hub.Hub.Song = songStub.ToSongStub()
		}
		if message.Type == "setTitle" {
			var title string
			err = json.Unmarshal([]byte(message.Message), &title)
			if err != nil {
				websocketError(err)
				break
			}
			c.Hub.Hub.Title = title
		}

		reply, err := json.Marshal(AuthedMessage{
			Privileged: c.Hub.Owner.EggsID == user.EggsID,
			Blocked:    user.EggsID == "" || c.Hub.Blocklist[user.EggsID],
			Sender:     user,
			Message:    string(bytes.TrimSpace(bytes.Replace(rawMessage, newline, space, -1))),
		})
		if err != nil {
			websocketError(err)
			break
		}
		c.Hub.Hub.Broadcast <- reply
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	Clients map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	// Owner of hub
	Owner queries.UserStub

	// Currently playing song
	Song SongStub

	// Room title
	Title string
}

type RawSongStub struct {
	MusicTitle    string `json:"musicTitle"`
	ImageDataPath string `json:"imageDataPath"`
	ArtistData    struct {
		DisplayName   string `json:"displayName"`
		ImageDataPath string `json:"imageDataPath"`
	} `json:"artistData"`
}

func (s RawSongStub) ToSongStub() SongStub {
	return SongStub{
		Title:               s.MusicTitle,
		Artist:              s.ArtistData.DisplayName,
		MusicImageDataPath:  s.ImageDataPath,
		ArtistImageDataPath: s.ArtistData.ImageDataPath,
	}
}

type SongStub struct {
	Title               string `json:"title"`
	Artist              string `json:"artist"`
	MusicImageDataPath  string `json:"musicImageDataPath"`
	ArtistImageDataPath string `json:"artistImageDataPath"`
}

func newHub(owner queries.UserStub) *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Owner:      owner,
		Song: SongStub{
			Artist:              "",
			Title:               "",
			MusicImageDataPath:  "",
			ArtistImageDataPath: "",
		},
		Title: owner.EggsID + "のルーム",
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)

				if len(h.Clients) == 0 {
					// cleanup
					close(h.Unregister)
					close(h.Broadcast)
					close(h.Register)
					delete(hubs, h.Owner.EggsID)
					h = nil
					return
				}
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

func AttachHub(userStub queries.UserStub) {
	hubs[userStub.EggsID] = &AuthedHub{
		Hub:       newHub(userStub),
		Owner:     userStub,
		Blocklist: make(map[string]bool),
	}
	go hubs[userStub.EggsID].Hub.run()
}

func GetHub(user string) *AuthedHub {
	return hubs[user]
}

type PublicHub struct {
	Owner     queries.UserStub `json:"owner"`
	Title     string           `json:"title"`
	Song      SongStub         `json:"song"`
	Listeners int              `json:"listeners"`
}

func GetHubs() []PublicHub {
	var publicHubs []PublicHub
	for _, hub := range hubs {
		publicHubs = append(publicHubs, PublicHub{
			Owner:     hub.Owner,
			Title:     hub.Hub.Title,
			Song:      hub.Hub.Song,
			Listeners: len(hub.Hub.Clients),
		})
	}

	return publicHubs
}
