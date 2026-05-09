package room

import (
	"encoding/json"
	"temp-ws/internal/client"
	"temp-ws/internal/message"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	room := NewRoom()
	go room.Run()
	client := &client.Client { Send: make(chan *message.Message) }

	room.register <- client

	msg := message.Message { Type: "TEST", Payload: json.RawMessage([]byte(`{ "title": "test" }`)), Timestamp: 0 }
	room.broadcast <- &msg

	received := <- client.Send

	if received != &msg {
		t.Error("클라이언트 추가 안됨")
	}
}

func TestUnRegister(t *testing.T) {
	room := NewRoom()
	go room.Run()
	client := &client.Client { Send: make(chan *message.Message) }

	room.register <- client

	room.unregister <- client
	
	msg := message.Message { Type: "TEST", Payload: json.RawMessage([]byte(`{ "title": "test" }`)), Timestamp: 0 }
	room.broadcast <- &msg

	select {
		case <-client.Send:
			t.Error("삭제된 클라이언트가 메시지를 받음")
		case <-time.After(50 * time.Millisecond):
			
	}
}

func TestBroadcast(t *testing.T) {
	room := NewRoom()
	go room.Run()

	firstClient := &client.Client { Send: make(chan *message.Message) }
	secondClient := &client.Client { Send: make(chan *message.Message) }
	thirdClient := &client.Client { Send: make(chan *message.Message) }

	room.register <- firstClient
	room.register <- secondClient
	room.register <- thirdClient

	msg := message.Message { Type: "TEST", Payload: json.RawMessage([]byte(`{ "title": "test" }`)), Timestamp: 0 }

	room.broadcast <- &msg


	received := <- firstClient.Send
	if received != &msg { t.Error("브로드캐스트 안됨") }
	
	received = <- secondClient.Send
	if received != &msg { t.Error("브로드캐스트 안됨") }

	received = <- thirdClient.Send
	if received != &msg { t.Error("브로드캐스트 안됨") }
}