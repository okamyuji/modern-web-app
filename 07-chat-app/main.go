package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"chat-app/internal/handlers"
	"chat-app/internal/models"

	"github.com/gorilla/mux"
)

func main() {
	// データベースディレクトリの作成
	if err := os.MkdirAll("db", 0755); err != nil {
		log.Fatal("データベースディレクトリの作成に失敗しました:", err)
	}

	// データベース接続
	repo, err := models.NewMessageRepository("db/chat.db")
	if err != nil {
		log.Fatal("データベース接続に失敗しました:", err)
	}
	defer repo.Close()

	// Hubの作成と起動
	hub := models.NewHub(repo)
	go hub.Run()

	// ハンドラーの初期化
	chatHandler := handlers.NewChatHandler(hub, repo)

	// ルーターの設定
	r := mux.NewRouter()

	// ルート定義
	r.HandleFunc("/", chatHandler.LoginPage).Methods("GET")
	r.HandleFunc("/chat", chatHandler.JoinChat).Methods("POST")
	r.HandleFunc("/chat/stream", chatHandler.Stream).Methods("GET")
	r.HandleFunc("/chat/send", chatHandler.SendMessage).Methods("POST")
	r.HandleFunc("/chat/stats", chatHandler.GetStats).Methods("GET")

	// 静的ファイルの配信
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	fmt.Println("💬 チャットアプリケーションを起動中...")
	fmt.Println("🌐 URL: http://localhost:8082")
	fmt.Println("🗄️  データベース: db/chat.db")
	fmt.Println("---")
	fmt.Println("💡 機能:")
	fmt.Println("  • リアルタイムメッセージング（Server-Sent Events）")
	fmt.Println("  • ユーザー参加・退出通知")
	fmt.Println("  • オンラインユーザー一覧")
	fmt.Println("  • メッセージ履歴の保存と表示")
	fmt.Println("  • 自動再接続機能")
	fmt.Println("  • ダークモード対応")
	fmt.Println("  • レスポンシブデザイン")
	fmt.Println("  • XSS攻撃対策")
	fmt.Println("🛑 終了するには Ctrl+C を押してください")

	log.Fatal(http.ListenAndServe(":8082", r))
}