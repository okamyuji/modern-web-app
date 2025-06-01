package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"test-debug-demo/internal/db"
	"test-debug-demo/internal/handlers"
	"test-debug-demo/internal/logger"
	"test-debug-demo/internal/middleware"
	"test-debug-demo/internal/models"
	"test-debug-demo/internal/templates"
)

func main() {
	// 環境変数の設定
	isDev := os.Getenv("ENV") != "production"
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	// ロガーの初期化
	logLevel := logger.INFO
	if isDev {
		logLevel = logger.DEBUG
	}
	appLogger := logger.NewLogger(os.Stdout, logLevel)

	// メトリクス初期化
	metrics := logger.NewMetrics()

	// データベース接続
	database, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// スキーマ初期化
	if err := database.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// テストデータの投入（開発環境のみ）
	if isDev {
		if err := database.SeedTestData(); err != nil {
			appLogger.Warn("Failed to seed test data", map[string]interface{}{
				"error": err.Error(),
			}, "")
		}
	}

	// リポジトリとハンドラーの初期化
	todoRepo := models.NewTodoRepositoryWithDriver(database.DB, database.GetDriver())
	todoHandler := handlers.NewTodoHandler(todoRepo, appLogger)

	// ルーター設定
	router := mux.NewRouter()

	// ミドルウェアの設定
	router.Use(middleware.RequestLogger(isDev))
	router.Use(middleware.ErrorHandler(isDev))
	router.Use(middleware.DebugPanel(isDev))
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics.IncrementActiveRequests()
			defer metrics.DecrementActiveRequests()

			start := time.Now()
			defer func() {
				duration := time.Since(start)
				isError := false
				// レスポンスライターから実際のステータスコードを取得するのは複雑なので
				// 簡易的な判定を行う
				metrics.RecordRequest(duration, isError)
			}()

			// トレースID生成
			traceID := fmt.Sprintf("trace-%d", time.Now().UnixNano())
			ctx := context.WithValue(r.Context(), "trace_id", traceID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	})

	// ルート定義
	router.HandleFunc("/", todoHandler.Home).Methods("GET")
	router.HandleFunc("/todos", todoHandler.List).Methods("GET")
	router.HandleFunc("/todos", todoHandler.Create).Methods("POST")
	router.HandleFunc("/todos/{id:[0-9]+}", todoHandler.Update).Methods("PUT", "PATCH")
	router.HandleFunc("/todos/{id:[0-9]+}", todoHandler.Delete).Methods("DELETE")
	router.HandleFunc("/todos/{id:[0-9]+}/toggle", todoHandler.ToggleComplete).Methods("PATCH")

	// メトリクスエンドポイント
	router.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		metricsData := metrics.GetStats()
		
		// データベース統計も追加
		if dbStats, err := database.GetStats(); err == nil {
			for k, v := range dbStats {
				metricsData[fmt.Sprintf("db_%s", k)] = v
			}
		}

		err := templates.MetricsDisplay(metricsData).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Failed to render metrics", http.StatusInternalServerError)
		}
	}).Methods("GET")

	// ヘルスチェックエンドポイント
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
		}

		// データベース接続確認
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		if err := database.PingContext(ctx); err != nil {
			health["status"] = "degraded"
			health["database"] = "failed: " + err.Error()
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			health["database"] = "ok"
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
	"status": "%s",
	"timestamp": "%s",
	"version": "%s",
	"database": "%s",
	"environment": "%s"
}`, health["status"], health["timestamp"], health["version"], health["database"], func() string {
			if isDev {
				return "development"
			}
			return "production"
		}())
	}).Methods("GET")

	// 開発環境用のパニックテストエンドポイント
	if isDev {
		router.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
			appLogger.Info("Panic test endpoint called", nil, getTraceID(r.Context()))
			panic("This is a test panic")
		}).Methods("GET")

		router.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
			appLogger.Error("Test error endpoint called", map[string]interface{}{
				"test_param": r.URL.Query().Get("test"),
			}, getTraceID(r.Context()))
			http.Error(w, "This is a test error", http.StatusInternalServerError)
		}).Methods("GET")
	}

	// サーバー起動
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	appLogger.Info("Starting Test & Debug Demo Application", map[string]interface{}{
		"port":        port,
		"environment": func() string {
			if isDev {
				return "development"
			}
			return "production"
		}(),
		"database": database.Stats(),
	}, "")

	fmt.Printf("🧪 テスト&デバッグデモアプリケーションを起動中...\n")
	fmt.Printf("🌐 URL: http://localhost:%s\n", port)
	fmt.Printf("📊 メトリクス: http://localhost:%s/metrics\n", port)
	fmt.Printf("💚 ヘルスチェック: http://localhost:%s/health\n", port)
	if isDev {
		fmt.Printf("🐛 デバッグモード: 有効\n")
		fmt.Printf("💥 パニックテスト: http://localhost:%s/panic\n", port)
		fmt.Printf("❌ エラーテスト: http://localhost:%s/error\n", port)
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		appLogger.ErrorWithStack("Server failed to start", map[string]interface{}{
			"port": port,
		}, "", err)
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getTraceID(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}