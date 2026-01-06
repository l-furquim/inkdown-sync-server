package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	UserID   string
	DeviceID string
	Conn     *websocket.Conn
	Manager  *Manager
	Send     chan []byte
}

func NewClient(id, userID, deviceID string, conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		ID:       id,
		UserID:   userID,
		DeviceID: deviceID,
		Conn:     conn,
		Manager:  manager,
		Send:     make(chan []byte, 256),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Manager.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(c.Manager.pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.Manager.pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		c.Manager.HandleMessage <- &ClientMessage{
			Client:  c,
			Message: message,
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(c.Manager.pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Manager.writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Manager.writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
