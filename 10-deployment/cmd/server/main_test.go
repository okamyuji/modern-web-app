package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"deployment-demo/internal/config"
)

const testToken = "not-a-real-key-but-long-enough-for-tests"

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := initDatabase("sqlite://" + filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("initDatabase: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func doRequest(h http.Handler, method, body, token, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/todos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestInitDatabase_OpensFileFromDatabaseURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom.db")

	db, err := initDatabase("sqlite://" + path)
	if err != nil {
		t.Fatalf("initDatabase: %v", err)
	}
	db.Close()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("database file not created at DATABASE_URL path: %v", err)
	}
}

func TestInitDatabase_RejectsNonSQLiteURL(t *testing.T) {
	if _, err := initDatabase("postgres://localhost/app"); err == nil {
		t.Fatal("expected error for non-sqlite DATABASE_URL")
	}
}

func TestInitDatabase_ErrorDoesNotLeakURLCredentials(t *testing.T) {
	_, err := initDatabase("postgres://app:not-a-real-password@db.example/app")

	if err == nil || strings.Contains(err.Error(), "not-a-real-password") {
		t.Fatalf("err = %v, want an error without the URL credentials", err)
	}
}

func TestTodosAPI_RejectsMissingToken(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 100)

	rec := doRequest(h, http.MethodGet, "", "", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestTodosAPI_RejectsWrongToken(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 100)

	rec := doRequest(h, http.MethodGet, "", testToken+"x", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestTodosAPI_RejectsNonBearerScheme(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 100)
	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	req.Header.Set("Authorization", "Basic "+testToken)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestTodosAPI_CreatesAndListsTodoWithValidToken(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 100)

	created := doRequest(h, http.MethodPost, `{"title":"buy milk"}`, testToken, "")
	listed := doRequest(h, http.MethodGet, "", testToken, "")

	if created.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want 201 (body %q)", created.Code, created.Body.String())
	}
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), "buy milk") {
		t.Fatalf("GET status = %d body = %q, want 200 containing the todo", listed.Code, listed.Body.String())
	}
}

func TestTodosAPI_AllowsRequestsWhenNoTokenConfigured(t *testing.T) {
	h := newAPIHandler(newTestDB(t), "", 100)

	rec := doRequest(h, http.MethodGet, "", "", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestTodosAPI_RejectsOversizedBody(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 100)
	body := `{"title":"x","description":"` + strings.Repeat("a", maxRequestBodyBytes) + `"}`

	rec := doRequest(h, http.MethodPost, body, testToken, "")

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
}

func TestTodosAPI_AcceptsBodyAtLimit(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 100)
	prefix, suffix := `{"title":"x","description":"`, `"}`
	body := prefix + strings.Repeat("a", maxRequestBodyBytes-len(prefix)-len(suffix)) + suffix

	rec := doRequest(h, http.MethodPost, body, testToken, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 for a body of exactly %d bytes", rec.Code, maxRequestBodyBytes)
	}
}

func TestTodosAPI_RejectsInvalidJSON(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 100)

	rec := doRequest(h, http.MethodPost, `{"title":`, testToken, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestTodosAPI_ValidatesTitle(t *testing.T) {
	cases := []struct {
		name  string
		title string
		want  int
	}{
		{"empty", "", http.StatusBadRequest},
		{"whitespace only", "   ", http.StatusBadRequest},
		{"at max length", strings.Repeat("あ", maxTitleLength), http.StatusCreated},
		{"over max length", strings.Repeat("あ", maxTitleLength+1), http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newAPIHandler(newTestDB(t), testToken, 100)

			rec := doRequest(h, http.MethodPost, `{"title":"`+tc.title+`"}`, testToken, "")

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestTodosAPI_RejectsUnsupportedMethod(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 100)

	rec := doRequest(h, http.MethodDelete, "", testToken, "")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestTodosAPI_HidesDatabaseErrors(t *testing.T) {
	db := newTestDB(t)
	h := newAPIHandler(db, testToken, 100)
	db.Close()

	get := doRequest(h, http.MethodGet, "", testToken, "")
	post := doRequest(h, http.MethodPost, `{"title":"x"}`, testToken, "")

	for name, rec := range map[string]*httptest.ResponseRecorder{"GET": get, "POST": post} {
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("%s status = %d, want 500", name, rec.Code)
		}
		if body := strings.TrimSpace(rec.Body.String()); body != internalErrorMessage {
			t.Fatalf("%s body = %q, want only %q", name, body, internalErrorMessage)
		}
	}
}

func TestTodosAPI_RateLimitsPerClientIP(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 2)

	first := doRequest(h, http.MethodGet, "", testToken, "192.0.2.1:1111")
	second := doRequest(h, http.MethodGet, "", testToken, "192.0.2.1:2222")
	third := doRequest(h, http.MethodGet, "", testToken, "192.0.2.1:3333")
	otherClient := doRequest(h, http.MethodGet, "", testToken, "192.0.2.2:1111")

	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("first two = %d, %d, want 200", first.Code, second.Code)
	}
	if third.Code != http.StatusTooManyRequests {
		t.Fatalf("third = %d, want 429", third.Code)
	}
	if third.Header().Get("Retry-After") == "" {
		t.Fatal("429 response has no Retry-After header")
	}
	if otherClient.Code != http.StatusOK {
		t.Fatalf("other client = %d, want 200", otherClient.Code)
	}
}

func TestTodosAPI_RateLimitCountsUnauthenticatedRequests(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 1)

	doRequest(h, http.MethodGet, "", "", "192.0.2.9:1")
	rec := doRequest(h, http.MethodGet, "", testToken, "192.0.2.9:2")

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429 after an unauthenticated request used the quota", rec.Code)
	}
}

func TestSetupRoutes_APIRoutesShareOneRateLimit(t *testing.T) {
	cfg := &config.Config{Env: "development", APIRateLimit: 1}
	h := setupRoutes(newTestDB(t), cfg)

	list := httptest.NewRecorder()
	h.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/todos", nil))
	byID := httptest.NewRecorder()
	h.ServeHTTP(byID, httptest.NewRequest(http.MethodGet, "/api/todos/1", nil))

	if list.Code != http.StatusOK {
		t.Fatalf("/api/todos = %d, want 200", list.Code)
	}
	if byID.Code != http.StatusTooManyRequests {
		t.Fatalf("/api/todos/1 = %d, want 429 because both routes share one quota", byID.Code)
	}
}

func TestSetupRoutes_RequiresTokenOnAPI(t *testing.T) {
	cfg := &config.Config{Env: "development", APIRateLimit: 100, APIToken: testToken}
	h := setupRoutes(newTestDB(t), cfg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/todos/1", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("/api/todos/1 without token = %d, want 401", rec.Code)
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rl := newRateLimiter(1, time.Minute)
	rl.now = func() time.Time { return now }

	if !rl.allow("a") {
		t.Fatal("first request denied")
	}
	if rl.allow("a") {
		t.Fatal("second request in the same window allowed")
	}
	now = now.Add(time.Minute)
	if !rl.allow("a") {
		t.Fatal("request after the window denied")
	}
}

func TestRateLimiter_RetryAfterIsTimeLeftInWindow(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rl := newRateLimiter(1, time.Minute)
	rl.now = func() time.Time { return now }
	rl.allow("a")

	now = now.Add(40 * time.Second)

	if got := rl.retryAfter(); got != 20*time.Second {
		t.Fatalf("retryAfter = %v, want 20s", got)
	}
}

func TestClientIP(t *testing.T) {
	cases := map[string]string{
		"192.0.2.1:1234":   "192.0.2.1",
		"[2001:db8::1]:80": "2001:db8::1",
		"no-port":          "no-port",
	}
	for remote, want := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = remote
		req.Header.Set("X-Forwarded-For", "203.0.113.7")

		if got := clientIP(req); got != want {
			t.Errorf("clientIP(%q) = %q, want %q", remote, got, want)
		}
	}
}
