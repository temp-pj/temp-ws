package main

import (
	"fmt"
	"net/http"
	"temp-ws/internal/client"

	"github.com/coder/websocket"
)

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil { return }

		defer conn.Close(websocket.StatusNormalClosure, "")

		cli := client.Client { Conn: conn }
		cli.Run()
	})

	fmt.Println("서버 시작: localhost:8080")
	http.ListenAndServe(":8080", nil)
}