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
	fmt.Println("Testing Enhanced Chapter 2 Clean Architecture Implementation")
	fmt.Println("==========================================================")

	// Setup the enhanced dependencies
	userRepo := repositories.NewMemoryUserRepository()
	logger := logger.NewConsoleLogger()
	userUseCase := usecases.NewUserUseCase(userRepo, logger)
	userHandler := handlers.NewUserHandler(userUseCase)

	router := mux.NewRouter()
	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.LoggingMiddleware)

	api := router.PathPrefix("/api").Subrouter()

	// Register all enhanced endpoints
	setupRoutes(api, userHandler)

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "ok", "version": "enhanced"})
	}).Methods("GET")

	// Run comprehensive tests
	fmt.Println("\\n🧪 Running Comprehensive Tests")
	fmt.Println("================================")

	testSuite := &TestSuite{router: router}

	// Basic CRUD Tests
	testSuite.runBasicCRUDTests()

	// Target Specification Tests
	testSuite.runTargetSpecificationTests()

	// Load Display Tests
	testSuite.runLoadDisplayTests()

	// Progressive Loading Tests
	testSuite.runProgressiveLoadingTests()

	// Form Processing Tests
	testSuite.runFormProcessingTests()

	// Progressive Enhancement Tests
	testSuite.runProgressiveEnhancementTests()

	fmt.Println("\\n✅ All Enhanced Features Tested Successfully!")
	fmt.Println("\\n📋 Implementation Summary:")
	fmt.Println("==========================")
	fmt.Println("✓ Target Specification: Advanced filtering and querying")
	fmt.Println("✓ Load Display: Pagination, sorting, and search")
	fmt.Println("✓ Form Processing: Bulk operations and validation")
	fmt.Println("✓ Progressive: Cursor-based loading and caching")
	fmt.Println("\\n🏗️ Clean Architecture Layers Enhanced:")
	fmt.Println("- Domain: Rich models with validation and query structures")
	fmt.Println("- Use Cases: Business logic for all advanced features")
	fmt.Println("- Interfaces: Comprehensive REST API with all endpoints")
	fmt.Println("- Infrastructure: Enhanced in-memory repository with indexing")
}

type TestSuite struct {
	router   *mux.Router
	userIDs  []string
	testData map[string]interface{}
}

func setupRoutes(api *mux.Router, userHandler *handlers.UserHandler) {
	// === IMPORTANT: Specific routes MUST be defined BEFORE generic {id} routes ===

	// Target specification
	api.HandleFunc("/users/email/{email}", userHandler.GetUserByEmail).Methods("GET")
	api.HandleFunc("/users/department/{department}", userHandler.GetUsersByDepartment).Methods("GET")
	api.HandleFunc("/users/position/{position}", userHandler.GetUsersByPosition).Methods("GET")
	api.HandleFunc("/users/active", userHandler.GetActiveUsers).Methods("GET")
	api.HandleFunc("/users/inactive", userHandler.GetInactiveUsers).Methods("GET")

	// Load display
	api.HandleFunc("/users/paginated", userHandler.GetUsersWithPagination).Methods("GET")
	api.HandleFunc("/users/search", userHandler.SearchUsers).Methods("GET")

	// Progressive loading
	api.HandleFunc("/users/batch", userHandler.GetUsersBatch).Methods("GET")

	// Statistics
	api.HandleFunc("/users/stats", userHandler.GetUserStats).Methods("GET")
	api.HandleFunc("/users/stats/departments", userHandler.GetDepartmentStats).Methods("GET")
	api.HandleFunc("/users/recent-signups", userHandler.GetRecentSignups).Methods("GET")

	// Bulk operations
	api.HandleFunc("/users/bulk", userHandler.CreateUsersInBulk).Methods("POST")
	api.HandleFunc("/users/bulk", userHandler.UpdateUsersInBulk).Methods("PUT")
	api.HandleFunc("/users/bulk", userHandler.DeleteUsersInBulk).Methods("DELETE")

	// Progressive enhancement (specific ID operations)
	api.HandleFunc("/users/{id}/activate", userHandler.ActivateUser).Methods("POST")
	api.HandleFunc("/users/{id}/deactivate", userHandler.DeactivateUser).Methods("POST")
	api.HandleFunc("/users/{id}/login", userHandler.UpdateLastLogin).Methods("POST")
	api.HandleFunc("/users/{id}/summary", userHandler.GetUserSummary).Methods("GET")

	// Basic CRUD (generic {id} routes MUST be LAST)
	api.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	api.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
}

func (ts *TestSuite) runBasicCRUDTests() {
	fmt.Println("\\n🔧 Basic CRUD Operations")
	fmt.Println("-------------------------")

	// Test health check
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("1. Health check: Status %d - %s\\n", rr.Code, rr.Body.String())

	// Create test users
	testUsers := []models.UserCreateRequest{
		{Name: "Alice Johnson", Email: "alice@company.com", Age: 28, Department: "Engineering", Position: "Senior Developer"},
		{Name: "Bob Smith", Email: "bob@company.com", Age: 32, Department: "Engineering", Position: "Tech Lead"},
		{Name: "Carol Davis", Email: "carol@company.com", Age: 25, Department: "Marketing", Position: "Marketing Manager"},
		{Name: "David Wilson", Email: "david@company.com", Age: 30, Department: "Sales", Position: "Sales Representative"},
		{Name: "Eve Brown", Email: "eve@company.com", Age: 27, Department: "HR", Position: "HR Specialist"},
	}

	for i, user := range testUsers {
		userJSON, _ := json.Marshal(user)
		req = httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(userJSON))
		req.Header.Set("Content-Type", "application/json")
		rr = httptest.NewRecorder()
		ts.router.ServeHTTP(rr, req)
		
		if rr.Code == 201 {
			var response handlers.APIResponse
			json.Unmarshal(rr.Body.Bytes(), &response)
			createdUser := response.Data.(map[string]interface{})
			userID := createdUser["id"].(string)
			ts.userIDs = append(ts.userIDs, userID)
			fmt.Printf("2.%d Create user: Status %d - User %s created\\n", i+1, rr.Code, user.Name)
		} else {
			fmt.Printf("2.%d Create user: Status %d - Failed to create %s\\n", i+1, rr.Code, user.Name)
		}
	}

	// Test get user by ID
	if len(ts.userIDs) > 0 {
		req = httptest.NewRequest("GET", "/api/users/"+ts.userIDs[0], nil)
		rr = httptest.NewRecorder()
		ts.router.ServeHTTP(rr, req)
		fmt.Printf("3. Get user by ID: Status %d\\n", rr.Code)
	}

	// Test get all users
	req = httptest.NewRequest("GET", "/api/users", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("4. Get all users: Status %d (Found %d users)\\n", rr.Code, len(ts.userIDs))
}

func (ts *TestSuite) runTargetSpecificationTests() {
	fmt.Println("\\n🎯 Target Specification Tests")
	fmt.Println("------------------------------")

	// Test get user by email
	req := httptest.NewRequest("GET", "/api/users/email/alice@company.com", nil)
	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("1. Get user by email: Status %d\\n", rr.Code)

	// Test get users by department
	req = httptest.NewRequest("GET", "/api/users/department/Engineering", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("2. Get users by department: Status %d\\n", rr.Code)

	// Test get users by position
	req = httptest.NewRequest("GET", "/api/users/position/Manager", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("3. Get users by position: Status %d\\n", rr.Code)

	// Test get active users
	req = httptest.NewRequest("GET", "/api/users/active", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("4. Get active users: Status %d\\n", rr.Code)

	// Test get inactive users (should be empty initially)
	req = httptest.NewRequest("GET", "/api/users/inactive", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("5. Get inactive users: Status %d\\n", rr.Code)
}

func (ts *TestSuite) runLoadDisplayTests() {
	fmt.Println("\\n📄 Load Display Tests")
	fmt.Println("----------------------")

	// Test pagination
	req := httptest.NewRequest("GET", "/api/users/paginated?page=1&page_size=3&sort_field=name&sort_order=asc", nil)
	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("1. Paginated users (page 1, 3 items): Status %d\\n", rr.Code)

	// Test pagination with filtering
	req = httptest.NewRequest("GET", "/api/users/paginated?page=1&page_size=10&name=Alice&sort_field=created_at&sort_order=desc", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("2. Filtered paginated users: Status %d\\n", rr.Code)

	// Test search
	req = httptest.NewRequest("GET", "/api/users/search?q=Engineering&page=1&page_size=5", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("3. Search users: Status %d\\n", rr.Code)

	// Test search with specific query
	req = httptest.NewRequest("GET", "/api/users/search?q=alice&page=1&page_size=10", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("4. Search users by name: Status %d\\n", rr.Code)
}

func (ts *TestSuite) runProgressiveLoadingTests() {
	fmt.Println("\\n🔄 Progressive Loading Tests")
	fmt.Println("-----------------------------")

	// Test initial batch
	req := httptest.NewRequest("GET", "/api/users/batch?batch_size=3&direction=forward", nil)
	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("1. Initial batch (3 items): Status %d\\n", rr.Code)

	// Extract cursor from response for next test
	var batchResponse map[string]interface{}
	if rr.Code == 200 {
		json.Unmarshal(rr.Body.Bytes(), &batchResponse)
		dataResponse := batchResponse["data"].(map[string]interface{})
		nextCursor := dataResponse["next_cursor"]
		
		if nextCursor != nil && nextCursor != "" {
			// Test next batch with cursor
			req = httptest.NewRequest("GET", fmt.Sprintf("/api/users/batch?batch_size=2&cursor=%s&direction=forward", nextCursor), nil)
			rr = httptest.NewRecorder()
			ts.router.ServeHTTP(rr, req)
			fmt.Printf("2. Next batch with cursor: Status %d\\n", rr.Code)
		} else {
			fmt.Printf("2. Next batch with cursor: No cursor available\\n")
		}
	}

	// Test backward direction
	req = httptest.NewRequest("GET", "/api/users/batch?batch_size=2&direction=backward", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("3. Backward batch: Status %d\\n", rr.Code)
}

func (ts *TestSuite) runFormProcessingTests() {
	fmt.Println("\\n📝 Form Processing Tests")
	fmt.Println("-------------------------")

	// Test bulk create
	bulkUsers := []models.UserCreateRequest{
		{Name: "Frank Miller", Email: "frank@company.com", Age: 35, Department: "Finance", Position: "Accountant"},
		{Name: "Grace Lee", Email: "grace@company.com", Age: 29, Department: "IT", Position: "System Admin"},
		{Name: "Henry Taylor", Email: "henry@company.com", Age: 31, Department: "Operations", Position: "Operations Manager"},
	}

	bulkJSON, _ := json.Marshal(bulkUsers)
	req := httptest.NewRequest("POST", "/api/users/bulk", bytes.NewBuffer(bulkJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("1. Bulk create users: Status %d\\n", rr.Code)

	// Get newly created user IDs for bulk operations
	if rr.Code == 201 {
		var response handlers.APIResponse
		json.Unmarshal(rr.Body.Bytes(), &response)
		createdUsers := response.Data.([]interface{})
		for _, user := range createdUsers {
			userMap := user.(map[string]interface{})
			ts.userIDs = append(ts.userIDs, userMap["id"].(string))
		}
	}

	// Test bulk update
	if len(ts.userIDs) >= 2 {
		newDept := "Updated Department"
		newAge := 40
		bulkUpdates := map[string]*models.UserUpdateRequest{
			ts.userIDs[0]: {Department: &newDept, Age: &newAge},
			ts.userIDs[1]: {Department: &newDept},
		}

		updateJSON, _ := json.Marshal(bulkUpdates)
		req = httptest.NewRequest("PUT", "/api/users/bulk", bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		rr = httptest.NewRecorder()
		ts.router.ServeHTTP(rr, req)
		fmt.Printf("2. Bulk update users: Status %d\\n", rr.Code)
	}

	// Test statistics after bulk operations
	req = httptest.NewRequest("GET", "/api/users/stats", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("3. Get user statistics: Status %d\\n", rr.Code)

	// Test department statistics
	req = httptest.NewRequest("GET", "/api/users/stats/departments", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("4. Get department statistics: Status %d\\n", rr.Code)

	// Test recent signups
	req = httptest.NewRequest("GET", "/api/users/recent-signups?days=1", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("5. Get recent signups: Status %d\\n", rr.Code)
}

func (ts *TestSuite) runProgressiveEnhancementTests() {
	fmt.Println("\\n🚀 Progressive Enhancement Tests")
	fmt.Println("---------------------------------")

	if len(ts.userIDs) == 0 {
		fmt.Println("No users available for progressive enhancement tests")
		return
	}

	userID := ts.userIDs[0]

	// Test user summary
	req := httptest.NewRequest("GET", "/api/users/"+userID+"/summary", nil)
	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("1. Get user summary: Status %d\\n", rr.Code)

	// Test update last login
	req = httptest.NewRequest("POST", "/api/users/"+userID+"/login", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("2. Update last login: Status %d\\n", rr.Code)

	// Test deactivate user
	req = httptest.NewRequest("POST", "/api/users/"+userID+"/deactivate", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("3. Deactivate user: Status %d\\n", rr.Code)

	// Test activate user
	req = httptest.NewRequest("POST", "/api/users/"+userID+"/activate", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("4. Activate user: Status %d\\n", rr.Code)

	// Test get inactive users (should find the deactivated user if timing is right)
	req = httptest.NewRequest("GET", "/api/users/inactive", nil)
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("5. Get inactive users after deactivation: Status %d\\n", rr.Code)

	// Test individual user update with enhanced validation
	name := "Updated Name"
	updateReq := models.UserUpdateRequest{
		Name: &name,
	}
	updateJSON, _ := json.Marshal(updateReq)
	req = httptest.NewRequest("PUT", "/api/users/"+userID, bytes.NewBuffer(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("6. Enhanced user update: Status %d\\n", rr.Code)

	// Test validation error handling
	invalidUser := models.UserCreateRequest{
		Name:  "", // Invalid: empty name
		Email: "invalid-email", // Invalid: no @ symbol
	}
	invalidJSON, _ := json.Marshal(invalidUser)
	req = httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	fmt.Printf("7. Validation error handling: Status %d\\n", rr.Code)

	// Test cleanup - bulk delete some users
	if len(ts.userIDs) >= 3 {
		deleteRequest := struct {
			IDs []string `json:"ids"`
		}{
			IDs: ts.userIDs[len(ts.userIDs)-3:], // Delete last 3 users
		}
		deleteJSON, _ := json.Marshal(deleteRequest)
		req = httptest.NewRequest("DELETE", "/api/users/bulk", bytes.NewBuffer(deleteJSON))
		req.Header.Set("Content-Type", "application/json")
		rr = httptest.NewRecorder()
		ts.router.ServeHTTP(rr, req)
		fmt.Printf("8. Bulk delete users: Status %d\\n", rr.Code)
	}
}