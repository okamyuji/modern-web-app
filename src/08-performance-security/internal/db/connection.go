package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

// N+1問題を回避する実装
func (r *OptimizedTodoRepository) GetTodosWithTags(ctx context.Context, limit int) ([]TodoWithTags, error) {
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
	LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos with tags: %w", err)
	}
	defer rows.Close()

	var todos []TodoWithTags
	for rows.Next() {
		var todo TodoWithTags
		var tagsJSON []byte

		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &tagsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}

		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &todo.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
			}
		}

		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return todos, nil
}

// インデックス最適化されたバッチ検索
func (r *OptimizedTodoRepository) BatchSearchTodos(ctx context.Context, queries []string) ([]TodoWithTags, error) {
	if len(queries) == 0 {
		return []TodoWithTags{}, nil
	}

	// PostgreSQLのGINインデックスを活用した全文検索
	query := `
	SELECT DISTINCT t.id, t.title, t.completed, t.created_at,
		   ts_rank(to_tsvector('english', t.title || ' ' || COALESCE(t.description, '')), plainto_tsquery('english', $1)) as rank
	FROM todos t
	WHERE to_tsvector('english', t.title || ' ' || COALESCE(t.description, '')) @@ plainto_tsquery('english', $1)
	ORDER BY rank DESC, t.created_at DESC
	LIMIT 100
	`

	// 複数クエリを結合
	searchTerm := ""
	for i, q := range queries {
		if i > 0 {
			searchTerm += " | "
		}
		searchTerm += q
	}

	rows, err := r.db.QueryContext(ctx, query, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to batch search todos: %w", err)
	}
	defer rows.Close()

	var todos []TodoWithTags
	for rows.Next() {
		var todo TodoWithTags
		var rank float64

		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &rank)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

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

// データベース統計情報の取得
func (r *OptimizedTodoRepository) GetDBStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// テーブルサイズ
	var todoCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM todos").Scan(&todoCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo count: %w", err)
	}
	stats["todo_count"] = todoCount

	// インデックス使用状況
	var indexUsage float64
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(
			(sum(idx_scan) * 100.0 / NULLIF(sum(idx_scan + seq_scan), 0)), 0
		) 
		FROM pg_stat_user_tables 
		WHERE relname = 'todos'
	`).Scan(&indexUsage)
	if err != nil {
		return nil, fmt.Errorf("failed to get index usage: %w", err)
	}
	stats["index_usage_percentage"] = indexUsage

	// データベース接続情報
	dbStats := r.db.Stats()
	stats["open_connections"] = dbStats.OpenConnections
	stats["idle_connections"] = dbStats.Idle
	stats["in_use_connections"] = dbStats.InUse

	return stats, nil
}

// スキーマ初期化（開発用）
func (r *OptimizedTodoRepository) InitSchema(ctx context.Context) error {
	schema := `
	-- TODOテーブル
	CREATE TABLE IF NOT EXISTS todos (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		completed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- タグテーブル
	CREATE TABLE IF NOT EXISTS tags (
		id SERIAL PRIMARY KEY,
		name VARCHAR(50) UNIQUE NOT NULL,
		color VARCHAR(7) DEFAULT '#007bff',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- TODO-タグ関連テーブル
	CREATE TABLE IF NOT EXISTS todo_tags_relation (
		todo_id INTEGER REFERENCES todos(id) ON DELETE CASCADE,
		tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
		PRIMARY KEY (todo_id, tag_id)
	);

	-- パフォーマンス最適化インデックス
	CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
	CREATE INDEX IF NOT EXISTS idx_todos_fulltext ON todos USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));
	CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

	-- サンプルデータの挿入
	INSERT INTO tags (name, color) VALUES 
		('重要', '#dc3545'),
		('仕事', '#007bff'),
		('個人', '#28a745'),
		('学習', '#ffc107')
	ON CONFLICT (name) DO NOTHING;
	`

	_, err := r.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}