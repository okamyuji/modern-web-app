package models

import (
	"database/sql"
	"errors"
	"time"
)

type Todo struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TodoRepository struct {
	db     *sql.DB
	driver string
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
	// ドライバーを検出
	driver := "sqlite3" // デフォルト
	if db != nil {
		// 接続文字列からドライバーを判定（簡易版）
		driver = detectDriver(db)
	}
	return &TodoRepository{db: db, driver: driver}
}

func NewTodoRepositoryWithDriver(db *sql.DB, driver string) *TodoRepository {
	return &TodoRepository{db: db, driver: driver}
}

func detectDriver(db *sql.DB) string {
	// PRAGMA文でSQLiteかどうかを判定
	_, err := db.Exec("PRAGMA schema_version")
	if err == nil {
		return "sqlite3"
	}
	return "postgres"
}

func (r *TodoRepository) Create(todo *Todo) (int64, error) {
	if todo.Title == "" {
		return 0, errors.New("title is required")
	}

	if r.driver == "postgres" {
		query := `
			INSERT INTO todos (title, description, priority, completed)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, updated_at`

		err := r.db.QueryRow(query, todo.Title, todo.Description, todo.Priority, todo.Completed).
			Scan(&todo.ID, &todo.CreatedAt, &todo.UpdatedAt)
		return todo.ID, err
	} else {
		// SQLite
		query := `
			INSERT INTO todos (title, description, priority, completed)
			VALUES (?, ?, ?, ?)`

		result, err := r.db.Exec(query, todo.Title, todo.Description, todo.Priority, todo.Completed)
		if err != nil {
			return 0, err
		}

		todo.ID, err = result.LastInsertId()
		if err != nil {
			return 0, err
		}

		// 作成されたレコードの時刻を取得
		err = r.db.QueryRow("SELECT created_at, updated_at FROM todos WHERE id = ?", todo.ID).
			Scan(&todo.CreatedAt, &todo.UpdatedAt)

		return todo.ID, err
	}
}

func (r *TodoRepository) GetByID(id int64) (*Todo, error) {
	todo := &Todo{}
	var query string
	
	if r.driver == "postgres" {
		query = `
			SELECT id, title, description, priority, completed, created_at, updated_at
			FROM todos
			WHERE id = $1`
	} else {
		query = `
			SELECT id, title, description, priority, completed, created_at, updated_at
			FROM todos
			WHERE id = ?`
	}

	err := r.db.QueryRow(query, id).Scan(
		&todo.ID, &todo.Title, &todo.Description,
		&todo.Priority, &todo.Completed,
		&todo.CreatedAt, &todo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("todo not found")
	}

	return todo, err
}

func (r *TodoRepository) List(filter string, sortBy string) ([]*Todo, error) {
	query := `
		SELECT id, title, description, priority, completed, created_at, updated_at
		FROM todos`

	var args []interface{}
	var whereClause string

	if filter != "" {
		if r.driver == "postgres" {
			whereClause = " WHERE title LIKE $1 OR description LIKE $1"
		} else {
			whereClause = " WHERE title LIKE ? OR description LIKE ?"
		}
		args = append(args, "%"+filter+"%")
	}

	orderBy := " ORDER BY created_at DESC"
	if sortBy == "priority" {
		orderBy = " ORDER BY CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 END"
	}

	query += whereClause + orderBy

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []*Todo
	for rows.Next() {
		todo := &Todo{}
		err := rows.Scan(
			&todo.ID, &todo.Title, &todo.Description,
			&todo.Priority, &todo.Completed,
			&todo.CreatedAt, &todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, rows.Err()
}

func (r *TodoRepository) Update(todo *Todo) error {
	if todo.Title == "" {
		return errors.New("title is required")
	}

	if r.driver == "postgres" {
		query := `
			UPDATE todos
			SET title = $1, description = $2, priority = $3, completed = $4, updated_at = CURRENT_TIMESTAMP
			WHERE id = $5
			RETURNING updated_at`

		err := r.db.QueryRow(query, todo.Title, todo.Description, todo.Priority, todo.Completed, todo.ID).
			Scan(&todo.UpdatedAt)
		return err
	} else {
		// SQLite
		query := `
			UPDATE todos
			SET title = ?, description = ?, priority = ?, completed = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`

		_, err := r.db.Exec(query, todo.Title, todo.Description, todo.Priority, todo.Completed, todo.ID)
		if err != nil {
			return err
		}

		// 更新された時刻を取得
		err = r.db.QueryRow("SELECT updated_at FROM todos WHERE id = ?", todo.ID).
			Scan(&todo.UpdatedAt)
		return err
	}
}

func (r *TodoRepository) Delete(id int64) error {
	var query string
	if r.driver == "postgres" {
		query = `DELETE FROM todos WHERE id = $1`
	} else {
		query = `DELETE FROM todos WHERE id = ?`
	}
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("todo not found")
	}

	return nil
}

func (r *TodoRepository) InitSchema() error {
	query := `
		CREATE TABLE IF NOT EXISTS todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			priority TEXT DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
			completed BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

	_, err := r.db.Exec(query)
	return err
}