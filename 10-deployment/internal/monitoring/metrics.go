package monitoring

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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

		// リクエストサイズの記録
		if r.ContentLength > 0 {
			httpRequestSize.WithLabelValues(r.Method, r.URL.Path).Observe(float64(r.ContentLength))
		}

		// カスタムレスポンスライター
		sw := &statusWriter{
			ResponseWriter: w,
			status:         200,
		}

		// リクエスト処理
		next.ServeHTTP(sw, r)

		// メトリクスの記録
		duration := time.Since(start).Seconds()
		endpoint := r.URL.Path
		method := r.Method
		statusCode := strconv.Itoa(sw.status)

		httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
		httpResponseSize.WithLabelValues(method, endpoint).Observe(float64(sw.size))
	})
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

// StartMetricsServer はPrometheusメトリクスサーバーを起動
func StartMetricsServer(port string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	
	// ヘルスチェック用エンドポイント
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("metrics server failed to start: " + err.Error())
		}
	}()
}

// GracefulShutdown はアプリケーションのグレースフルシャットダウンを実行
func GracefulShutdown(server *http.Server, cfg *config.Config, cleanup func()) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// 新しい接続の受付を停止
	if err := server.Shutdown(ctx); err != nil {
		// 強制終了
		server.Close()
	}

	// 追加のクリーンアップ処理
	if cleanup != nil {
		cleanup()
	}
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