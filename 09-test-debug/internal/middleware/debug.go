package middleware

import (
	"bytes"
	"fmt"
	"html"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

type debugResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	written    bool
}

func (drw *debugResponseWriter) WriteHeader(code int) {
	if !drw.written {
		drw.statusCode = code
		drw.ResponseWriter.WriteHeader(code)
	}
}

func (drw *debugResponseWriter) Write(b []byte) (int, error) {
	// バッファにのみ書き込み、実際のレスポンスには書き込まない
	return drw.body.Write(b)
}

// AppError represents an application error with additional context
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Err        error
	StackTrace string
}

func NewAppError(code, message string, statusCode int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
		StackTrace: string(debug.Stack()),
	}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// DebugPanel adds debug information to HTML responses in development mode
func DebugPanel(isDev bool) func(http.Handler) http.Handler {
	if !isDev {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// パフォーマンス計測
			start := time.Now()

			// メモリ使用量の記録
			var memStatsBefore runtime.MemStats
			runtime.ReadMemStats(&memStatsBefore)

			// カスタムレスポンスライター
			drw := &debugResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
			}

			// リクエスト処理
			next.ServeHTTP(drw, r)

			// メモリ使用量の計算
			var memStatsAfter runtime.MemStats
			runtime.ReadMemStats(&memStatsAfter)

			duration := time.Since(start)

			// デバッグ情報をヘッダーに追加
			drw.Header().Set("X-Debug-Duration", duration.String())
			drw.Header().Set("X-Debug-Memory", fmt.Sprintf("%d KB", (memStatsAfter.Alloc-memStatsBefore.Alloc)/1024))
			drw.Header().Set("X-Debug-Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine()))

			// HTMLレスポンスの場合、デバッグパネルを挿入
			contentType := drw.Header().Get("Content-Type")
			if strings.Contains(contentType, "text/html") && drw.body.Len() > 0 {
				debugHTML := fmt.Sprintf(`
				<div id="debug-panel" style="position: fixed; bottom: 10px; right: 10px; background: rgba(0,0,0,0.8); color: #fff; padding: 10px; font-size: 12px; z-index: 9999; border-radius: 5px; font-family: monospace;">
					<div style="margin-bottom: 5px;"><strong>🐛 Debug Info</strong></div>
					<div>⏱️ Duration: %s</div>
					<div>🧠 Memory: %d KB</div>
					<div>📊 Status: %d</div>
					<div>🔄 Goroutines: %d</div>
					<div>🛤️ Path: %s</div>
					<div>📝 Method: %s</div>
					<button onclick="this.parentElement.remove()" style="background: #ff4444; color: white; border: none; padding: 2px 6px; border-radius: 3px; cursor: pointer; float: right; margin-top: 5px;">×</button>
				</div>
				`, 
				duration, 
				(memStatsAfter.Alloc-memStatsBefore.Alloc)/1024, 
				drw.statusCode, 
				runtime.NumGoroutine(),
				html.EscapeString(r.URL.Path),
				html.EscapeString(r.Method))

				// レスポンスボディに追加
				body := drw.body.String()
				if strings.Contains(body, "</body>") {
					body = strings.Replace(body, "</body>", debugHTML+"</body>", 1)
				} else if strings.Contains(body, "</html>") {
					body = strings.Replace(body, "</html>", debugHTML+"</html>", 1)
				} else {
					body += debugHTML
				}

				// コンテンツ長を更新
				drw.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
				// 実際のレスポンスに書き込む
				drw.ResponseWriter.Write([]byte(body))
			} else {
				// HTML以外の場合はそのまま出力
				if drw.body.Len() > 0 {
					drw.ResponseWriter.Write(drw.body.Bytes())
				}
			}
		})
	}
}

// ErrorHandler provides comprehensive error handling with environment-specific responses
func ErrorHandler(isDev bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					appErr := &AppError{
						Code:       "PANIC",
						Message:    "Internal server error",
						StatusCode: http.StatusInternalServerError,
						Err:        fmt.Errorf("%v", err),
						StackTrace: string(debug.Stack()),
					}

					// エラーログ
					fmt.Printf("PANIC: %s\nPath: %s %s\nStack: %s\n", 
						appErr.Error(), r.Method, r.URL.Path, appErr.StackTrace)

					// エラーレスポンス
					if r.Header.Get("HX-Request") == "true" {
						// HTMXエラーレスポンス
						w.Header().Set("HX-Retarget", "#error-container")
						w.Header().Set("HX-Reswap", "innerHTML")
					}

					w.WriteHeader(appErr.StatusCode)

					if isDev {
						// 開発環境では詳細を表示
						fmt.Fprintf(w, `
						<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
							<div class="flex items-center">
								<div class="flex-shrink-0">
									<span class="text-red-500 text-xl">⚠️</span>
								</div>
								<div class="ml-3">
									<h3 class="text-lg font-bold">Error: %s</h3>
									<p class="mt-1">%s</p>
									<p class="mt-2 text-xs text-red-500">スタックトレースはサーバーログを参照してください。</p>
								</div>
							</div>
						</div>
						`, appErr.Code, appErr.Message)
					} else {
						// 本番環境では一般的なメッセージ
						fmt.Fprintf(w, `
						<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
							<div class="flex items-center">
								<div class="flex-shrink-0">
									<span class="text-red-500 text-xl">⚠️</span>
								</div>
								<div class="ml-3">
									<p>エラーが発生しました。しばらく経ってから再度お試しください。</p>
								</div>
							</div>
						</div>
						`)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger logs HTTP requests with detailed information
func RequestLogger(isDev bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// カスタムレスポンスライター
			lrw := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// リクエスト処理
			next.ServeHTTP(lrw, r)

			// ログ出力
			duration := time.Since(start)
			
			logLevel := "INFO"
			if lrw.statusCode >= 400 {
				logLevel = "ERROR"
			}

			logEntry := fmt.Sprintf("[%s] %s %s %s %d %s",
				logLevel,
				start.Format("2006-01-02 15:04:05"),
				r.Method,
				r.URL.Path,
				lrw.statusCode,
				duration,
			)

			// 開発環境では詳細情報を追加
			if isDev {
				logEntry += fmt.Sprintf(" | UserAgent: %s | RemoteAddr: %s", 
					r.UserAgent(), r.RemoteAddr)
				
				if r.Header.Get("HX-Request") == "true" {
					logEntry += " | HTMX: true"
					if target := r.Header.Get("HX-Target"); target != "" {
						logEntry += fmt.Sprintf(" | Target: %s", target)
					}
				}
			}

			fmt.Println(logEntry)
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// HTMXDebugger returns JavaScript code for HTMX debugging
func HTMXDebugger() string {
	return `
	<script>
	// HTMXイベントのロギング（開発環境のみ）
	if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
		const htmxEvents = [
			'htmx:configRequest',
			'htmx:beforeRequest',
			'htmx:afterRequest',
			'htmx:responseError',
			'htmx:sendError',
			'htmx:timeout',
			'htmx:afterSettle',
			'htmx:afterSwap',
			'htmx:beforeSwap'
		];
		
		htmxEvents.forEach(event => {
			document.body.addEventListener(event, (e) => {
				console.group('🌐 HTMX Event: ' + event);
				console.log('Target:', e.detail.target);
				console.log('Detail:', e.detail);
				if (e.detail.xhr) {
					console.log('Status:', e.detail.xhr.status);
					console.log('Response:', e.detail.xhr.responseText.substring(0, 200) + '...');
				}
				console.groupEnd();
			});
		});
		
		// Alpine.jsのデバッグ
		document.addEventListener('alpine:init', () => {
			if (window.Alpine) {
				console.log('🏔️ Alpine.js initialized');
				
				// コンポーネント初期化のロギング
				window.Alpine.onBeforeComponentInit((component) => {
					console.log('🧩 Alpine Component Init:', component.$el, component.$data);
				});
			}
		});
		
		// パフォーマンス監視
		if (window.PerformanceObserver) {
			const observer = new PerformanceObserver((list) => {
				list.getEntries().forEach((entry) => {
					if (entry.entryType === 'navigation') {
						console.log('📊 Page Load Performance:', {
							'DOM Content Loaded': entry.domContentLoadedEventEnd - entry.domContentLoadedEventStart,
							'Load Complete': entry.loadEventEnd - entry.loadEventStart,
							'Total Time': entry.loadEventEnd - entry.fetchStart
						});
					}
				});
			});
			observer.observe({entryTypes: ['navigation']});
		}
	}
	</script>
	`
}

// DevToolsCSS returns CSS for development tools styling
func DevToolsCSS() string {
	return `
	<style>
	/* 開発環境用のデバッグスタイル */
	.debug-info {
		position: fixed;
		top: 10px;
		left: 10px;
		background: rgba(0, 0, 0, 0.8);
		color: white;
		padding: 10px;
		border-radius: 5px;
		font-family: monospace;
		font-size: 12px;
		z-index: 10000;
		max-width: 300px;
	}
	
	.debug-info h4 {
		margin: 0 0 5px 0;
		font-size: 14px;
	}
	
	.debug-info ul {
		margin: 0;
		padding: 0;
		list-style: none;
	}
	
	.debug-info li {
		margin: 2px 0;
	}
	
	/* HTMX要素のハイライト */
	[hx-get], [hx-post], [hx-put], [hx-delete], [hx-patch] {
		outline: 1px dashed #00f !important;
		outline-offset: 1px;
	}
	
	[hx-get]:hover, [hx-post]:hover, [hx-put]:hover, [hx-delete]:hover, [hx-patch]:hover {
		outline-color: #f00 !important;
		outline-width: 2px !important;
	}
	
	/* Alpine.js要素のハイライト */
	[x-data], [x-show], [x-if], [x-for] {
		outline: 1px dashed #0f0 !important;
		outline-offset: 1px;
	}
	
	[x-data]:hover, [x-show]:hover, [x-if]:hover, [x-for]:hover {
		outline-color: #ff0 !important;
		outline-width: 2px !important;
	}
	</style>
	`
}