package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

// websocket bullshit

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 0,
	ReadBufferSize:   0,
	WriteBufferSize:  0,
	WriteBufferPool:  nil,
	Subprotocols:     []string{},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
	},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

type Event struct {
	Event string `json:"event"`
	Data json.RawMessage `json:"data"`
}

var wsBroadcast = make(chan Event)

var connectedClients = []*websocket.Conn{}
 
func socketHandler(w http.ResponseWriter, r *http.Request) {
    // Upgrade our raw HTTP connection to a websocket based one
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
		return
    }

	connectedClients = append(connectedClients, conn)
}

func channelReceiver() {
	for {
		enc := <- wsBroadcast
		data, _ := json.Marshal(enc)
		newChans := []*websocket.Conn{}

		for _, conn := range connectedClients {
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				conn.Close()
				continue
			}
			newChans = append(newChans, conn)
		}

		connectedClients = newChans
	}
}
