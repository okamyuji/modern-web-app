package models

import (
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // text, image, system
	CreatedAt time.Time `json:"created_at"`
}

type Hub struct {
	clients    map[string]*Client
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
	repo       *MessageRepository
}

type Client struct {
	ID       string
	Username string
	Send     chan *Message
	Done     chan bool
}

func NewHub(repo *MessageRepository) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		Broadcast:  make(chan *Message, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		repo:       repo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()

			// システムメッセージを送信
			message := &Message{
				ID:        uuid.New().String(),
				Type:      "system",
				Content:   client.Username + " が参加しました",
				CreatedAt: time.Now(),
			}
			h.repo.SaveMessage(message)
			h.broadcastMessage(message)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				h.mu.Unlock()

				message := &Message{
					ID:        uuid.New().String(),
					Type:      "system",
					Content:   client.Username + " が退出しました",
					CreatedAt: time.Now(),
				}
				h.repo.SaveMessage(message)
				h.broadcastMessage(message)
			} else {
				h.mu.Unlock()
			}

		case message := <-h.Broadcast:
			// メッセージを保存
			h.repo.SaveMessage(message)
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for id, client := range h.clients {
		select {
		case client.Send <- message:
			// 送信成功
		default:
			// バッファが満杯の場合、クライアントを削除
			close(client.Send)
			delete(h.clients, id)
		}
	}
}

func (h *Hub) GetConnectedClients() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []string
	for _, client := range h.clients {
		clients = append(clients, client.Username)
	}
	return clients
}

func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// MessageRepository - メッセージの永続化
type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(dbPath string) (*MessageRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &MessageRepository{db: db}
	if err := repo.InitSchema(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *MessageRepository) InitSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		user_id TEXT,
		username TEXT NOT NULL,
		content TEXT NOT NULL,
		type TEXT DEFAULT 'text',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
	CREATE INDEX IF NOT EXISTS idx_messages_type ON messages(type);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *MessageRepository) SaveMessage(msg *Message) error {
	query := `
	INSERT INTO messages (id, user_id, username, content, type, created_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, msg.ID, msg.UserID, msg.Username, msg.Content, msg.Type, msg.CreatedAt)
	return err
}

func (r *MessageRepository) GetRecentMessages(limit int) ([]Message, error) {
	query := `
	SELECT id, user_id, username, content, type, created_at
	FROM messages
	ORDER BY created_at DESC
	LIMIT ?
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var userID sql.NullString
		err := rows.Scan(&msg.ID, &userID, &msg.Username, &msg.Content, &msg.Type, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		if userID.Valid {
			msg.UserID = userID.String
		}
		messages = append(messages, msg)
	}

	// 時系列順に並び替え
	for i := len(messages)/2 - 1; i >= 0; i-- {
		opp := len(messages) - 1 - i
		messages[i], messages[opp] = messages[opp], messages[i]
	}

	return messages, nil
}

func (r *MessageRepository) GetMessagesByTimeRange(start, end time.Time) ([]Message, error) {
	query := `
	SELECT id, user_id, username, content, type, created_at
	FROM messages
	WHERE created_at BETWEEN ? AND ?
	ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var userID sql.NullString
		err := rows.Scan(&msg.ID, &userID, &msg.Username, &msg.Content, &msg.Type, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		if userID.Valid {
			msg.UserID = userID.String
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *MessageRepository) DeleteOldMessages(days int) error {
	query := `DELETE FROM messages WHERE created_at < datetime('now', '-' || ? || ' days')`
	_, err := r.db.Exec(query, days)
	return err
}

func (r *MessageRepository) GetMessageCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&count)
	return count, err
}

func (r *MessageRepository) Close() error {
	return r.db.Close()
}