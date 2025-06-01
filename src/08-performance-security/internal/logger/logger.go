package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "debug"
	case INFO:
		return "info"
	case WARN:
		return "warn"
	case ERROR:
		return "error"
	default:
		return "unknown"
	}
}

type Logger struct {
	output io.Writer
	level  LogLevel
}

type LogEntry struct {
	Time       time.Time              `json:"time"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	TraceID    string                 `json:"trace_id,omitempty"`
	Duration   *float64               `json:"duration_ms,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
}

func NewLogger(output io.Writer, level LogLevel) *Logger {
	if output == nil {
		output = os.Stdout
	}
	return &Logger{
		output: output,
		level:  level,
	}
}

func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}, traceID string, duration *float64, err error) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Time:    time.Now().UTC(),
		Level:   level.String(),
		Message: message,
		Fields:  fields,
		TraceID: traceID,
	}

	if duration != nil {
		entry.Duration = duration
	}

	if err != nil {
		entry.Error = err.Error()
		if level == ERROR {
			entry.StackTrace = getStackTrace()
		}
	}

	data, _ := json.Marshal(entry)
	l.output.Write(append(data, '\n'))
}

func (l *Logger) Debug(message string, fields map[string]interface{}, traceID string, duration *float64) {
	l.log(DEBUG, message, fields, traceID, duration, nil)
}

func (l *Logger) Info(message string, fields map[string]interface{}, traceID string, duration *float64) {
	l.log(INFO, message, fields, traceID, duration, nil)
}

func (l *Logger) Warn(message string, fields map[string]interface{}, traceID string, duration *float64) {
	l.log(WARN, message, fields, traceID, duration, nil)
}

func (l *Logger) Error(message string, fields map[string]interface{}, traceID string, duration *float64) {
	l.log(ERROR, message, fields, traceID, duration, nil)
}

func (l *Logger) ErrorWithStack(message string, fields map[string]interface{}, traceID string, err error) {
	l.log(ERROR, message, fields, traceID, nil, err)
}

// getStackTrace - スタックトレースを取得
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// generateTraceID - トレースIDを生成
func generateTraceID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// LoggingResponseWriter - ログ用レスポンスライター
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}

// RequestLogger - リクエストロギングミドルウェア
func RequestLogger(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// トレースIDの生成/取得
			traceID := r.Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = generateTraceID()
			}

			// レスポンスライターのラップ
			lrw := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// コンテキストにトレースIDを追加
			ctx := context.WithValue(r.Context(), "trace_id", traceID)
			r = r.WithContext(ctx)

			// レスポンスヘッダーにトレースIDを追加
			w.Header().Set("X-Trace-ID", traceID)

			// リクエスト処理
			next.ServeHTTP(lrw, r)

			// ログ記録
			duration := float64(time.Since(start).Nanoseconds()) / 1e6 // ミリ秒

			fields := map[string]interface{}{
				"method":       r.Method,
				"path":         r.URL.Path,
				"query":        r.URL.RawQuery,
				"status":       lrw.statusCode,
				"size":         lrw.size,
				"ip":           getClientIP(r),
				"user_agent":   r.UserAgent(),
				"referer":      r.Referer(),
				"htmx_request": r.Header.Get("HX-Request") == "true",
				"htmx_target":  r.Header.Get("HX-Target"),
				"htmx_trigger": r.Header.Get("HX-Trigger"),
			}

			// エラーレスポンスの場合は詳細を記録
			if lrw.statusCode >= 400 {
				if lrw.statusCode >= 500 {
					logger.Error("Request failed", fields, traceID, &duration)
				} else {
					logger.Warn("Client error", fields, traceID, &duration)
				}
			} else {
				logger.Info("Request completed", fields, traceID, &duration)
			}
		})
	}
}

// Metrics - メトリクス収集
type Metrics struct {
	RequestCount    uint64
	ErrorCount      uint64
	TotalDuration   int64 // nanoseconds
	ActiveRequests  int32
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) RecordRequest(duration time.Duration, isError bool) {
	atomic.AddUint64(&m.RequestCount, 1)
	atomic.AddInt64(&m.TotalDuration, int64(duration))

	if isError {
		atomic.AddUint64(&m.ErrorCount, 1)
	}
}

func (m *Metrics) IncActiveRequests() {
	atomic.AddInt32(&m.ActiveRequests, 1)
}

func (m *Metrics) DecActiveRequests() {
	atomic.AddInt32(&m.ActiveRequests, -1)
}

func (m *Metrics) GetStats() map[string]interface{} {
	requestCount := atomic.LoadUint64(&m.RequestCount)
	errorCount := atomic.LoadUint64(&m.ErrorCount)
	totalDuration := atomic.LoadInt64(&m.TotalDuration)
	activeRequests := atomic.LoadInt32(&m.ActiveRequests)

	avgDuration := float64(0)
	if requestCount > 0 {
		avgDuration = float64(totalDuration) / float64(requestCount) / 1e6 // ms
	}

	errorRate := float64(0)
	if requestCount > 0 {
		errorRate = float64(errorCount) / float64(requestCount) * 100
	}

	return map[string]interface{}{
		"request_count":    requestCount,
		"error_count":      errorCount,
		"error_rate":       errorRate,
		"avg_duration_ms":  avgDuration,
		"active_requests":  activeRequests,
	}
}

// MetricsMiddleware - メトリクス収集ミドルウェア
func MetricsMiddleware(metrics *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			metrics.IncActiveRequests()
			defer metrics.DecActiveRequests()

			lrw := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(lrw, r)

			duration := time.Since(start)
			isError := lrw.statusCode >= 400

			metrics.RecordRequest(duration, isError)
		})
	}
}

// getClientIP - クライアントIPを取得
func getClientIP(r *http.Request) string {
	// プロキシヘッダーをチェック
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// 最初のIPを返す
		if idx := strings.Index(ip, ","); idx != -1 {
			return strings.TrimSpace(ip[:idx])
		}
		return strings.TrimSpace(ip)
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	// RemoteAddrからポート番号を除去
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}