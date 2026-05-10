package room

import (
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	rooms map[string]*Room
	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub { rooms: make(map[string]*Room) }
}

func (h *Hub) CreateRoom() string {
	roomID := uuid.NewString()
	newRoom := NewRoom(roomID)

	h.mu.Lock()
	h.rooms[roomID] = newRoom
	h.mu.Unlock()

	go newRoom.Run()

	return roomID
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