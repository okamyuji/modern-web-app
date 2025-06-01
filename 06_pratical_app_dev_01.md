# 第6章 実践アプリケーション開発(1) - TODOリスト

## 1. アプリケーション設計の基礎

### アーキテクチャの全体像

TODOリストアプリケーションを通じて、Golang/HTMX/Alpine.js/Tailwind CSSの実践的な統合方法を学びます。このシンプルなアプリケーションには、実際のプロジェクトで必要となる要素がすべて含まれています。

```text
// プロジェクト構造
todo-app/
├── main.go
├── handlers/
│   └── todo.go
├── models/
│   └── todo.go
├── templates/
│   ├── layout.html
│   ├── index.html
│   └── components/
│       └── todo-item.html
├── static/
│   └── css/
│       └── app.css
└── db/
    └── sqlite.db
```

**💡 設計の要点:** 小規模なアプリケーションでも、将来の拡張を見据えた構造にすることが重要です。各レイヤーの責務を明確に分離しましょう。

### データモデルの設計

```go
// models/todo.go
package models

import (
    "database/sql"
    "time"
)

type Todo struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    Priority    string    `json:"priority"` // low, medium, high
    DueDate     *time.Time `json:"due_date"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type TodoRepository struct {
    db *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
    return &TodoRepository{db: db}
}

// データベース初期化
func (r *TodoRepository) InitSchema() error {
    query := `
    CREATE TABLE IF NOT EXISTS todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        description TEXT,
        completed BOOLEAN DEFAULT FALSE,
        priority TEXT DEFAULT 'medium',
        due_date DATETIME,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE TRIGGER IF NOT EXISTS update_todos_updated_at 
    AFTER UPDATE ON todos
    BEGIN
        UPDATE todos SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
    `
    _, err := r.db.Exec(query)
    return err
}

// 検索とフィルタリング
func (r *TodoRepository) List(filter string, search string) ([]Todo, error) {
    query := `
    SELECT id, title, description, completed, priority, due_date, created_at, updated_at 
    FROM todos 
    WHERE 1=1
    `
    args := []interface{}{}
    
    // フィルター条件の追加
    switch filter {
    case "active":
        query += " AND completed = FALSE"
    case "completed":
        query += " AND completed = TRUE"
    case "overdue":
        query += " AND due_date < datetime('now') AND completed = FALSE"
    }
    
    // 検索条件の追加
    if search != "" {
        query += " AND (title LIKE ? OR description LIKE ?)"
        searchPattern := "%" + search + "%"
        args = append(args, searchPattern, searchPattern)
    }
    
    query += " ORDER BY completed ASC, priority DESC, created_at DESC"
    
    rows, err := r.db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var todos []Todo
    for rows.Next() {
        var todo Todo
        err := rows.Scan(
            &todo.ID, &todo.Title, &todo.Description, 
            &todo.Completed, &todo.Priority, &todo.DueDate,
            &todo.CreatedAt, &todo.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        todos = append(todos, todo)
    }
    
    return todos, nil
}
```

**⚠️ データベース設計の注意点:** SQLiteは開発には最適ですが、本番環境では並行性の制限があります。初期段階ではSQLiteで開発し、必要に応じてPostgreSQLなどに移行する計画を立てましょう。

## 2. ハンドラーの実装

```go
// handlers/todo.go
package handlers

import (
    "html/template"
    "net/http"
    "strconv"
    "time"
    
    "todo-app/models"
)

type TodoHandler struct {
    repo      *models.TodoRepository
    templates *template.Template
}

func NewTodoHandler(repo *models.TodoRepository, templates *template.Template) *TodoHandler {
    return &TodoHandler{
        repo:      repo,
        templates: templates,
    }
}

// メインページの表示
func (h *TodoHandler) Index(w http.ResponseWriter, r *http.Request) {
    filter := r.URL.Query().Get("filter")
    search := r.URL.Query().Get("search")
    
    todos, err := h.repo.List(filter, search)
    if err != nil {
        http.Error(w, "データの取得に失敗しました", http.StatusInternalServerError)
        return
    }
    
    // 統計情報の計算
    stats := struct {
        Total     int
        Active    int
        Completed int
        Overdue   int
    }{
        Total: len(todos),
    }
    
    for _, todo := range todos {
        if todo.Completed {
            stats.Completed++
        } else {
            stats.Active++
            if todo.DueDate != nil && todo.DueDate.Before(time.Now()) {
                stats.Overdue++
            }
        }
    }
    
    data := map[string]interface{}{
        "Todos":  todos,
        "Filter": filter,
        "Search": search,
        "Stats":  stats,
    }
    
    // HTMXリクエストの場合は部分的なHTMLを返す
    if r.Header.Get("HX-Request") == "true" {
        target := r.Header.Get("HX-Target")
        if target == "todo-list" {
            err = h.templates.ExecuteTemplate(w, "todo-list", data)
        } else {
            err = h.templates.ExecuteTemplate(w, "index-content", data)
        }
    } else {
        err = h.templates.ExecuteTemplate(w, "layout", data)
    }
    
    if err != nil {
        http.Error(w, "テンプレートの実行に失敗しました", http.StatusInternalServerError)
    }
}

// TODO作成
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
        return
    }
    
    // フォームデータの解析
    err := r.ParseForm()
    if err != nil {
        h.sendError(w, "フォームデータの解析に失敗しました", http.StatusBadRequest)
        return
    }
    
    // バリデーション
    title := r.FormValue("title")
    if title == "" {
        h.sendError(w, "タイトルは必須です", http.StatusBadRequest)
        return
    }
    
    todo := models.Todo{
        Title:       title,
        Description: r.FormValue("description"),
        Priority:    r.FormValue("priority"),
    }
    
    // 期限の解析
    if dueDateStr := r.FormValue("due_date"); dueDateStr != "" {
        dueDate, err := time.Parse("2006-01-02", dueDateStr)
        if err == nil {
            todo.DueDate = &dueDate
        }
    }
    
    // データベースに保存
    id, err := h.repo.Create(&todo)
    if err != nil {
        h.sendError(w, "TODOの作成に失敗しました", http.StatusInternalServerError)
        return
    }
    
    todo.ID = id
    todo.CreatedAt = time.Now()
    todo.UpdatedAt = time.Now()
    
    // 新しいTODOアイテムのHTMLを返す
    w.Header().Set("HX-Trigger", "todoAdded")
    err = h.templates.ExecuteTemplate(w, "todo-item", todo)
    if err != nil {
        http.Error(w, "テンプレートの実行に失敗しました", http.StatusInternalServerError)
    }
}

// エラーレスポンスの送信
func (h *TodoHandler) sendError(w http.ResponseWriter, message string, status int) {
    w.Header().Set("HX-Retarget", "#error-message")
    w.Header().Set("HX-Reswap", "innerHTML")
    w.WriteHeader(status)
    w.Write([]byte(`<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" x-data x-init="setTimeout(() => $el.remove(), 5000)">` + message + `</div>`))
}
```

**💡 HTMXの活用ポイント:** `HX-Request`ヘッダーをチェックすることで、同じエンドポイントで通常のページレンダリングとHTMXの部分更新の両方に対応できます。

## 3. テンプレートの実装

```html
<!-- templates/index.html -->
<div x-data="todoApp()" class="container mx-auto px-4 py-8 max-w-4xl">
    <h1 class="text-3xl font-bold mb-8">TODOリスト</h1>
    
    <!-- エラーメッセージ表示エリア -->
    <div id="error-message" class="mb-4"></div>
    
    <!-- 統計情報 -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div class="bg-white p-4 rounded-lg shadow">
            <div class="text-2xl font-bold">{{.Stats.Total}}</div>
            <div class="text-gray-600 text-sm">全タスク</div>
        </div>
        <div class="bg-blue-50 p-4 rounded-lg shadow">
            <div class="text-2xl font-bold text-blue-600">{{.Stats.Active}}</div>
            <div class="text-gray-600 text-sm">未完了</div>
        </div>
        <div class="bg-green-50 p-4 rounded-lg shadow">
            <div class="text-2xl font-bold text-green-600">{{.Stats.Completed}}</div>
            <div class="text-gray-600 text-sm">完了</div>
        </div>
        <div class="bg-red-50 p-4 rounded-lg shadow">
            <div class="text-2xl font-bold text-red-600">{{.Stats.Overdue}}</div>
            <div class="text-gray-600 text-sm">期限切れ</div>
        </div>
    </div>
    
    <!-- 新規TODO作成フォーム -->
    <form 
        hx-post="/todos"
        hx-target="#todo-list"
        hx-swap="afterbegin"
        @htmx:after-request="if($event.detail.successful) this.reset()"
        class="bg-white p-6 rounded-lg shadow mb-6"
    >
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                    タイトル <span class="text-red-500">*</span>
                </label>
                <input 
                    type="text" 
                    name="title" 
                    required
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    placeholder="タスクのタイトル"
                >
            </div>
            
            <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                    優先度
                </label>
                <select 
                    name="priority"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                >
                    <option value="low">低</option>
                    <option value="medium" selected>中</option>
                    <option value="high">高</option>
                </select>
            </div>
            
            <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                    期限
                </label>
                <input 
                    type="date" 
                    name="due_date"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    :min="new Date().toISOString().split('T')[0]"
                >
            </div>
            
            <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                    説明
                </label>
                <textarea 
                    name="description"
                    rows="2"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    placeholder="タスクの詳細（任意）"
                ></textarea>
            </div>
        </div>
        
        <button 
            type="submit"
            class="mt-4 w-full md:w-auto px-6 py-2 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700 transition-colors"
        >
            タスクを追加
        </button>
    </form>
    
    <!-- フィルターと検索 -->
    <div class="bg-white p-4 rounded-lg shadow mb-6">
        <div class="flex flex-col md:flex-row gap-4">
            <div class="flex gap-2">
                <button 
                    hx-get="/todos?filter="
                    hx-target="#todo-list"
                    class="px-4 py-2 rounded-md {{if eq .Filter ""}}bg-blue-600 text-white{{else}}bg-gray-200{{end}}"
                >
                    すべて
                </button>
                <button 
                    hx-get="/todos?filter=active"
                    hx-target="#todo-list"
                    class="px-4 py-2 rounded-md {{if eq .Filter "active"}}bg-blue-600 text-white{{else}}bg-gray-200{{end}}"
                >
                    未完了
                </button>
                <button 
                    hx-get="/todos?filter=completed"
                    hx-target="#todo-list"
                    class="px-4 py-2 rounded-md {{if eq .Filter "completed"}}bg-blue-600 text-white{{else}}bg-gray-200{{end}}"
                >
                    完了
                </button>
            </div>
            
            <div class="flex-1">
                <input 
                    type="search"
                    name="search"
                    placeholder="検索..."
                    hx-get="/todos"
                    hx-trigger="keyup changed delay:500ms"
                    hx-target="#todo-list"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md"
                    value="{{.Search}}"
                >
            </div>
        </div>
    </div>
    
    <!-- TODOリスト -->
    <div id="todo-list" class="space-y-2">
        {{template "todo-list" .}}
    </div>
</div>

<script>
function todoApp() {
    return {
        // ローカルの状態管理
        selectedTodos: [],
        
        // 一括操作
        toggleAll() {
            // 実装
        },
        
        deleteSelected() {
            if (confirm(`${this.selectedTodos.length}件のタスクを削除しますか？`)) {
                // 実装
            }
        }
    }
}
</script>
```

**⚠️ フォームの注意点:** HTMXでフォーム送信後に自動的にリセットするには、`@htmx:after-request`イベントを活用します。成功時のみリセットすることで、エラー時は入力内容を保持できます。

## 復習問題

1. HTMXリクエストと通常のリクエストを同一エンドポイントで判別する方法と、その利点を説明してください。

2. 以下のコードに潜在的な問題があります。何が問題で、どう修正すべきですか？

    ```go
    func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Query().Get("id")
        h.repo.Delete(id)
        w.Write([]byte("OK"))
    }
    ```

3. TODOアプリケーションにリアルタイム同期機能を追加する場合、どのような実装方法が考えられますか？

## 模範解答

1. 判別方法と利点
   - `HX-Request`ヘッダーの確認で判別
   - 利点：同じビジネスロジックで、フルページとパーシャルレスポンスの両方に対応可能
   - プログレッシブエンハンスメントの実現

2. 問題点と修正
   - GETメソッドで削除操作（RESTful原則違反）
   - エラーハンドリングなし
   - CSRF対策なし

   ```go
   func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
       if r.Method != http.MethodDelete {
           http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
           return
       }
       
       id, err := strconv.Atoi(r.URL.Path[len("/todos/"):])
       if err != nil {
           http.Error(w, "Invalid ID", http.StatusBadRequest)
           return
       }
       
       if err := h.repo.Delete(id); err != nil {
           http.Error(w, "Failed to delete", http.StatusInternalServerError)
           return
       }
       
       w.Header().Set("HX-Trigger", "todoDeleted")
       w.WriteHeader(http.StatusOK)
   }
   ```

3. リアルタイム同期の実装方法
   - Server-Sent Events (SSE) を使用したHTMXの自動更新
   - WebSocketを使用した双方向通信
   - ポーリングによる定期的な更新（最も簡単だが効率は劣る）
