package room

import (
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	rooms map[string]*Room
	musicProvider MusicProvider
	mu sync.RWMutex
}

func NewHub(mp MusicProvider) *Hub {
	return &Hub { rooms: make(map[string]*Room), musicProvider: mp }
}

func (h *Hub) CreateRoom() (*Room, string) {
	roomID := uuid.NewString()
	newRoom := NewRoom(roomID, h.musicProvider)

	h.mu.Lock()
	h.rooms[roomID] = newRoom
	h.mu.Unlock()

	go newRoom.Run()

	return newRoom, roomID
}

func (h *Hub) FindRoom(roomID string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[roomID]
}

func (h *Hub) DeleteRoom(roomID string) {
	h.mu.Lock()
	room := h.rooms[roomID]
	defer h.mu.Unlock()
	delete(h.rooms, roomID)

	if room != nil {
		room.Close()
	}
}