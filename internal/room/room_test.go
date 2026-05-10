package room

import (
	"encoding/json"
	"temp-ws/internal/client"
	"temp-ws/internal/message"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRegister(t *testing.T) {
	uuid := uuid.NewString()
	room := NewRoom(uuid)
	go room.Run()
	client := &client.Client { Send: make(chan *message.Message, 3) }

	room.register <- client

	select {
		case <-client.Send:
			
		case <-time.After(50 * time.Millisecond):
			t.Error("방 참여 메시지 못 받음")
	}
}

func TestUnRegister(t *testing.T) {
	uuid := uuid.NewString()
	room := NewRoom(uuid)
	go room.Run()
	client := &client.Client { Send: make(chan *message.Message, 3) }

	room.register <- client
	<- client.Send

	room.unregister <- client

	select {
		case <-client.Send:
			t.Error("삭제된 클라이언트가 메시지를 받음")
		case <-time.After(50 * time.Millisecond):
			
	}
}

func TestBroadcast(t *testing.T) {
	uuid := uuid.NewString()
	room := NewRoom(uuid)
	go room.Run()

	firstClient := &client.Client { Send: make(chan *message.Message, 3) }
	secondClient := &client.Client { Send: make(chan *message.Message, 3) }
	thirdClient := &client.Client { Send: make(chan *message.Message, 3) }

	room.register <- firstClient
	<-firstClient.Send
	room.register <- secondClient
	<-firstClient.Send
	<-secondClient.Send
	room.register <- thirdClient
	<-firstClient.Send
	<-secondClient.Send
	<-thirdClient.Send

	msg := message.Message { Type: "TEST", Payload: json.RawMessage([]byte(`{ "title": "test" }`)), Timestamp: 0 }

	room.broadcast <- &msg

	

	received := <- firstClient.Send
	if received != &msg { t.Error("브로드캐스트 안됨") }
	
	received = <- secondClient.Send
	if received != &msg { t.Error("브로드캐스트 안됨") }

	received = <- thirdClient.Send
	if received != &msg { t.Error("브로드캐스트 안됨") }
}