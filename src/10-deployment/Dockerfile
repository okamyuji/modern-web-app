# マルチステージビルド
FROM golang:1.24.3-alpine AS builder

# 必要なツールのインストール
RUN apk add --no-cache \
    git \
    nodejs \
    npm \
    make \
    curl

WORKDIR /app

# Goの依存関係のキャッシュ
COPY go.mod go.sum ./
RUN go mod download

# Node.jsの依存関係のキャッシュ
COPY package*.json ./
RUN npm ci --only=production

# ソースコードのコピー
COPY . .

# ビルド引数
ARG VERSION=dev
ARG BUILD_TIME=unknown

# Templテンプレートの生成
RUN go install github.com/a-h/templ@latest
RUN templ generate

# Tailwind CSSのビルド
RUN npm run build:css || echo "npm script not found, skipping CSS build"

# Goアプリケーションのビルド
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w" \
    -o server ./cmd/server

# 実行用の軽量イメージ
FROM alpine:3.19

# セキュリティアップデートとタイムゾーンデータ
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# 非rootユーザーの作成
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

# 必要なディレクトリの作成
RUN mkdir -p static logs && \
    chown -R appuser:appgroup /app

# 必要なファイルのコピー
COPY --from=builder --chown=appuser:appgroup /app/server ./
COPY --from=builder --chown=appuser:appgroup /app/internal/templates ./internal/templates
COPY --from=builder --chown=appuser:appgroup /app/static ./static

# 実行権限の設定
RUN chmod +x ./server

# 実行ユーザーの切り替え
USER appuser

# 環境変数の設定
ENV ENV=production
ENV PORT=8080
ENV LOG_LEVEL=info

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/server", "-health-check"] || exit 1

# ポートの公開
EXPOSE 8080

# エントリーポイント
ENTRYPOINT ["/app/server"]

# メタデータラベル
LABEL maintainer="deployment-demo"
LABEL version="${VERSION}"
LABEL description="Modern Web App with Go/HTMX/Alpine.js/Tailwind CSS"