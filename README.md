# Modern Web Application with Go

> モダンなWebアプリケーション開発の実践的デモプロジェクト

## 📖 概要

このプロジェクトは、**Go言語を使用したモダンなWebアプリケーション開発**の実践的な学習を目的としたデモアプリケーション集です。

### 🎯 学習目標

- パフォーマンス最適化: 高速なWebアプリケーションの構築手法
- セキュリティ対策: 実用的なセキュリティ機能の実装
- テスト・デバッグ: 品質保証のための開発手法
- デプロイメント・運用: 本番環境での安定運用

### ✨ 特徴

- 📊 実践的なコード例: 実際のプロジェクトで使える実装
- 🔒 セキュリティ重視: CSRF、XSS対策などを標準実装
- 📈 監視・メトリクス: Prometheusによる包括的な監視
- 🐳 コンテナ対応: Dockerによる一貫した開発・本番環境
- 🚀 自動化: CI/CDパイプラインとデプロイメント自動化

## 🏗️ プロジェクト構成

```text
modern-web-app/
├── README.md                    # このファイル
├── src/
│   ├── 01-env-setup/            # 環境セットアップ
│   ├── 02-golang-patterns/      # Goパターン実装
│   ├── 03-htmx/                 # HTMX統合
│   ├── 04-alpinejs/             # Alpine.js実装
│   ├── 05-tailwind/             # Tailwind CSS
│   ├── 06-todo-app/             # Todoアプリケーション
│   ├── 07-chat-app/             # チャットアプリケーション
│   ├── 08-performance-security/ # パフォーマンス・セキュリティ
│   ├── 09-test-debug/           # テスト・デバッグ
│   └── 10-deployment/           # デプロイメント・運用
└── books/                       # Zennの本
```

### 📁 各ディレクトリの役割

| ディレクトリ | 説明 | 主な学習内容 |
|-------------|------|-------------|
| `01-env-setup` | 環境セットアップ | 開発環境構築、ツール設定、プロジェクト初期化 |
| `02-golang-patterns` | Goパターン実装 | Goの設計パターン、ベストプラクティス |
| `03-htmx` | HTMX統合 | HTMXライブラリ、インタラクティブUI |
| `04-alpinejs` | Alpine.js実装 | Alpine.js、軽量JavaScript |
| `05-tailwind` | Tailwind CSS | CSSフレームワーク、スタイリング |
| `06-todo-app` | Todoアプリケーション | 実践的なWebアプリ開発 |
| `07-chat-app` | チャットアプリケーション | リアルタイム通信、WebSocket |
| `08-performance-security` | パフォーマンス・セキュリティ | gzip圧縮、キャッシュ、CSRF/XSS対策、レート制限 |
| `09-test-debug` | テスト・デバッグ | ユニットテスト、統合テスト、デバッグツール |
| `10-deployment` | デプロイメント・運用 | Blue-Greenデプロイ、監視、設定管理 |

## 🚀 クイックスタート

### 前提条件

- Go 1.24.3以上
- Node.js 18以上 (Tailwind CSS用)
- Make
- Docker (オプション)

### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd modern-web-app
```

### 2. 任意の章を選択して実行

#### 基礎編（推奨開始点）

```bash
cd src/01-env-setup
make setup
make dev
# 環境セットアップとツール確認
```

#### フロントエンド統合

```bash
cd src/05-tailwind
make setup
make dev
# ブラウザで http://localhost:8080 を開く
```

#### 実践アプリ開発

```bash
cd src/06-todo-app
make setup
make dev
# ブラウザで http://localhost:8080 を開く
```

## 📚 各章の詳細

### 🌱 01-env-setup: 環境セットアップ

Go開発環境の構築、必要なツールのインストール、プロジェクト初期化手順を学習します。

### 🛠️ 02-golang-patterns: Goパターン実装

Go言語の設計パターン、ベストプラクティス、効率的なコード記述法を学習します。

### 🔗 03-htmx: HTMX統合

HTMXライブラリを使用したインタラクティブなWebアプリケーション開発を学習します。

### 🗻 04-alpinejs: Alpine.js実装

Alpine.jsによる軽量なJavaScriptフレームワークの活用方法を学習します。

### 🎨 05-tailwind: Tailwind CSS

- 学習内容
    - ユーティリティファーストCSS: Tailwind CSSの基本原則と活用法
    - レスポンシブデザイン: ブレークポイントとモバイルファースト設計
    - ダークモード: テーマ切り替えとカラーパレット管理
    - アニメーション: トランジション、ホバーエフェクト、カスタムアニメーション
- 主要機能
    - コンポーネントライブラリ（ボタン、カード、フォーム、バッジ等）
    - レスポンシブ対応の完全なページレイアウト
    - インタラクティブなアニメーション
    - ダークモード対応

### 📝 06-todo-app: Todoアプリケーション

- 学習内容
    - HTMX統合: ページリロード無しでの動的UI更新
    - Alpine.js活用: 軽量なJavaScript機能の実装
    - データベース操作: SQLiteを使用したCRUD操作
    - リアルタイムUI: 楽観的UI更新とエラーハンドリング
- 主要機能
    - タスクの作成・編集・削除・完了切り替え
    - フィルタリング（全て・未完了・完了・期限切れ）
    - リアルタイム検索機能
    - 統計情報表示とダークモード

### 💬 07-chat-app: チャットアプリケーション

- 学習内容
    - Server-Sent Events: WebSocketより軽量なリアルタイム通信
    - Hubパターン: 効率的なメッセージブロードキャスト
    - セッション管理: ユーザー識別と状態管理
    - リアルタイムUI: 自動スクロール、接続状態管理
- 主要機能
    - リアルタイムメッセージング
    - ユーザー参加・退出通知
    - メッセージ履歴の永続化
    - オンラインユーザー一覧

### 🔒 08-performance-security: パフォーマンス・セキュリティ

- 学習内容
- パフォーマンス最適化
    - gzip圧縮による転送量削減
    - HTTPキャッシュ戦略
    - 効率的なデータベースクエリ
    - 静的ファイルの最適化

- セキュリティ機能
    - CSRF（Cross-Site Request Forgery）対策
    - XSS（Cross-Site Scripting）防止
    - セキュアヘッダーの設定
    - レート制限の実装

- 主要エンドポイント
    - `/` - メインページ
    - `/performance` - パフォーマンステスト
    - `/security` - セキュリティデモ
    - `/health` - ヘルスチェック

### 🧪 09-test-debug: テスト・デバッグ

- 学習内容
    - テスト戦略
        - ユニットテストの実装
        - 統合テストの設計
        - テストカバレッジの計測
        - モックとスタブの活用
    - デバッグ技術
        - ログ出力の最適化
        - デバッグパネルの実装
        - パフォーマンス分析
        - エラーハンドリング
- 主要機能
    - リアルタイムデバッグパネル
    - テスト結果の可視化
    - パフォーマンス監視
    - エラー追跡

### 🚀 10-deployment: デプロイメント・運用

- 学習内容
    - デプロイメント戦略
        - Blue-Greenデプロイメント
        - ローリングアップデート
        - 自動ロールバック
        - ヘルスチェック連携
    - 運用・監視
        - Prometheusメトリクス
        - 構造化ログ
        - 設定管理
        - グレースフルシャットダウン
- 主要機能
    - 自動デプロイスクリプト
    - メトリクス監視ダッシュボード
    - 環境別設定管理
    - Dockerコンテナ化

## 🛠️ 技術スタック

### バックエンド

- [Go](https://golang.org/) - メインプログラミング言語
- [Templ](https://templ.guide/) - 型安全なHTMLテンプレート
- [SQLite/PostgreSQL](https://www.sqlite.org/) - データベース
- [Prometheus](https://prometheus.io/) - メトリクス監視

### フロントエンド

- [Tailwind CSS](https://tailwindcss.com/) - CSSフレームワーク
- HTML5 - マークアップ
- JavaScript - インタラクション

### インフラ・ツール

- [Docker](https://www.docker.com/) - コンテナ化
- [Make](https://www.gnu.org/software/make/) - ビルド自動化
- [Node.js](https://nodejs.org/) - フロントエンドビルドツール

### テスト・品質

- Go testing - 標準テストパッケージ
- [golangci-lint](https://golangci-lint.run/) - 静的解析

## 📖 学習リソース

### 🔗 有用なリンク

- [Go公式ドキュメント](https://golang.org/doc/) - Go言語の学習
- [Tailwind CSS](https://tailwindcss.com/docs) - CSSフレームワーク
- [Docker入門](https://docs.docker.com/get-started/) - コンテナ技術

## 🤝 コントリビューション

このプロジェクトへの貢献を歓迎します！

### 貢献方法

1. Issue報告: バグや改善提案
2. プルリクエスト: 新機能や修正の提案
3. ドキュメント: README やコメントの改善
4. テスト: 新しいテストケースの追加

### 開発環境のセットアップ

```bash
# 各章のセットアップスクリプトを実行
make setup
make dev

# テスト実行
make test

# コード品質チェック
make lint
```

## 📄 ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 🙏 謝辞

このプロジェクトは、モダンなWebアプリケーション開発のベストプラクティスを学習するために作成されました。Go言語コミュニティ、Tailwind CSS、Prometheusなどのオープンソースプロジェクトに感謝いたします。

---

Happy Learning! 🚀
