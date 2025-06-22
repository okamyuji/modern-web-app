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
	router.HandleFunc("/targets", pageHandler.Targets).Methods("GET")
	router.HandleFunc("/indicators", pageHandler.Indicators).Methods("GET")
	router.HandleFunc("/forms", pageHandler.Forms).Methods("GET")
	router.HandleFunc("/progressive", pageHandler.Progressive).Methods("GET")

	// API routes for HTMX
	api := router.PathPrefix("/api").Subrouter()
	
	// Basic requests APIs
	api.HandleFunc("/users", apiHandler.GetUsers).Methods("GET")
	api.HandleFunc("/users", apiHandler.CreateUser).Methods("POST")
	api.HandleFunc("/echo", apiHandler.Echo).Methods("PUT", "DELETE")
	api.HandleFunc("/search", apiHandler.Search).Methods("GET", "POST")
	api.HandleFunc("/time", apiHandler.GetTime).Methods("GET")
	api.HandleFunc("/focus-data", apiHandler.FocusData).Methods("GET")
	api.HandleFunc("/special-action", apiHandler.SpecialAction).Methods("POST")
	api.HandleFunc("/custom-response", apiHandler.CustomResponse).Methods("GET")
	
	// Target specification APIs
	api.HandleFunc("/target-content", apiHandler.TargetContent).Methods("GET")
	api.HandleFunc("/multi-target", apiHandler.MultiTarget).Methods("GET")
	api.HandleFunc("/selector-target", apiHandler.SelectorTarget).Methods("GET")
	api.HandleFunc("/relative-target", apiHandler.RelativeTarget).Methods("GET")
	api.HandleFunc("/swap-demo", apiHandler.SwapDemo).Methods("GET")
	
	// Indicators APIs
	api.HandleFunc("/slow-response", apiHandler.SlowResponse).Methods("GET")
	api.HandleFunc("/progress-response", apiHandler.ProgressResponse).Methods("GET")
	api.HandleFunc("/skeleton-response", apiHandler.SkeletonResponse).Methods("GET")
	
	// Forms APIs
	api.HandleFunc("/form-submit", apiHandler.FormSubmit).Methods("POST")
	api.HandleFunc("/validate-username", apiHandler.ValidateUsername).Methods("POST")
	api.HandleFunc("/validate-password", apiHandler.ValidatePassword).Methods("POST")
	api.HandleFunc("/validate-password-confirm", apiHandler.ValidatePasswordConfirm).Methods("POST")
	api.HandleFunc("/validate-form", apiHandler.ValidateForm).Methods("POST")
	api.HandleFunc("/subcategories", apiHandler.Subcategories).Methods("GET")
	api.HandleFunc("/dynamic-form", apiHandler.DynamicForm).Methods("POST")
	api.HandleFunc("/file-upload", apiHandler.FileUpload).Methods("POST")
	api.HandleFunc("/edit-field", apiHandler.EditField).Methods("GET")
	api.HandleFunc("/save-field", apiHandler.SaveField).Methods("POST")
	api.HandleFunc("/cancel-edit", apiHandler.CancelEdit).Methods("GET")
	api.HandleFunc("/bulk-action", apiHandler.BulkAction).Methods("POST")
	
	// Progressive APIs
	api.HandleFunc("/load-more", apiHandler.LoadMore).Methods("GET")
	api.HandleFunc("/live-counter", apiHandler.LiveCounter).Methods("GET")
	api.HandleFunc("/live-stats", apiHandler.LiveStats).Methods("GET")
	api.HandleFunc("/progressive-load", apiHandler.ProgressiveLoad).Methods("GET")
	api.HandleFunc("/lazy-content", apiHandler.LazyContent).Methods("GET")
	api.HandleFunc("/autocomplete", apiHandler.Autocomplete).Methods("GET")
	api.HandleFunc("/select-city", apiHandler.SelectCity).Methods("GET")
	api.HandleFunc("/save-order", apiHandler.SaveOrder).Methods("POST")
	api.HandleFunc("/start-notifications", apiHandler.StartNotifications).Methods("POST")
	api.HandleFunc("/stop-notifications", apiHandler.StopNotifications).Methods("POST")
	api.HandleFunc("/notification-updates", apiHandler.NotificationUpdates).Methods("GET")

	log.Printf("HTMX Demo Server starting on port %s", port)
	log.Printf("Open http://localhost:%s to view the demo", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}