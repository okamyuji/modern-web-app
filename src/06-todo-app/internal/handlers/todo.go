package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"todo-app/internal/models"
	"todo-app/internal/templates"

	"github.com/gorilla/mux"
)

type TodoHandler struct {
	repo *models.TodoRepository
}

func NewTodoHandler(repo *models.TodoRepository) *TodoHandler {
	return &TodoHandler{
		repo: repo,
	}
}

// メインページの表示
func (h *TodoHandler) Index(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	search := r.URL.Query().Get("search")

	todos, err := h.repo.List(filter, search)
	if err != nil {
		h.sendError(w, "データの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	stats, err := h.repo.GetStats()
	if err != nil {
		h.sendError(w, "統計情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// HTMXリクエストの場合は部分的なHTMLを返す
	if r.Header.Get("HX-Request") == "true" {
		target := r.Header.Get("HX-Target")
		if target == "todo-list" || target == "#todo-list" {
			w.Header().Set("Content-Type", "text/html")
			err = templates.TodoList(todos).Render(r.Context(), w)
		} else {
			// 全体の再レンダリング
			w.Header().Set("Content-Type", "text/html")
			err = templates.Home(todos, stats, filter, search).Render(r.Context(), w)
		}
	} else {
		w.Header().Set("Content-Type", "text/html")
		err = templates.Home(todos, stats, filter, search).Render(r.Context(), w)
	}

	if err != nil {
		h.sendError(w, "テンプレートの実行に失敗しました", http.StatusInternalServerError)
		return
	}
}

// TODO作成
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// フォームデータの解析
	err := r.ParseForm()
	if err != nil {
		h.sendError(w, "フォームデータの解析に失敗しました", http.StatusBadRequest)
		return
	}

	// バリデーション
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		h.sendError(w, "タイトルは必須です", http.StatusBadRequest)
		return
	}

	todo := models.Todo{
		Title:       title,
		Description: strings.TrimSpace(r.FormValue("description")),
		Priority:    r.FormValue("priority"),
	}

	// 優先度のバリデーション
	if todo.Priority != "low" && todo.Priority != "medium" && todo.Priority != "high" {
		todo.Priority = "medium"
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

	// 作成されたTODOを取得
	createdTodo, err := h.repo.GetByID(id)
	if err != nil {
		h.sendError(w, "作成されたTODOの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 新しいTODOアイテムのHTMLを返す
	w.Header().Set("HX-Trigger", "todoAdded")
	w.Header().Set("Content-Type", "text/html")
	err = templates.TodoItem(*createdTodo).Render(r.Context(), w)
	if err != nil {
		h.sendError(w, "テンプレートの実行に失敗しました", http.StatusInternalServerError)
	}
}

// TODO更新
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.sendError(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// IDの取得
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendError(w, "無効なIDです", http.StatusBadRequest)
		return
	}

	// 既存のTODOを取得
	todo, err := h.repo.GetByID(id)
	if err != nil {
		h.sendError(w, "TODOが見つかりません", http.StatusNotFound)
		return
	}

	// フォームデータの解析
	err = r.ParseForm()
	if err != nil {
		h.sendError(w, "フォームデータの解析に失敗しました", http.StatusBadRequest)
		return
	}

	// バリデーション
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		h.sendError(w, "タイトルは必須です", http.StatusBadRequest)
		return
	}

	// データ更新
	todo.Title = title
	todo.Description = strings.TrimSpace(r.FormValue("description"))
	todo.Priority = r.FormValue("priority")

	// 優先度のバリデーション
	if todo.Priority != "low" && todo.Priority != "medium" && todo.Priority != "high" {
		todo.Priority = "medium"
	}

	// 期限の解析
	dueDateStr := r.FormValue("due_date")
	if dueDateStr != "" {
		dueDate, err := time.Parse("2006-01-02", dueDateStr)
		if err == nil {
			todo.DueDate = &dueDate
		} else {
			todo.DueDate = nil
		}
	} else {
		todo.DueDate = nil
	}

	// データベース更新
	err = h.repo.Update(todo)
	if err != nil {
		h.sendError(w, "TODOの更新に失敗しました", http.StatusInternalServerError)
		return
	}

	// 更新後のTODOを取得
	updatedTodo, err := h.repo.GetByID(id)
	if err != nil {
		h.sendError(w, "更新されたTODOの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 更新されたTODOアイテムのHTMLを返す
	w.Header().Set("HX-Trigger", "todoUpdated")
	w.Header().Set("Content-Type", "text/html")
	err = templates.TodoItem(*updatedTodo).Render(r.Context(), w)
	if err != nil {
		h.sendError(w, "テンプレートの実行に失敗しました", http.StatusInternalServerError)
	}
}

// TODO削除
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.sendError(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// IDの取得
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendError(w, "無効なIDです", http.StatusBadRequest)
		return
	}

	// データベースから削除
	err = h.repo.Delete(id)
	if err != nil {
		h.sendError(w, "TODOの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	// 削除成功のレスポンス（空のレスポンス）
	w.Header().Set("HX-Trigger", "todoDeleted")
	w.WriteHeader(http.StatusOK)
}

// 完了状態の切り替え
func (h *TodoHandler) ToggleCompleted(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		h.sendError(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// IDの取得
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendError(w, "無効なIDです", http.StatusBadRequest)
		return
	}

	// 完了状態を切り替え
	err = h.repo.ToggleCompleted(id)
	if err != nil {
		h.sendError(w, "完了状態の切り替えに失敗しました", http.StatusInternalServerError)
		return
	}

	// 更新後のTODOを取得
	todo, err := h.repo.GetByID(id)
	if err != nil {
		h.sendError(w, "TODOの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 更新されたTODOアイテムのHTMLを返す
	w.Header().Set("HX-Trigger", "todoToggled")
	w.Header().Set("Content-Type", "text/html")
	err = templates.TodoItem(*todo).Render(r.Context(), w)
	if err != nil {
		h.sendError(w, "テンプレートの実行に失敗しました", http.StatusInternalServerError)
	}
}

// 編集フォーム表示
func (h *TodoHandler) Edit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// IDの取得
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendError(w, "無効なIDです", http.StatusBadRequest)
		return
	}

	// TODOを取得
	todo, err := h.repo.GetByID(id)
	if err != nil {
		h.sendError(w, "TODOが見つかりません", http.StatusNotFound)
		return
	}

	// 編集フォームのHTMLを返す
	w.Header().Set("Content-Type", "text/html")
	err = templates.TodoEditForm(*todo).Render(r.Context(), w)
	if err != nil {
		h.sendError(w, "テンプレートの実行に失敗しました", http.StatusInternalServerError)
	}
}

// 単一TODO表示（編集キャンセル用）
func (h *TodoHandler) Show(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// IDの取得
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendError(w, "無効なIDです", http.StatusBadRequest)
		return
	}

	// TODOを取得
	todo, err := h.repo.GetByID(id)
	if err != nil {
		h.sendError(w, "TODOが見つかりません", http.StatusNotFound)
		return
	}

	// TODOアイテムのHTMLを返す
	w.Header().Set("Content-Type", "text/html")
	err = templates.TodoItem(*todo).Render(r.Context(), w)
	if err != nil {
		h.sendError(w, "テンプレートの実行に失敗しました", http.StatusInternalServerError)
	}
}

// エラーレスポンスの送信
func (h *TodoHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("HX-Retarget", "#error-message")
	w.Header().Set("HX-Reswap", "innerHTML")
	w.WriteHeader(status)
	errorHTML := `<div class="bg-red-100 dark:bg-red-900 border border-red-400 dark:border-red-600 text-red-700 dark:text-red-300 px-4 py-3 rounded relative" x-data x-init="setTimeout(() => $el.remove(), 5000)">` + message + `</div>`
	w.Write([]byte(errorHTML))
}