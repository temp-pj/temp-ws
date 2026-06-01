package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" { 
		log.Fatal("DATABASE_URL 환경변수 설정 안됨")
	 }

	 db, err := sql.Open("postgres", dsn)

	 if err != nil {
		log.Fatal("DB 연결 실패:", err)
	 }

	 if err := db.Ping(); err != nil {
        log.Fatal("DB Ping 실패:", err)
    }

	defer func() { _ = db.Close() }()

	total := 0

	for {
		var mbids []string

		query := "SELECT mbid FROM recordings WHERE listen_count IS NULL OR listen_count = 0 LIMIT 100"

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var mbid string
			if err := rows.Scan(&mbid); err != nil {
				log.Fatal(err)
			}
			mbids = append(mbids, mbid)
		}

		_ = rows.Close()

		time.Sleep(50 * time.Millisecond)

		if len(mbids) == 0 {
			log.Println("전부 완료!")
			break
		}

		total += len(mbids)
		log.Printf("%d곡 처리 완료", total)

		body := map[string][]string {
			"recording_mbids": mbids,
		}

		jsonBody, err := json.Marshal(body)

		if err != nil { log.Fatal(err) }
		
		resp, err := http.Post(
			"https://api.listenbrainz.org/1/popularity/recording",
			"application/json",
			bytes.NewReader(jsonBody),
		)

		if err != nil {
			log.Fatal(err)
		}

		var results []struct {
			RecordingMBID   string `json:"recording_mbid"`
			TotalListenCount *int  `json:"total_listen_count"`
			TotalUserCount   *int  `json:"total_user_count"`
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
			log.Fatal(err)
		}

		_ = resp.Body.Close()

		for _, r := range results {
			if r.TotalListenCount == nil {
				_, _ = db.Exec("UPDATE recordings SET listen_count = -1 WHERE mbid = $1", r.RecordingMBID)
				continue
			}
		
			_, err := db.Exec(
				"UPDATE recordings SET listen_count = $1 WHERE mbid = $2",
				*r.TotalListenCount,
				r.RecordingMBID,
			)
			if err != nil {
				log.Printf("UPDATE 실패 %s: %v", r.RecordingMBID, err)
			}
		}
	}
}