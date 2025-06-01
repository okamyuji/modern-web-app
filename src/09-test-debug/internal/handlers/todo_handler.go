package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"test-debug-demo/internal/logger"
	"test-debug-demo/internal/models"
	"test-debug-demo/internal/templates"
)

type TodoHandler struct {
	repo   *models.TodoRepository
	logger *logger.Logger
}

func NewTodoHandler(repo *models.TodoRepository, log *logger.Logger) *TodoHandler {
	return &TodoHandler{
		repo:   repo,
		logger: log,
	}
}

// Home - TODOアプリのホーム画面
func (h *TodoHandler) Home(w http.ResponseWriter, r *http.Request) {
	todos, err := h.repo.List("", "")
	if err != nil {
		h.logger.Error("Failed to get todos", map[string]interface{}{
			"error": err.Error(),
		}, getTraceID(r.Context()))
		http.Error(w, "Failed to load todos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = templates.HomePage(todos).Render(r.Context(), w)
	if err != nil {
		h.logger.Error("Template render failed", map[string]interface{}{
			"template": "HomePage",
			"error":    err.Error(),
		}, getTraceID(r.Context()))
		http.Error(w, "Template render failed", http.StatusInternalServerError)
	}
}

// List - TODOリストの取得
func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	sortBy := r.URL.Query().Get("sort")

	todos, err := h.repo.List(filter, sortBy)
	if err != nil {
		h.logger.Error("Failed to get todos", map[string]interface{}{
			"filter": filter,
			"sort":   sortBy,
			"error":  err.Error(),
		}, getTraceID(r.Context()))
		http.Error(w, "Failed to load todos", http.StatusInternalServerError)
		return
	}

	// HTMXリクエストの場合はパーシャルHTMLを返す
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		err = templates.TodoList(todos).Render(r.Context(), w)
		if err != nil {
			h.logger.Error("Template render failed", map[string]interface{}{
				"template": "TodoList",
				"error":    err.Error(),
			}, getTraceID(r.Context()))
			http.Error(w, "Template render failed", http.StatusInternalServerError)
		}
		return
	}

	// JSON APIとしてのレスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// Create - 新しいTODOの作成
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	priority := r.FormValue("priority")

	// バリデーション
	var errors []string
	if title == "" {
		errors = append(errors, "タイトルは必須です")
	}
	if len(title) > 100 {
		errors = append(errors, "タイトルは100文字以内で入力してください")
	}
	if priority != "" && !isValidPriority(priority) {
		errors = append(errors, "優先度は low, medium, high のいずれかを指定してください")
	}

	if len(errors) > 0 {
		// HTMXリクエストの場合はエラー表示
		if r.Header.Get("HX-Request") == "true" {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/html")
			err = templates.ValidationErrors(errors).Render(r.Context(), w)
			if err != nil {
				h.logger.Error("Template render failed", map[string]interface{}{
					"template": "ValidationErrors",
					"error":    err.Error(),
				}, getTraceID(r.Context()))
			}
			return
		}

		// 通常のフォーム送信の場合はリダイレクト
		http.Error(w, strings.Join(errors, ", "), http.StatusBadRequest)
		return
	}

	// TODO作成
	todo := &models.Todo{
		Title:       title,
		Description: description,
		Priority:    priority,
	}

	if priority == "" {
		todo.Priority = "medium"
	}

	id, err := h.repo.Create(todo)
	if err != nil {
		h.logger.Error("Failed to create todo", map[string]interface{}{
			"title":       title,
			"description": description,
			"priority":    priority,
			"error":       err.Error(),
		}, getTraceID(r.Context()))

		if r.Header.Get("HX-Request") == "true" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">TODOの作成に失敗しました</div>`)
			return
		}

		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	// 成功ログ
	h.logger.Info("TODO created successfully", map[string]interface{}{
		"todo_id": id,
		"title":   title,
	}, getTraceID(r.Context()))

	// HTMXリクエストの場合は新しいTODOアイテムを返す
	if r.Header.Get("HX-Request") == "true" {
		// 作成されたTODOを取得
		createdTodo, err := h.repo.GetByID(id)
		if err != nil {
			h.logger.Error("Failed to get created todo", map[string]interface{}{
				"todo_id": id,
				"error":   err.Error(),
			}, getTraceID(r.Context()))
			http.Error(w, "Failed to get created todo", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("HX-Trigger", "todoCreated")
		err = templates.TodoItem(createdTodo).Render(r.Context(), w)
		if err != nil {
			h.logger.Error("Template render failed", map[string]interface{}{
				"template": "TodoItem",
				"error":    err.Error(),
			}, getTraceID(r.Context()))
			http.Error(w, "Template render failed", http.StatusInternalServerError)
		}
		return
	}

	// 通常のフォーム送信の場合はリダイレクト
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Update - TODOの更新
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	// 既存のTODOを取得
	todo, err := h.repo.GetByID(id)
	if err != nil {
		h.logger.Error("Todo not found", map[string]interface{}{
			"todo_id": id,
			"error":   err.Error(),
		}, getTraceID(r.Context()))
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// 更新データの設定
	if title := strings.TrimSpace(r.FormValue("title")); title != "" {
		todo.Title = title
	}
	if description := r.FormValue("description"); r.Form.Has("description") {
		todo.Description = description
	}
	if priority := r.FormValue("priority"); priority != "" {
		if isValidPriority(priority) {
			todo.Priority = priority
		}
	}
	if completed := r.FormValue("completed"); completed != "" {
		todo.Completed = completed == "true" || completed == "on"
	}

	err = h.repo.Update(todo)
	if err != nil {
		h.logger.Error("Failed to update todo", map[string]interface{}{
			"todo_id": id,
			"error":   err.Error(),
		}, getTraceID(r.Context()))
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	// 成功ログ
	h.logger.Info("TODO updated successfully", map[string]interface{}{
		"todo_id": id,
		"title":   todo.Title,
	}, getTraceID(r.Context()))

	// HTMXリクエストの場合は更新されたアイテムを返す
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("HX-Trigger", "todoUpdated")
		err = templates.TodoItem(todo).Render(r.Context(), w)
		if err != nil {
			h.logger.Error("Template render failed", map[string]interface{}{
				"template": "TodoItem",
				"error":    err.Error(),
			}, getTraceID(r.Context()))
			http.Error(w, "Template render failed", http.StatusInternalServerError)
		}
		return
	}

	// 通常のリクエストの場合はJSONで返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

// Delete - TODOの削除
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	err = h.repo.Delete(id)
	if err != nil {
		h.logger.Error("Failed to delete todo", map[string]interface{}{
			"todo_id": id,
			"error":   err.Error(),
		}, getTraceID(r.Context()))
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	// 成功ログ
	h.logger.Info("TODO deleted successfully", map[string]interface{}{
		"todo_id": id,
	}, getTraceID(r.Context()))

	// HTMXリクエストの場合は空のレスポンス（要素が削除される）
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "todoDeleted")
		w.WriteHeader(http.StatusOK)
		return
	}

	// 通常のリクエストの場合は204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// ToggleComplete - TODO完了状態の切り替え
func (h *TodoHandler) ToggleComplete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	todo, err := h.repo.GetByID(id)
	if err != nil {
		h.logger.Error("Todo not found", map[string]interface{}{
			"todo_id": id,
			"error":   err.Error(),
		}, getTraceID(r.Context()))
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	// 完了状態を切り替え
	todo.Completed = !todo.Completed

	err = h.repo.Update(todo)
	if err != nil {
		h.logger.Error("Failed to toggle todo completion", map[string]interface{}{
			"todo_id": id,
			"error":   err.Error(),
		}, getTraceID(r.Context()))
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	// HTMXリクエストの場合は更新されたアイテムを返す
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("HX-Trigger", "todoToggled")
		err = templates.TodoItem(todo).Render(r.Context(), w)
		if err != nil {
			h.logger.Error("Template render failed", map[string]interface{}{
				"template": "TodoItem",
				"error":    err.Error(),
			}, getTraceID(r.Context()))
			http.Error(w, "Template render failed", http.StatusInternalServerError)
		}
		return
	}

	// JSON APIレスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

// ユーティリティ関数
func isValidPriority(priority string) bool {
	return priority == "low" || priority == "medium" || priority == "high"
}

func getTraceID(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}