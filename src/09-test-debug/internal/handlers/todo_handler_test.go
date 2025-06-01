package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"test-debug-demo/internal/logger"
	"test-debug-demo/internal/models"
)

func setupTestHandler(t testing.TB) (*TodoHandler, *sql.DB, func()) {
	// テスト用データベース
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// スキーマ作成
	repo := models.NewTodoRepository(db)
	err = repo.InitSchema()
	require.NoError(t, err)

	// ロガー
	log := logger.NewLogger(&bytes.Buffer{}, logger.INFO)

	handler := NewTodoHandler(repo, log)

	cleanup := func() {
		db.Close()
	}

	return handler, db, cleanup
}

func TestTodoHandler_Create_HTMX(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	tests := []struct {
		name           string
		formData       url.Values
		headers        map[string]string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "HTMXリクエストでの正常な作成",
			formData: url.Values{
				"title":       {"新しいタスク"},
				"description": {"タスクの説明"},
				"priority":    {"high"},
			},
			headers: map[string]string{
				"HX-Request": "true",
				"HX-Target":  "todo-list",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "新しいタスク")
				assert.Contains(t, body, "data-todo-id")
				assert.Contains(t, body, "bg-red-50") // high priority
			},
		},
		{
			name: "通常のフォーム送信",
			formData: url.Values{
				"title":    {"通常のタスク"},
				"priority": {"medium"},
			},
			headers:        map[string]string{},
			expectedStatus: http.StatusSeeOther, // リダイレクト
		},
		{
			name: "バリデーションエラー - 空のタイトル",
			formData: url.Values{
				"title": {""},
			},
			headers: map[string]string{
				"HX-Request": "true",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "タイトルは必須です")
				assert.Contains(t, body, "bg-red-50") // エラー表示
			},
		},
		{
			name: "バリデーションエラー - 長すぎるタイトル",
			formData: url.Values{
				"title": {strings.Repeat("a", 101)},
			},
			headers: map[string]string{
				"HX-Request": "true",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "100文字以内")
			},
		},
		{
			name: "無効な優先度",
			formData: url.Values{
				"title":    {"テストタスク"},
				"priority": {"invalid"},
			},
			headers: map[string]string{
				"HX-Request": "true",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "low, medium, high")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// リクエストの作成
			req := httptest.NewRequest(
				http.MethodPost,
				"/todos",
				strings.NewReader(tt.formData.Encode()),
			)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// ヘッダーの設定
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// レスポンスの記録
			rec := httptest.NewRecorder()

			// ハンドラーの実行
			handler.Create(rec, req)

			// ステータスコードの確認
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// レスポンスボディの確認
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.String())
			}

			// HTMXトリガーの確認
			if tt.headers["HX-Request"] == "true" && tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "todoCreated", rec.Header().Get("HX-Trigger"))
			}
		})
	}
}

func TestTodoHandler_List(t *testing.T) {
	handler, db, cleanup := setupTestHandler(t)
	defer cleanup()

	repo := models.NewTodoRepository(db)

	// テストデータの準備
	testTodos := []models.Todo{
		{Title: "高優先度タスク", Priority: "high", Description: "重要なタスク"},
		{Title: "中優先度タスク", Priority: "medium", Description: "普通のタスク"},
		{Title: "検索テスト", Priority: "low", Description: "検索用のキーワード"},
	}

	for _, todo := range testTodos {
		_, err := repo.Create(&todo)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		query          string
		headers        map[string]string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "全件取得",
			query:          "",
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				// JSON レスポンスの確認
				assert.Contains(t, body, "高優先度タスク")
				assert.Contains(t, body, "中優先度タスク")
				assert.Contains(t, body, "検索テスト")
			},
		},
		{
			name:  "HTMXリクエスト",
			query: "",
			headers: map[string]string{
				"HX-Request": "true",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				// HTML レスポンスの確認
				assert.Contains(t, body, "data-todo-id")
				assert.Contains(t, body, "高優先度タスク")
			},
		},
		{
			name:  "フィルター検索",
			query: "?filter=検索",
			headers: map[string]string{
				"HX-Request": "true",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "検索テスト")
				assert.NotContains(t, body, "高優先度タスク")
			},
		},
		{
			name:  "優先度ソート",
			query: "?sort=priority",
			headers: map[string]string{
				"HX-Request": "true",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				// 高優先度が最初に来ることを確認
				pos1 := strings.Index(body, "高優先度タスク")
				pos2 := strings.Index(body, "中優先度タスク")
				assert.True(t, pos1 < pos2, "高優先度タスクが最初に表示されるべき")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/todos"+tt.query, nil)

			// ヘッダーの設定
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			handler.List(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.String())
			}
		})
	}
}

func TestTodoHandler_Update(t *testing.T) {
	handler, db, cleanup := setupTestHandler(t)
	defer cleanup()

	repo := models.NewTodoRepository(db)

	// テストデータの作成
	original := &models.Todo{
		Title:       "更新前タスク",
		Description: "更新前の説明",
		Priority:    "low",
		Completed:   false,
	}

	_, err := repo.Create(original)
	require.NoError(t, err)

	tests := []struct {
		name           string
		id             string
		formData       url.Values
		headers        map[string]string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "正常な更新",
			id:   "1",
			formData: url.Values{
				"title":       {"更新後タスク"},
				"description": {"更新後の説明"},
				"priority":    {"high"},
				"completed":   {"true"},
			},
			headers: map[string]string{
				"HX-Request": "true",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "更新後タスク")
				assert.Contains(t, body, "bg-red-50") // high priority
			},
		},
		{
			name: "存在しないTODO",
			id:   "9999",
			formData: url.Values{
				"title": {"更新タスク"},
			},
			headers:        map[string]string{},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "無効なID",
			id:   "invalid",
			formData: url.Values{
				"title": {"更新タスク"},
			},
			headers:        map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPut,
				"/todos/"+tt.id,
				strings.NewReader(tt.formData.Encode()),
			)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// ヘッダーの設定
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// mux.Vars をシミュレート
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})

			rec := httptest.NewRecorder()
			handler.Update(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.String())
			}

			// HTMXトリガーの確認
			if tt.headers["HX-Request"] == "true" && tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "todoUpdated", rec.Header().Get("HX-Trigger"))
			}
		})
	}
}

func TestTodoHandler_Delete(t *testing.T) {
	handler, db, cleanup := setupTestHandler(t)
	defer cleanup()

	repo := models.NewTodoRepository(db)

	// テストデータの作成
	todo := &models.Todo{
		Title:    "削除テスト",
		Priority: "medium",
	}

	_, err := repo.Create(todo)
	require.NoError(t, err)

	tests := []struct {
		name           string
		id             string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name: "HTMXでの削除",
			id:   "1",
			headers: map[string]string{
				"HX-Request": "true",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "通常の削除",
			id:             "2", // 存在しないが、このテストではステータスコードのみ確認
			headers:        map[string]string{},
			expectedStatus: http.StatusInternalServerError, // 存在しないので失敗
		},
		{
			name:           "無効なID",
			id:             "invalid",
			headers:        map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/todos/"+tt.id, nil)

			// ヘッダーの設定
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// mux.Vars をシミュレート
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})

			rec := httptest.NewRecorder()
			handler.Delete(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			// HTMXトリガーの確認
			if tt.headers["HX-Request"] == "true" && tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "todoDeleted", rec.Header().Get("HX-Trigger"))
			}
		})
	}
}

func TestTodoHandler_ToggleComplete(t *testing.T) {
	handler, db, cleanup := setupTestHandler(t)
	defer cleanup()

	repo := models.NewTodoRepository(db)

	// テストデータの作成
	todo := &models.Todo{
		Title:     "完了切り替えテスト",
		Priority:  "medium",
		Completed: false,
	}

	_, err := repo.Create(todo)
	require.NoError(t, err)

	tests := []struct {
		name           string
		id             string
		headers        map[string]string
		expectedStatus int
		checkCompleted bool
	}{
		{
			name: "HTMXでの完了切り替え",
			id:   "1",
			headers: map[string]string{
				"HX-Request": "true",
			},
			expectedStatus: http.StatusOK,
			checkCompleted: true,
		},
		{
			name:           "存在しないTODO",
			id:             "9999",
			headers:        map[string]string{},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/todos/"+tt.id+"/toggle", nil)

			// ヘッダーの設定
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// mux.Vars をシミュレート
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})

			rec := httptest.NewRecorder()
			handler.ToggleComplete(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			// 完了状態の確認
			if tt.expectedStatus == http.StatusOK && tt.checkCompleted {
				// データベースから確認
				updated, err := repo.GetByID(1)
				assert.NoError(t, err)
				assert.True(t, updated.Completed, "TODOが完了状態になっているべき")
			}

			// HTMXトリガーの確認
			if tt.headers["HX-Request"] == "true" && tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "todoToggled", rec.Header().Get("HX-Trigger"))
			}
		})
	}
}

// E2Eテストのサンプル
func TestTodoFlow_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	// テストサーバーの設定
	router := mux.NewRouter()
	router.HandleFunc("/", handler.Home).Methods("GET")
	router.HandleFunc("/todos", handler.Create).Methods("POST")
	router.HandleFunc("/todos", handler.List).Methods("GET")

	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 1. ホームページの取得
	resp, err := client.Get(server.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 2. 新しいTODOの作成（HTMX）
	form := url.Values{
		"title":    {"E2Eテストタスク"},
		"priority": {"medium"},
	}

	req, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/todos",
		strings.NewReader(form.Encode()),
	)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// レスポンスの確認
	var body bytes.Buffer
	body.ReadFrom(resp.Body)
	assert.Contains(t, body.String(), "E2Eテストタスク")

	// 3. TODOリストの取得
	req, err = http.NewRequest(http.MethodGet, server.URL+"/todos", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body.Reset()
	body.ReadFrom(resp.Body)
	assert.Contains(t, body.String(), "E2Eテストタスク")
}

// モックのトレースIDコンテキストを作成
func createTestContext() context.Context {
	return context.WithValue(context.Background(), "trace_id", "test-trace-123")
}