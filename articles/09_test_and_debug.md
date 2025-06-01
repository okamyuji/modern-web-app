---
title: "第9章 テストとデバッグ"
emoji: "😸" 
type: "tech" 
topics: ["golang","go","alpinejs","htmx"] 
published: false
---

# 第9章 テストとデバッグ

## 1. 包括的なテスト戦略

### ユニットテストの実装

Golang/HTMXアプリケーションでは、バックエンドのロジックとHTMLレスポンスの両方をテストする必要があります。

```go
// models/todo_test.go
package models

import (
    "database/sql"
    "testing"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// テスト用のデータベース設定
func setupTestDB(t *testing.T) (*sql.DB, func()) {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    
    // スキーマの作成
    _, err = db.Exec(`
        CREATE TABLE todos (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            completed BOOLEAN DEFAULT FALSE,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    require.NoError(t, err)
    
    cleanup := func() {
        db.Close()
    }
    
    return db, cleanup
}

func TestTodoRepository_Create(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()
    
    repo := NewTodoRepository(db)
    
    tests := []struct {
        name    string
        todo    Todo
        wantErr bool
    }{
        {
            name: "正常なTODO作成",
            todo: Todo{
                Title:     "テストタスク",
                Completed: false,
            },
            wantErr: false,
        },
        {
            name: "空のタイトル",
            todo: Todo{
                Title: "",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            id, err := repo.Create(&tt.todo)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Greater(t, id, int64(0))
            
            // 作成されたTODOを確認
            created, err := repo.GetByID(id)
            assert.NoError(t, err)
            assert.Equal(t, tt.todo.Title, created.Title)
        })
    }
}

// ベンチマークテスト
func BenchmarkTodoRepository_List(b *testing.B) {
    db, cleanup := setupTestDB(b)
    defer cleanup()
    
    repo := NewTodoRepository(db)
    
    // テストデータの準備
    for i := 0; i < 1000; i++ {
        repo.Create(&Todo{
            Title: fmt.Sprintf("Task %d", i),
        })
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := repo.List("", "")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**💡 テストのポイント:** テーブル駆動テストを使用することで、様々なケースを網羅的にテストできます。また、`testify`パッケージを使用すると、アサーションが読みやすくなります。

### HTTPハンドラーのテスト

```go
// handlers/todo_handler_test.go
package handlers

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "net/url"
    "strings"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestTodoHandler_Create_HTMX(t *testing.T) {
    // モックリポジトリの準備
    mockRepo := &MockTodoRepository{
        CreateFunc: func(todo *Todo) (int64, error) {
            return 1, nil
        },
    }
    
    handler := NewTodoHandler(mockRepo, templates)
    
    tests := []struct {
        name           string
        formData       url.Values
        headers        map[string]string
        expectedStatus int
        checkResponse  func(t *testing.T, body string)
    }{
        {
            name: "HTMXリクエストでの正常な作成",
            formData: url.Values{
                "title":       {"新しいタスク"},
                "description": {"説明文"},
                "priority":    {"high"},
            },
            headers: map[string]string{
                "HX-Request": "true",
                "HX-Target":  "todo-list",
            },
            expectedStatus: http.StatusOK,
            checkResponse: func(t *testing.T, body string) {
                assert.Contains(t, body, "新しいタスク")
                assert.Contains(t, body, "data-todo-id=\"1\"")
                assert.Contains(t, body, "bg-red-100") // high priority
            },
        },
        {
            name: "通常のフォーム送信",
            formData: url.Values{
                "title": {"通常のタスク"},
            },
            headers:        map[string]string{},
            expectedStatus: http.StatusSeeOther, // リダイレクト
        },
        {
            name: "バリデーションエラー",
            formData: url.Values{
                "title": {""},
            },
            headers: map[string]string{
                "HX-Request": "true",
            },
            expectedStatus: http.StatusBadRequest,
            checkResponse: func(t *testing.T, body string) {
                assert.Contains(t, body, "タイトルは必須です")
                assert.Contains(t, body, "bg-red-100") // エラー表示
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // リクエストの作成
            req := httptest.NewRequest(
                http.MethodPost,
                "/todos",
                strings.NewReader(tt.formData.Encode()),
            )
            req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
            
            // ヘッダーの設定
            for k, v := range tt.headers {
                req.Header.Set(k, v)
            }
            
            // レスポンスの記録
            rec := httptest.NewRecorder()
            
            // ハンドラーの実行
            handler.Create(rec, req)
            
            // ステータスコードの確認
            assert.Equal(t, tt.expectedStatus, rec.Code)
            
            // レスポンスボディの確認
            if tt.checkResponse != nil {
                tt.checkResponse(t, rec.Body.String())
            }
            
            // HTMXトリガーの確認
            if tt.headers["HX-Request"] == "true" && tt.expectedStatus == http.StatusOK {
                assert.NotEmpty(t, rec.Header().Get("HX-Trigger"))
            }
        })
    }
}

// E2Eテストのヘルパー
func TestTodoFlow_E2E(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }
    
    // テストサーバーの起動
    server := setupTestServer(t)
    defer server.Close()
    
    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }
    
    // 1. TODOリストページの取得
    resp, err := client.Get(server.URL + "/todos")
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // 2. 新しいTODOの作成
    form := url.Values{
        "title":    {"E2Eテストタスク"},
        "priority": {"medium"},
    }
    
    req, _ := http.NewRequest(
        http.MethodPost,
        server.URL+"/todos",
        strings.NewReader(form.Encode()),
    )
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("HX-Request", "true")
    
    resp, err = client.Do(req)
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // レスポンスの確認
    body := readBody(t, resp)
    assert.Contains(t, body, "E2Eテストタスク")
}
```

**⚠️ テストの注意点:** HTMXのヘッダーを適切に設定することで、実際の動作を正確にテストできます。E2Eテストは時間がかかるため、`testing.Short()`でスキップできるようにしましょう。

## 2. デバッグツールと手法

### デバッグミドルウェアの実装

```go
// middleware/debug.go
package middleware

import (
    "fmt"
    "net/http"
    "runtime"
    "time"
)

// 開発環境用のデバッグパネル
func DebugPanel(isDev bool) func(http.Handler) http.Handler {
    if !isDev {
        return func(next http.Handler) http.Handler {
            return next
        }
    }
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // パフォーマンス計測
            start := time.Now()
            
            // メモリ使用量の記録
            var memStatsBefore runtime.MemStats
            runtime.ReadMemStats(&memStatsBefore)
            
            // カスタムレスポンスライター
            drw := &debugResponseWriter{
                ResponseWriter: w,
                statusCode:     http.StatusOK,
            }
            
            // リクエスト処理
            next.ServeHTTP(drw, r)
            
            // メモリ使用量の計算
            var memStatsAfter runtime.MemStats
            runtime.ReadMemStats(&memStatsAfter)
            
            duration := time.Since(start)
            
            // デバッグ情報をヘッダーに追加
            drw.Header().Set("X-Debug-Duration", duration.String())
            drw.Header().Set("X-Debug-Memory", fmt.Sprintf("%d KB", (memStatsAfter.Alloc-memStatsBefore.Alloc)/1024))
            drw.Header().Set("X-Debug-Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine()))
            
            // HTMLレスポンスの場合、デバッグパネルを挿入
            if strings.Contains(drw.Header().Get("Content-Type"), "text/html") {
                debugHTML := fmt.Sprintf(`
                <div id="debug-panel" style="position: fixed; bottom: 0; right: 0; background: #000; color: #fff; padding: 10px; font-size: 12px; z-index: 9999;">
                    <div>Duration: %s</div>
                    <div>Memory: %d KB</div>
                    <div>Status: %d</div>
                    <div>Goroutines: %d</div>
                    <button onclick="this.parentElement.remove()">×</button>
                </div>
                `, duration, (memStatsAfter.Alloc-memStatsBefore.Alloc)/1024, drw.statusCode, runtime.NumGoroutine())
                
                // レスポンスボディに追加
                body := drw.body.String()
                body = strings.Replace(body, "</body>", debugHTML+"</body>", 1)
                w.Write([]byte(body))
            }
        })
    }
}

// HTMXデバッグヘルパー
func HTMXDebugger() string {
    return `
    <script>
    // HTMXイベントのロギング
    if (window.location.hostname === 'localhost') {
        const htmxEvents = [
            'htmx:configRequest',
            'htmx:beforeRequest',
            'htmx:afterRequest',
            'htmx:responseError',
            'htmx:sendError',
            'htmx:timeout',
            'htmx:afterSettle',
            'htmx:afterSwap'
        ];
        
        htmxEvents.forEach(event => {
            document.body.addEventListener(event, (e) => {
                console.group('HTMX Event: ' + event);
                console.log('Target:', e.detail.target);
                console.log('Detail:', e.detail);
                console.groupEnd();
            });
        });
        
        // Alpine.jsのデバッグ
        if (window.Alpine) {
            window.Alpine.onBeforeComponentInit((component) => {
                console.log('Alpine Component Init:', component.$el, component.$data);
            });
        }
    }
    </script>
    `
}
```

**💡 デバッグのコツ:** 開発環境でのみ有効化されるデバッグツールを用意することで、問題の早期発見が可能になります。本番環境では必ず無効化しましょう。

### エラートレースとロギング

```go
// エラーハンドリングの改善
type AppError struct {
    Code       string
    Message    string
    StatusCode int
    Err        error
    StackTrace string
}

func NewAppError(code, message string, statusCode int, err error) *AppError {
    return &AppError{
        Code:       code,
        Message:    message,
        StatusCode: statusCode,
        Err:        err,
        StackTrace: string(debug.Stack()),
    }
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// エラーハンドリングミドルウェア
func ErrorHandler(logger *Logger, isDev bool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    appErr := &AppError{
                        Code:       "PANIC",
                        Message:    "Internal server error",
                        StatusCode: http.StatusInternalServerError,
                        Err:        fmt.Errorf("%v", err),
                        StackTrace: string(debug.Stack()),
                    }
                    
                    // ログ記録
                    logger.Error("Panic recovered", map[string]interface{}{
                        "error":       appErr.Error(),
                        "stack_trace": appErr.StackTrace,
                        "path":        r.URL.Path,
                        "method":      r.Method,
                    })
                    
                    // エラーレスポンス
                    if r.Header.Get("HX-Request") == "true" {
                        // HTMXエラーレスポンス
                        w.Header().Set("HX-Retarget", "#error-container")
                        w.Header().Set("HX-Reswap", "innerHTML")
                    }
                    
                    w.WriteHeader(appErr.StatusCode)
                    
                    if isDev {
                        // 開発環境では詳細を表示
                        fmt.Fprintf(w, `
                        <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                            <h3 class="font-bold">Error: %s</h3>
                            <p>%s</p>
                            <pre class="mt-2 text-xs">%s</pre>
                        </div>
                        `, appErr.Code, appErr.Message, appErr.StackTrace)
                    } else {
                        // 本番環境では一般的なメッセージ
                        fmt.Fprintf(w, `
                        <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                            <p>エラーが発生しました。しばらく経ってから再度お試しください。</p>
                        </div>
                        `)
                    }
                }
            }()
            
            next.ServeHTTP(w, r)
        })
    }
}
```

**⚠️ エラー処理の重要性:** パニックは必ずリカバーし、適切なエラーレスポンスを返しましょう。スタックトレースは開発環境でのみ表示し、本番環境では機密情報の漏洩を防ぎます。

## 復習問題

1. HTMXアプリケーションのテストにおいて、なぜHTMXヘッダーの設定が重要なのか説明してください。

2. 以下のテストコードの問題点を指摘し、改善してください。

    ```go
    func TestHandler(t *testing.T) {
        handler := NewHandler()
        req, _ := http.NewRequest("GET", "/test", nil)
        rec := httptest.NewRecorder()
        handler.ServeHTTP(rec, req)
        if rec.Code != 200 {
            t.Fail()
        }
    }
    ```

3. 開発環境と本番環境でエラーハンドリングを分ける理由と方法を説明してください。

## 模範解答

1. HTMXヘッダーの重要性
   - HTMXリクエストと通常リクエストで異なるレスポンスを返すため
   - 部分的なHTMLレンダリングのテストが必要
   - `HX-Target`や`HX-Trigger`などの動作確認

2. 改善版

    ```go
    func TestHandler(t *testing.T) {
        handler := NewHandler()
        
        tests := []struct {
            name           string
            method         string
            path           string
            expectedStatus int
            expectedBody   string
        }{
            {
                name:           "正常なGETリクエスト",
                method:         "GET",
                path:           "/test",
                expectedStatus: http.StatusOK,
                expectedBody:   "expected content",
            },
        }
        
        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                req, err := http.NewRequest(tt.method, tt.path, nil)
                require.NoError(t, err)
                
                rec := httptest.NewRecorder()
                handler.ServeHTTP(rec, req)
                
                assert.Equal(t, tt.expectedStatus, rec.Code)
                assert.Contains(t, rec.Body.String(), tt.expectedBody)
            })
        }
    }
    ```

3. 環境別エラーハンドリング
   - 開発環境：詳細なスタックトレース、変数の状態、SQLクエリなどを表示
   - 本番環境：一般的なエラーメッセージのみ、詳細はログに記録
   - 理由：セキュリティ（情報漏洩防止）とユーザビリティの両立
