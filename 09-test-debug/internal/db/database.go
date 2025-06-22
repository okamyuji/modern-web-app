package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	*sql.DB
	driver string
}

func NewDatabase() (*Database, error) {
	// PostgreSQL接続を試行
	if pgDB, err := connectPostgreSQL(); err == nil {
		log.Println("Connected to PostgreSQL database")
		return &Database{DB: pgDB, driver: "postgres"}, nil
	}

	// PostgreSQLに接続できない場合はSQLiteを使用
	log.Println("PostgreSQL connection failed, using SQLite")
	sqliteDB, err := connectSQLite()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	log.Println("Connected to SQLite database")
	return &Database{DB: sqliteDB, driver: "sqlite3"}, nil
}

func (db *Database) GetDriver() string {
	return db.driver
}

func connectPostgreSQL() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "test_debug_demo")

	if password == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func connectSQLite() (*sql.DB, error) {
	dbPath := getEnv("SQLITE_PATH", "./test_debug_demo.db")
	
	db, err := sql.Open("sqlite3", dbPath+"?cache=shared&mode=rwc")
	if err != nil {
		return nil, err
	}

	// SQLiteの設定
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA temp_store = memory",
		"PRAGMA mmap_size = 268435456", // 256MB
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set pragma %s: %w", pragma, err)
		}
	}

	return db, nil
}

func (db *Database) InitSchema() error {
	var schema string

	if db.driver == "postgres" {
		schema = `
		CREATE TABLE IF NOT EXISTS todos (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			priority VARCHAR(10) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
			completed BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
		CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority);
		CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at DESC);
		`
	} else {
		schema = `
		CREATE TABLE IF NOT EXISTS todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			priority TEXT DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
			completed BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
		CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority);
		CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at DESC);
		`
	}

	_, err := db.Exec(schema)
	return err
}

func (db *Database) SeedTestData() error {
	// まず既存のテストデータがあるかチェック
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM todos").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	// 既にデータがある場合はスキップ
	if count > 0 {
		return nil
	}

	testTodos := []struct {
		title       string
		description string
		priority    string
		completed   bool
	}{
		{
			title:       "プロジェクト設計",
			description: "新しいプロジェクトのアーキテクチャを設計する",
			priority:    "high",
			completed:   false,
		},
		{
			title:       "ユニットテスト作成",
			description: "すべてのモデルとハンドラーのテストを作成",
			priority:    "high",
			completed:   true,
		},
		{
			title:       "ドキュメント更新",
			description: "README とAPI仕様書の更新",
			priority:    "medium",
			completed:   false,
		},
		{
			title:       "パフォーマンス最適化",
			description: "データベースクエリとHTTPレスポンスの最適化",
			priority:    "medium",
			completed:   false,
		},
		{
			title:       "コードレビュー",
			description: "チームメンバーのコードレビューを実施",
			priority:    "low",
			completed:   true,
		},
	}

	for _, todo := range testTodos {
		var query string
		if db.driver == "postgres" {
			query = `
			INSERT INTO todos (title, description, priority, completed)
			VALUES ($1, $2, $3, $4)`
		} else {
			query = `
			INSERT INTO todos (title, description, priority, completed)
			VALUES (?, ?, ?, ?)`
		}

		_, err := db.Exec(query, todo.title, todo.description, todo.priority, todo.completed)
		if err != nil {
			return fmt.Errorf("failed to insert test data: %w", err)
		}
	}

	return nil
}

func (db *Database) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 基本統計
	var totalTodos, completedTodos, pendingTodos int
	err := db.QueryRow("SELECT COUNT(*) FROM todos").Scan(&totalTodos)
	if err != nil {
		return nil, err
	}

	err = db.QueryRow("SELECT COUNT(*) FROM todos WHERE completed = true").Scan(&completedTodos)
	if err != nil {
		return nil, err
	}

	pendingTodos = totalTodos - completedTodos

	stats["total_todos"] = totalTodos
	stats["completed_todos"] = completedTodos
	stats["pending_todos"] = pendingTodos

	// 優先度別統計
	priorityStats := make(map[string]int)
	rows, err := db.Query("SELECT priority, COUNT(*) FROM todos GROUP BY priority")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var priority string
		var count int
		if err := rows.Scan(&priority, &count); err != nil {
			return nil, err
		}
		priorityStats[priority] = count
	}

	stats["priority_stats"] = priorityStats

	// データベース接続統計
	dbStats := db.Stats()
	stats["db_stats"] = map[string]interface{}{
		"open_connections": dbStats.OpenConnections,
		"in_use":          dbStats.InUse,
		"idle":            dbStats.Idle,
	}

	stats["driver"] = db.driver

	return stats, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}