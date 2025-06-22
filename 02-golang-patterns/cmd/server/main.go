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

	// === IMPORTANT: Specific routes MUST be defined BEFORE generic {id} routes ===

	// === Target specification - Advanced query operations ===
	api.HandleFunc("/users/email/{email}", userHandler.GetUserByEmail).Methods("GET")
	api.HandleFunc("/users/department/{department}", userHandler.GetUsersByDepartment).Methods("GET")
	api.HandleFunc("/users/position/{position}", userHandler.GetUsersByPosition).Methods("GET")
	api.HandleFunc("/users/active", userHandler.GetActiveUsers).Methods("GET")
	api.HandleFunc("/users/inactive", userHandler.GetInactiveUsers).Methods("GET")

	// === Load display - Pagination and sorting ===
	api.HandleFunc("/users/paginated", userHandler.GetUsersWithPagination).Methods("GET")
	api.HandleFunc("/users/search", userHandler.SearchUsers).Methods("GET")

	// === Progressive loading ===
	api.HandleFunc("/users/batch", userHandler.GetUsersBatch).Methods("GET")

	// === Statistics and analytics ===
	api.HandleFunc("/users/stats", userHandler.GetUserStats).Methods("GET")
	api.HandleFunc("/users/stats/departments", userHandler.GetDepartmentStats).Methods("GET")
	api.HandleFunc("/users/recent-signups", userHandler.GetRecentSignups).Methods("GET")

	// === Form processing - Bulk operations ===
	api.HandleFunc("/users/bulk", userHandler.CreateUsersInBulk).Methods("POST")
	api.HandleFunc("/users/bulk", userHandler.UpdateUsersInBulk).Methods("PUT")
	api.HandleFunc("/users/bulk", userHandler.DeleteUsersInBulk).Methods("DELETE")

	// === Progressive enhancement features (specific ID operations) ===
	api.HandleFunc("/users/{id}/activate", userHandler.ActivateUser).Methods("POST")
	api.HandleFunc("/users/{id}/deactivate", userHandler.DeactivateUser).Methods("POST")
	api.HandleFunc("/users/{id}/login", userHandler.UpdateLastLogin).Methods("POST")
	api.HandleFunc("/users/{id}/summary", userHandler.GetUserSummary).Methods("GET")

	// === Basic CRUD operations (generic {id} routes MUST be LAST) ===
	api.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	api.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "ok", "version": "enhanced"})
	}).Methods("GET")

	log.Printf("Enhanced server starting on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  Basic CRUD:")
	log.Printf("    POST   /api/users                    - Create user")
	log.Printf("    GET    /api/users                    - Get all users")
	log.Printf("    GET    /api/users/{id}               - Get user by ID")
	log.Printf("    PUT    /api/users/{id}               - Update user")
	log.Printf("    DELETE /api/users/{id}               - Delete user")
	log.Printf("  Target Specification:")
	log.Printf("    GET    /api/users/email/{email}      - Get user by email")
	log.Printf("    GET    /api/users/department/{dept}  - Get users by department")
	log.Printf("    GET    /api/users/position/{pos}     - Get users by position")
	log.Printf("    GET    /api/users/active             - Get active users")
	log.Printf("    GET    /api/users/inactive           - Get inactive users")
	log.Printf("  Load Display:")
	log.Printf("    GET    /api/users/paginated          - Paginated users with filtering")
	log.Printf("    GET    /api/users/search             - Search users")
	log.Printf("  Progressive Loading:")
	log.Printf("    GET    /api/users/batch              - Batch loading with cursor")
	log.Printf("  Statistics:")
	log.Printf("    GET    /api/users/stats              - User statistics")
	log.Printf("    GET    /api/users/stats/departments  - Department statistics")
	log.Printf("    GET    /api/users/recent-signups     - Recent signups")
	log.Printf("  Form Processing:")
	log.Printf("    POST   /api/users/bulk               - Bulk create users")
	log.Printf("    PUT    /api/users/bulk               - Bulk update users")
	log.Printf("    DELETE /api/users/bulk               - Bulk delete users")
	log.Printf("  Progressive Enhancement:")
	log.Printf("    POST   /api/users/{id}/activate      - Activate user")
	log.Printf("    POST   /api/users/{id}/deactivate    - Deactivate user")
	log.Printf("    POST   /api/users/{id}/login         - Update last login")
	log.Printf("    GET    /api/users/{id}/summary       - User summary")

	log.Fatal(http.ListenAndServe(":"+port, router))
}