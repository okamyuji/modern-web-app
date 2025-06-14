package main

import (
	"fmt"
	"log"
	"net/http"

	"tailwind-demo/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	// Route handlers
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/components", handlers.ComponentsHandler).Methods("GET")
	r.HandleFunc("/responsive", handlers.ResponsiveHandler).Methods("GET")
	r.HandleFunc("/animations", handlers.AnimationsHandler).Methods("GET")

	fmt.Println("🎨 Tailwind CSS デモサーバーを起動中...")
	fmt.Println("📱 ホーム: http://localhost:8080")
	fmt.Println("🧩 コンポーネント: http://localhost:8080/components")
	fmt.Println("📱 レスポンシブ: http://localhost:8080/responsive")
	fmt.Println("✨ アニメーション: http://localhost:8080/animations")
	fmt.Println("---")
	fmt.Println("💡 ダークモードの切り替えボタンを試してみてください！")
	fmt.Println("📱 ブラウザのサイズを変更してレスポンシブデザインを確認してください！")
	fmt.Println("🛑 終了するには Ctrl+C を押してください")

	log.Fatal(http.ListenAndServe(":8080", r))
}