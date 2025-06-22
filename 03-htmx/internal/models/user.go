package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type SearchResult struct {
	Query   string `json:"query"`
	Results []string `json:"results"`
}

type TimeResponse struct {
	CurrentTime string `json:"current_time"`
	Timestamp   int64  `json:"timestamp"`
}