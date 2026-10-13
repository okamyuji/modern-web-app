package main

import (
	"database/sql"
	"io"
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

// localRequest トークンなしの API は Host がループバックの要求だけを通すので、Host を localhost にした要求を作る
func localRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Host = "localhost:8080"
	return req
}

func doRequest(h http.Handler, method, body, token, remoteAddr string) *httptest.ResponseRecorder {
	req := localRequest(method, "/api/todos", strings.NewReader(body))
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

func TestInitDatabase_RejectsEmptySQLitePath(t *testing.T) {
	if _, err := initDatabase("sqlite://"); err == nil {
		t.Fatal("expected error for sqlite:// without a path")
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

func TestTodosAPI_ValidatesPriority(t *testing.T) {
	cases := map[string]int{
		"":           http.StatusCreated,
		"low":        http.StatusCreated,
		"medium":     http.StatusCreated,
		"high":       http.StatusCreated,
		"urgent":     http.StatusBadRequest,
		"High":       http.StatusBadRequest,
		"low-random": http.StatusBadRequest,
	}
	for priority, want := range cases {
		t.Run(priority, func(t *testing.T) {
			h := newAPIHandler(newTestDB(t), testToken, 100)

			rec := doRequest(h, http.MethodPost, `{"title":"x","priority":"`+priority+`"}`, testToken, "")

			if rec.Code != want {
				t.Fatalf("priority %q status = %d, want %d", priority, rec.Code, want)
			}
		})
	}
}

func TestTodosAPI_DefaultsMissingPriorityToMedium(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 100)

	created := doRequest(h, http.MethodPost, `{"title":"no priority"}`, testToken, "")
	listed := doRequest(h, http.MethodGet, "", testToken, "")

	if !strings.Contains(created.Body.String(), `"priority":"medium"`) {
		t.Fatalf("POST body = %q, want priority medium", created.Body.String())
	}
	if !strings.Contains(listed.Body.String(), `"priority":"medium"`) {
		t.Fatalf("GET body = %q, want stored priority medium", listed.Body.String())
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
	h.ServeHTTP(list, localRequest(http.MethodGet, "/api/todos", nil))
	byID := httptest.NewRecorder()
	h.ServeHTTP(byID, localRequest(http.MethodGet, "/api/todos/1", nil))

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
		"192.0.2.1:1234":        "192.0.2.1",
		"[2001:db8::1]:80":      "2001:db8::/64",
		"[2001:db8::ffff:1]:80": "2001:db8::/64",
		"[2001:db8:0:1::1]:80":  "2001:db8:0:1::/64",
		"[::ffff:192.0.2.1]:80": "192.0.2.1",
		"no-port":               "no-port",
		"host.example:80":       "host.example",
		"[fe80::1%en0]:80":      "fe80::/64",
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

func TestTodosAPI_PostRequiresJSONContentType(t *testing.T) {
	cases := []struct {
		contentType string
		want        int
	}{
		{"text/plain", http.StatusUnsupportedMediaType},
		{"application/x-www-form-urlencoded", http.StatusUnsupportedMediaType},
		{"multipart/form-data; boundary=x", http.StatusUnsupportedMediaType},
		{"", http.StatusUnsupportedMediaType},
		{"application/json; =bad", http.StatusUnsupportedMediaType},
		{"application/json", http.StatusCreated},
		{"Application/JSON; charset=utf-8", http.StatusCreated},
	}
	for _, tc := range cases {
		t.Run(tc.contentType, func(t *testing.T) {
			h := newAPIHandler(newTestDB(t), "", 100)
			req := localRequest(http.MethodPost, "/api/todos", strings.NewReader(`{"title":"drive-by"}`))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestTodosAPI_NoTokenAllowsOnlyLoopbackHost(t *testing.T) {
	cases := []struct {
		host string
		want int
	}{
		{"rebind.attacker.example", http.StatusForbidden},
		{"rebind.attacker.example:8080", http.StatusForbidden},
		{"localhost.attacker.example:8080", http.StatusForbidden},
		{"192.168.1.10:8080", http.StatusForbidden},
		{"", http.StatusForbidden},
		{"localhost", http.StatusOK},
		{"LOCALHOST:8080", http.StatusOK},
		{"127.0.0.1:8080", http.StatusOK},
		{"127.0.0.2", http.StatusOK},
		{"[::1]:8080", http.StatusOK},
		{"[::1]", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.host, func(t *testing.T) {
			h := newAPIHandler(newTestDB(t), "", 100)
			req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
			req.Host = tc.host
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Fatalf("Host %q: status = %d, want %d", tc.host, rec.Code, tc.want)
			}
		})
	}
}

func TestTodosAPI_TokenAuthDoesNotCheckHost(t *testing.T) {
	h := newAPIHandler(newTestDB(t), testToken, 100)
	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	req.Host = "api.example.com"
	req.Header.Set("Authorization", "Bearer "+testToken)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestCheckHealth(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{"ok", func(w http.ResponseWriter, r *http.Request) {}, false},
		{"unhealthy", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }, true},
		{"stalled", func(w http.ResponseWriter, r *http.Request) { time.Sleep(time.Second) }, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()
			start := time.Now()

			err := checkHealth(srv.URL, 100*time.Millisecond)

			if tc.wantErr != (err != nil) {
				t.Fatalf("checkHealth = %v, wantErr %v", err, tc.wantErr)
			}
			if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
				t.Fatalf("checkHealth took %v, want it to give up after the 100ms timeout", elapsed)
			}
		})
	}
}
