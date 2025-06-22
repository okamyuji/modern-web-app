package main

import (
	"alpinejs-demo/internal/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize handlers
	pageHandler := handlers.NewPageHandler()

	// Setup router
	router := mux.NewRouter()

	// Page routes
	router.HandleFunc("/", pageHandler.Home).Methods("GET")
	router.HandleFunc("/basic-state", pageHandler.BasicState).Methods("GET")
	router.HandleFunc("/todo-app", pageHandler.TodoApp).Methods("GET")
	router.HandleFunc("/event-handling", pageHandler.EventHandling).Methods("GET")
	router.HandleFunc("/global-state", pageHandler.GlobalState).Methods("GET")
	router.HandleFunc("/htmx-integration", pageHandler.HTMXIntegration).Methods("GET")
	router.HandleFunc("/advanced-patterns", pageHandler.AdvancedPatterns).Methods("GET")

	log.Printf("Alpine.js Demo Server starting on port %s", port)
	log.Printf("Open http://localhost:%s to view the demo", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}