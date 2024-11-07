package server

import (
	"fmt"

	ws "github.com/gorilla/websocket"
)

type Client struct {
	Conn            *ws.Conn          `json:"-"`
	Id              string            `json:"id"`
	Username        string            `json:"username"`
	Distance        float64           `json:"distance"`
	IsReady         bool              `json:"isReady"`
	MsgCh           chan ServerPacket `json:"-"`
	droppedMessages int               `json:"-"`
}

type ClientPacket struct {
	Header  Header                 `json:"header"`
	Message map[string]interface{} `json:"message"`
	Hiker   *Client
}

func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
		close(c.MsgCh)
	}()

	for {
		select {
		case msg, ok := <-c.MsgCh:
			if !ok {
				fmt.Printf("Message channel closed for %v\n", c.Username)
				return
			}
			if err := c.Conn.WriteJSON(msg); err != nil {
				fmt.Printf("Error in writePump for %v: %v\n", c.Username, err)
				return
			}
		}
	}
}
