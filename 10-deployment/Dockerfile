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
# Alpine の nodejs には corepack が含まれないため npm で入れる。pnpm 11 は Node 22 以上が必要
RUN npm install -g corepack@0.36.0 && corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

# ソースコードのコピー
COPY . .

# ビルド引数
ARG VERSION=dev
ARG BUILD_TIME=unknown

# Tailwind CSSのビルド
RUN pnpm run build:css

# Goアプリケーションのビルド（go-sqlite3 は cgo 必須）
RUN CGO_ENABLED=1 go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w" \
    -o server ./cmd/server

# 実行用の軽量イメージ
# cgo バイナリが musl に動的リンクするため builder と同じ Alpine 版にそろえる
FROM alpine:3.24@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

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