package client

import (
	"context"
	"temp-ws/internal/message"

	"github.com/coder/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Send chan *message.Message
}

func (c *Client) Run() { 
	ctx := context.Background()
	for {
		msgType, data, err := c.Conn.Read(ctx)
		if err != nil { return }
		_ = c.Conn.Write(ctx, msgType, data)
	}
}