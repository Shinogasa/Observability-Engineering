package main

import (
	"log"
	"net/http"

	"github.com/Shinogasa/Observability-Engineering/ch06/trace-by-hand/authservice"
	"github.com/Shinogasa/Observability-Engineering/ch06/trace-by-hand/frontend"
	"github.com/Shinogasa/Observability-Engineering/ch06/trace-by-hand/nameservice"
)

func main() {
	// 認証サービスを起動（ポート8081）
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/auth", authservice.AuthHandler)
		log.Println("認証サービス起動: :8081")
		log.Fatal(http.ListenAndServe(":8081", mux))
	}()

	// 名前サービスを起動（ポート8082）
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/name", nameservice.NameHandler)
		log.Println("名前サービス起動: :8082")
		log.Fatal(http.ListenAndServe(":8082", mux))
	}()

	// フロントエンドを起動（ポート8080、メインgoroutine）
	mux := http.NewServeMux()
	mux.HandleFunc("/", frontend.RootHandler)
	log.Println("フロントエンド起動: :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
