---
title: "パフォーマンス最適化とセキュリティ"
---

# 第8章 パフォーマンス最適化とセキュリティ

ここでは以下のようなアプリケーションを作成します。

![画面1](/images/08-00.png)
![画面2](/images/08-01.png)
![画面3](/images/08-02.png)

## 1. パフォーマンス最適化の実践

### データベースの最適化

Golang/HTMXアプリケーションのパフォーマンスは、多くの場合データベースがボトルネックになります。適切な最適化により、レスポンスタイムを大幅に改善できます。

```go
// db/connection.go
package db

import (
    "database/sql"
    "time"
    _ "github.com/lib/pq"
)

type DBConfig struct {
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}

func NewDB(dsn string, config DBConfig) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    // コネクションプールの設定
    db.SetMaxOpenConns(config.MaxOpenConns)
    db.SetMaxIdleConns(config.MaxIdleConns)
    db.SetConnMaxLifetime(config.ConnMaxLifetime)
    db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
    
    // 接続確認
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return nil, err
    }
    
    return db, nil
}

// クエリの最適化例
type OptimizedTodoRepository struct {
    db *sql.DB
}

// N+1問題を回避する実装
func (r *OptimizedTodoRepository) GetTodosWithTags() ([]TodoWithTags, error) {
    query := `
    WITH todo_tags AS (
        SELECT 
            t.id,
            t.title,
            t.completed,
            t.created_at,
            COALESCE(
                json_agg(
                    json_build_object(
                        'id', tg.id,
                        'name', tg.name,
                        'color', tg.color
                    ) ORDER BY tg.name
                ) FILTER (WHERE tg.id IS NOT NULL),
                '[]'::json
            ) as tags
        FROM todos t
        LEFT JOIN todo_tags_relation ttr ON t.id = ttr.todo_id
        LEFT JOIN tags tg ON ttr.tag_id = tg.id
        GROUP BY t.id, t.title, t.completed, t.created_at
    )
    SELECT id, title, completed, created_at, tags
    FROM todo_tags
    ORDER BY created_at DESC
    LIMIT 100
    `
    
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var todos []TodoWithTags
    for rows.Next() {
        var todo TodoWithTags
        var tagsJSON []byte
        
        err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &tagsJSON)
        if err != nil {
            return nil, err
        }
        
        if err := json.Unmarshal(tagsJSON, &todo.Tags); err != nil {
            return nil, err
        }
        
        todos = append(todos, todo)
    }
    
    return todos, nil
}
```

**💡 最適化のポイント:** N+1問題は最も一般的なパフォーマンス問題です。JOINやサブクエリを使用して、1回のクエリで必要なデータをすべて取得しましょう。

### HTTPレスポンスの最適化

```go
// middleware/compression.go
package middleware

import (
    "compress/gzip"
    "io"
    "net/http"
    "strings"
)

type gzipResponseWriter struct {
    io.Writer
    http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
    return w.Writer.Write(b)
}

func Gzip(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // gzipをサポートしているか確認
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }
        
        // HTMXの部分更新は圧縮しない（小さいため）
        if r.Header.Get("HX-Request") == "true" {
            next.ServeHTTP(w, r)
            return
        }
        
        gz := gzip.NewWriter(w)
        defer gz.Close()
        
        w.Header().Set("Content-Encoding", "gzip")
        w.Header().Del("Content-Length") // 圧縮後のサイズは不明
        
        gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
        next.ServeHTTP(gzw, r)
    })
}

// キャッシュミドルウェア
func Cache(duration time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 静的リソースのみキャッシュ
            if strings.HasPrefix(r.URL.Path, "/static/") {
                w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(duration.Seconds())))
                w.Header().Set("Vary", "Accept-Encoding")
            } else if r.Header.Get("HX-Request") == "true" {
                // HTMXリクエストはキャッシュしない
                w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

**⚠️ キャッシュの注意点:** HTMXの動的コンテンツは絶対にキャッシュしないでください。予期しない動作の原因となります。静的リソースのみに適用しましょう。

## 2. セキュリティの実装

### CSRF対策

```go
// middleware/csrf.go
package middleware

import (
    "crypto/rand"
    "encoding/base64"
    "net/http"
    "sync"
)

type CSRFManager struct {
    tokens sync.Map
}

func NewCSRFManager() *CSRFManager {
    return &CSRFManager{}
}

func (m *CSRFManager) GenerateToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    token := base64.URLEncoding.EncodeToString(b)
    m.tokens.Store(token, true)
    return token
}

func (m *CSRFManager) ValidateToken(token string) bool {
    _, exists := m.tokens.LoadAndDelete(token)
    return exists
}

func (m *CSRFManager) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // GETリクエストはスキップ
        if r.Method == http.MethodGet || r.Method == http.MethodHead {
            next.ServeHTTP(w, r)
            return
        }
        
        // HTMXリクエストの場合、ヘッダーからトークンを取得
        token := r.Header.Get("X-CSRF-Token")
        if token == "" {
            token = r.FormValue("csrf_token")
        }
        
        if !m.ValidateToken(token) {
            http.Error(w, "Invalid CSRF token", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// テンプレートヘルパー
func (m *CSRFManager) TemplateFunc() template.FuncMap {
    return template.FuncMap{
        "csrfToken": m.GenerateToken,
        "csrfField": func() template.HTML {
            token := m.GenerateToken()
            return template.HTML(fmt.Sprintf(`<input type="hidden" name="csrf_token" value="%s">`, token))
        },
    }
}
```

**💡 HTMXでのCSRF対策:**:

- HTMXではメタタグからCSRFトークンを自動的に送信できます。

    ```html
    <meta name="csrf-token" content="{{csrfToken}}">
    <script>
    document.body.addEventListener('htmx:configRequest', (event) => {
        event.detail.headers['X-CSRF-Token'] = document.querySelector('meta[name="csrf-token"]').content;
    });
    </script>
    ```

### XSS対策とコンテンツセキュリティポリシー

```go
// middleware/security.go
package middleware

import (
    "html/template"
    "net/http"
    "regexp"
)

func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 基本的なセキュリティヘッダー
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // CSP（Content Security Policy）
        csp := []string{
            "default-src 'self'",
            "script-src 'self' 'unsafe-inline' https://unpkg.com", // HTMXとAlpine.js用
            "style-src 'self' 'unsafe-inline'", // Tailwind CSS用
            "img-src 'self' data: https:",
            "connect-src 'self'", // HTMXのリクエスト用
            "font-src 'self'",
            "frame-ancestors 'none'",
        }
        w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))
        
        next.ServeHTTP(w, r)
    })
}

// 入力のサニタイゼーション
type Sanitizer struct {
    allowedTags *regexp.Regexp
}

func NewSanitizer() *Sanitizer {
    return &Sanitizer{
        allowedTags: regexp.MustCompile(`<(b|i|u|strong|em|br)(\s[^>]*)?>|</(b|i|u|strong|em)>`),
    }
}

func (s *Sanitizer) Sanitize(input string) string {
    // HTMLエスケープ
    escaped := template.HTMLEscapeString(input)
    
    // 許可されたタグのみ復元
    return s.allowedTags.ReplaceAllStringFunc(escaped, func(match string) string {
        // エスケープを解除
        return strings.ReplaceAll(
            strings.ReplaceAll(match, "&lt;", "<"),
            "&gt;", ">",
        )
    })
}

// SQLインジェクション対策
func (r *Repository) SafeSearch(query string) ([]Result, error) {
    // 必ずプレースホルダを使用
    stmt := `
    SELECT id, title, content 
    FROM items 
    WHERE title ILIKE $1 OR content ILIKE $2
    ORDER BY created_at DESC
    LIMIT 100
    `
    
    searchPattern := "%" + query + "%"
    rows, err := r.db.Query(stmt, searchPattern, searchPattern)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    // 結果の処理...
}
```

**⚠️ セキュリティの重要性:** `'unsafe-inline'`の使用は最小限に留めましょう。可能な限り、外部スクリプトファイルを使用し、nonceベースのCSPを検討してください。

## 3. 監視とロギング

### 構造化ログの実装

```go
// logger/logger.go
package logger

import (
    "context"
    "encoding/json"
    "io"
    "os"
    "time"
)

type Logger struct {
    output io.Writer
    level  LogLevel
}

type LogEntry struct {
    Time       time.Time              `json:"time"`
    Level      string                 `json:"level"`
    Message    string                 `json:"message"`
    Fields     map[string]interface{} `json:"fields,omitempty"`
    TraceID    string                 `json:"trace_id,omitempty"`
    Duration   *float64               `json:"duration_ms,omitempty"`
    Error      string                 `json:"error,omitempty"`
    StackTrace string                 `json:"stack_trace,omitempty"`
}

// リクエストロギングミドルウェア
func RequestLogger(logger *Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // トレースIDの生成
            traceID := r.Header.Get("X-Trace-ID")
            if traceID == "" {
                traceID = generateTraceID()
            }
            
            // レスポンスライターのラップ
            lrw := &loggingResponseWriter{
                ResponseWriter: w,
                statusCode:     http.StatusOK,
            }
            
            // コンテキストにトレースIDを追加
            ctx := context.WithValue(r.Context(), "trace_id", traceID)
            r = r.WithContext(ctx)
            
            // リクエスト処理
            next.ServeHTTP(lrw, r)
            
            // ログ記録
            duration := time.Since(start).Milliseconds()
            
            fields := map[string]interface{}{
                "method":       r.Method,
                "path":         r.URL.Path,
                "status":       lrw.statusCode,
                "ip":           r.RemoteAddr,
                "user_agent":   r.UserAgent(),
                "htmx_request": r.Header.Get("HX-Request") == "true",
                "htmx_target":  r.Header.Get("HX-Target"),
            }
            
            // エラーレスポンスの場合は詳細を記録
            if lrw.statusCode >= 400 {
                logger.Error("Request failed", fields, traceID, &duration)
            } else {
                logger.Info("Request completed", fields, traceID, &duration)
            }
        })
    }
}

// メトリクスの収集
type Metrics struct {
    RequestCount    uint64
    ErrorCount      uint64
    TotalDuration   time.Duration
    ActiveRequests  int32
}

func (m *Metrics) RecordRequest(duration time.Duration, isError bool) {
    atomic.AddUint64(&m.RequestCount, 1)
    atomic.AddInt64((*int64)(&m.TotalDuration), int64(duration))
    
    if isError {
        atomic.AddUint64(&m.ErrorCount, 1)
    }
}

// ヘルスチェックエンドポイント
func HealthCheck(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        health := struct {
            Status   string            `json:"status"`
            Checks   map[string]string `json:"checks"`
            Version  string            `json:"version"`
            Uptime   string            `json:"uptime"`
        }{
            Status: "ok",
            Checks: make(map[string]string),
        }
        
        // データベース接続確認
        ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
        defer cancel()
        
        if err := db.PingContext(ctx); err != nil {
            health.Status = "degraded"
            health.Checks["database"] = "failed: " + err.Error()
        } else {
            health.Checks["database"] = "ok"
        }
        
        // レスポンス
        w.Header().Set("Content-Type", "application/json")
        if health.Status != "ok" {
            w.WriteHeader(http.StatusServiceUnavailable)
        }
        json.NewEncoder(w).Encode(health)
    }
}
```

**💡 監視のベストプラクティス:** 構造化ログを使用することで、ログの検索と分析が容易になります。また、トレースIDにより、リクエストの全体的な流れを追跡できます。

## 復習問題

1. N+1問題とは何か、そしてGolangのアプリケーションでどのように解決するか説明してください。

2. 以下のコードのセキュリティ問題を指摘し、修正してください。

    ```go
    func SearchHandler(w http.ResponseWriter, r *http.Request) {
        query := r.URL.Query().Get("q")
        sql := fmt.Sprintf("SELECT * FROM products WHERE name LIKE '%%%s%%'", query)
        rows, _ := db.Query(sql)
        // ... 結果の処理
    }
    ```

3. HTMXアプリケーションにおいて、なぜ動的コンテンツをキャッシュしてはいけないのか説明してください。

## 模範解答

1. N+1問題の解決
   - N+1問題：1回のクエリで取得したN件のデータに対して、関連データを取得するためにN回の追加クエリが発生する問題
   - 解決方法：JOINを使用した単一クエリ、IN句を使用したバッチ取得、データローダーパターンの実装

2. 修正版

    ```go
    func SearchHandler(w http.ResponseWriter, r *http.Request) {
        query := r.URL.Query().Get("q")
        
        // プレースホルダを使用
        stmt := "SELECT * FROM products WHERE name ILIKE $1"
        searchPattern := "%" + query + "%"
        
        rows, err := db.Query(stmt, searchPattern)
        if err != nil {
            http.Error(w, "Search failed", http.StatusInternalServerError)
            log.Printf("Search error: %v", err)
            return
        }
        defer rows.Close()
        // ... 結果の処理
    }
    ```

3. 動的コンテンツのキャッシュ問題
   - HTMXは部分的なHTML更新を行うため、古いキャッシュが返されると不整合が発生
   - ユーザー固有のデータが他のユーザーに表示される可能性
   - 状態の同期が取れなくなり、予期しない動作の原因となる
