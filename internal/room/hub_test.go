package room

import "testing"

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