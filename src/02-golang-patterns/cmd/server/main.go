package main

import (
	"golang-patterns/internal/infrastructure/logger"
	"golang-patterns/internal/infrastructure/middleware"
	"golang-patterns/internal/infrastructure/repositories"
	"golang-patterns/internal/interfaces/handlers"
	"golang-patterns/internal/usecases"
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

	// Initialize dependencies following clean architecture
	// Infrastructure layer
	userRepo := repositories.NewMemoryUserRepository()
	logger := logger.NewConsoleLogger()
	
	// Use case layer
	userUseCase := usecases.NewUserUseCase(userRepo, logger)
	
	// Interface layer (handlers)
	userHandler := handlers.NewUserHandler(userUseCase)
	
	// Setup routes
	router := mux.NewRouter()
	
	// Apply middleware
	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.LoggingMiddleware)
	
	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	api.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	
	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
	}).Methods("GET")
	
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}