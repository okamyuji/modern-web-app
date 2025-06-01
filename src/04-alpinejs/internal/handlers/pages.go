package handlers

import (
	"context"
	"log"
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
	log.Printf("Home page accessed: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.Home().Render(context.Background(), w)
	if err != nil {
		log.Printf("Error rendering home template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Home page rendered successfully")
}

// BasicState serves the basic state management demo
func (h *PageHandler) BasicState(w http.ResponseWriter, r *http.Request) {
	log.Printf("BasicState page accessed: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.BasicState().Render(context.Background(), w)
	if err != nil {
		log.Printf("Error rendering BasicState template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("BasicState page rendered successfully")
}

// TodoApp serves the TODO application demo
func (h *PageHandler) TodoApp(w http.ResponseWriter, r *http.Request) {
	log.Printf("TodoApp page accessed: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.TodoApp().Render(context.Background(), w)
	if err != nil {
		log.Printf("Error rendering TodoApp template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("TodoApp page rendered successfully")
}

// EventHandling serves the event handling demo
func (h *PageHandler) EventHandling(w http.ResponseWriter, r *http.Request) {
	log.Printf("EventHandling page accessed: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.EventHandling().Render(context.Background(), w)
	if err != nil {
		log.Printf("Error rendering EventHandling template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("EventHandling page rendered successfully")
}

// GlobalState serves the global state demo
func (h *PageHandler) GlobalState(w http.ResponseWriter, r *http.Request) {
	log.Printf("GlobalState page accessed: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.GlobalState().Render(context.Background(), w)
	if err != nil {
		log.Printf("Error rendering GlobalState template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("GlobalState page rendered successfully")
}

// HTMXIntegration serves the HTMX integration demo
func (h *PageHandler) HTMXIntegration(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTMXIntegration page accessed: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.HTMXIntegration().Render(context.Background(), w)
	if err != nil {
		log.Printf("Error rendering HTMXIntegration template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("HTMXIntegration page rendered successfully")
}

// AdvancedPatterns serves the advanced patterns demo
func (h *PageHandler) AdvancedPatterns(w http.ResponseWriter, r *http.Request) {
	log.Printf("AdvancedPatterns page accessed: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.AdvancedPatterns().Render(context.Background(), w)
	if err != nil {
		log.Printf("Error rendering AdvancedPatterns template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("AdvancedPatterns page rendered successfully")
}