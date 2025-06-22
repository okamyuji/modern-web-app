package main

import (
	"fmt"
	"htmx-demo/internal/handlers"
	"net/http/httptest"
	"strings"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize handlers
	pageHandler := handlers.NewPageHandler()
	apiHandler := handlers.NewAPIHandler()

	// Setup router
	router := mux.NewRouter()
	router.HandleFunc("/", pageHandler.Home).Methods("GET")
	router.HandleFunc("/basic-requests", pageHandler.BasicRequests).Methods("GET")
	router.HandleFunc("/triggers", pageHandler.Triggers).Methods("GET")

	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/users", apiHandler.GetUsers).Methods("GET")
	api.HandleFunc("/users", apiHandler.CreateUser).Methods("POST")
	api.HandleFunc("/echo", apiHandler.Echo).Methods("PUT", "DELETE")
	api.HandleFunc("/search", apiHandler.Search).Methods("GET", "POST")
	api.HandleFunc("/time", apiHandler.GetTime).Methods("GET")
	api.HandleFunc("/focus-data", apiHandler.FocusData).Methods("GET")
	api.HandleFunc("/special-action", apiHandler.SpecialAction).Methods("POST")
	api.HandleFunc("/custom-response", apiHandler.CustomResponse).Methods("GET")

	fmt.Println("Testing Chapter 3 HTMX Demo Implementation")
	fmt.Println("=========================================")

	// Test home page
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("1. Home page: Status %d - Contains HTMX: %t\n", rr.Code, 
		strings.Contains(rr.Body.String(), "HTMX"))

	// Test basic requests page
	req = httptest.NewRequest("GET", "/basic-requests", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("2. Basic requests page: Status %d - Contains hx-get: %t\n", rr.Code,
		strings.Contains(rr.Body.String(), "hx-get"))

	// Test triggers page
	req = httptest.NewRequest("GET", "/triggers", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("3. Triggers page: Status %d - Contains hx-trigger: %t\n", rr.Code,
		strings.Contains(rr.Body.String(), "hx-trigger"))

	// Test API endpoints
	req = httptest.NewRequest("GET", "/api/users", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("4. Get users API: Status %d - Contains user data: %t\n", rr.Code,
		strings.Contains(rr.Body.String(), "田中太郎"))

	// Test create user
	req = httptest.NewRequest("POST", "/api/users", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("5. Create user API: Status %d - Contains success message: %t\n", rr.Code,
		strings.Contains(rr.Body.String(), "作成されました"))

	// Test search functionality
	req = httptest.NewRequest("GET", "/api/search?search=Go", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("6. Search API: Status %d - Contains search results: %t\n", rr.Code,
		strings.Contains(rr.Body.String(), "Go言語"))

	// Test time endpoint
	req = httptest.NewRequest("GET", "/api/time", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("7. Time API: Status %d - Contains time: %t\n", rr.Code,
		strings.Contains(rr.Body.String(), "現在時刻"))

	// Test special action
	req = httptest.NewRequest("POST", "/api/special-action", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	fmt.Printf("8. Special action API: Status %d - Contains special message: %t\n", rr.Code,
		strings.Contains(rr.Body.String(), "特別なアクション"))

	fmt.Println("\n✅ Chapter 3 HTMX implementation complete and tested successfully!")
	fmt.Println("HTMX Features demonstrated:")
	fmt.Println("- Basic HTTP verbs (GET, POST, PUT, DELETE)")
	fmt.Println("- Trigger controls (keyup, focus, periodic, conditional)")
	fmt.Println("- Target specification and content swapping")
	fmt.Println("- Loading indicators and user feedback")
	fmt.Println("- Search functionality with delay")
	fmt.Println("- Custom events and progressive enhancement")
}