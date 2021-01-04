package web

import (
	"encoding/json"
	"fmt"

	"github.com/momper14/msw/wrapper"
	wrappermodel "github.com/momper14/msw/wrapper/model"
	"github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	clients    map[*Client]bool
	msw        chan *wrappermodel.Message
	register   chan *Client
	unregister chan *Client
	command    chan *wrappermodel.Command
}

// NewHub initialises a new Hub
func NewHub() *Hub {
	return &Hub{
		msw:        make(chan *wrappermodel.Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run runs the Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.msw:
			json, _ := json.Marshal(message)
			for client := range h.clients {
				select {
				case client.send <- json:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// SendCommand sends a command to the MSW
func (h *Hub) SendCommand(c []byte) {
	var cs = new(wrappermodel.Command)

	fmt.Printf("%s\n", c)
	err := json.Unmarshal(c, cs)
	if err != nil {
		logrus.Warn(err)
		return
	}

	h.command <- cs
}

// Subscribe subscribes to the MSW
func (h *Hub) Subscribe(w *wrapper.Wrapper) {
	h.command = w.Subscribe(h.msw)
}
