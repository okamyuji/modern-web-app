package main

import (
	"htmx-demo/internal/handlers"
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
	apiHandler := handlers.NewAPIHandler()

	// Setup router
	router := mux.NewRouter()

	// Page routes
	router.HandleFunc("/", pageHandler.Home).Methods("GET")
	router.HandleFunc("/basic-requests", pageHandler.BasicRequests).Methods("GET")
	router.HandleFunc("/triggers", pageHandler.Triggers).Methods("GET")

	// API routes for HTMX
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/users", apiHandler.GetUsers).Methods("GET")
	api.HandleFunc("/users", apiHandler.CreateUser).Methods("POST")
	api.HandleFunc("/echo", apiHandler.Echo).Methods("PUT", "DELETE")
	api.HandleFunc("/search", apiHandler.Search).Methods("GET", "POST")
	api.HandleFunc("/time", apiHandler.GetTime).Methods("GET")
	api.HandleFunc("/focus-data", apiHandler.FocusData).Methods("GET")
	api.HandleFunc("/special-action", apiHandler.SpecialAction).Methods("POST")
	api.HandleFunc("/custom-response", apiHandler.CustomResponse).Methods("GET")

	log.Printf("HTMX Demo Server starting on port %s", port)
	log.Printf("Open http://localhost:%s to view the demo", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}