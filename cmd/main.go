package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("LOCAL TEST")
	})

	fmt.Println("서버 시작: localhost:8080")
	http.ListenAndServe(":8080", nil)
}