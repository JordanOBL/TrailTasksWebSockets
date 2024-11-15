package server

import (
	"fmt"
	"sync"

	ws "github.com/gorilla/websocket"
)

type Client struct {
	mux             sync.RWMutex      `json:"-"`
	Conn            *ws.Conn          `json:"-"`
	Id              string            `json:"id"`
	IsHost          bool              `json:"isHost"`
	Username        string            `json:"username"`
	Distance        float64           `json:"distance"`
	IsReady         bool              `json:"isReady"`
	IsPaused        bool              `json:"isPaused"`
	Strikes         uint8             `json:"strikes"`
	MsgCh           chan ServerPacket `json:"-"`
	droppedMessages uint8             `json:"-"`
	TokensEarned    uint8             `json:"tokensEarned"`
	BonusTokens     uint8             `json:"bonusTokens"`
	RoomId          string            `json:"roomId"`
}

type ClientPacket struct {
	Header  Header                 `json:"header"`
	Message map[string]interface{} `json:"message"`
	Hiker   *Client
}

func (c *Client) writePump() {

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
