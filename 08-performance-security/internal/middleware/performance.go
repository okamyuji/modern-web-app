package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GzipResponseWriter - gzip圧縮レスポンスライター
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Gzip - gzip圧縮ミドルウェア
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// gzipをサポートしているか確認
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// HTMXの部分更新は圧縮しない（小さいため）
		if r.Header.Get("HX-Request") == "true" {
			next.ServeHTTP(w, r)
			return
		}

		// 既に圧縮されているファイルはスキップ
		if isCompressedContent(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzip.NewWriter(w)
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length") // 圧縮後のサイズは不明

		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzw, r)
	})
}

// isCompressedContent - 既に圧縮されているコンテンツかチェック
func isCompressedContent(path string) bool {
	compressedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".zip", ".gz", ".br"}
	for _, ext := range compressedExts {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}
	return false
}

// Cache - キャッシュミドルウェア
func Cache(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 静的リソースのみキャッシュ
			if strings.HasPrefix(r.URL.Path, "/static/") {
				w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(duration.Seconds())))
				w.Header().Set("Vary", "Accept-Encoding")
				
				// ETags for better caching
				if etag := r.Header.Get("If-None-Match"); etag != "" {
					w.Header().Set("ETag", etag)
					w.WriteHeader(http.StatusNotModified)
					return
				}
			} else if r.Header.Get("HX-Request") == "true" {
				// HTMXリクエストはキャッシュしない
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestSizeLimit - リクエストサイズ制限
func RequestSizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// ResponseHeaders - レスポンスヘッダー最適化
func ResponseHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Keep-Alive設定
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Keep-Alive", "timeout=5, max=1000")

		// HTMX専用ヘッダーの最適化
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Location", r.URL.Path)
		}

		next.ServeHTTP(w, r)
	})
}

// StaticFileOptimizer - 静的ファイル最適化
func StaticFileOptimizer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			// ファイルタイプに応じたContent-Type設定
			ext := strings.ToLower(r.URL.Path[strings.LastIndex(r.URL.Path, "."):])
			switch ext {
			case ".css":
				w.Header().Set("Content-Type", "text/css; charset=utf-8")
			case ".js":
				w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			case ".svg":
				w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
			case ".woff2":
				w.Header().Set("Content-Type", "font/woff2")
			case ".woff":
				w.Header().Set("Content-Type", "font/woff")
			}

			// 長期キャッシュの設定
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}

		next.ServeHTTP(w, r)
	})
}