# マルチステージビルド
FROM golang:1.27-alpine3.24@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS builder

# 必要なツールのインストール
RUN apk add --no-cache \
    git \
    nodejs \
    npm \
    make \
    curl \
    gcc \
    musl-dev

WORKDIR /app

# Goの依存関係のキャッシュ
COPY go.mod go.sum ./
RUN go mod download

# Node.jsの依存関係のキャッシュ
COPY package*.json ./
RUN npm ci

# ソースコードのコピー
COPY . .

# ビルド引数
ARG VERSION=dev
ARG BUILD_TIME=unknown

# Tailwind CSSのビルド
RUN npm run build:css

# Goアプリケーションのビルド（go-sqlite3 は cgo 必須）
RUN CGO_ENABLED=1 go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w" \
    -o server ./cmd/server

# 実行用の軽量イメージ
# cgo バイナリが musl に動的リンクするため builder と同じ Alpine 版にそろえる
FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

# セキュリティアップデートとタイムゾーンデータ
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# 非rootユーザーの作成
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

# 実行ユーザーが書き込めるのはデータとログだけにし、バイナリと静的ファイルは root 所有のままにする
RUN mkdir -p data logs && \
    chown appuser:appgroup data logs

COPY --from=builder /app/server ./
COPY --from=builder /app/static ./static

# 実行ユーザーの切り替え
USER appuser

# 環境変数の設定
ENV ENV=production
ENV PORT=8080
ENV LOG_LEVEL=info
ENV DATABASE_URL=sqlite:///app/data/deployment_demo.db

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/server", "-health-check"]

# ポートの公開
EXPOSE 8080

# エントリーポイント
ENTRYPOINT ["/app/server"]

# メタデータラベル
ARG VERSION=dev
LABEL maintainer="deployment-demo"
LABEL version="${VERSION}"
LABEL description="Modern Web App with Go/HTMX/Alpine.js/Tailwind CSS"