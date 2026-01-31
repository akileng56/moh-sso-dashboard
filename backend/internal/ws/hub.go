package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID   string
	Role     string
	ClientID string
	Conn     *websocket.Conn
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
	}
}

func (h *Hub) Add(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *Hub) Remove(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
}

func (h *Hub) Broadcast(match func(*Client) bool, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		if match(c) {
			_ = c.Conn.WriteMessage(websocket.TextMessage, payload)
		}
	}
}
