package client

import (
	"context"
	"encoding/json"
	"temp-ws/internal/message"

	"github.com/coder/websocket"
)

type Client struct {
	ID string
	Conn *websocket.Conn
	Send chan *message.Message
}

func (c *Client) Run() { 
	ctx := context.Background()

	go func() {
		for msg := range c.Send {
			data, _ := json.Marshal(msg)
			c.Conn.Write(ctx, websocket.MessageText, data)
		}
	}()

	for {
		_, _, err := c.Conn.Read(ctx)
        if err != nil { return }
	}
}