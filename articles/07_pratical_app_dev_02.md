# 第7章 実践アプリケーション開発(2) - チャットアプリケーション

## 1. リアルタイムアプリケーションの設計

### WebSocketとSSEの選択

チャットアプリケーションでは、リアルタイム通信が必須です。GoとHTMXの組み合わせでは、Server-Sent Events (SSE) が最も自然な選択となります。

```go
// models/message.go
package models

import (
    "sync"
    "time"
)

type Message struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Username  string    `json:"username"`
    Content   string    `json:"content"`
    Type      string    `json:"type"` // text, image, system
    CreatedAt time.Time `json:"created_at"`
}

type Hub struct {
    clients    map[string]*Client
    broadcast  chan *Message
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

type Client struct {
    ID       string
    Username string
    Send     chan *Message
    Done     chan bool
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[string]*Client),
        broadcast:  make(chan *Message),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client.ID] = client
            h.mu.Unlock()
            
            // システムメッセージを送信
            h.broadcast <- &Message{
                Type:      "system",
                Content:   client.Username + " が参加しました",
                CreatedAt: time.Now(),
            }
            
        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client.ID]; ok {
                delete(h.clients, client.ID)
                close(client.Send)
                h.mu.Unlock()
                
                h.broadcast <- &Message{
                    Type:      "system",
                    Content:   client.Username + " が退出しました",
                    CreatedAt: time.Now(),
                }
            } else {
                h.mu.Unlock()
            }
            
        case message := <-h.broadcast:
            h.mu.RLock()
            for _, client := range h.clients {
                select {
                case client.Send <- message:
                default:
                    // クライアントへの送信がブロックされた場合
                    close(client.Send)
                }
            }
            h.mu.RUnlock()
        }
    }
}
```

**💡 設計の要点:** SSEは単方向通信ですが、HTMXと組み合わせることで、双方向のチャット体験を実現できます。WebSocketよりもシンプルで、プロキシやファイアウォールの問題も少ないです。

### ハンドラーの実装

```go
// handlers/chat.go
package handlers

import (
    "encoding/json"
    "fmt"
    "html/template"
    "net/http"
    "time"
    
    "github.com/google/uuid"
)

type ChatHandler struct {
    hub       *models.Hub
    templates *template.Template
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
    
    client := &Client{
        ID:       clientID,
        Username: username,
        Send:     make(chan *Message, 256),
        Done:     make(chan bool),
    }
    
    h.hub.register <- client
    
    // クリーンアップの設定
    defer func() {
        h.hub.unregister <- client
    }()
    
    // コンテキストの監視
    notify := r.Context().Done()
    
    for {
        select {
        case message := <-client.Send:
            // メッセージをHTMLとして送信
            var html string
            err := h.templates.ExecuteTemplate(&html, "message", message)
            if err == nil {
                fmt.Fprintf(w, "event: message\n")
                fmt.Fprintf(w, "data: %s\n\n", html)
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
    
    // ユーザー情報の取得（セッションから）
    username := r.FormValue("username")
    if username == "" {
        username = "ゲスト"
    }
    
    message := &Message{
        ID:        uuid.New().String(),
        Username:  username,
        Content:   template.HTMLEscapeString(content),
        Type:      "text",
        CreatedAt: time.Now(),
    }
    
    // メッセージをブロードキャスト
    h.hub.broadcast <- message
    
    // 成功レスポンス
    w.WriteHeader(http.StatusOK)
}
```

**⚠️ セキュリティの注意点:** ユーザー入力は必ずエスケープしてXSS攻撃を防ぎます。また、SSE接続には適切なタイムアウトとキープアライブを設定しましょう。

## 2. フロントエンドの実装

```html
<!-- templates/chat.html -->
<div x-data="chatApp()" x-init="init()" class="flex flex-col h-screen max-w-4xl mx-auto">
    <!-- ヘッダー -->
    <header class="bg-blue-600 text-white p-4 shadow-lg">
        <div class="flex items-center justify-between">
            <h1 class="text-xl font-bold">チャットルーム</h1>
            <div class="flex items-center gap-4">
                <span class="text-sm">
                    接続中: <span x-text="username"></span>
                </span>
                <div 
                    class="w-3 h-3 rounded-full"
                    :class="connected ? 'bg-green-400' : 'bg-red-400'"
                ></div>
            </div>
        </div>
    </header>
    
    <!-- メッセージエリア -->
    <main class="flex-1 overflow-y-auto bg-gray-50 p-4">
        <div 
            id="messages"
            hx-ext="sse"
            sse-connect="/chat/stream?username={{.Username}}"
            sse-swap="message"
            hx-swap="beforeend"
            class="space-y-2"
        >
            <!-- 初期メッセージ -->
            <div class="text-center text-gray-500 text-sm py-4">
                チャットルームへようこそ！
            </div>
        </div>
    </main>
    
    <!-- 入力エリア -->
    <footer class="bg-white border-t p-4">
        <form 
            hx-post="/chat/send"
            hx-trigger="submit"
            hx-swap="none"
            @htmx:after-request="if($event.detail.successful) this.reset(); $refs.input.focus()"
            class="flex gap-2"
        >
            <input type="hidden" name="username" :value="username">
            
            <input 
                type="text"
                name="content"
                x-ref="input"
                placeholder="メッセージを入力..."
                required
                maxlength="500"
                @keydown.enter.meta="$el.form.requestSubmit()"
                class="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                :disabled="!connected"
            >
            
            <button 
                type="submit"
                :disabled="!connected"
                class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
                送信
            </button>
        </form>
        
        <!-- 文字数カウンター -->
        <div class="text-xs text-gray-500 mt-1 text-right">
            <span x-text="$refs.input?.value?.length || 0"></span> / 500
        </div>
    </footer>
</div>

<script>
function chatApp() {
    return {
        username: '{{.Username}}',
        connected: false,
        reconnectTimer: null,
        
        init() {
            // SSE接続の監視
            document.body.addEventListener('htmx:sseOpen', () => {
                this.connected = true;
                this.showNotification('接続しました', 'success');
            });
            
            document.body.addEventListener('htmx:sseError', () => {
                this.connected = false;
                this.showNotification('接続が切断されました', 'error');
                this.scheduleReconnect();
            });
            
            // メッセージ受信時の処理
            document.body.addEventListener('htmx:sseMessage', (event) => {
                // 自動スクロール
                this.$nextTick(() => {
                    const messages = document.getElementById('messages');
                    messages.scrollTop = messages.scrollHeight;
                });
                
                // 通知音（オプション）
                if (this.notificationsEnabled && !document.hasFocus()) {
                    this.playNotificationSound();
                }
            });
            
            // ページ離脱時の処理
            window.addEventListener('beforeunload', () => {
                // SSE接続のクリーンアップ
            });
        },
        
        scheduleReconnect() {
            if (this.reconnectTimer) return;
            
            this.reconnectTimer = setTimeout(() => {
                location.reload(); // 簡単な再接続方法
            }, 5000);
        },
        
        showNotification(message, type) {
            // Alpine.jsのストアを使用した通知
            Alpine.store('notifications').add(message, type);
        }
    }
}
</script>

<!-- メッセージテンプレート -->
{{define "message"}}
<div class="flex {{if eq .Type "system"}}justify-center{{else if eq .Username $.CurrentUser}}justify-end{{else}}justify-start{{end}}">
    {{if eq .Type "system"}}
        <div class="text-sm text-gray-500 italic">
            {{.Content}}
        </div>
    {{else}}
        <div class="max-w-xs lg:max-w-md">
            <div class="text-xs text-gray-500 mb-1 {{if eq .Username $.CurrentUser}}text-right{{end}}">
                {{.Username}} • {{.CreatedAt.Format "15:04"}}
            </div>
            <div class="
                px-4 py-2 rounded-lg
                {{if eq .Username $.CurrentUser}}
                    bg-blue-600 text-white
                {{else}}
                    bg-white border border-gray-200
                {{end}}
            ">
                {{.Content}}
            </div>
        </div>
    {{end}}
</div>
{{end}}
```

**💡 UXの工夫:** 自動スクロール、接続状態の表示、文字数カウンターなど、小さな工夫がユーザビリティを大きく向上させます。

## 3. パフォーマンスとスケーラビリティ

### メッセージの永続化とページネーション

```go
// メッセージの保存と取得
func (r *MessageRepository) SaveMessage(msg *Message) error {
    query := `
    INSERT INTO messages (id, user_id, username, content, type, created_at)
    VALUES (?, ?, ?, ?, ?, ?)
    `
    _, err := r.db.Exec(query, msg.ID, msg.UserID, msg.Username, msg.Content, msg.Type, msg.CreatedAt)
    return err
}

func (r *MessageRepository) GetRecentMessages(limit int) ([]Message, error) {
    query := `
    SELECT id, user_id, username, content, type, created_at
    FROM messages
    ORDER BY created_at DESC
    LIMIT ?
    `
    rows, err := r.db.Query(query, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var messages []Message
    for rows.Next() {
        var msg Message
        err := rows.Scan(&msg.ID, &msg.UserID, &msg.Username, &msg.Content, &msg.Type, &msg.CreatedAt)
        if err != nil {
            return nil, err
        }
        messages = append(messages, msg)
    }
    
    // 時系列順に並び替え
    for i := len(messages)/2-1; i >= 0; i-- {
        opp := len(messages)-1-i
        messages[i], messages[opp] = messages[opp], messages[i]
    }
    
    return messages, nil
}
```

**⚠️ スケーラビリティの考慮:** メッセージの永続化により、新規参加者に過去のメッセージを表示できます。ただし、大量のメッセージがある場合はページネーションやメッセージの有効期限を検討しましょう。

## 復習問題

1. SSEとWebSocketの違いと、チャットアプリケーションでSSEを選ぶメリットを説明してください。

2. 以下のコードのメモリリークの可能性を指摘し、修正してください。

    ```go
    func (h *Hub) broadcast(message *Message) {
        for _, client := range h.clients {
            client.Send <- message
        }
    }
    ```

3. チャットアプリケーションに「入力中...」インジケーターを実装する方法を説明してください。

## 模範解答

1. SSEとWebSocketの違い
   - SSE：サーバーからクライアントへの単方向通信、HTTPベース、自動再接続
   - WebSocket：双方向通信、独自プロトコル、手動での再接続処理が必要
   - SSEのメリット：HTTPインフラとの互換性、自動再接続、実装の簡単さ

2. 修正版

    ```go
    func (h *Hub) broadcast(message *Message) {
        h.mu.RLock()
        defer h.mu.RUnlock()
        
        for id, client := range h.clients {
            select {
            case client.Send <- message:
                // 送信成功
            default:
                // バッファが満杯の場合、クライアントを削除
                close(client.Send)
                delete(h.clients, id)
            }
        }
    }
    ```

3. 入力中インジケーターの実装
   - デバウンス付きのキー入力イベントでサーバーに通知
   - 一定時間入力がなければ自動的にクリア
   - SSEで他のユーザーに配信し、UIに表示
