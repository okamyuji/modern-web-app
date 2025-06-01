package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"strings"
	"time"

	"performance-security-demo/internal/db"
	"performance-security-demo/internal/logger"
	"performance-security-demo/internal/middleware"
	"performance-security-demo/internal/templates"
)

type DemoHandler struct {
	repo      *db.OptimizedTodoRepository
	logger    *logger.Logger
	metrics   *logger.Metrics
	sanitizer *middleware.Sanitizer
	validator *middleware.InputValidator
}

func NewDemoHandler(database *sql.DB, log *logger.Logger, metrics *logger.Metrics) *DemoHandler {
	return &DemoHandler{
		repo:      db.NewOptimizedTodoRepository(database),
		logger:    log,
		metrics:   metrics,
		sanitizer: middleware.NewSanitizer(),
		validator: middleware.NewInputValidator(),
	}
}

// Home - デモホームページ
func (h *DemoHandler) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := templates.HomePage().Render(r.Context(), w)
	if err != nil {
		h.logger.ErrorWithStack("Template render failed", map[string]interface{}{
			"template": "HomePage",
		}, getTraceID(r.Context()), err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// PerformanceDemo - パフォーマンス最適化のデモ
func (h *DemoHandler) PerformanceDemo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	// N+1問題を回避した効率的なクエリ
	todos, err := h.repo.GetTodosWithTags(ctx, 10)
	if err != nil {
		h.logger.ErrorWithStack("Failed to get todos with tags", map[string]interface{}{
			"limit": 10,
		}, getTraceID(ctx), err)
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}

	duration := float64(time.Since(start).Nanoseconds()) / 1e6

	// HTMXリクエストの場合はパーシャルHTMLを返す
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		err = templates.TodoList(todos).Render(ctx, w)
		if err != nil {
			h.logger.ErrorWithStack("Template render failed", map[string]interface{}{
				"template": "TodoList",
			}, getTraceID(ctx), err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// 完全なページを返す
	w.Header().Set("Content-Type", "text/html")
	err = templates.PerformancePage(todos, duration).Render(ctx, w)
	if err != nil {
		h.logger.ErrorWithStack("Template render failed", map[string]interface{}{
			"template": "PerformancePage",
		}, getTraceID(ctx), err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// SecurityDemo - セキュリティ機能のデモ
func (h *DemoHandler) SecurityDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.handleSecurityForm(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err := templates.SecurityPage().Render(r.Context(), w)
	if err != nil {
		h.logger.ErrorWithStack("Template render failed", map[string]interface{}{
			"template": "SecurityPage",
		}, getTraceID(r.Context()), err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleSecurityForm - セキュリティフォームの処理
func (h *DemoHandler) handleSecurityForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// フォームデータの解析
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// 入力の検証とサニタイゼーション
	username := h.validator.SanitizeInput(r.FormValue("username"), 50)
	email := h.validator.SanitizeInput(r.FormValue("email"), 100)
	comment := h.sanitizer.Sanitize(r.FormValue("comment"))

	// バリデーション
	var errors []string
	if !h.validator.ValidateUsername(username) {
		errors = append(errors, "ユーザー名は3-20文字の英数字・アンダースコアで入力してください")
	}
	if !h.validator.ValidateEmail(email) {
		errors = append(errors, "有効なメールアドレスを入力してください")
	}

	// HTMXリクエストの場合はパーシャルHTMLを返す
	if r.Header.Get("HX-Request") == "true" {
		if len(errors) > 0 {
			w.Header().Set("Content-Type", "text/html")
			err = templates.ValidationErrors(errors).Render(ctx, w)
			if err != nil {
				h.logger.ErrorWithStack("Template render failed", nil, getTraceID(ctx), err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// 成功時のレスポンス
		result := map[string]string{
			"username": username,
			"email":    email,
			"comment":  comment,
		}

		w.Header().Set("Content-Type", "text/html")
		err = templates.SecurityResult(result).Render(ctx, w)
		if err != nil {
			h.logger.ErrorWithStack("Template render failed", nil, getTraceID(ctx), err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// 通常のフォーム送信の場合はリダイレクト
	http.Redirect(w, r, "/security", http.StatusSeeOther)
}

// SearchDemo - 安全な検索のデモ
func (h *DemoHandler) SearchDemo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("q")

	if query == "" {
		w.Header().Set("Content-Type", "text/html")
		err := templates.EmptySearchResult().Render(ctx, w)
		if err != nil {
			h.logger.ErrorWithStack("Template render failed", nil, getTraceID(ctx), err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// 入力のサニタイゼーション
	cleanQuery := h.validator.SanitizeInput(query, 100)

	// 簡単な検索実行（デモ用）
	todos := []db.TodoWithTags{
		{
			ID:        1,
			Title:     "データベース最適化の学習",
			Completed: false,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			Tags: []db.Tag{
				{ID: 1, Name: "学習", Color: "#ffc107"},
			},
		},
		{
			ID:        2,
			Title:     "データベースパフォーマンステスト",
			Completed: true,
			CreatedAt: time.Now().Add(-12 * time.Hour),
			Tags: []db.Tag{
				{ID: 2, Name: "仕事", Color: "#007bff"},
			},
		},
	}
	
	// クエリにマッチするものだけをフィルタリング
	var filteredTodos []db.TodoWithTags
	for _, todo := range todos {
		if strings.Contains(strings.ToLower(todo.Title), strings.ToLower(cleanQuery)) {
			filteredTodos = append(filteredTodos, todo)
		}
	}
	todos = filteredTodos

	// HTMXリクエストの場合はパーシャルHTMLを返す
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		err := templates.SearchResults(todos, cleanQuery).Render(ctx, w)
		if err != nil {
			h.logger.ErrorWithStack("Template render failed", nil, getTraceID(ctx), err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// 通常のページ
	w.Header().Set("Content-Type", "text/html")
	err := templates.SearchPage(todos, cleanQuery).Render(ctx, w)
	if err != nil {
		h.logger.ErrorWithStack("Template render failed", nil, getTraceID(ctx), err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HealthCheck - ヘルスチェックエンドポイント
func (h *DemoHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health := struct {
		Status    string                 `json:"status"`
		Checks    map[string]string      `json:"checks"`
		Version   string                 `json:"version"`
		Uptime    string                 `json:"uptime"`
		Metrics   map[string]interface{} `json:"metrics"`
		Timestamp time.Time              `json:"timestamp"`
	}{
		Status:    "ok",
		Checks:    make(map[string]string),
		Version:   "1.0.0",
		Metrics:   h.metrics.GetStats(),
		Timestamp: time.Now().UTC(),
	}

	// データベース接続確認
	dbStats, err := h.repo.GetDBStats(ctx)
	if err != nil {
		health.Status = "degraded"
		health.Checks["database"] = "failed: " + err.Error()
	} else {
		health.Checks["database"] = "ok"
		health.Metrics["database"] = dbStats
	}

	// システム情報
	health.Checks["system"] = "ok"
	health.Metrics["system"] = map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		"memory":     getMemoryStats(),
	}

	// レスポンス
	w.Header().Set("Content-Type", "application/json")
	if health.Status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(health)
	if err != nil {
		h.logger.ErrorWithStack("Health check encoding failed", nil, getTraceID(r.Context()), err)
	}
}

// MetricsEndpoint - メトリクス情報の取得
func (h *DemoHandler) MetricsEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(h.metrics.GetStats())
	if err != nil {
		h.logger.ErrorWithStack("Metrics encoding failed", nil, getTraceID(r.Context()), err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// getTraceID - コンテキストからトレースIDを取得
func getTraceID(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// getMemoryStats - メモリ統計情報を取得
func getMemoryStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"alloc_mb":       m.Alloc / 1024 / 1024,
		"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
		"sys_mb":         m.Sys / 1024 / 1024,
		"num_gc":         m.NumGC,
	}
}