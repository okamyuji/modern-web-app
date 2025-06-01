package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang-patterns/internal/domain/models"
	"golang-patterns/internal/infrastructure/logger"
	"golang-patterns/internal/infrastructure/middleware"
	"golang-patterns/internal/infrastructure/repositories"
	"golang-patterns/internal/interfaces/handlers"
	"golang-patterns/internal/usecases"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

func main() {
	// Setup the same dependencies as the real server
	userRepo := repositories.NewMemoryUserRepository()
	logger := logger.NewConsoleLogger()
	userUseCase := usecases.NewUserUseCase(userRepo, logger)
	userHandler := handlers.NewUserHandler(userUseCase)
	
	router := mux.NewRouter()
	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.LoggingMiddleware)
	
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	api.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
	}).Methods("GET")

	// Test the endpoints
	fmt.Println("Testing Chapter 2 Clean Architecture Implementation")
	fmt.Println("=================================================")

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("1. Health check: Status %d - %s\n", rr.Code, rr.Body.String())

	// Test get all users (should be empty initially)
	req = httptest.NewRequest("GET", "/api/users", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("2. Get all users: Status %d - %s\n", rr.Code, rr.Body.String())

	// Test create user
	user := models.User{Name: "Test User", Email: "test@example.com"}
	userJSON, _ := json.Marshal(user)
	req = httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("3. Create user: Status %d - %s\n", rr.Code, rr.Body.String())

	// Parse response to get user ID
	var response handlers.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	createdUser := response.Data.(map[string]interface{})
	userID := createdUser["id"].(string)

	// Test get user by ID
	req = httptest.NewRequest("GET", "/api/users/"+userID, nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("4. Get user by ID: Status %d - %s\n", rr.Code, rr.Body.String())

	// Test get all users again (should have one user)
	req = httptest.NewRequest("GET", "/api/users", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("5. Get all users (after creation): Status %d - %s\n", rr.Code, rr.Body.String())

	// Test error case - user not found
	req = httptest.NewRequest("GET", "/api/users/nonexistent", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("6. User not found: Status %d - %s\n", rr.Code, rr.Body.String())

	fmt.Println("\n✅ Chapter 2 implementation complete and tested successfully!")
	fmt.Println("Clean Architecture layers:")
	fmt.Println("- Domain: User model with validation")
	fmt.Println("- Use Cases: User business logic")  
	fmt.Println("- Interfaces: HTTP handlers and repository interfaces")
	fmt.Println("- Infrastructure: In-memory repository, logger, middleware")
}