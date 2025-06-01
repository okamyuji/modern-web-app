package handlers

import (
	"context"
	"alpinejs-demo/internal/templates"
	"net/http"
)

// PageHandler handles page requests
type PageHandler struct{}

// NewPageHandler creates a new PageHandler
func NewPageHandler() *PageHandler {
	return &PageHandler{}
}

// Home serves the home page
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	err := templates.Home().Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// BasicState serves the basic state management demo
func (h *PageHandler) BasicState(w http.ResponseWriter, r *http.Request) {
	err := templates.BasicState().Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TodoApp serves the TODO application demo
func (h *PageHandler) TodoApp(w http.ResponseWriter, r *http.Request) {
	err := templates.TodoApp().Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}