package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"todo-app/internal/handlers"
	"todo-app/internal/models"

	"github.com/gorilla/mux"
)

func main() {
	// データベースディレクトリの作成
	if err := os.MkdirAll("db", 0755); err != nil {
		log.Fatal("データベースディレクトリの作成に失敗しました:", err)
	}

	// データベース接続
	repo, err := models.NewTodoRepository("db/todo.db")
	if err != nil {
		log.Fatal("データベース接続に失敗しました:", err)
	}
	defer repo.Close()

	// ハンドラーの初期化
	todoHandler := handlers.NewTodoHandler(repo)

	// ルーターの設定
	r := mux.NewRouter()

	// ルート定義
	r.HandleFunc("/", todoHandler.Index).Methods("GET")
	r.HandleFunc("/todos", todoHandler.Create).Methods("POST")
	r.HandleFunc("/todos/{id:[0-9]+}", todoHandler.Update).Methods("PUT")
	r.HandleFunc("/todos/{id:[0-9]+}", todoHandler.Delete).Methods("DELETE")
	r.HandleFunc("/todos/{id:[0-9]+}", todoHandler.Show).Methods("GET")
	r.HandleFunc("/todos/{id:[0-9]+}/edit", todoHandler.Edit).Methods("GET")
	r.HandleFunc("/todos/{id:[0-9]+}/toggle", todoHandler.ToggleCompleted).Methods("PATCH")

	// 静的ファイルの配信
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	fmt.Println("📝 TODO アプリケーションを起動中...")
	fmt.Println("🌐 URL: http://localhost:8081")
	fmt.Println("🗄️  データベース: db/todo.db")
	fmt.Println("---")
	fmt.Println("💡 機能:")
	fmt.Println("  • タスクの作成・編集・削除")
	fmt.Println("  • 完了状態の切り替え")
	fmt.Println("  • 優先度と期限の設定")
	fmt.Println("  • フィルタリング（全て・未完了・完了・期限切れ）")
	fmt.Println("  • リアルタイム検索")
	fmt.Println("  • ダークモード対応")
	fmt.Println("  • レスポンシブデザイン")
	fmt.Println("🛑 終了するには Ctrl+C を押してください")

	log.Fatal(http.ListenAndServe(":8081", r))
}