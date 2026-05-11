package room

import (
	"temp-ws/internal/message"
	"testing"
	"time"
)

func receiveWithTimeout(t *testing.T, ch <-chan *message.Message) *message.Message {
    t.Helper()
    select {
    case msg := <-ch:
        return msg
    case <-time.After(time.Second):
        t.Fatal("메시지 수신 타임아웃")
        return nil
    }
}