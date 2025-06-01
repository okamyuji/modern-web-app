package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"performance-security-demo/internal/db"
	"performance-security-demo/internal/handlers"
	"performance-security-demo/internal/logger"
	"performance-security-demo/internal/middleware"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 環境変数の読み込み（デフォルト値付き）
	dbConfig := db.DBConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            5432,
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "password"),
		Database:        getEnv("DB_NAME", "performance_security_demo"),
		SSLMode:         getEnv("DB_SSL_MODE", "disable"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	// データベース接続（PostgreSQL）
	database, err := db.NewDB(dbConfig)
	if err != nil {
		// PostgreSQLに接続できない場合は、SQLiteを使用
		log.Printf("PostgreSQL connection failed, falling back to SQLite: %v", err)
		database, err = initSQLite()
		if err != nil {
			log.Fatal("Database initialization failed:", err)
		}
	}
	defer database.Close()

	// スキーマ初期化
	repo := db.NewOptimizedTodoRepository(database)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := repo.InitSchema(ctx); err != nil {
		log.Printf("Schema initialization warning: %v", err)
	}

	// ログ設定
	appLogger := logger.NewLogger(os.Stdout, logger.INFO)
	metrics := logger.NewMetrics()

	// CSRF管理
	csrfManager := middleware.NewCSRFManager()

	// レート制限
	rateLimiter := middleware.NewRateLimiter()

	// ハンドラーの初期化
	demoHandler := handlers.NewDemoHandler(database, appLogger, metrics)

	// ルーターの設定
	r := mux.NewRouter()

	// ミドルウェアの適用
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.Gzip)
	r.Use(middleware.Cache(24 * time.Hour))
	r.Use(middleware.ResponseHeaders)
	r.Use(middleware.RequestSizeLimit(10 * 1024 * 1024)) // 10MB
	r.Use(logger.RequestLogger(appLogger))
	r.Use(logger.MetricsMiddleware(metrics))
	r.Use(rateLimiter.Middleware(100, 60)) // 100 requests per minute

	// パブリックルート
	r.HandleFunc("/", demoHandler.Home).Methods("GET")
	r.HandleFunc("/performance", demoHandler.PerformanceDemo).Methods("GET")
	r.HandleFunc("/security", demoHandler.SecurityDemo).Methods("GET", "POST")
	r.HandleFunc("/search", demoHandler.SearchDemo).Methods("GET")

	// ヘルスチェック・メトリクス
	r.HandleFunc("/health", demoHandler.HealthCheck).Methods("GET")
	r.HandleFunc("/metrics", demoHandler.MetricsEndpoint).Methods("GET")

	// 静的ファイルの配信
	r.PathPrefix("/static/").Handler(
		middleware.StaticFileOptimizer(
			http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))),
		),
	)

	// CSRF保護が必要なルート
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(csrfManager.Middleware)

	fmt.Println("🔒 パフォーマンス・セキュリティデモアプリケーションを起動中...")
	fmt.Println("🌐 URL: http://localhost:8083")
	fmt.Println("🗄️  データベース設定:")
	fmt.Printf("   Host: %s:%d\n", dbConfig.Host, dbConfig.Port)
	fmt.Printf("   Database: %s\n", dbConfig.Database)
	fmt.Printf("   Max Connections: %d\n", dbConfig.MaxOpenConns)
	fmt.Println("---")
	fmt.Println("💡 機能:")
	fmt.Println("  ⚡ パフォーマンス最適化:")
	fmt.Println("    • N+1問題解決")
	fmt.Println("    • インデックス最適化")
	fmt.Println("    • gzip圧縮")
	fmt.Println("    • 静的リソースキャッシュ")
	fmt.Println("    • 構造化ログ")
	fmt.Println("    • メトリクス収集")
	fmt.Println("  🔒 セキュリティ機能:")
	fmt.Println("    • CSRF対策")
	fmt.Println("    • XSS対策")
	fmt.Println("    • SQLインジェクション対策")
	fmt.Println("    • 入力検証・サニタイゼーション")
	fmt.Println("    • セキュリティヘッダー")
	fmt.Println("    • レート制限")
	fmt.Println("🛑 終了するには Ctrl+C を押してください")

	log.Fatal(http.ListenAndServe(":8083", r))
}

// getEnv - 環境変数の取得（デフォルト値付き）
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// initSQLite - SQLite フォールバック初期化
func initSQLite() (*sql.DB, error) {
	// SQLite用のスキーマを簡略化
	database, err := sql.Open("sqlite3", "performance_security_demo.db")
	if err != nil {
		return nil, err
	}

	// 基本的なテーブル作成
	schema := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	INSERT OR IGNORE INTO todos (title, completed) VALUES 
		('データベース最適化の学習', false),
		('セキュリティテストの実行', false),
		('パフォーマンステストの実装', true),
		('ログ監視の設定', false),
		('CSRF対策の確認', true);
	`

	_, err = database.Exec(schema)
	return database, err
}