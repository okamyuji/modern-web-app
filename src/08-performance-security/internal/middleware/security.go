package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// CSRFManager - CSRF攻撃対策
type CSRFManager struct {
	tokens sync.Map
}

func NewCSRFManager() *CSRFManager {
	return &CSRFManager{}
}

func (m *CSRFManager) GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := base64.URLEncoding.EncodeToString(b)
	m.tokens.Store(token, true)
	return token
}

func (m *CSRFManager) ValidateToken(token string) bool {
	_, exists := m.tokens.LoadAndDelete(token)
	return exists
}

func (m *CSRFManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GETリクエストはスキップ
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		// HTMXリクエストの場合、ヘッダーからトークンを取得
		token := r.Header.Get("X-CSRF-Token")
		if token == "" {
			token = r.FormValue("csrf_token")
		}

		if !m.ValidateToken(token) {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// テンプレートヘルパー
func (m *CSRFManager) TemplateFunc() template.FuncMap {
	return template.FuncMap{
		"csrfToken": m.GenerateToken,
		"csrfField": func() template.HTML {
			token := m.GenerateToken()
			return template.HTML(fmt.Sprintf(`<input type="hidden" name="csrf_token" value="%s">`, token))
		},
	}
}

// SecurityHeaders - セキュリティヘッダーの設定
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 基本的なセキュリティヘッダー
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// CSP（Content Security Policy）
		csp := []string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' https://unpkg.com https://cdn.tailwindcss.com", // HTMX、Alpine.js、Tailwind用
			"style-src 'self' 'unsafe-inline'",                                                 // Tailwind CSS用
			"img-src 'self' data: https:",
			"connect-src 'self'", // HTMXのリクエスト用
			"font-src 'self'",
			"frame-ancestors 'none'",
		}
		w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))

		// HSTS（HTTPSの場合のみ）
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// Sanitizer - 入力のサニタイゼーション
type Sanitizer struct {
	allowedTags *regexp.Regexp
}

func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		allowedTags: regexp.MustCompile(`<(b|i|u|strong|em|br)(\s[^>]*)?>|</(b|i|u|strong|em)>`),
	}
}

func (s *Sanitizer) Sanitize(input string) string {
	// HTMLエスケープ
	escaped := template.HTMLEscapeString(input)

	// 許可されたタグのみ復元
	return s.allowedTags.ReplaceAllStringFunc(escaped, func(match string) string {
		// エスケープを解除
		return strings.ReplaceAll(
			strings.ReplaceAll(match, "&lt;", "<"),
			"&gt;", ">",
		)
	})
}

// RateLimiter - レート制限
type RateLimiter struct {
	requests sync.Map // IP -> *RequestCounter
}

type RequestCounter struct {
	count     int
	resetTime int64
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

func (rl *RateLimiter) IsAllowed(ip string, limit int, windowSeconds int64) bool {
	now := time.Now().Unix()
	
	value, _ := rl.requests.LoadOrStore(ip, &RequestCounter{
		count:     0,
		resetTime: now + windowSeconds,
	})
	
	counter := value.(*RequestCounter)
	
	// ウィンドウがリセットされた場合
	if now > counter.resetTime {
		counter.count = 0
		counter.resetTime = now + windowSeconds
	}
	
	counter.count++
	
	return counter.count <= limit
}

func (rl *RateLimiter) Middleware(limit int, windowSeconds int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			
			if !rl.IsAllowed(ip, limit, windowSeconds) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP - クライアントIPを取得
func getClientIP(r *http.Request) string {
	// プロキシヘッダーをチェック
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// InputValidator - 入力検証
type InputValidator struct {
	patterns map[string]*regexp.Regexp
}

func NewInputValidator() *InputValidator {
	return &InputValidator{
		patterns: map[string]*regexp.Regexp{
			"email":    regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
			"username": regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`),
			"alphanumeric": regexp.MustCompile(`^[a-zA-Z0-9]+$`),
		},
	}
}

func (v *InputValidator) ValidateEmail(email string) bool {
	return v.patterns["email"].MatchString(email)
}

func (v *InputValidator) ValidateUsername(username string) bool {
	return v.patterns["username"].MatchString(username)
}

func (v *InputValidator) ValidateAlphanumeric(input string) bool {
	return v.patterns["alphanumeric"].MatchString(input)
}

func (v *InputValidator) SanitizeInput(input string, maxLength int) string {
	// 長さ制限
	if len(input) > maxLength {
		input = input[:maxLength]
	}
	
	// 危険な文字の除去
	input = strings.ReplaceAll(input, "<script", "")
	input = strings.ReplaceAll(input, "</script>", "")
	input = strings.ReplaceAll(input, "javascript:", "")
	input = strings.ReplaceAll(input, "vbscript:", "")
	
	return strings.TrimSpace(input)
}

// AuthMiddleware - 認証ミドルウェア（簡易版）
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// セッションから認証情報を確認
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			sessionID = cookie.Value
		}
		
		// セッションの検証（実際の実装では Redis や データベースを使用）
		if !isValidSession(sessionID) {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// isValidSession - セッションの有効性チェック（簡易実装）
func isValidSession(sessionID string) bool {
	// 実際の実装では、セッションストアから検証
	return sessionID != "" && len(sessionID) > 10
}

// CORSMiddleware - CORS設定
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// 許可されたオリジンかチェック
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token, HX-Request")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			
			// プリフライトリクエストの処理
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}