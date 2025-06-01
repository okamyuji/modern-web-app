package handlers

import (
	"fmt"
	"htmx-demo/internal/models"
	"net/http"
	"strings"
	"time"
)

// APIHandler handles API requests
type APIHandler struct {
	users []models.User
}

// NewAPIHandler creates a new APIHandler
func NewAPIHandler() *APIHandler {
	// Initialize with some sample users
	users := []models.User{
		{ID: "1", Name: "田中太郎", Email: "tanaka@example.com", CreatedAt: time.Now()},
		{ID: "2", Name: "佐藤花子", Email: "sato@example.com", CreatedAt: time.Now()},
		{ID: "3", Name: "鈴木一郎", Email: "suzuki@example.com", CreatedAt: time.Now()},
	}
	return &APIHandler{users: users}
}

// GetUsers returns all users as HTML
func (h *APIHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	// Simulate loading delay
	time.Sleep(500 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	html := "<div class='space-y-2'>"
	for _, user := range h.users {
		html += fmt.Sprintf(`
			<div class='flex justify-between items-center p-2 bg-white rounded border'>
				<div>
					<strong>%s</strong><br>
					<small class='text-gray-600'>%s</small>
				</div>
				<span class='text-xs text-gray-500'>%s</span>
			</div>
		`, user.Name, user.Email, user.ID)
	}
	html += "</div>"
	
	fmt.Fprint(w, html)
}

// CreateUser creates a new user and returns HTML
func (h *APIHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	time.Sleep(300 * time.Millisecond)
	
	newUser := models.User{
		ID:        fmt.Sprintf("%d", len(h.users)+1),
		Name:      "新規ユーザー",
		Email:     "new@example.com",
		CreatedAt: time.Now(),
	}
	h.users = append(h.users, newUser)
	
	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`
		<div class='p-4 bg-green-100 border border-green-300 rounded'>
			<strong>✅ ユーザーが作成されました</strong><br>
			ID: %s, 名前: %s<br>
			作成日時: %s
		</div>
	`, newUser.ID, newUser.Name, newUser.CreatedAt.Format("2006-01-02 15:04:05"))
	
	fmt.Fprint(w, html)
}

// Echo handles echo requests for PUT/DELETE demos
func (h *APIHandler) Echo(w http.ResponseWriter, r *http.Request) {
	time.Sleep(200 * time.Millisecond)
	
	method := r.Method
	var message string
	
	switch method {
	case "PUT":
		message = "✏️ データが更新されました"
	case "DELETE":
		message = "🗑️ データが削除されました"
	default:
		message = fmt.Sprintf("📨 %s リクエストを受信しました", method)
	}
	
	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`
		<div class='p-4 bg-blue-100 border border-blue-300 rounded'>
			<strong>%s</strong><br>
			時刻: %s<br>
			メソッド: %s
		</div>
	`, message, time.Now().Format("15:04:05"), method)
	
	fmt.Fprint(w, html)
}

// Search handles search requests
func (h *APIHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("search")
	if query == "" {
		query = strings.TrimSpace(r.FormValue("search"))
	}
	
	// Extract from form data if available
	if query == "" {
		r.ParseForm()
		for key, values := range r.Form {
			if key != "search" && len(values) > 0 && values[0] != "" {
				query = values[0]
				break
			}
		}
	}
	
	time.Sleep(200 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	if query == "" {
		fmt.Fprint(w, "<p class='text-gray-500'>検索キーワードを入力してください...</p>")
		return
	}
	
	// Simple search simulation
	results := []string{}
	searchTerms := []string{"Go言語", "HTMX", "Alpine.js", "Tailwind CSS", "プログラミング", "Web開発"}
	
	for _, term := range searchTerms {
		if strings.Contains(strings.ToLower(term), strings.ToLower(query)) {
			results = append(results, term)
		}
	}
	
	html := fmt.Sprintf("<p><strong>検索キーワード:</strong> %s</p>", query)
	if len(results) > 0 {
		html += "<ul class='mt-2 space-y-1'>"
		for _, result := range results {
			html += fmt.Sprintf("<li class='p-2 bg-yellow-100 rounded'>📄 %s</li>", result)
		}
		html += "</ul>"
	} else {
		html += "<p class='mt-2 text-gray-500'>該当する結果が見つかりませんでした。</p>"
	}
	
	fmt.Fprint(w, html)
}

// GetTime returns current time
func (h *APIHandler) GetTime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`
		<div class='p-3 bg-green-100 border border-green-300 rounded'>
			🕐 現在時刻: <strong>%s</strong>
		</div>
	`, time.Now().Format("2006-01-02 15:04:05"))
	
	fmt.Fprint(w, html)
}

// FocusData returns data when input is focused
func (h *APIHandler) FocusData(w http.ResponseWriter, r *http.Request) {
	time.Sleep(100 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	html := `
		<div class='p-3 bg-purple-100 border border-purple-300 rounded'>
			📋 フォーカス時のデータが読み込まれました！<br>
			<small>このデータは入力欄にフォーカスした時に取得されます。</small>
		</div>
	`
	
	fmt.Fprint(w, html)
}

// SpecialAction handles special action (Ctrl+click)
func (h *APIHandler) SpecialAction(w http.ResponseWriter, r *http.Request) {
	time.Sleep(150 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	html := `
		<div class='p-3 bg-orange-100 border border-orange-300 rounded'>
			⚡ 特別なアクションが実行されました！<br>
			<small>Ctrl+クリックでのみ実行される特別な処理です。</small>
		</div>
	`
	
	fmt.Fprint(w, html)
}

// CustomResponse handles custom event response
func (h *APIHandler) CustomResponse(w http.ResponseWriter, r *http.Request) {
	time.Sleep(100 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	html := `
		<div class='p-3 bg-indigo-100 border border-indigo-300 rounded'>
			🎯 カスタムイベントに応答しました！<br>
			<small>JavaScriptから発生したカスタムイベントを受信しました。</small>
		</div>
	`
	
	fmt.Fprint(w, html)
}