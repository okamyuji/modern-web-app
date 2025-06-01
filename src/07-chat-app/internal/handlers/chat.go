package handlers

import (
	"fmt"
	"html"
	"net/http"
	"time"

	"chat-app/internal/models"
	"chat-app/internal/templates"

	"github.com/google/uuid"
)

type ChatHandler struct {
	hub  *models.Hub
	repo *models.MessageRepository
}

func NewChatHandler(hub *models.Hub, repo *models.MessageRepository) *ChatHandler {
	return &ChatHandler{
		hub:  hub,
		repo: repo,
	}
}

// ログインページの表示
func (h *ChatHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := templates.LoginPage().Render(r.Context(), w)
	if err != nil {
		http.Error(w, "テンプレートの実行に失敗しました", http.StatusInternalServerError)
	}
}

// チャットルームへの参加
func (h *ChatHandler) JoinChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// フォームデータの解析
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "フォームデータの解析に失敗しました", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "ユーザー名は必須です", http.StatusBadRequest)
		return
	}

	// ユーザー名の検証
	if len(username) > 20 {
		http.Error(w, "ユーザー名は20文字以内で入力してください", http.StatusBadRequest)
		return
	}

	// 過去のメッセージを取得
	messages, err := h.repo.GetRecentMessages(50)
	if err != nil {
		messages = []models.Message{} // エラーが発生した場合は空のスライス
	}

	// 接続中のユーザー一覧を取得
	connectedUsers := h.hub.GetConnectedClients()

	// チャットルームページを表示
	w.Header().Set("Content-Type", "text/html")
	err = templates.ChatRoom(username, messages, connectedUsers).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "テンプレートの実行に失敗しました", http.StatusInternalServerError)
	}
}

// SSEストリームの実装
func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
	// SSEのヘッダー設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Nginxでのバッファリング無効化

	// クライアントの作成
	clientID := uuid.New().String()
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "ゲスト"
	}

	client := &models.Client{
		ID:       clientID,
		Username: username,
		Send:     make(chan *models.Message, 256),
		Done:     make(chan bool),
	}

	h.hub.Register <- client

	// クリーンアップの設定
	defer func() {
		h.hub.Unregister <- client
	}()

	// コンテキストの監視
	notify := r.Context().Done()

	// 初期接続確認メッセージ
	fmt.Fprintf(w, "event: connected\n")
	fmt.Fprintf(w, "data: connected\n\n")
	w.(http.Flusher).Flush()

	for {
		select {
		case message := <-client.Send:
			// メッセージをHTMLとして送信
			var htmlBuffer = &HTMLBuffer{}
			err := templates.MessageComponent(*message, username).Render(r.Context(), htmlBuffer)
			if err == nil {
				fmt.Fprintf(w, "event: message\n")
				fmt.Fprintf(w, "data: %s\n\n", htmlBuffer.String())
				w.(http.Flusher).Flush()
			}

		case <-notify:
			// クライアントが切断された
			return

		case <-time.After(30 * time.Second):
			// キープアライブの送信
			fmt.Fprintf(w, ": keepalive\n\n")
			w.(http.Flusher).Flush()
		}
	}
}

// メッセージ送信の処理
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// フォームデータの解析
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	// メッセージの長さ制限
	if len(content) > 500 {
		http.Error(w, "Message too long", http.StatusBadRequest)
		return
	}

	// ユーザー情報の取得
	username := r.FormValue("username")
	if username == "" {
		username = "ゲスト"
	}

	message := &models.Message{
		ID:        uuid.New().String(),
		Username:  username,
		Content:   html.EscapeString(content), // XSS対策
		Type:      "text",
		CreatedAt: time.Now(),
	}

	// メッセージをブロードキャスト
	h.hub.Broadcast <- message

	// 成功レスポンス
	w.WriteHeader(http.StatusOK)
}

// 統計情報の取得
func (h *ChatHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	connectedCount := h.hub.GetClientCount()
	messageCount, err := h.repo.GetMessageCount()
	if err != nil {
		messageCount = 0
	}

	stats := map[string]interface{}{
		"connected_users": connectedCount,
		"total_messages":  messageCount,
	}

	w.Header().Set("Content-Type", "application/json")
	// JSON response (simplified)
	fmt.Fprintf(w, `{"connected_users": %d, "total_messages": %d}`, stats["connected_users"], stats["total_messages"])
}

// HTMLBuffer - テンプレートレンダリング用
type HTMLBuffer struct {
	content string
}

func (b *HTMLBuffer) Write(p []byte) (n int, err error) {
	b.content += string(p)
	return len(p), nil
}

func (b *HTMLBuffer) String() string {
	return b.content
}