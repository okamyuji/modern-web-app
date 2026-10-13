package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"unicode/utf8"

	_ "github.com/mattn/go-sqlite3"

	"deployment-demo/internal/config"
	"deployment-demo/internal/monitoring"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	Priority    string `json:"priority"`
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Version   string                 `json:"version"`
	BuildTime string                 `json:"build_time"`
	Data      map[string]interface{} `json:"data"`
}

func main() {
	// ヘルスチェックモード
	if len(os.Args) > 1 && os.Args[1] == "-health-check" {
		resp, err := http.Get("http://localhost:8080/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	log.Printf("アプリケーション開始: %s v%s (ビルド時刻: %s)", cfg.AppName, Version, BuildTime)

	// データベースの初期化
	db, err := initDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("データベースの初期化に失敗しました: %v", err)
	}
	defer db.Close()

	// メトリクスの初期化
	monitoring.InitMetrics(cfg)

	// Prometheusメトリクスサーバーの起動
	if cfg.PrometheusEnabled {
		monitoring.StartMetricsServer(cfg.PrometheusPort)
		log.Printf("Prometheusメトリクスサーバーが起動しました: http://localhost:%s/metrics", cfg.PrometheusPort)
	}

	// ルーターの設定
	mux := setupRoutes(db, cfg)

	// HTTPサーバーの設定
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Printf("サーバーが起動しました: http://localhost:%s", cfg.Port)

	// グレースフルシャットダウンの設定
	go monitoring.GracefulShutdown(server, cfg, func() {
		log.Println("データベース接続を閉じています...")
		db.Close()
	})

	// サーバー起動
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}

	log.Println("サーバーが正常に停止しました")
}

func initDatabase(databaseURL string) (*sql.DB, error) {
	path, ok := strings.CutPrefix(databaseURL, "sqlite://")
	// 空のパスは SQLite では一時 DB になり、接続を閉じるとデータが消える
	if !ok || path == "" {
		// URL は資格情報を含みうるので、ログに出るエラーへ載せない
		return nil, errors.New("未対応のDATABASE_URLです（sqlite://のみ対応）")
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("データベース接続に失敗: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("データベースのPingに失敗: %w", err)
	}

	// テーブルの作成
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			completed BOOLEAN DEFAULT FALSE,
			priority TEXT DEFAULT 'medium',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("テーブル作成に失敗: %w", err)
	}

	return db, nil
}

func setupRoutes(db *sql.DB, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// 静的ファイルの配信
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// ヘルスチェック
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		healthData := monitoring.HealthCheck()
		healthData["version"] = Version
		healthData["build_time"] = BuildTime
		healthData["environment"] = cfg.Env
		healthData["database"] = "connected"

		response := HealthResponse{
			Status:    "healthy",
			Version:   Version,
			BuildTime: BuildTime,
			Data:      healthData,
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(response)
	})

	// メトリクス（メインアプリケーション用）
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		metrics := map[string]interface{}{
			"version":    Version,
			"build_time": BuildTime,
			"uptime":     monitoring.HealthCheck()["uptime"],
		}
		json.NewEncoder(w).Encode(metrics)
	})

	// Todo API
	api := newAPIHandler(db, cfg.APIToken, cfg.APIRateLimit)
	mux.Handle("/api/todos", api)
	mux.Handle("/api/todos/", api)

	// メインページ
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		html := `<!DOCTYPE html>
<html lang="ja" class="h-full">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Deployment Demo</title>
    <link href="/static/css/app.css" rel="stylesheet">
    <style>
        /* フォールバックスタイル - Tailwindが読み込まれない場合に備えて */
        .btn {
            display: inline-block;
            padding: 0.75rem 1.5rem;
            margin: 0.25rem;
            font-weight: 500;
            text-align: center;
            text-decoration: none;
            border-radius: 0.5rem;
            transition: all 0.2s ease-in-out;
            border: none;
            cursor: pointer;
        }
        .btn-primary {
            background-color: #2563eb;
            color: white;
        }
        .btn-primary:hover {
            background-color: #1d4ed8;
        }
        .btn-secondary {
            background-color: #6b7280;
            color: white;
        }
        .btn-secondary:hover {
            background-color: #4b5563;
        }
        body {
            font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif;
            background-color: #f9fafb;
            margin: 0;
            padding: 2rem;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 1rem;
            border: 4px dashed #d1d5db;
            padding: 2rem;
            min-height: 24rem;
            text-align: center;
        }
        h1 {
            font-size: 2.5rem;
            font-weight: bold;
            color: #111827;
            margin-bottom: 1rem;
        }
        p {
            font-size: 1.125rem;
            color: #6b7280;
            margin-bottom: 2rem;
        }
        .button-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }
        .version-info {
            margin-top: 2rem;
            font-size: 0.875rem;
            color: #9ca3af;
        }
        @media (prefers-color-scheme: dark) {
            body { background-color: #111827; }
            .container { background-color: #1f2937; border-color: #374151; }
            h1 { color: #f9fafb; }
            p { color: #d1d5db; }
            .version-info { color: #9ca3af; }
        }
    </style>
</head>
<body class="h-full bg-gray-50 dark:bg-gray-900">
    <div class="min-h-full">
        <div class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
            <div class="px-4 py-6 sm:px-0">
                <div class="container border-4 border-dashed border-gray-200 dark:border-gray-700 rounded-lg h-96 p-8">
                    <div class="text-center">
                        <h1 class="text-4xl font-bold text-gray-900 dark:text-gray-100 mb-4">
                            🚀 Deployment Demo
                        </h1>
                        <p class="text-lg text-gray-600 dark:text-gray-300 mb-8">
                            デプロイメントと運用機能のデモアプリケーション
                        </p>
                        <div class="space-y-4">
                            <div class="button-grid grid grid-cols-1 md:grid-cols-3 gap-4">
                                <a href="/health" class="btn btn-primary">
                                    💚 ヘルスチェック
                                </a>
                                <a href="/metrics" class="btn btn-secondary">
                                    📊 メトリクス
                                </a>
                                <a href="/api/todos" class="btn btn-secondary">
                                    📝 Todo API
                                </a>
                            </div>
                            <div class="version-info mt-8 text-sm text-gray-500 dark:text-gray-400">
                                <p><strong>Version:</strong> ` + Version + `</p>
                                <p><strong>Build Time:</strong> ` + BuildTime + `</p>
                                <p><strong>Environment:</strong> development</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// メトリクス収集ミドルウェアを適用
	return monitoring.PrometheusMiddleware(mux)
}

func handleTodos(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			todos, err := getTodos(db)
			if err != nil {
				log.Printf("Todo一覧の取得に失敗: %v", err)
				http.Error(w, internalErrorMessage, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(todos)

		case http.MethodPost:
			var todo Todo
			r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
			if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
				var tooLarge *http.MaxBytesError
				if errors.As(err, &tooLarge) {
					http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
					return
				}
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(todo.Title) == "" || utf8.RuneCountInString(todo.Title) > maxTitleLength {
				http.Error(w, fmt.Sprintf("title is required and must be at most %d characters", maxTitleLength), http.StatusBadRequest)
				return
			}
			// priority はメトリクスのラベルになるので、値の種類を限る
			if !slices.Contains([]string{"", "low", "medium", "high"}, todo.Priority) {
				http.Error(w, "priority must be one of low, medium, high", http.StatusBadRequest)
				return
			}

			id, err := createTodo(db, todo)
			if err != nil {
				log.Printf("Todoの作成に失敗: %v", err)
				http.Error(w, internalErrorMessage, http.StatusInternalServerError)
				return
			}

			todo.ID = id
			monitoring.RecordTodoCreated(todo.Priority)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(todo)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleTodoByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simplified ID extraction for demo
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Todo by ID endpoint"})
	}
}

func getTodos(db *sql.DB) ([]Todo, error) {
	rows, err := db.Query("SELECT id, title, description, completed, priority FROM todos ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.Priority)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

func createTodo(db *sql.DB, todo Todo) (int, error) {
	result, err := db.Exec(
		"INSERT INTO todos (title, description, priority) VALUES (?, ?, ?)",
		todo.Title, todo.Description, todo.Priority,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}
