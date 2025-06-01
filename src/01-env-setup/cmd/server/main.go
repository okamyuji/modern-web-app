package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"myapp/internal/templates"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 静的ファイルの提供
	fs := http.FileServer(http.Dir("./ui/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// ルートハンドラー
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/greeting", greetingHandler)

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	err := templates.Home().Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func greetingHandler(w http.ResponseWriter, r *http.Request) {
	// 少し遅延を入れてHTMXの動作を確認
	time.Sleep(500 * time.Millisecond)
	fmt.Fprintf(w, "こんにちは！現在の時刻は %s です", time.Now().Format("15:04:05"))
}