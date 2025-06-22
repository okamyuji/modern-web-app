package handlers

import (
	"context"
	"htmx-demo/internal/templates"
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

// BasicRequests serves the basic requests demo page
func (h *PageHandler) BasicRequests(w http.ResponseWriter, r *http.Request) {
	err := templates.BasicRequests().Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Triggers serves the triggers demo page
func (h *PageHandler) Triggers(w http.ResponseWriter, r *http.Request) {
	err := templates.Triggers().Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Targets serves the targets demo page
func (h *PageHandler) Targets(w http.ResponseWriter, r *http.Request) {
	err := templates.Targets().Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Indicators serves the indicators demo page
func (h *PageHandler) Indicators(w http.ResponseWriter, r *http.Request) {
	err := templates.Indicators().Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Forms serves the forms demo page
func (h *PageHandler) Forms(w http.ResponseWriter, r *http.Request) {
	err := templates.Forms().Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Progressive serves the progressive demo page
func (h *PageHandler) Progressive(w http.ResponseWriter, r *http.Request) {
	err := templates.Progressive().Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}