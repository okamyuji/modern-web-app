# 第7章 実践アプリケーション開発② - チャットアプリケーション

## 概要

このディレクトリには、Server-Sent Events (SSE) を活用したリアルタイムチャットアプリケーションが含まれています。Golang、HTMX、Alpine.js、Tailwind CSSを組み合わせて、モダンなリアルタイム通信アプリケーションを実現します。

## 主要機能

### 💬 リアルタイムメッセージング

- **Server-Sent Events**: WebSocketより軽量でHTTPインフラと互換性が高い
- **自動再接続**: 接続が切断された場合の自動復旧機能
- **リアルタイム配信**: 即座にメッセージが全ユーザーに配信

### 👥 ユーザー管理

- **参加・退出通知**: ユーザーの入退室をリアルタイムで通知
- **オンラインユーザー一覧**: 現在接続中のユーザーをサイドバーに表示
- **ユーザー名検証**: 適切な文字制限とパターンチェック

### 🗄️ メッセージ永続化

- **履歴保存**: SQLiteによるメッセージの永続化
- **初期表示**: 新規参加者への過去メッセージ表示（直近50件）
- **統計情報**: 接続ユーザー数・総メッセージ数の表示

### 🎨 UI/UX機能

- **ダークモード**: ワンクリックでテーマ切替
- **レスポンシブデザイン**: モバイル・デスクトップ対応
- **自動スクロール**: 新着メッセージ時の自動スクロール
- **文字数制限**: 500文字制限とリアルタイムカウンター
- **通知システム**: 接続状態やエラーの視覚的フィードバック

### 🔒 セキュリティ機能

- **XSS対策**: HTMLエスケープによる安全なコンテンツ表示
- **入力検証**: フォームデータの適切なバリデーション
- **メッセージ制限**: 長すぎるメッセージの防止

## 技術スタック

### バックエンド

- **Go**: 1.24.3
- **Gorilla Mux**: 1.8.1 (HTTP ルーティング)
- **SQLite**: 軽量データベース
- **UUID**: メッセージ・クライアント識別
- **Templ**: 0.3.887 (型安全なテンプレートエンジン)

### フロントエンド

- **HTMX**: 1.9.5 + SSE Extension (Server-Sent Events)
- **Alpine.js**: 3.x (リアクティブUI)
- **Tailwind CSS**: 3.x (モダンスタイリング)

## アーキテクチャ

### プロジェクト構造

```text
07-chat-app/
├── main.go                     # アプリケーションエントリーポイント
├── go.mod                      # Go モジュール定義
├── internal/
│   ├── models/
│   │   └── message.go          # メッセージモデル・Hub・リポジトリ
│   ├── handlers/
│   │   └── chat.go             # チャットハンドラー
│   └── templates/
│       └── base.templ          # テンプレート定義
├── db/
│   └── chat.db                 # SQLite データベース
└── README.md                   # このファイル
```

### リアルタイム通信アーキテクチャ

#### Hub パターン

```go
type Hub struct {
    clients    map[string]*Client
    Broadcast  chan *Message
    Register   chan *Client
    Unregister chan *Client
    mu         sync.RWMutex
    repo       *MessageRepository
}
```

#### SSE ストリーム

```text
[クライアント] ←─── SSE Stream ←─── [Hub] ←─── [メッセージ送信]
     ↓                                ↑
[HTMX POST] ──→ [Handler] ──→ [Broadcast] ──→ [全クライアント配信]
```

## データモデル

### Message 構造体

```go
type Message struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Username  string    `json:"username"`
    Content   string    `json:"content"`
    Type      string    `json:"type"` // text, image, system
    CreatedAt time.Time `json:"created_at"`
}
```

### データベーススキーマ

```sql
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    username TEXT NOT NULL,
    content TEXT NOT NULL,
    type TEXT DEFAULT 'text',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 実行方法

### 1. 依存関係のインストール

```bash
cd 07-chat-app
go mod tidy
```

### 2. テンプレート生成

```bash
templ generate
```

### 3. アプリケーション起動

```bash
go run main.go
```

### 4. ブラウザでアクセス

http://localhost:8082

## API エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/` | ログインページ |
| POST | `/chat` | チャットルーム参加 |
| GET | `/chat/stream` | SSEストリーム接続 |
| POST | `/chat/send` | メッセージ送信 |
| GET | `/chat/stats` | 統計情報取得 |

## Server-Sent Events 実装

### 1. SSE 接続設定

```html
<div 
    hx-ext="sse"
    sse-connect="/chat/stream?username=username"
    sse-swap="message"
    hx-swap="beforeend"
>
```

### 2. サーバー側SSE実装

```go
func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
    // SSEヘッダー設定
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    // クライアント登録
    client := &models.Client{
        ID:       uuid.New().String(),
        Username: username,
        Send:     make(chan *models.Message, 256),
    }
    
    h.hub.Register <- client
    defer func() { h.hub.Unregister <- client }()
    
    // メッセージ配信ループ
    for {
        select {
        case message := <-client.Send:
            fmt.Fprintf(w, "event: message\n")
            fmt.Fprintf(w, "data: %s\n\n", htmlTemplate)
            w.(http.Flusher).Flush()
        case <-time.After(30 * time.Second):
            fmt.Fprintf(w, ": keepalive\n\n")
            w.(http.Flusher).Flush()
        }
    }
}
```

### 3. メッセージブロードキャスト

```go
func (h *Hub) Run() {
    for {
        select {
        case message := <-h.Broadcast:
            h.repo.SaveMessage(message)
            h.broadcastMessage(message)
        }
    }
}
```

## HTMX + Alpine.js 活用パターン

### 1. リアルタイム接続状態管理

```javascript
function chatApp() {
    return {
        connected: false,
        
        init() {
            document.body.addEventListener('htmx:sseOpen', () => {
                this.connected = true;
                this.showNotification('接続しました', 'success');
            });
            
            document.body.addEventListener('htmx:sseError', () => {
                this.connected = false;
                this.scheduleReconnect();
            });
        }
    }
}
```

### 2. 自動スクロール機能

```javascript
document.body.addEventListener('htmx:sseMessage', (event) => {
    this.$nextTick(() => {
        const messages = document.getElementById('messages');
        messages.scrollTop = messages.scrollHeight;
    });
});
```

### 3. フォーム送信と自動リセット

```html
<form 
    hx-post="/chat/send"
    hx-on:htmx:after-request="if(event.detail.successful) { this.reset(); $refs.input.focus(); }"
>
```

## パフォーマンス最適化

### 1. 効率的なメッセージ配信

```go
func (h *Hub) broadcastMessage(message *Message) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    for id, client := range h.clients {
        select {
        case client.Send <- message:
            // 送信成功
        default:
            // バッファ満杯時のクライアント切断
            close(client.Send)
            delete(h.clients, id)
        }
    }
}
```

### 2. バッファリング戦略

- **クライアントバッファ**: 256メッセージの送信キュー
- **ノンブロッキング送信**: デッドロック防止
- **自動クライアント除去**: 応答しないクライアントの自動削除

### 3. データベース最適化

```sql
-- 効率的なインデックス設定
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_type ON messages(type);
```

## セキュリティ対策

### 1. XSS対策

```go
// HTMLエスケープ
message.Content = html.EscapeString(content)
```

### 2. 入力検証

```go
// メッセージ長制限
if len(content) > 500 {
    http.Error(w, "Message too long", http.StatusBadRequest)
    return
}

// ユーザー名パターン検証
pattern="[a-zA-Z0-9_\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FAF]+"
```

### 3. リソース保護

```go
// 適切なタイムアウト設定
case <-time.After(30 * time.Second):
    // キープアライブ送信

// ゴルーチンリークの防止
defer func() {
    h.hub.Unregister <- client
}()
```

## スケーラビリティ考慮

### 1. 接続数制限

```go
const MaxClients = 1000

func (h *Hub) GetClientCount() int {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return len(h.clients)
}
```

### 2. メッセージ履歴管理

```go
// 古いメッセージの自動削除
func (r *MessageRepository) DeleteOldMessages(days int) error {
    query := `DELETE FROM messages WHERE created_at < datetime('now', '-' || ? || ' days')`
    _, err := r.db.Exec(query, days)
    return err
}
```

### 3. 水平スケーリング対応

- Redis Pub/Sub による複数インスタンス間通信
- ロードバランサー配下での Sticky Session
- 分散データベースへの移行計画

## 運用・監視

### 1. ヘルスチェック

```go
func (h *ChatHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    stats := map[string]interface{}{
        "status":          "healthy",
        "connected_users": h.hub.GetClientCount(),
        "uptime":          time.Since(startTime),
    }
    json.NewEncoder(w).Encode(stats)
}
```

### 2. ログ出力

```go
// 構造化ログ
log.Printf("Client connected: id=%s, username=%s", client.ID, client.Username)
log.Printf("Message broadcast: id=%s, users=%d", message.ID, len(h.clients))
```

### 3. メトリクス収集

- 接続ユーザー数の推移
- メッセージ送信レート
- 接続持続時間
- エラー発生率

## テスト戦略

### 1. ユニットテスト

```go
func TestHubBroadcast(t *testing.T) {
    hub := models.NewHub(repo)
    // ブロードキャスト機能のテスト
}
```

### 2. 統合テスト

```go
func TestSSEConnection(t *testing.T) {
    // SSE接続とメッセージ受信のテスト
}
```

### 3. 負荷テスト

```go
// 大量クライアント接続テスト
// メッセージ送信レートテスト
// 長時間接続安定性テスト
```

## 拡張アイデア

### 1. 高度な機能

- **チャットルーム**: 複数部屋の作成と管理
- **プライベートメッセージ**: 1対1メッセージング
- **ファイル共有**: 画像・ファイルのアップロード
- **メッセージ検索**: 過去メッセージの全文検索

### 2. UI/UX改善

- **入力中表示**: タイピングインジケーター
- **メッセージ既読**: 既読状態の管理
- **絵文字リアクション**: メッセージへのリアクション
- **通知音**: カスタマイズ可能な通知音

### 3. 管理機能

- **管理者パネル**: ユーザー管理とモデレーション
- **メッセージ削除**: 不適切コンテンツの削除
- **ユーザーブロック**: 問題ユーザーの排除
- **統計ダッシュボード**: 利用状況の可視化

## トラブルシューティング

### よくある問題

1. **SSE接続が切断される**

   ```bash
   # キープアライブ間隔の調整
   # プロキシ設定の確認
   # ファイアウォール設定の確認
   ```

2. **メッセージが重複する**

   ```bash
   # UUID重複の確認
   # ブラウザキャッシュのクリア
   # データベース制約の確認
   ```

3. **パフォーマンス劣化**

   ```bash
   # 接続クライアント数の監視
   # メッセージ送信頻度の制限
   # データベースインデックスの最適化
   ```

### デバッグ方法

1. **ログレベル設定**

   ```go
   log.SetLevel(log.DebugLevel)
   ```

2. **接続状態確認**

   ```javascript
   // ブラウザ開発者ツールのネットワークタブ
   // EventSource接続の監視
   ```

3. **データベース確認**

   ```sql
   -- メッセージ数確認
   SELECT COUNT(*) FROM messages;
   
   -- 最新メッセージ確認
   SELECT * FROM messages ORDER BY created_at DESC LIMIT 10;
   ```

---

このチャットアプリケーションは、リアルタイム通信の基礎技術を学ぶ実践的な教材として設計されています。
Server-Sent Events による効率的な通信と、モダンなフロントエンド技術の組み合わせにより、
スケーラブルで保守性の高いリアルタイムアプリケーションの構築方法を習得できます。
