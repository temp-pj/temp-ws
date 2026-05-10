package main

import (
	"fmt"
	"net/http"
	"temp-ws/internal/client"
	"temp-ws/internal/message"
	"temp-ws/internal/room"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

func main() {
	hub := room.NewHub()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil { return }

		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()

		roomID := r.URL.Query().Get("room")
		var currentRoom *room.Room

		if roomID == "" {
			currentRoom, _ = hub.CreateRoom()
		} else {
			currentRoom = hub.FindRoom(roomID)
			if currentRoom == nil {
				_ = conn.Close(websocket.StatusPolicyViolation, "room not found")
				return
			}
		}

		cli := client.Client { ID: uuid.NewString(), Conn: conn, Send: make(chan *message.Message, 1) }
		currentRoom.Register(&cli)
		cli.Run()
	})

	fmt.Println("서버 시작: localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("서버 에러:", err)
	}
}