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

func (c *Client) Run(incoming chan <- *message.ClientMessage) { 
	ctx := context.Background()

	go func() {
		for msg := range c.Send {
			data, err := json.Marshal(msg)
			if err != nil { continue }
			if err := c.Conn.Write(ctx, websocket.MessageText, data); err != nil {
				return
			}
		}
	}()

	for {
		_, data, err := c.Conn.Read(ctx)
        if err != nil { return }

		var msg message.Message
		if err := json.Unmarshal(data, &msg); err != nil { continue }

		incoming <- &message.ClientMessage { From: c.ID, Message: &msg }
	}
}