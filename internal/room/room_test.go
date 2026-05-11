package room

import (
	"encoding/json"
	"temp-ws/internal/client"
	"temp-ws/internal/message"
	"testing"

	"github.com/google/uuid"
)

func TestRegister(t *testing.T) {
	id := uuid.NewString()
	room := NewRoom(id)
	go room.Run()
	cli := &client.Client{Send: make(chan *message.Message, 3)}

	room.register <- cli

	msg := receiveWithTimeout(t, cli.Send)
	if msg == nil {
		t.Error("방 참여 메시지 못 받음")
	}
}

func TestUnRegister(t *testing.T) {
	id := uuid.NewString()
	room := NewRoom(id)
	go room.Run()
	cli := &client.Client{Send: make(chan *message.Message, 3)}

	room.register <- cli
	receiveWithTimeout(t, cli.Send)

	room.unregister <- cli
	_, ok := <-cli.Send
	if ok {
		t.Error("Send 채널이 닫히지 않음")
	}
}

func TestBroadcast(t *testing.T) {
	id := uuid.NewString()
	room := NewRoom(id)
	go room.Run()

	firstClient := &client.Client{Send: make(chan *message.Message, 3)}
	secondClient := &client.Client{Send: make(chan *message.Message, 3)}
	thirdClient := &client.Client{Send: make(chan *message.Message, 3)}

	room.register <- firstClient
	receiveWithTimeout(t, firstClient.Send)
	room.register <- secondClient
	receiveWithTimeout(t, firstClient.Send)
	receiveWithTimeout(t, secondClient.Send)
	room.register <- thirdClient
	receiveWithTimeout(t, firstClient.Send)
	receiveWithTimeout(t, secondClient.Send)
	receiveWithTimeout(t, thirdClient.Send)

	msg := message.Message{Type: "TEST", Payload: json.RawMessage([]byte(`{"title":"test"}`)), Timestamp: 0}

	room.broadcast <- &msg

	if receiveWithTimeout(t, firstClient.Send) != &msg { t.Error("브로드캐스트 안됨") }
	if receiveWithTimeout(t, secondClient.Send) != &msg { t.Error("브로드캐스트 안됨") }
	if receiveWithTimeout(t, thirdClient.Send) != &msg { t.Error("브로드캐스트 안됨") }
}