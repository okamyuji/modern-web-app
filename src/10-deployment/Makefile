.PHONY: build clean test deploy docker-build help dev setup lint fmt security

# 変数定義
APP_NAME := deployment-demo
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev-$(shell date +%Y%m%d-%H%M%S)")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)

# Go関連
GO_VERSION := 1.24.3
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
BINARY_DIR := bin
BINARY_PATH := $(BINARY_DIR)/$(APP_NAME)

# Tailwind CSS関連
CSS_INPUT := static/css/input.css
CSS_OUTPUT := static/css/app.css

# Docker関連
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := $(VERSION)

# デプロイ関連
DEPLOY_USER := deploy
DEPLOY_HOST := localhost
DEPLOY_PATH := /var/www/$(APP_NAME)

help: ## このヘルプを表示
	@echo "使用可能なコマンド:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

setup: ## 開発環境のセットアップ
	@echo "開発環境をセットアップ中..."
	go mod download
	@if command -v npm >/dev/null 2>&1; then \
		npm install; \
	else \
		echo "npmが見つかりません。Node.jsをインストールしてください。"; \
	fi
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env ファイルを作成しました。必要に応じて編集してください。"; \
	fi
	@echo "セットアップ完了"

dev: ## 開発サーバーを起動
	@echo "開発サーバーを起動中..."
	@if command -v templ >/dev/null 2>&1; then \
		templ generate; \
	else \
		echo "templがインストールされていません。go install github.com/a-h/templ@latest でインストールしてください。"; \
	fi
	ENV=development go run cmd/server/main.go

css-dev: ## Tailwind CSSを開発モードでビルド（ウォッチモード）
	@if command -v npx >/dev/null 2>&1; then \
		npx tailwindcss -i $(CSS_INPUT) -o $(CSS_OUTPUT) --watch; \
	else \
		echo "npxが見つかりません。Node.jsをインストールしてください。"; \
	fi

css-build: ## Tailwind CSSを本番用にビルド
	@echo "Tailwind CSSをビルド中..."
	@if command -v npx >/dev/null 2>&1; then \
		NODE_ENV=production npx tailwindcss -i $(CSS_INPUT) -o $(CSS_OUTPUT) --minify; \
	else \
		echo "npxが見つかりません。CSSの最適化をスキップします。"; \
	fi

templ-generate: ## Templテンプレートを生成
	@echo "Templテンプレートを生成中..."
	@if command -v templ >/dev/null 2>&1; then \
		templ generate; \
	else \
		echo "templがインストールされていません。go install github.com/a-h/templ@latest でインストールしてください。"; \
	fi

fmt: ## Goコードのフォーマット
	@echo "Goコードをフォーマット中..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w $(GO_FILES); \
	fi

lint: ## コードの静的解析
	@echo "コードを解析中..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		go vet ./...; \
		echo "より詳細な解析にはgolangci-lintをインストールしてください。"; \
	fi

security: ## セキュリティスキャン
	@echo "セキュリティスキャン中..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosecがインストールされていません。go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest でインストールしてください。"; \
	fi

test: ## テストを実行
	@echo "テストを実行中..."
	go test -v ./...

test-coverage: ## テストカバレッジを計測
	@echo "テストカバレッジを計測中..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "カバレッジレポートが coverage.html に生成されました"

benchmark: ## ベンチマークテストを実行
	@echo "ベンチマークテストを実行中..."
	go test -bench=. -benchmem ./...

build-go: ## Goアプリケーションをビルド
	@echo "Goアプリケーションをビルド中... ($(VERSION))"
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS) -s -w" \
		-o $(BINARY_PATH) \
		./cmd/server

optimize-static: ## 静的ファイルを最適化
	@echo "静的ファイルを最適化中..."
	@if command -v html-minifier >/dev/null 2>&1; then \
		find ./internal/templates -name "*.html" -exec html-minifier \
			--collapse-whitespace \
			--remove-comments \
			--minify-css \
			--minify-js \
			{} -o {} \; 2>/dev/null || true; \
	fi
	@if command -v imagemin >/dev/null 2>&1; then \
		find ./static/images -name "*.jpg" -o -name "*.png" | \
			xargs -I {} imagemin {} --out-dir=./static/images 2>/dev/null || true; \
	fi

build: test templ-generate css-build build-go optimize-static ## 本番用ビルド（テスト→CSS→Go→最適化）
	@echo "ビルド完了: $(VERSION)"
	@echo "バイナリ: $(BINARY_PATH)"
	@ls -lh $(BINARY_PATH)

clean: ## ビルド成果物をクリーンアップ
	@echo "クリーンアップ中..."
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html
	rm -f $(CSS_OUTPUT)
	go clean -cache
	@echo "クリーンアップ完了"

docker-build: ## Dockerイメージをビルド
	@echo "Dockerイメージをビルド中... ($(DOCKER_TAG))"
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		.
	@echo "Dockerイメージビルド完了"

docker-run: ## Dockerコンテナを起動
	@echo "Dockerコンテナを起動中..."
	docker run --rm -p 8080:8080 \
		-e ENV=production \
		$(DOCKER_IMAGE):latest

docker-push: docker-build ## Dockerイメージをプッシュ
	@echo "Dockerイメージをプッシュ中..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

health-check: ## ヘルスチェックを実行
	@echo "ヘルスチェック中..."
	@if curl -f -s http://localhost:8080/health >/dev/null; then \
		echo "✅ アプリケーションは正常に動作しています"; \
	else \
		echo "❌ アプリケーションにアクセスできません"; \
		exit 1; \
	fi

deploy-check: ## デプロイ前チェック
	@echo "デプロイ前チェック中..."
	@echo "バージョン: $(VERSION)"
	@echo "ビルド時刻: $(BUILD_TIME)"
	@if [ ! -f $(BINARY_PATH) ]; then \
		echo "❌ バイナリが見つかりません。'make build' を実行してください"; \
		exit 1; \
	fi
	@echo "✅ デプロイ準備完了"

deploy: deploy-check ## 本番環境にデプロイ
	@echo "本番環境にデプロイ中..."
	@echo "⚠️  このコマンドは実際のデプロイ環境に合わせて調整してください"
	# ./scripts/deploy.sh $(VERSION)

install-tools: ## 開発ツールをインストール
	@echo "開発ツールをインストール中..."
	go install github.com/a-h/templ@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@if command -v npm >/dev/null 2>&1; then \
		npm install -g html-minifier imagemin-cli; \
	fi
	@echo "開発ツールのインストール完了"

version: ## バージョン情報を表示
	@echo "アプリケーション: $(APP_NAME)"
	@echo "バージョン: $(VERSION)"
	@echo "ビルド時刻: $(BUILD_TIME)"
	@echo "Goバージョン: $(GO_VERSION)"

# 開発環境用のタスク
.PHONY: dev-deps dev-watch

dev-deps: ## 開発依存関係のインストール
	@echo "開発依存関係をインストール中..."
	go mod tidy
	@if command -v npm >/dev/null 2>&1; then \
		npm install --include=dev; \
	fi

dev-watch: ## ファイル変更を監視して自動リビルド
	@echo "ファイル監視を開始中..."
	@echo "Ctrl+C で停止"
	@if command -v fswatch >/dev/null 2>&1; then \
		fswatch -o . --exclude='.*\.(git|node_modules|bin).*' | xargs -n1 -I{} make templ-generate build-go; \
	else \
		echo "fswatch がインストールされていません。手動でリビルドしてください。"; \
	fi

# CI/CD用のタスク
ci-test: ## CI環境でのテスト実行
	@echo "CI環境でテストを実行中..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

ci-build: ## CI環境でのビルド
	@echo "CI環境でビルド中..."
	make templ-generate
	make build-go