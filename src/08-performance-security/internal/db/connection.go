package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

func NewDB(config DBConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// コネクションプールの最適化設定
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 接続確認
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// OptimizedTodoRepository - パフォーマンス最適化されたリポジトリ
type OptimizedTodoRepository struct {
	db *sql.DB
}

type TodoWithTags struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	Tags      []Tag     `json:"tags"`
}

type Tag struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func NewOptimizedTodoRepository(db *sql.DB) *OptimizedTodoRepository {
	return &OptimizedTodoRepository{db: db}
}

// N+1問題を回避する実装（SQLite用）
func (r *OptimizedTodoRepository) GetTodosWithTags(ctx context.Context, limit int) ([]TodoWithTags, error) {
	// SQLite用のクエリ（JSON集約の代わりにGROUP_CONCATを使用）
	query := `
	SELECT 
		t.id,
		t.title,
		t.completed,
		t.created_at,
		GROUP_CONCAT(
			CASE WHEN tg.id IS NOT NULL 
				THEN json_object('id', tg.id, 'name', tg.name, 'color', tg.color)
				ELSE NULL 
			END
		) as tags_json
	FROM todos t
	LEFT JOIN todo_tags_relation ttr ON t.id = ttr.todo_id
	LEFT JOIN tags tg ON ttr.tag_id = tg.id
	GROUP BY t.id, t.title, t.completed, t.created_at
	ORDER BY t.created_at DESC
	LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos with tags: %w", err)
	}
	defer rows.Close()

	var todos []TodoWithTags
	for rows.Next() {
		var todo TodoWithTags
		var tagsJSON sql.NullString

		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &tagsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}

		// タグデータの解析（SQLite GROUP_CONCAT形式）
		if tagsJSON.Valid && tagsJSON.String != "" {
			tagStrings := strings.Split(tagsJSON.String, ",")
			for _, tagStr := range tagStrings {
				if strings.TrimSpace(tagStr) != "" {
					var tag Tag
					if err := json.Unmarshal([]byte(tagStr), &tag); err == nil {
						todo.Tags = append(todo.Tags, tag)
					}
				}
			}
		}

		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return todos, nil
}

// インデックス最適化されたバッチ検索（SQLite用）
func (r *OptimizedTodoRepository) BatchSearchTodos(ctx context.Context, queries []string) ([]TodoWithTags, error) {
	if len(queries) == 0 {
		return []TodoWithTags{}, nil
	}

	// SQLite用の簡単な全文検索
	query := `
	SELECT DISTINCT t.id, t.title, t.completed, t.created_at
	FROM todos t
	WHERE t.title LIKE ?
	ORDER BY t.created_at DESC
	LIMIT 100
	`

	// 複数クエリを結合
	searchTerm := ""
	for i, q := range queries {
		if i > 0 {
			searchTerm += " "
		}
		searchTerm += q
	}
	searchPattern := "%" + searchTerm + "%"

	rows, err := r.db.QueryContext(ctx, query, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to batch search todos: %w", err)
	}
	defer rows.Close()

	var todos []TodoWithTags
	for rows.Next() {
		var todo TodoWithTags

		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		// タグは検索結果では省略（パフォーマンスのため）
		todo.Tags = []Tag{}
		todos = append(todos, todo)
	}

	return todos, nil
}

// プリペアドステートメントを使用した安全な更新
func (r *OptimizedTodoRepository) SafeUpdateTodo(ctx context.Context, id int, title string, completed bool) error {
	query := `
	UPDATE todos 
	SET title = $2, completed = $3, updated_at = CURRENT_TIMESTAMP
	WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, title, completed)
	if err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("todo with id %d not found", id)
	}

	return nil
}

// データベース統計情報の取得（SQLite用）
func (r *OptimizedTodoRepository) GetDBStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// テーブルサイズ
	var todoCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM todos").Scan(&todoCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo count: %w", err)
	}
	stats["todo_count"] = todoCount

	// タグ数
	var tagCount int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tags").Scan(&tagCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag count: %w", err)
	}
	stats["tag_count"] = tagCount

	// 完了済み TODO数
	var completedCount int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM todos WHERE completed = 1").Scan(&completedCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed count: %w", err)
	}
	stats["completed_todos"] = completedCount
	stats["completion_rate"] = float64(completedCount) / float64(todoCount) * 100

	// データベース接続情報
	dbStats := r.db.Stats()
	stats["open_connections"] = dbStats.OpenConnections
	stats["idle_connections"] = dbStats.Idle
	stats["in_use_connections"] = dbStats.InUse
	stats["max_open_connections"] = dbStats.MaxOpenConnections

	return stats, nil
}

// スキーマ初期化（開発用）
func (r *OptimizedTodoRepository) InitSchema(ctx context.Context) error {
	// SQLite用のスキーマ
	schema := `
	-- TODOテーブル
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		completed BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- タグテーブル
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(50) UNIQUE NOT NULL,
		color VARCHAR(7) DEFAULT '#007bff',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- TODO-タグ関連テーブル
	CREATE TABLE IF NOT EXISTS todo_tags_relation (
		todo_id INTEGER,
		tag_id INTEGER,
		PRIMARY KEY (todo_id, tag_id),
		FOREIGN KEY (todo_id) REFERENCES todos(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);

	-- SQLite用インデックス
	CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
	CREATE INDEX IF NOT EXISTS idx_todos_title ON todos(title);
	CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
	`

	_, err := r.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	// サンプルデータの挿入
	sampleData := `
	INSERT OR IGNORE INTO tags (name, color) VALUES 
		('重要', '#dc3545'),
		('仕事', '#007bff'),
		('個人', '#28a745'),
		('学習', '#ffc107');

	INSERT OR IGNORE INTO todos (title, description, completed) VALUES 
		('データベース最適化の学習', 'N+1問題の解決とインデックス最適化について', false),
		('セキュリティテストの実行', 'CSRF、XSS、SQLインジェクション対策のテスト', false),
		('パフォーマンステストの実装', 'ロードテストとベンチマークの実装', true),
		('ログ監視の設定', '構造化ログとメトリクス収集の設定', false),
		('CSRF対策の確認', 'フォームとAJAXリクエストでのCSRF対策確認', true),
		('レート制限の実装', 'API エンドポイントのレート制限実装', false),
		('セキュリティヘッダーの設定', 'CSP、HSTS、XSSフィルタの設定', true),
		('入力検証の強化', 'バリデーションとサニタイゼーションの実装', false);
	`

	_, err = r.db.ExecContext(ctx, sampleData)
	if err != nil {
		return fmt.Errorf("failed to insert sample data: %w", err)
	}

	// タグ関連の挿入
	tagRelations := `
	INSERT OR IGNORE INTO todo_tags_relation (todo_id, tag_id)
	SELECT t.id, tg.id FROM todos t, tags tg WHERE 
		(t.title LIKE '%データベース%' AND tg.name = '学習') OR
		(t.title LIKE '%セキュリティ%' AND tg.name = '重要') OR
		(t.title LIKE '%パフォーマンス%' AND tg.name = '仕事') OR
		(t.title LIKE '%ログ%' AND tg.name = '仕事') OR
		(t.title LIKE '%CSRF%' AND tg.name = '重要') OR
		(t.title LIKE '%レート%' AND tg.name = '仕事') OR
		(t.title LIKE '%ヘッダー%' AND tg.name = '重要') OR
		(t.title LIKE '%入力%' AND tg.name = '個人');
	`

	_, err = r.db.ExecContext(ctx, tagRelations)
	if err != nil {
		return fmt.Errorf("failed to insert tag relations: %w", err)
	}

	return nil
}