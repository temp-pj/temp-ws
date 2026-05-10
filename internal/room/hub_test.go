package room

import (
	"sync"
	"testing"
)

func TestCreateRoom(t *testing.T) {
	hub := NewHub()
	roomID := hub.CreateRoom()

	room := hub.FindRoom(roomID)

	if room == nil {
		t.Error("방 생성 안됨")
	}
}

func TestDeleteRoom(t *testing.T) {
	hub := NewHub()

	roomID := hub.CreateRoom()

	hub.DeleteRoom(roomID)

	room := hub.FindRoom(roomID)

	if room != nil {
		t.Error("방 제거 안됨")
	}

}

func TestFindRoom_NotFound(t *testing.T) {
	hub := NewHub()

	room := hub.FindRoom("testroomid")

	if room != nil {
		t.Error("있을리 없는 방이 찾아짐 FindRoom 로직 자체에 문제 있음")
	}
}

func TestConcurrentAccess(t *testing.T) {
	hub := NewHub()
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            roomID := hub.CreateRoom()
            room := hub.FindRoom(roomID)
            if room == nil {
                t.Error("동시 생성된 방을 찾지 못함")
            }
            hub.DeleteRoom(roomID)
        }()
    }

    wg.Wait()
}