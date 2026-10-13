package main

import (
	"crypto/subtle"
	"database/sql"
	"math"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxRequestBodyBytes  = 64 << 10
	maxTitleLength       = 200
	internalErrorMessage = "internal server error"
)

// newAPIHandler /api/todos 配下のハンドラを、レート制限とトークン認証で包んで返す
func newAPIHandler(db *sql.DB, token string, ratePerMinute int) http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("/api/todos", handleTodos(db))
	api.HandleFunc("/api/todos/", handleTodoByID(db))
	// 認証より先に数えて、トークンの総当たりも制限する
	return limitRate(newRateLimiter(ratePerMinute, time.Minute), requireToken(token, api))
}

// requireToken Authorization: Bearer <token> が一致しない要求を 401 で拒否する。token が空なら Host がループバックの要求だけを通す
func requireToken(token string, next http.Handler) http.Handler {
	if token == "" {
		// DNS rebinding では、ブラウザが攻撃者のドメイン名を Host に入れて送ってくる
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isLoopbackHost(r.Host) {
				http.Error(w, "forbidden host", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	want := []byte("Bearer " + token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), want) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// isLoopbackHost Host ヘッダーが localhost かループバックアドレスを指すかを返す
func isLoopbackHost(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	addr, err := netip.ParseAddr(strings.Trim(host, "[]"))
	return err == nil && addr.IsLoopback()
}

// limitRate クライアント IP ごとの上限を超えた要求を 429 で拒否する
func limitRate(rl *rateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientIP(r)) {
			seconds := int(math.Ceil(rl.retryAfter().Seconds()))
			w.Header().Set("Retry-After", strconv.Itoa(max(seconds, 1)))
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimiter 全クライアント共通の固定ウィンドウで要求数を数える
type rateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	now    func() time.Time
	start  time.Time
	counts map[string]int
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{limit: limit, window: window, now: time.Now, counts: map[string]int{}}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if now := rl.now(); now.Sub(rl.start) >= rl.window {
		rl.start = now
		// ウィンドウごとに作り直し、見かけたクライアント数でメモリが増え続けないようにする
		rl.counts = map[string]int{}
	}
	rl.counts[key]++
	return rl.counts[key] <= rl.limit
}

func (rl *rateLimiter) retryAfter() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.window - rl.now().Sub(rl.start)
}

// clientIP レート制限のキーとして接続元を返す。X-Forwarded-For は偽装できるので使わない
// shortcut: リバースプロキシの後ろでは全クライアントが 1 つの上限を共有する。プロキシを置くときは信頼するプロキシ経由の X-Forwarded-For だけを読むようにする
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return host
	}
	addr = addr.Unmap().WithZone("")
	if addr.Is4() {
		return addr.String()
	}
	// IPv6 は利用者 1 人に /64 が割り当てられるのが普通なので、アドレス単位だと上限を回避できる
	prefix, _ := addr.Prefix(64)
	return prefix.String()
}
