package room

import (
	"context"
	"sync"
	"temp-ws/internal/client"
	"temp-ws/internal/message"
	"temp-ws/internal/music"
	"testing"
	"time"
)

type mockMusicProvider struct{}

func (m *mockMusicProvider) FetchSongs(ctx context.Context, category music.Category, count int, timeLimit int) ([]music.Song, error) {
    return []music.Song{
        {Title: "테스트곡", Artist: "테스트", ISRC: "TEST001"},
    }, nil
}

func TestTwoClientsInRoom(t *testing.T) {
	hub := NewHub(&mockMusicProvider{})
	room, _ := hub.CreateRoom()

	cli1 := client.Client { ID: "t1", Send: make(chan *message.Message, 3) }
	cli2 := client.Client { ID: "t2", Send: make(chan *message.Message, 3) }

	room.Register(&cli1)
	receiveWithTimeout(t, cli1.Send)
	room.Register(&cli2)
	receiveWithTimeout(t, cli1.Send)
	receiveWithTimeout(t, cli2.Send)

	msg, _ := message.New("Test", nil)
	room.Broadcast(msg)

	if receiveWithTimeout(t, cli1.Send) != msg { t.Error("cli1 브로드캐스트 안됨") }
	if receiveWithTimeout(t, cli2.Send) != msg { t.Error("cli2 브로드캐스트 안됨") }
}

func TestTwoClientsInDifferentRoom(t *testing.T) {
	hub := NewHub(&mockMusicProvider{})
	room1, _ := hub.CreateRoom()
	room2, _ := hub.CreateRoom()

	cli1 := client.Client { ID: "t1", Send: make(chan *message.Message, 3) }
	cli2 := client.Client { ID: "t2", Send: make(chan *message.Message, 3) }

	room1.Register(&cli1)
	receiveWithTimeout(t, cli1.Send)
	room2.Register(&cli2)
	receiveWithTimeout(t, cli2.Send)

	msg, _ := message.New("Test", nil)
	room1.Broadcast(msg)

	if receiveWithTimeout(t, cli1.Send) != msg { t.Error("cli1 브로드캐스트 안됨") }

    select {
		case <-cli2.Send:
			t.Error("cli2가 다른 방 메시지를 받음")
		case <-time.After(50 * time.Millisecond):
    }


}

func TestCreateRoom(t *testing.T) {
	hub := NewHub(&mockMusicProvider{})

	room, _ := hub.CreateRoom()

	if room == nil {
		t.Error("방 생성 안됨")
	}
}

func TestDeleteRoom(t *testing.T) {
	hub := NewHub(&mockMusicProvider{})
	_, roomID := hub.CreateRoom()

	hub.DeleteRoom(roomID)
	room := hub.FindRoom(roomID)

	if room != nil {
		t.Error("방 제거 안됨")
	}

}

func TestFindRoom_NotFound(t *testing.T) {
	hub := NewHub(&mockMusicProvider{})

	room := hub.FindRoom("testroomid")

	if room != nil {
		t.Error("있을리 없는 방이 찾아짐 FindRoom 로직 자체에 문제 있음")
	}
}

func TestConcurrentAccess(t *testing.T) {
	hub := NewHub(&mockMusicProvider{})
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            room, roomID := hub.CreateRoom()
            if room == nil {
                t.Error("동시 생성된 방을 찾지 못함")
            }
            hub.DeleteRoom(roomID)
        }()
    }

    wg.Wait()
}