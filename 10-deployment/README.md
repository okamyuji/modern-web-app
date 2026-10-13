# 第10章 デプロイメント・運用機能デモアプリケーション

本章では、モダンなWebアプリケーションの本番運用に必要なデプロイメント戦略、監視機能、設定管理、コンテナ化を実装したデモアプリケーションを紹介します。

## 目次

- [機能概要](#機能概要)
- [アーキテクチャ](#アーキテクチャ)
- [セットアップ](#セットアップ)
- [ビルドとデプロイ](#ビルドとデプロイ)
- [監視とメトリクス](#監視とメトリクス)
- [設定管理](#設定管理-1)
- [デプロイメント戦略](#デプロイメント戦略)
- [Docker活用](#docker活用)
- [技術的解説](#技術的解説)

## 機能概要

### デプロイメント機能

- Blue-Greenデプロイメント アプリケーションを止めずに新しいバージョンへ切り替えます。
- ローリングアップデート 複数インスタンスの環境で、インスタンスを1台ずつ順に更新します。
- 自動ロールバック デプロイが失敗したとき、前のバージョンへ自動で戻します。
- ヘルスチェック デプロイが成功したかどうかを自動で判定します。

### 監視・メトリクス

- Prometheusメトリクス HTTPリクエスト、レスポンス時間、システムリソースを監視します。
- ビジネスメトリクス Todoの作成数や完了数など、アプリケーション固有の指標を集計します。
- ヘルスチェックAPI アプリケーションが応答できる状態かどうかを返します。
- 構造化ログ ログをJSON形式にそろえて出力します。

### 設定管理

- 環境別設定 開発、ステージング、本番の各環境で設定を分けます。
- 設定バリデーション 起動時に、本番環境で必須の設定がそろっているかを検証します。
- シークレット管理 セッションとCSRFの秘密値を環境変数から受け取り、本番環境では既定値のままの起動を拒否します。
- 12-Factor Appの方針 設定を環境変数で受け取るので、アプリケーションを別の実行環境へ移しやすくなります。

### コンテナ化

- マルチステージビルド ビルド用と実行用のステージを分け、実行用イメージにビルドツールを含めません。
- セキュリティ強化 コンテナを非rootユーザーで実行します。
- ヘルスチェック統合 DockerのHEALTHCHECKがアプリケーションのヘルスチェックを呼び出します。
- レイヤー最適化 依存関係の取得をソースのコピーより前に置き、ソースだけを変えたときは依存関係のレイヤーを再利用します。

## アーキテクチャ

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

## セットアップ

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

## ビルドとデプロイ

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

## 監視とメトリクス

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

# Todo API（API_TOKEN を設定したときは Bearer トークンが必要）
curl -H "Authorization: Bearer $API_TOKEN" http://localhost:8080/api/todos

# メトリクス（JSON形式）
curl http://localhost:8080/metrics

# Prometheusメトリクス
curl http://localhost:9090/metrics
```

## 設定管理

### 環境変数

アプリケーションの設定は、次の環境変数で変えられます。

#### 基本設定

```bash
APP_NAME=deployment-demo
ENV=development              # development, test, testing, staging, production
PORT=8080
VERSION=v1.0.0
```

#### セキュリティ設定

```bash
SESSION_SECRET=<64文字のランダム文字列>
CSRF_SECRET=<32文字のランダム文字列>
API_TOKEN=<32文字以上のランダム文字列>
API_RATE_LIMIT=100           # クライアントIPごとの1分あたりの要求数
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

アプリケーションは、次の優先度の順に設定ファイルを読み込みます。

1. `.env.{環境}.local`（例：`.env.production.local`）
2. `.env.{環境}`（例：`.env.production`）
3. `.env.local`
4. `.env`

### 本番環境での注意点

本番環境では、次の設定を必ず見直してください。

- `SESSION_SECRET`: デフォルト値から変更必須（`ENV`が`development`、`test`、`testing`以外のとき）
- `CSRF_SECRET`: デフォルト値から変更必須（同上）
- `API_TOKEN`: 32文字以上が必須（`ENV`が`development`、`test`、`testing`以外では、未設定だと起動しません）
- `ENV`: プロセスの環境変数の値が、`.env`系のファイルの値より優先されます。どちらにも指定がないときは`production`として扱います。使える値は`development`、`test`、`testing`、`staging`、`production`で、ほかの値では起動しません
- `LOG_LEVEL`: `debug`以外を推奨
- `TLS_ENABLED=true`: HTTPS環境での運用推奨

## デプロイメント戦略

### Blue-Greenデプロイメント

#### 特徴

- ダウンタイムゼロ
- 即座にロールバック可能
- リソース消費量が多い

#### 手順

1. 新バージョンを一時ポートで起動
2. ヘルスチェック実行
3. 成功時に本番ポートに切り替え
4. 旧バージョンを停止

```bash
# Blue-Greenデプロイメント実行
DEPLOY_STRATEGY=blue-green ./scripts/deploy.sh v1.0.0 production
```

### ローリングアップデート

#### 特徴

- リソース効率が良い
- 段階的な更新
- 部分的な障害の影響を限定

#### 手順

1. 複数インスタンスを順次更新
2. ロードバランサーから一時的に除外
3. 更新完了後にトラフィック復帰

```bash
# ローリングアップデート実行
DEPLOY_STRATEGY=rolling ./scripts/deploy.sh v1.0.0 production
```

### 自動ロールバック

デプロイメント中にヘルスチェックが失敗した場合、自動的に前のバージョンにロールバックします。

#### ロールバックの条件

- ヘルスチェックタイムアウト
- アプリケーション起動失敗
- 重要なエラーレスポンス

## Docker活用

### マルチステージビルド

Dockerfileはマルチステージビルドを使い、ビルド用と実行用のステージを分けています。

```dockerfile
# ビルダーステージ
FROM golang:1.27-alpine3.24 AS builder
# 依存関係とソースコードのコピー
# ビルド実行

# 本番ステージ
FROM alpine:3.24
# 最小限のファイルのみコピー
# 非rootユーザーで実行
```

### Docker操作

```bash
# イメージビルド
make docker-build

# コンテナ実行（本番モードで起動するため秘密値が必要）
SESSION_SECRET=<64文字のランダム文字列> CSRF_SECRET=<32文字のランダム文字列> API_TOKEN=<32文字以上のランダム文字列> make docker-run

# イメージのプッシュ
make docker-push
```

### セキュリティ対策

- 非rootユーザー（appuser）での実行
- 最小限のベースイメージ（Alpine Linux）
- セキュリティアップデートの適用
- 不要なファイルの除外

## 技術的解説

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

#### 実装のポイント

- 環境変数の優先的な使用
- 開発環境では設定なしでも起動できるデフォルト値
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

#### 実装のポイント

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

#### 実装のポイント

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

#### 実装のポイント

- 段階的なデプロイメント
- 各段階でのエラーハンドリング
- ロールバック機能の実装

## 開発とデバッグ

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

開発環境では、次のデバッグ機能を使えます。

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

## パフォーマンス最適化

### ビルド最適化

- Go: `-ldflags \"-s -w\"` によるバイナリサイズ削減
- CSS: Tailwind CSSがテンプレートで使うクラスだけを出力します。
- 静的ファイル: gzip圧縮による転送サイズ削減

### 実行時最適化

- データベース接続プール
- HTTP Keep-Alive
- 読み込み、書き込み、アイドル、シャットダウンのタイムアウト設定
- メモリ使用量の監視

## セキュリティ

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

## 運用チェックリスト

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

## まとめ

本デモアプリケーションは、Webアプリケーションを本番で運用するために次の要素を備えています。

1. 自動化されたデプロイメント Blue-Greenデプロイメントで、アプリケーションを止めずに更新します。
2. 包括的な監視 Prometheusメトリクスで、リクエストとリソースの状態を細かく監視します。
3. 柔軟な設定管理 環境ごとに設定を分け、本番環境では起動時に設定値を厳しく検証します。
4. コンテナ化 Dockerで、どの環境でも同じ構成のアプリケーションを動かせます。
5. 運用自動化 スクリプトで、同じ手順のデプロイメントを何度でも繰り返せます。

これらの実装により、安全で効率的な本番運用が可能な基盤を構築できました。実際のプロジェクトでは、要件に応じてこれらの機能をカスタマイズし、さらなる改善を行うことが推奨されます。

---

## サポート

質問や問題があるときは、次の情報を確認してください。

- ドキュメント このREADMEファイルです。
- ログ `logs/`ディレクトリにログファイルを出力します。
- メトリクス `http://localhost:9090/metrics`で取得できます。
- ヘルスチェック `http://localhost:8080/health`で確認できます。

---
