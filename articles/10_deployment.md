---
title: "第10章 デプロイメントと運用"
emoji: "😸" 
type: "tech" 
topics: ["golang","go","alpinejs","htmx"] 
published: false
---

# 第10章 デプロイメントと運用

## 1. 本番環境への準備

### ビルドとリリース戦略

Golang/HTMX/Alpine.js/Tailwind CSSアプリケーションを本番環境にデプロイする際は、最適化されたビルドプロセスが不可欠です。

```makefile
# Makefile
.PHONY: build clean test deploy

# 変数定義
APP_NAME := myapp
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)

# 開発環境のセットアップ
setup:
    go mod download
    npm install
    cp .env.example .env

# Tailwind CSSのビルド
css-dev:
    npx tailwindcss -i ./assets/input.css -o ./static/css/app.css --watch

css-build:
    NODE_ENV=production npx tailwindcss -i ./assets/input.css -o ./static/css/app.css --minify

# Goアプリケーションのビルド
build-go:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags "$(LDFLAGS) -s -w" \
        -o bin/$(APP_NAME) \
        ./cmd/server

# 静的ファイルの最適化
optimize-static:
    # HTMLの圧縮
    find ./templates -name "*.html" -exec html-minifier \
        --collapse-whitespace \
        --remove-comments \
        --minify-css \
        --minify-js \
        {} -o {} \;
    
    # 画像の最適化
    find ./static/images -name "*.jpg" -o -name "*.png" | \
        xargs -I {} imagemin {} --out-dir=./static/images

# 本番ビルド
build: test css-build build-go optimize-static
    @echo "Build complete: $(VERSION)"

# Dockerイメージのビルド
docker-build:
    docker build \
        --build-arg VERSION=$(VERSION) \
        --build-arg BUILD_TIME=$(BUILD_TIME) \
        -t $(APP_NAME):$(VERSION) \
        -t $(APP_NAME):latest \
        .

# ヘルスチェック
health-check:
    @curl -f http://localhost:8080/health || exit 1
```

**💡 ビルドの最適化:** `-ldflags "-s -w"`でバイナリサイズを削減し、`CGO_ENABLED=0`で静的バイナリを生成します。これにより、Dockerイメージのサイズが大幅に削減されます。

### Dockerコンテナ化

```dockerfile
# Dockerfile
# マルチステージビルド
FROM golang:1.21-alpine AS builder

# 必要なツールのインストール
RUN apk add --no-cache git nodejs npm

WORKDIR /app

# 依存関係のキャッシュ
COPY go.mod go.sum ./
RUN go mod download

COPY package.json package-lock.json ./
RUN npm ci

# ソースコードのコピー
COPY . .

# ビルド引数
ARG VERSION
ARG BUILD_TIME

# アプリケーションのビルド
RUN npm run build:css
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w" \
    -o server ./cmd/server

# 実行用の軽量イメージ
FROM alpine:latest

# セキュリティアップデート
RUN apk --no-cache add ca-certificates tzdata

# 非rootユーザーの作成
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

# 必要なファイルのコピー
COPY --from=builder --chown=appuser:appgroup /app/server .
COPY --from=builder --chown=appuser:appgroup /app/templates ./templates
COPY --from=builder --chown=appuser:appgroup /app/static ./static

# 実行ユーザーの切り替え
USER appuser

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/server", "-health-check"]

EXPOSE 8080

ENTRYPOINT ["/app/server"]
```

**⚠️ セキュリティの注意点:** 本番環境では必ず非rootユーザーで実行し、最小権限の原則を守りましょう。また、定期的にベースイメージを更新してセキュリティパッチを適用します。

## 2. インフラストラクチャと設定

### リバースプロキシの設定

```nginx
# nginx.conf
upstream app_backend {
    server app1:8080 weight=1 max_fails=3 fail_timeout=30s;
    server app2:8080 weight=1 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

server {
    listen 80;
    server_name example.com;
    
    # HTTPSへのリダイレクト
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name example.com;
    
    # SSL設定
    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    
    # セキュリティヘッダー
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    
    # 静的ファイルの配信
    location /static/ {
        alias /var/www/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
        
        # Brotli圧縮
        brotli on;
        brotli_comp_level 6;
        brotli_types text/css application/javascript image/svg+xml;
    }
    
    # WebSocket/SSE対応
    location /chat/stream {
        proxy_pass http://app_backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        
        # SSE用の設定
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 86400s;
        
        # Nginxのバッファリングを無効化
        proxy_set_header X-Accel-Buffering no;
    }
    
    # アプリケーションへのプロキシ
    location / {
        proxy_pass http://app_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Keep-alive接続
        proxy_set_header Connection "";
        
        # タイムアウト設定
        proxy_connect_timeout 5s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # バッファサイズ
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;
    }
    
    # レート制限
    limit_req_zone $binary_remote_addr zone=general:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=api:10m rate=5r/s;
    
    location /api/ {
        limit_req zone=api burst=10 nodelay;
        proxy_pass http://app_backend;
    }
}
```

**💡 パフォーマンスのポイント:** Keep-alive接続を有効にし、静的ファイルには長期キャッシュを設定します。Brotli圧縮はgzipよりも効率的です。

### 環境変数と設定管理

```go
// config/config.go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
    
    "github.com/joho/godotenv"
)

type Config struct {
    // サーバー設定
    Port            string
    Environment     string
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    IdleTimeout     time.Duration
    ShutdownTimeout time.Duration
    
    // データベース設定
    DatabaseURL     string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    
    // セキュリティ設定
    SessionSecret   string
    CSRFSecret      string
    AllowedOrigins  []string
    
    // 外部サービス
    RedisURL        string
    SMTPHost        string
    SMTPPort        int
    SMTPUser        string
    SMTPPassword    string
    
    // 監視
    SentryDSN       string
    LogLevel        string
}

func Load() (*Config, error) {
    // 環境に応じて.envファイルを読み込み
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }
    
    envFile := fmt.Sprintf(".env.%s", env)
    if err := godotenv.Load(envFile); err != nil {
        // ファイルが存在しない場合は環境変数のみを使用
        if !os.IsNotExist(err) {
            return nil, err
        }
    }
    
    config := &Config{
        Port:            getEnv("PORT", "8080"),
        Environment:     env,
        ReadTimeout:     getDuration("READ_TIMEOUT", 10*time.Second),
        WriteTimeout:    getDuration("WRITE_TIMEOUT", 10*time.Second),
        IdleTimeout:     getDuration("IDLE_TIMEOUT", 60*time.Second),
        ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
        
        DatabaseURL:     getEnvRequired("DATABASE_URL"),
        MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
        MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 5),
        ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
        
        SessionSecret:  getEnvRequired("SESSION_SECRET"),
        CSRFSecret:     getEnvRequired("CSRF_SECRET"),
        AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"https://example.com"}),
        
        RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379"),
        SMTPHost:     getEnv("SMTP_HOST", ""),
        SMTPPort:     getInt("SMTP_PORT", 587),
        SMTPUser:     getEnv("SMTP_USER", ""),
        SMTPPassword: getEnv("SMTP_PASSWORD", ""),
        
        SentryDSN: getEnv("SENTRY_DSN", ""),
        LogLevel:  getEnv("LOG_LEVEL", "info"),
    }
    
    // 設定の検証
    if err := config.Validate(); err != nil {
        return nil, err
    }
    
    return config, nil
}

func (c *Config) Validate() error {
    if c.Environment == "production" {
        // 本番環境での必須チェック
        if c.SessionSecret == "default-secret" {
            return fmt.Errorf("SESSION_SECRET must be changed in production")
        }
        if c.CSRFSecret == "default-secret" {
            return fmt.Errorf("CSRF_SECRET must be changed in production")
        }
        if c.LogLevel == "debug" {
            return fmt.Errorf("LOG_LEVEL should not be 'debug' in production")
        }
    }
    return nil
}
```

**⚠️ 設定管理の重要性:** 環境変数は12 Factor Appの原則に従い、設定とコードを分離します。秘密情報は絶対にコードにハードコーディングしないでください。

## 3. 監視と運用

### アプリケーションメトリクス

```go
// monitoring/metrics.go
package monitoring

import (
    "net/http"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    // HTTPメトリクス
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
    
    // ビジネスメトリクス
    todosCreated = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "todos_created_total",
            Help: "Total number of todos created",
        },
    )
    
    activeUsers = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_users",
            Help: "Number of active users",
        },
    )
)

func init() {
    // メトリクスの登録
    prometheus.MustRegister(
        httpRequestsTotal,
        httpRequestDuration,
        todosCreated,
        activeUsers,
    )
}

// Prometheusミドルウェア
func PrometheusMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(
            r.Method,
            r.URL.Path,
        ))
        
        wrapped := &statusWriter{ResponseWriter: w}
        next.ServeHTTP(wrapped, r)
        
        timer.ObserveDuration()
        
        httpRequestsTotal.WithLabelValues(
            r.Method,
            r.URL.Path,
            strconv.Itoa(wrapped.status),
        ).Inc()
    })
}

// グレースフルシャットダウン
func GracefulShutdown(server *http.Server, config *Config) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    
    <-quit
    log.Println("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
    defer cancel()
    
    // 新しい接続の受付を停止
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Server forced to shutdown: %v", err)
    }
    
    // データベース接続のクローズ
    if db != nil {
        db.Close()
    }
    
    log.Println("Server shutdown complete")
}

// デプロイメントスクリプト
#!/bin/bash
# deploy.sh

set -e

# 変数
APP_NAME="myapp"
DEPLOY_USER="deploy"
DEPLOY_HOST="example.com"
DEPLOY_PATH="/var/www/${APP_NAME}"

echo "Starting deployment..."

# ビルド
make build

# ヘルスチェック用の一時ポート
TEMP_PORT=8081

# サーバーへのデプロイ
ssh ${DEPLOY_USER}@${DEPLOY_HOST} << EOF
    cd ${DEPLOY_PATH}
    
    # バックアップ
    cp ${APP_NAME} ${APP_NAME}.backup
    
    # 新しいバイナリをコピー
    # (実際はrsyncやscpを使用)
    
    # 新しいインスタンスを一時ポートで起動
    ./${APP_NAME} -port=${TEMP_PORT} &
    NEW_PID=\$!
    
    # ヘルスチェック
    sleep 5
    if curl -f http://localhost:${TEMP_PORT}/health; then
        # 古いプロセスにSIGTERMを送信
        pkill -TERM -f "${APP_NAME} -port=8080" || true
        
        # 新しいプロセスを本番ポートに切り替え
        kill -TERM \$NEW_PID
        ./${APP_NAME} -port=8080 &
        
        echo "Deployment successful"
    else
        # ロールバック
        kill -9 \$NEW_PID
        echo "Deployment failed, keeping old version"
        exit 1
    fi
EOF
```

**💡 運用のベストプラクティス:** ブルーグリーンデプロイメントやローリングアップデートを使用して、ダウンタイムなしでデプロイを行います。必ずヘルスチェックで新バージョンの動作を確認してから切り替えましょう。

## 復習問題

1. マルチステージDockerビルドを使用する利点を3つ挙げてください。

2. 以下の設定に潜在的なセキュリティリスクがあります。何が問題で、どう修正すべきですか？

    ```go
    func LoadConfig() *Config {
        return &Config{
            SessionSecret: "my-secret-key",
            DatabaseURL:   "postgres://user:pass@localhost/db",
        }
    }
    ```

3. グレースフルシャットダウンが重要な理由と、その実装方法を説明してください。

## 模範解答

1. マルチステージビルドの利点
   - 最終イメージサイズの削減（ビルドツールを含まない）
   - セキュリティの向上（ソースコードや開発ツールが本番イメージに含まれない）
   - ビルドキャッシュの効率化（依存関係の層を分離）

2. 修正版

    ```go
    func LoadConfig() (*Config, error) {
        sessionSecret := os.Getenv("SESSION_SECRET")
        if sessionSecret == "" {
            return nil, errors.New("SESSION_SECRET is required")
        }
        
        dbURL := os.Getenv("DATABASE_URL")
        if dbURL == "" {
            return nil, errors.New("DATABASE_URL is required")
        }
        
        return &Config{
            SessionSecret: sessionSecret,
            DatabaseURL:   dbURL,
        }, nil
    }
    ```

3. グレースフルシャットダウンの重要性
   - 処理中のリクエストを完了させてデータロスを防ぐ
   - データベース接続を適切にクローズしてリソースリークを防ぐ
   - 実装：SIGTERMシグナルを受信→新規接続の受付停止→既存接続の完了待機→リソースのクリーンアップ
