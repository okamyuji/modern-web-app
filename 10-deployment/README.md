# 第10章 デプロイメント・運用機能デモアプリケーション

本章では、モダンなWebアプリケーションの本番運用に必要なデプロイメント戦略、監視機能、設定管理、コンテナ化を実装したデモアプリケーションを紹介します。

## 📋 目次

- [🎯 機能概要](#🎯 機能概要)
- [🏗️ アーキテクチャ](#🏗️ アーキテクチャ)
- [🚀 セットアップ](#🚀 セットアップ)
- [🔨 ビルドとデプロイ](#🔨 ビルドとデプロイ)
- [📊 監視とメトリクス](#📊 監視とメトリクス)
- [⚙️ 設定管理](#⚙️ 設定管理)
- [🚀 デプロイメント戦略](#🚀 デプロイメント戦略)
- [🐳 Docker活用](#🐳 Docker活用)
- [📚 技術的解説](#📚 技術的解説)

## 🎯 機能概要

### デプロイメント機能

- **Blue-Greenデプロイメント**: ダウンタイムゼロでのアプリケーション更新
- **ローリングアップデート**: 複数インスタンス環境での段階的更新
- **自動ロールバック**: デプロイ失敗時の自動復旧機能
- **ヘルスチェック**: デプロイ成功/失敗の自動判定

### 監視・メトリクス

- **Prometheusメトリクス**: HTTPリクエスト、レスポンス時間、システムリソースの監視
- **ビジネスメトリクス**: Todo作成数、完了数などアプリケーション固有の指標
- **ヘルスチェックAPI**: アプリケーションの生存状態監視
- **構造化ログ**: JSON形式での統一されたログ出力

### 設定管理

- **環境別設定**: 開発/ステージング/本番環境の設定分離
- **設定バリデーション**: 本番環境での必須設定チェック
- **シークレット管理**: セキュリティ情報の適切な管理
- **12-Factor App**: 設定の環境変数化による移植性向上

### コンテナ化

- **マルチステージビルド**: 最適化されたDockerイメージ
- **セキュリティ強化**: 非rootユーザーでの実行
- **ヘルスチェック統合**: Dockerレベルでのヘルスチェック
- **レイヤー最適化**: イメージサイズの最小化

## 🏗️ アーキテクチャ

```text
deployment-demo/
├── cmd/server/           # アプリケーションエントリーポイント
├── internal/
│   ├── config/          # 設定管理
│   └── monitoring/      # メトリクス・監視
├── scripts/             # デプロイメントスクリプト
├── static/              # 静的ファイル
├── Dockerfile           # コンテナ定義
├── Makefile            # ビルド自動化
└── .env.example        # 設定テンプレート
```

### 主要コンポーネント

#### 設定管理 (`internal/config/`)

- 環境変数ベースの設定読み込み
- 環境別設定ファイル（.env.development、.env.production）
- 設定値バリデーション
- デフォルト値管理

#### 監視システム (`internal/monitoring/`)

- Prometheusメトリクス収集
- HTTPリクエスト・レスポンス監視
- システムリソース監視
- グレースフルシャットダウン

#### デプロイメントシステム (`scripts/`)

- Blue-Greenデプロイメント
- ヘルスチェック統合
- 自動ロールバック
- ログ・バックアップ管理

## 🚀 セットアップ

### 前提条件

- Go 1.24.3以上
- Node.js 18以上（Tailwind CSS用）
- Make
- Docker（コンテナ化する場合）

### 初期セットアップ

```bash
# リポジトリのクローンと移動
git clone <repository-url>
cd src/10-deployment

# 開発環境のセットアップ
make setup

# 設定ファイルの編集
cp .env.example .env
# .envファイルを環境に合わせて編集
```

### 依存関係のインストール

```bash
# Go依存関係
go mod download

# Node.js依存関係
npm install

# 必要なツールのインストール
go install github.com/a-h/templ@latest
```

## 🔨 ビルドとデプロイ

### 開発環境での実行

```bash
# 開発サーバーの起動
make dev

# または直接実行
ENV=development PORT=8080 go run cmd/server/main.go
```

### 本番ビルド

```bash
# 完全ビルド（テスト→CSS→Go→最適化）
make build

# 個別ビルド
make test           # テスト実行
make css-build      # CSS最適化ビルド
make build-go       # Goアプリケーションビルド
```

### デプロイメント

```bash
# Blue-Greenデプロイメント
./scripts/deploy.sh v1.0.0 production

# ローリングアップデート
DEPLOY_STRATEGY=rolling ./scripts/deploy.sh v1.0.0 production

# ロールバック
./scripts/deploy.sh --rollback
```

## 📊 監視とメトリクス

### Prometheusメトリクス

アプリケーションは以下のメトリクスを提供します。

#### HTTPメトリクス

- `http_requests_total`: リクエスト総数（method, endpoint, status_code別）
- `http_request_duration_seconds`: リクエスト処理時間
- `http_request_size_bytes`: リクエストサイズ
- `http_response_size_bytes`: レスポンスサイズ

#### ビジネスメトリクス

- `todos_created_total`: Todo作成数（priority別）
- `todos_completed_total`: Todo完了数
- `active_users_current`: 現在のアクティブユーザー数

#### システムメトリクス

- `database_connections`: データベース接続数（状態別）
- `database_operations_total`: データベース操作数
- `app_info`: アプリケーション情報（バージョン、ビルド時刻）

### エンドポイント

```bash
# ヘルスチェック
curl http://localhost:8080/health

# メトリクス（JSON形式）
curl http://localhost:8080/metrics

# Prometheusメトリクス
curl http://localhost:9090/metrics
```

## ⚙️ 設定管理

### 環境変数

アプリケーションは以下の環境変数で設定可能です：

#### 基本設定

```bash
APP_NAME=deployment-demo
ENV=development              # development, staging, production
PORT=8080
VERSION=v1.0.0
```

#### セキュリティ設定

```bash
SESSION_SECRET=<64文字のランダム文字列>
CSRF_SECRET=<32文字のランダム文字列>
TLS_ENABLED=false
```

#### データベース設定

```bash
DATABASE_URL=sqlite://deployment_demo.db
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
```

#### 監視設定

```bash
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
LOG_LEVEL=info
```

### 設定ファイル

優先度順に以下のファイルから設定を読み込みます：

1. `.env.{環境}.local`（例：`.env.production.local`）
2. `.env.{環境}`（例：`.env.production`）
3. `.env.local`
4. `.env`

### 本番環境での注意点

本番環境では以下の設定が必須です：

- `SESSION_SECRET`: デフォルト値から変更必須
- `CSRF_SECRET`: デフォルト値から変更必須
- `LOG_LEVEL`: `debug`以外を推奨
- `TLS_ENABLED=true`: HTTPS環境での運用推奨

## 🚀 デプロイメント戦略

### Blue-Greenデプロイメント

**特徴:**

- ダウンタイムゼロ
- 即座にロールバック可能
- リソース消費量が多い

**プロセス:**

1. 新バージョンを一時ポートで起動
2. ヘルスチェック実行
3. 成功時に本番ポートに切り替え
4. 旧バージョンを停止

```bash
# Blue-Greenデプロイメント実行
DEPLOY_STRATEGY=blue-green ./scripts/deploy.sh v1.0.0 production
```

### ローリングアップデート

**特徴:**

- リソース効率が良い
- 段階的な更新
- 部分的な障害の影響を限定

**プロセス:**

1. 複数インスタンスを順次更新
2. ロードバランサーから一時的に除外
3. 更新完了後にトラフィック復帰

```bash
# ローリングアップデート実行
DEPLOY_STRATEGY=rolling ./scripts/deploy.sh v1.0.0 production
```

### 自動ロールバック

デプロイメント中にヘルスチェックが失敗した場合、自動的に前のバージョンにロールバックします。

**トリガー条件:**

- ヘルスチェックタイムアウト
- アプリケーション起動失敗
- 重要なエラーレスポンス

## 🐳 Docker活用

### マルチステージビルド

効率的なDockerイメージを作成するため、マルチステージビルドを採用：

```dockerfile
# ビルダーステージ
FROM golang:1.24.3-alpine AS builder
# 依存関係とソースコードのコピー
# ビルド実行

# 本番ステージ
FROM alpine:3.19
# 最小限のファイルのみコピー
# 非rootユーザーで実行
```

### Docker操作

```bash
# イメージビルド
make docker-build

# コンテナ実行
make docker-run

# 本番用イメージビルド
make docker-build-prod

# イメージのプッシュ
make docker-push
```

### セキュリティ対策

- 非rootユーザー（appuser）での実行
- 最小限のベースイメージ（Alpine Linux）
- セキュリティアップデートの適用
- 不要なファイルの除外

## 📚 技術的解説

### 1. 設定管理の実装

```go
// internal/config/config.go
type Config struct {
    AppName    string
    Port       string
    Env        string
    Version    string
    // ... その他の設定
}

func Load() (*Config, error) {
    // 環境変数とファイルからの設定読み込み
    // バリデーション実行
}
```

**ポイント:**

- 環境変数の優先的な使用
- デフォルト値の適切な設定
- 本番環境での厳格なバリデーション

### 2. メトリクス収集

```go
// internal/monitoring/metrics.go
func PrometheusMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        // リクエスト処理
        next.ServeHTTP(sw, r)
        // メトリクス記録
        httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
    })
}
```

**ポイント:**

- ミドルウェアパターンによる透過的な計測
- ラベルによるメトリクスの分類
- パフォーマンスへの影響を最小化

### 3. グレースフルシャットダウン

```go
func GracefulShutdown(server *http.Server, cfg *config.Config, cleanup func()) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit
    
    ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
    defer cancel()
    
    server.Shutdown(ctx)
}
```

**ポイント:**

- シグナルハンドリング
- タイムアウト付きシャットダウン
- リソースのクリーンアップ

### 4. デプロイメントスクリプト

```bash
# scripts/deploy.sh
blue_green_deploy() {
    # 新インスタンス起動
    # ヘルスチェック
    # 切り替え実行
    # 旧インスタンス停止
}
```

**ポイント:**

- 段階的なデプロイメント
- 各段階でのエラーハンドリング
- ロールバック機能の実装

## 🛠️ 開発とデバッグ

### ローカル開発

```bash
# 開発サーバー起動（ホットリロード対応）
make dev

# CSS監視モード
make css-dev

# テスト実行
make test

# コード品質チェック
make lint
```

### デバッグ機能

開発環境では以下のデバッグ機能が利用可能：

- リクエスト処理時間の表示
- メモリ使用量の監視
- SQLクエリのロギング
- デバッグパネルの表示

### ログ出力

```bash
# アプリケーションログ
tail -f logs/app.log

# エラーログ
tail -f logs/error.log

# デプロイメントログ
tail -f logs/deploy.log
```

## 📈 パフォーマンス最適化

### ビルド最適化

- Go: `-ldflags \"-s -w\"` によるバイナリサイズ削減
- CSS: Tailwind CSS の未使用クラス除去
- 静的ファイル: gzip圧縮による転送サイズ削減

### 実行時最適化

- データベース接続プール
- HTTP Keep-Alive
- 適切なタイムアウト設定
- メモリ使用量の監視

## 🔒 セキュリティ

### 実装済みセキュリティ機能

- CSRF保護
- XSS防止（Content Security Policy）
- セキュアヘッダーの設定
- 入力値検証
- SQLインジェクション対策

### 本番環境での推奨事項

- HTTPS（TLS）の有効化
- セキュリティヘッダーの設定
- ファイアウォール設定
- 定期的なセキュリティアップデート
- 監査ログの記録

## 📋 運用チェックリスト

### デプロイ前

- [ ] 全テストの通過確認
- [ ] 設定ファイルの確認
- [ ] シークレット情報の設定
- [ ] バックアップの実行
- [ ] ロールバック手順の確認

### デプロイ後

- [ ] ヘルスチェックの確認
- [ ] メトリクスの監視開始
- [ ] ログの確認
- [ ] 主要機能の動作確認
- [ ] パフォーマンス指標の確認

### 定期運用

- [ ] ログローテーション
- [ ] バックアップの確認
- [ ] セキュリティアップデート
- [ ] メトリクス分析
- [ ] 容量監視

## 🎯 まとめ

本デモアプリケーションでは、モダンなWebアプリケーションの本番運用に必要な以下の要素を実装しました：

1. **自動化されたデプロイメント**: Blue-Greenデプロイメントによるダウンタイムゼロの更新
2. **包括的な監視**: Prometheusメトリクスによる詳細な監視とアラート
3. **柔軟な設定管理**: 環境別設定と厳格なバリデーション
4. **コンテナ化**: Dockerによる移植性と一貫性の確保
5. **運用自動化**: スクリプトによる反復可能なデプロイメント

これらの実装により、安全で効率的な本番運用が可能な基盤を構築できました。実際のプロジェクトでは、要件に応じてこれらの機能をカスタマイズし、さらなる改善を行うことが推奨されます。

---

## 📞 サポート

質問や問題がある場合は、以下のリソースを参照してください：

- **ドキュメント**: このREADMEファイル
- **ログ**: `logs/`ディレクトリ内のログファイル
- **メトリクス**: `http://localhost:9090/metrics`
- **ヘルスチェック**: `http://localhost:8080/health`

---
