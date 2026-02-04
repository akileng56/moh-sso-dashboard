package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 🔒 tighten later
	},
}

func NotificationSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		// TODO: extract from session / JWT
		client := &Client{
			UserID:   c.Query("user_id"),
			Role:     c.Query("role"),
			ClientID: c.Query("client_id"),
			Conn:     conn,
		}

		hub.Add(client)
		defer hub.Remove(client)

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}
