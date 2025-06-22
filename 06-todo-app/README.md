# 第6章 実践アプリケーション開発① - TODOリスト

## 概要

このディレクトリには、Golang、HTMX、Alpine.js、Tailwind CSS を統合した完全なTODOアプリケーションが含まれています。第1章から第5章で学んだ技術を実際のアプリケーションで活用する実践例です。

## 主要機能

### 📝 タスク管理

- **作成**: タイトル、説明、優先度、期限を設定してタスク作成
- **編集**: インライン編集でタスク内容を更新
- **削除**: 確認ダイアログ付きのタスク削除
- **完了切替**: チェックボックスクリックで瞬時に状態変更

### 🔍 フィルタリング・検索

- **フィルター**: 全て・未完了・完了・期限切れの4つのビュー
- **リアルタイム検索**: 500ms遅延でタイトル・説明文を検索
- **URL連動**: フィルター状態がURLに反映され、ブックマーク可能

### 📊 統計情報

- **ダッシュボード**: 全タスク・未完了・完了・期限切れの件数表示
- **リアルタイム更新**: タスク操作と連動して統計が自動更新

### 🎨 UI/UX機能

- **ダークモード**: ワンクリックでライト/ダーク切替
- **レスポンシブ**: モバイル・タブレット・デスクトップ対応
- **アニメーション**: Tailwind CSS による滑らかなトランジション
- **エラーハンドリング**: ユーザーフレンドリーなエラーメッセージ

### ⚡ リアルタイム体験

- **HTMX活用**: ページリロード無しでの全操作
- **楽観的UI**: 即座のフィードバック表示
- **プログレッシブエンハンスメント**: JavaScript無効でも基本機能動作

## 技術スタック

### バックエンド

- **Go**: 1.24.3
- **Gorilla Mux**: 1.8.1 (HTTP ルーティング)
- **SQLite**: 軽量データベース (本番では PostgreSQL 推奨)
- **Templ**: 0.3.887 (型安全なテンプレートエンジン)

### フロントエンド  

- **HTMX**: 1.9.5 (動的UI)
- **Alpine.js**: 3.x (軽量JavaScript)
- **Tailwind CSS**: 3.x (ユーティリティファーストCSS)

## アーキテクチャ

### プロジェクト構造

```text
06-todo-app/
├── main.go                     # アプリケーションエントリーポイント
├── go.mod                      # Go モジュール定義
├── internal/
│   ├── models/
│   │   └── todo.go             # データモデル・リポジトリ
│   ├── handlers/
│   │   └── todo.go             # HTTP ハンドラー
│   └── templates/
│       ├── base.templ          # ベーステンプレート
│       └── todo.templ          # TODOコンポーネント
├── db/
│   └── todo.db                 # SQLite データベース
└── README.md                   # このファイル
```

### レイヤー構成

1. **プレゼンテーション層**: Templ テンプレート + HTMX/Alpine.js
2. **アプリケーション層**: HTTP ハンドラー (handlers/)
3. **ドメイン層**: ビジネスロジック (models/)
4. **インフラ層**: SQLite データベース

## データモデル

### Todo 構造体

```go
type Todo struct {
    ID          int        `json:"id"`
    Title       string     `json:"title"`
    Description string     `json:"description"`
    Completed   bool       `json:"completed"`
    Priority    string     `json:"priority"` // low, medium, high
    DueDate     *time.Time `json:"due_date"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}
```

### データベーススキーマ

```sql
CREATE TABLE todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    priority TEXT DEFAULT 'medium',
    due_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 実行方法

### 1. 依存関係のインストール

```bash
cd 06-todo-app
go mod tidy
```

### 2. テンプレート生成

```bash
templ generate
```

### 3. アプリケーション起動

```bash
go run main.go
```

### 4. ブラウザでアクセス

http://localhost:8081

## API エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/` | メインページ（フィルター・検索対応） |
| POST | `/todos` | 新規TODO作成 |
| GET | `/todos/{id}` | TODO詳細表示 |
| PUT | `/todos/{id}` | TODO更新 |
| DELETE | `/todos/{id}` | TODO削除 |
| GET | `/todos/{id}/edit` | TODO編集フォーム |
| PATCH | `/todos/{id}/toggle` | 完了状態切替 |

## HTMX 活用パターン

### 1. フォーム送信後の部分更新

```html
<form 
    hx-post="/todos"
    hx-target="#todo-list"
    hx-swap="afterbegin"
    hx-on:htmx:after-request="if(event.detail.successful) this.reset()"
>
```

### 2. リアルタイム検索

```html
<input 
    hx-get="/"
    hx-trigger="keyup changed delay:500ms"
    hx-target="#todo-list"
>
```

### 3. インライン編集

```html
<button
    hx-get="/todos/{id}/edit"
    hx-target="#todo-{id}"
    hx-swap="outerHTML"
>
```

## Alpine.js 活用例

### 1. ダークモード切替

```html
<html x-data="{ darkMode: false }" :class="{ 'dark': darkMode }">
<button @click="darkMode = !darkMode">
```

### 2. フォームバリデーション

```html
<input x-bind:min="new Date().toISOString().split('T')[0]">
```

### 3. 動的UI表示

```html
<span x-show="!darkMode">🌙</span>
<span x-show="darkMode">☀️</span>
```

## Tailwind CSS 設計

### 1. コンポーネント指向

- カード、ボタン、バッジなどの再利用可能なデザイン
- ダークモード対応の一貫したカラーパレット

### 2. レスポンシブデザイン

```html
<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
<input class="w-full md:w-auto">
```

### 3. 状態表現

```html
<div class="border-l-4 border-l-red-500"> <!-- 期限切れ -->
<span class="bg-green-50 text-green-800"> <!-- 完了状態 -->
```

## 学習ポイント

### 1. **プログレッシブエンハンスメント**

- JavaScript無効でも基本機能が動作
- HTMXリクエストと通常リクエストの両方に対応

### 2. **型安全性**

- Templ による型安全なテンプレート
- Go の静的型チェックによるバグ防止

### 3. **ユーザビリティ**

- 楽観的UI更新による高速なフィードバック
- 適切なエラーハンドリングとローディング表示

### 4. **保守性**

- 明確なレイヤー分離
- 単一責任の原則に基づくコンポーネント設計

## パフォーマンス考慮

### 1. **データベース最適化**

- 適切なインデックス設定
- クエリ最適化（ORDER BY、WHERE句）

### 2. **フロントエンド最適化**

- 遅延ローディング（検索500ms遅延）
- 部分更新によるデータ転送量削減

### 3. **キャッシング**

- 静的アセットのブラウザキャッシュ
- データベース接続プール（将来拡張）

## セキュリティ対策

### 1. **入力検証**

- タイトル必須チェック
- 優先度値の検証
- 日付フォーマット検証

### 2. **SQL インジェクション対策**

- プリペアドステートメント使用
- パラメータ化クエリ

### 3. **XSS 対策**

- Templ による自動エスケープ
- 適切なContent-Type設定

## テスト方針

### 1. **ユニットテスト**

- モデル層のビジネスロジック
- ハンドラーのHTTPレスポンス

### 2. **統合テスト**

- データベース操作
- テンプレートレンダリング

### 3. **E2Eテスト**

- ブラウザ自動化テスト
- HTMX動作確認

## 本番運用考慮

### 1. **データベース移行**

```go
// PostgreSQL 移行例
import _ "github.com/lib/pq"
db, err := sql.Open("postgres", "postgres://...")
```

### 2. **設定外部化**

```go
// 環境変数での設定
port := os.Getenv("PORT")
dbURL := os.Getenv("DATABASE_URL")
```

### 3. **ロギング**

```go
// 構造化ログ
log.Info("todo created", "id", todo.ID, "user", userID)
```

## 拡張アイデア

### 1. **ユーザー認証**

- JWT認証
- ソーシャルログイン

### 2. **リアルタイム同期**

- WebSocket
- Server-Sent Events

### 3. **高度な機能**

- タスクのカテゴリ分け
- ファイル添付
- コメント機能

## トラブルシューティング

### よくある問題

1. **データベース接続エラー**

   ```bash
   mkdir -p db  # データベースディレクトリ作成
   ```

2. **テンプレート生成エラー**

   ```bash
   templ generate  # テンプレート再生成
   ```

3. **依存関係エラー**

   ```bash
   go mod tidy     # 依存関係解決
   ```

### パフォーマンス問題

1. **検索が遅い**
   - SQLiteのFTSインデックス追加
   - PostgreSQLへの移行検討

2. **レスポンスが遅い**
   - データベース接続プール設定
   - 静的ファイルのCDN配信

---

このTODOアプリケーションは、モダンなウェブ開発の実践的な学習教材として設計されています。
シンプルな機能を通じて、スケーラブルで保守性の高いアプリケーション設計の基礎を学ぶことができます。
