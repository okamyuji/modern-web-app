package models

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	Priority    string     `json:"priority"` // low, medium, high
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TodoRepository struct {
	db *sql.DB
}

func NewTodoRepository(dbPath string) (*TodoRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &TodoRepository{db: db}
	if err := repo.InitSchema(); err != nil {
		return nil, err
	}

	return repo, nil
}

// データベース初期化
func (r *TodoRepository) InitSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		completed BOOLEAN DEFAULT FALSE,
		priority TEXT DEFAULT 'medium',
		due_date DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TRIGGER IF NOT EXISTS update_todos_updated_at 
	AFTER UPDATE ON todos
	BEGIN
		UPDATE todos SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;
	`
	_, err := r.db.Exec(query)
	return err
}

// 検索とフィルタリング
func (r *TodoRepository) List(filter string, search string) ([]Todo, error) {
	query := `
	SELECT id, title, description, completed, priority, due_date, created_at, updated_at 
	FROM todos 
	WHERE 1=1
	`
	args := []interface{}{}

	// フィルター条件の追加
	switch filter {
	case "active":
		query += " AND completed = FALSE"
	case "completed":
		query += " AND completed = TRUE"
	case "overdue":
		query += " AND due_date < datetime('now') AND completed = FALSE"
	}

	// 検索条件の追加
	if search != "" {
		query += " AND (title LIKE ? OR description LIKE ?)"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query += " ORDER BY completed ASC, priority DESC, created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(
			&todo.ID, &todo.Title, &todo.Description,
			&todo.Completed, &todo.Priority, &todo.DueDate,
			&todo.CreatedAt, &todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

// TODO作成
func (r *TodoRepository) Create(todo *Todo) (int, error) {
	query := `
	INSERT INTO todos (title, description, priority, due_date)
	VALUES (?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, todo.Title, todo.Description, todo.Priority, todo.DueDate)
	if err != nil {
		return 0, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	
	return int(id), nil
}

// TODO更新
func (r *TodoRepository) Update(todo *Todo) error {
	query := `
	UPDATE todos 
	SET title = ?, description = ?, completed = ?, priority = ?, due_date = ?
	WHERE id = ?
	`
	
	_, err := r.db.Exec(query, todo.Title, todo.Description, todo.Completed, todo.Priority, todo.DueDate, todo.ID)
	return err
}

// TODO削除
func (r *TodoRepository) Delete(id int) error {
	query := "DELETE FROM todos WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// IDでTODO取得
func (r *TodoRepository) GetByID(id int) (*Todo, error) {
	query := `
	SELECT id, title, description, completed, priority, due_date, created_at, updated_at 
	FROM todos 
	WHERE id = ?
	`
	
	var todo Todo
	err := r.db.QueryRow(query, id).Scan(
		&todo.ID, &todo.Title, &todo.Description,
		&todo.Completed, &todo.Priority, &todo.DueDate,
		&todo.CreatedAt, &todo.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &todo, nil
}

// 完了状態の切り替え
func (r *TodoRepository) ToggleCompleted(id int) error {
	query := "UPDATE todos SET completed = NOT completed WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// 統計情報取得
func (r *TodoRepository) GetStats() (map[string]int, error) {
	query := `
	SELECT 
		COUNT(*) as total,
		COUNT(CASE WHEN completed = FALSE THEN 1 END) as active,
		COUNT(CASE WHEN completed = TRUE THEN 1 END) as completed,
		COUNT(CASE WHEN due_date < datetime('now') AND completed = FALSE THEN 1 END) as overdue
	FROM todos
	`
	
	var total, active, completed, overdue int
	err := r.db.QueryRow(query).Scan(&total, &active, &completed, &overdue)
	if err != nil {
		return nil, err
	}
	
	return map[string]int{
		"total":     total,
		"active":    active,
		"completed": completed,
		"overdue":   overdue,
	}, nil
}

// データベースクローズ
func (r *TodoRepository) Close() error {
	return r.db.Close()
}