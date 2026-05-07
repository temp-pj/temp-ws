package client

import (
	"context"

	"github.com/coder/websocket"
)

type Client struct {
	Conn *websocket.Conn
}

func (c *Client) Run() { 
	ctx := context.Background()
	for {
		msgType, data, err := c.Conn.Read(ctx)
		if err != nil { return }
		_ = c.Conn.Write(ctx, msgType, data)
	}
}