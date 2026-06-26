package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"temp-ws/internal/client"
	"temp-ws/internal/message"
	"temp-ws/internal/music"
	"temp-ws/internal/room"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	db := mustConnectDB()
    defer func() { _ = db.Close() }()

    store := music.NewStore(db)
    hub := room.NewHub(store)

    mux := http.NewServeMux()
    mux.HandleFunc("/ws", handleWebSocket(hub))

    log.Println("서버 시작: localhost:8080")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatal("서버 에러:", err)
    }
}

func mustConnectDB() *sql.DB {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatal("DATABASE_URL 환경변수가 설정되지 않음")
    }

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("DB 연결 실패:", err)
    }

    if err := db.Ping(); err != nil {
        log.Fatal("DB Ping 실패:", err)
    }

    log.Println("DB 연결 성공!")
    return db
}

func handleWebSocket(hub *room.Hub) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
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

        playerID := uuid.NewString()
        cli := client.Client{ID: playerID, Conn: conn, Send: make(chan *message.Message, 16)}

        currentRoom.Register(&cli)
        defer currentRoom.Unregister(&cli)

        cli.Run(currentRoom.IncomingChannel())
    }
}