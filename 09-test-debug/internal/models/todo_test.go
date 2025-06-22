package models

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t testing.TB) (*sql.DB, func()) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// スキーマの作成
	_, err = db.Exec(`
		CREATE TABLE todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			priority TEXT DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
			completed BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestTodoRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTodoRepository(db)

	tests := []struct {
		name    string
		todo    Todo
		wantErr bool
	}{
		{
			name: "正常なTODO作成",
			todo: Todo{
				Title:       "テストタスク",
				Description: "これはテスト用のタスクです",
				Priority:    "high",
				Completed:   false,
			},
			wantErr: false,
		},
		{
			name: "空のタイトル",
			todo: Todo{
				Title: "",
			},
			wantErr: true,
		},
		{
			name: "説明なしでも作成可能",
			todo: Todo{
				Title:    "説明なしタスク",
				Priority: "low",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := repo.Create(&tt.todo)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Greater(t, id, int64(0))

			// 作成されたTODOを確認
			created, err := repo.GetByID(id)
			assert.NoError(t, err)
			assert.Equal(t, tt.todo.Title, created.Title)
			assert.Equal(t, tt.todo.Description, created.Description)
			assert.Equal(t, tt.todo.Priority, created.Priority)
		})
	}
}

func TestTodoRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTodoRepository(db)

	// テストデータの作成
	todo := &Todo{
		Title:       "取得テスト",
		Description: "取得テスト用のタスク",
		Priority:    "medium",
	}

	id, err := repo.Create(todo)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "存在するTODOの取得",
			id:      id,
			wantErr: false,
		},
		{
			name:    "存在しないTODOの取得",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.id, result.ID)
		})
	}
}

func TestTodoRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTodoRepository(db)

	// テストデータの準備
	testTodos := []Todo{
		{Title: "高優先度タスク", Priority: "high", Description: "重要なタスク"},
		{Title: "中優先度タスク", Priority: "medium", Description: "普通のタスク"},
		{Title: "低優先度タスク", Priority: "low", Description: "後回しタスク"},
		{Title: "検索テスト", Priority: "medium", Description: "検索用のキーワード"},
	}

	for _, todo := range testTodos {
		_, err := repo.Create(&todo)
		require.NoError(t, err)
	}

	tests := []struct {
		name         string
		filter       string
		sortBy       string
		expectedMin  int
		shouldContain string
	}{
		{
			name:        "フィルターなし",
			filter:      "",
			sortBy:      "",
			expectedMin: 4,
		},
		{
			name:         "タイトルフィルター",
			filter:       "検索",
			sortBy:       "",
			expectedMin:  1,
			shouldContain: "検索テスト",
		},
		{
			name:        "説明フィルター",
			filter:      "重要",
			sortBy:      "",
			expectedMin: 1,
		},
		{
			name:        "優先度ソート",
			filter:      "",
			sortBy:      "priority",
			expectedMin: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todos, err := repo.List(tt.filter, tt.sortBy)

			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(todos), tt.expectedMin)

			if tt.shouldContain != "" {
				found := false
				for _, todo := range todos {
					if todo.Title == tt.shouldContain {
						found = true
						break
					}
				}
				assert.True(t, found, "期待されるTODOが見つかりませんでした: %s", tt.shouldContain)
			}

			// 優先度ソートの確認
			if tt.sortBy == "priority" && len(todos) > 1 {
				// 最初のアイテムは高優先度であるべき
				assert.Equal(t, "high", todos[0].Priority)
			}
		})
	}
}

func TestTodoRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTodoRepository(db)

	// テストデータの作成
	original := &Todo{
		Title:       "更新前タスク",
		Description: "更新前の説明",
		Priority:    "low",
		Completed:   false,
	}

	id, err := repo.Create(original)
	require.NoError(t, err)

	tests := []struct {
		name    string
		todo    Todo
		wantErr bool
	}{
		{
			name: "正常な更新",
			todo: Todo{
				ID:          id,
				Title:       "更新後タスク",
				Description: "更新後の説明",
				Priority:    "high",
				Completed:   true,
			},
			wantErr: false,
		},
		{
			name: "空のタイトルでの更新",
			todo: Todo{
				ID:    id,
				Title: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(&tt.todo)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// 更新されたデータを確認
			updated, err := repo.GetByID(tt.todo.ID)
			assert.NoError(t, err)
			assert.Equal(t, tt.todo.Title, updated.Title)
			assert.Equal(t, tt.todo.Description, updated.Description)
			assert.Equal(t, tt.todo.Priority, updated.Priority)
			assert.Equal(t, tt.todo.Completed, updated.Completed)
		})
	}
}

func TestTodoRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTodoRepository(db)

	// テストデータの作成
	todo := &Todo{
		Title:    "削除テスト",
		Priority: "medium",
	}

	id, err := repo.Create(todo)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "存在するTODOの削除",
			id:      id,
			wantErr: false,
		},
		{
			name:    "存在しないTODOの削除",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// 削除されたことを確認
			_, err = repo.GetByID(tt.id)
			assert.Error(t, err)
		})
	}
}

// ベンチマークテスト
func BenchmarkTodoRepository_List(b *testing.B) {
	db, cleanup := setupTestDB(b)
	defer cleanup()

	repo := NewTodoRepository(db)

	// テストデータの準備
	for i := 0; i < 1000; i++ {
		_, err := repo.Create(&Todo{
			Title:       fmt.Sprintf("Task %d", i),
			Description: fmt.Sprintf("Description for task %d", i),
			Priority:    []string{"low", "medium", "high"}[i%3],
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.List("", "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTodoRepository_Create(b *testing.B) {
	db, cleanup := setupTestDB(b)
	defer cleanup()

	repo := NewTodoRepository(db)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		todo := &Todo{
			Title:       fmt.Sprintf("Benchmark Task %d", i),
			Description: "Benchmark task description",
			Priority:    "medium",
		}
		_, err := repo.Create(todo)
		if err != nil {
			b.Fatal(err)
		}
	}
}