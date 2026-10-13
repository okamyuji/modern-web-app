package monitoring

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"deployment-demo/internal/config"
)

var (
	// HTTPメトリクス
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	httpRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	// ビジネスメトリクス
	todosCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "todos_created_total",
			Help: "Total number of todos created",
		},
		[]string{"priority"},
	)

	todosCompleted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "todos_completed_total",
			Help: "Total number of todos completed",
		},
	)

	activeUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_current",
			Help: "Current number of active users",
		},
	)

	// システムメトリクス
	databaseConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "Current number of database connections",
		},
		[]string{"state"}, // open, idle, in_use
	)

	databaseOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "table", "status"}, // operation: select, insert, update, delete
	)

	cacheOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "status"}, // operation: hit, miss, set, delete
	)

	// アプリケーション情報
	appInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Application information",
		},
		[]string{"version", "build_time", "environment"},
	)
)

type statusWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

func init() {
	// メトリクスの登録
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		httpRequestSize,
		httpResponseSize,
		todosCreated,
		todosCompleted,
		activeUsers,
		databaseConnections,
		databaseOperations,
		cacheOperations,
		appInfo,
	)
}

// InitMetrics はアプリケーション情報を設定
func InitMetrics(cfg *config.Config) {
	appInfo.WithLabelValues(cfg.Version, cfg.BuildTime, cfg.Env).Set(1)
}

// PrometheusMiddleware はHTTPリクエストのメトリクスを収集
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// カスタムレスポンスライター
		sw := &statusWriter{
			ResponseWriter: w,
			status:         200,
		}

		// リクエスト処理
		next.ServeHTTP(sw, r)

		// メトリクスの記録
		duration := time.Since(start).Seconds()
		// パスやメソッドをそのままラベルにすると、未認証の要求だけで系列が際限なく増える。
		// ServeMux がルーティング時に設定する登録済みパターンを使う
		endpoint := cmp.Or(r.Pattern, "other")
		method := metricMethod(r.Method)
		statusCode := strconv.Itoa(sw.status)

		if r.ContentLength > 0 {
			httpRequestSize.WithLabelValues(method, endpoint).Observe(float64(r.ContentLength))
		}

		httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
		httpResponseSize.WithLabelValues(method, endpoint).Observe(float64(sw.size))
	})
}

// metricMethod 標準のメソッド以外を OTHER にまとめる
func metricMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodOptions, http.MethodConnect, http.MethodTrace:
		return method
	}
	return "OTHER"
}

// ビジネスメトリクスのヘルパー関数
func RecordTodoCreated(priority string) {
	todosCreated.WithLabelValues(priority).Inc()
}

func RecordTodoCompleted() {
	todosCompleted.Inc()
}

func SetActiveUsers(count float64) {
	activeUsers.Set(count)
}

func RecordDatabaseConnections(open, idle, inUse int) {
	databaseConnections.WithLabelValues("open").Set(float64(open))
	databaseConnections.WithLabelValues("idle").Set(float64(idle))
	databaseConnections.WithLabelValues("in_use").Set(float64(inUse))
}

func RecordDatabaseOperation(operation, table, status string) {
	databaseOperations.WithLabelValues(operation, table, status).Inc()
}

func RecordCacheOperation(operation, status string) {
	cacheOperations.WithLabelValues(operation, status).Inc()
}

// StartMetricsServer はPrometheusメトリクスサーバーを起動し、待ち受けたアドレスを返す
func StartMetricsServer(addr string, idleTimeout time.Duration) (net.Addr, error) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// ヘルスチェック用エンドポイント
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 待ち受けの失敗を起動時のエラーとして返すため、Listen は同期で呼ぶ
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("metrics server failed to listen on %s: %w", addr, err)
	}

	// 遅いヘッダーや休止した keep-alive 接続でファイル記述子を使い切られないよう、待ち時間に上限を設ける
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: idleTimeout}
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic("metrics server stopped: " + err.Error())
		}
	}()
	return ln.Addr(), nil
}

// ServeUntilSignal ln で要求を処理し、quit を受けたら処理中の要求を timeout まで待って止める。cleanup の完了後に戻る
func ServeUntilSignal(server *http.Server, ln net.Listener, quit <-chan os.Signal, timeout time.Duration, cleanup func()) error {
	serveDone := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Serve が異常終了したときや外から Shutdown されたときも、このゴルーチンを残さない
		select {
		case <-quit:
		case <-serveDone:
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			server.Close()
		}
		if cleanup != nil {
			cleanup()
		}
	}()

	// Shutdown を呼ぶと Serve はすぐ戻るので、処理中の要求と cleanup の完了は done で待つ
	err := server.Serve(ln)
	close(serveDone)
	<-done
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// HealthCheck はアプリケーションの健全性をチェック
func HealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
		"metrics": map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}
}

var startTime = time.Now()

// SystemMetrics はシステムメトリクスを収集
type SystemMetrics struct {
	startTime time.Time
}

func NewSystemMetrics() *SystemMetrics {
	return &SystemMetrics{
		startTime: time.Now(),
	}
}

func (sm *SystemMetrics) GetUptime() time.Duration {
	return time.Since(sm.startTime)
}

func (sm *SystemMetrics) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"uptime_seconds": time.Since(sm.startTime).Seconds(),
		"start_time":     sm.startTime.Format(time.RFC3339),
	}
}

// CustomMetrics はアプリケーション固有のメトリクスを収集
type CustomMetrics struct {
	requestCount    prometheus.Counter
	errorCount      prometheus.Counter
	responseTimeSum prometheus.Counter
}

func NewCustomMetrics() *CustomMetrics {
	return &CustomMetrics{
		requestCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "custom_requests_total",
			Help: "Total custom requests",
		}),
		errorCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "custom_errors_total",
			Help: "Total custom errors",
		}),
		responseTimeSum: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "custom_response_time_sum",
			Help: "Sum of custom response times",
		}),
	}
}

func (cm *CustomMetrics) Register() {
	prometheus.MustRegister(cm.requestCount, cm.errorCount, cm.responseTimeSum)
}

func (cm *CustomMetrics) RecordRequest() {
	cm.requestCount.Inc()
}

func (cm *CustomMetrics) RecordError() {
	cm.errorCount.Inc()
}

func (cm *CustomMetrics) RecordResponseTime(duration time.Duration) {
	cm.responseTimeSum.Add(duration.Seconds())
}
