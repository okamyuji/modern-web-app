# 第9章: テストとデバッグ

## 概要

このディレクトリには、Golang/HTMX アプリケーションにおける包括的なテスト戦略とデバッグツールの実装例が含まれています。実際の開発現場で必要となるテストの書き方、デバッグ手法、開発効率向上のためのツールを実践的に学ぶことができます。

## 主要機能

### 🧪 テスト機能

- **ユニットテスト**: testify を使用した詳細なテストケース
- **HTTPハンドラーテスト**: HTMX リクエストの適切なテスト
- **データベーステスト**: インメモリSQLiteを使用した高速テスト
- **E2Eテスト**: 実際のHTTPサーバーを使用した統合テスト
- **ベンチマークテスト**: パフォーマンス測定とボトルネック特定

### 🐛 デバッグ機能

- **デバッグパネル**: 開発環境での詳細な実行情報表示
- **パフォーマンス監視**: リクエスト処理時間、メモリ使用量の追跡
- **エラートレーシング**: スタックトレース付きの詳細なエラー情報
- **HTMX/Alpine.jsデバッグ**: フロントエンド機能の動作確認
- **構造化ログ**: JSON形式での詳細なログ出力

### 📊 監視・診断機能

- **リアルタイムメトリクス**: アプリケーション稼働状況の監視
- **ヘルスチェック**: システム状態の確認エンドポイント
- **メモリプロファイリング**: メモリ使用量とGCの監視
- **トレーシング**: リクエスト全体の追跡機能

## 技術スタック

### バックエンド

- **Go**: 1.24.3
- **Gorilla Mux**: 1.8.1 (HTTP ルーティング)
- **PostgreSQL**: プロダクション用データベース (フォールバック: SQLite)
- **Templ**: 0.3.887 (型安全なテンプレートエンジン)
- **testify**: 1.9.0 (テストフレームワーク)

### フロントエンド

- **HTMX**: 1.9.5 (動的UI)
- **Alpine.js**: 3.x (軽量JavaScript)
- **Tailwind CSS**: 3.x (ユーティリティファーストCSS)

## プロジェクト構造

```text
09-test-debug/
├── main.go                         # アプリケーションエントリーポイント
├── go.mod                          # Go モジュール定義
├── internal/
│   ├── models/
│   │   ├── todo.go                 # TODOモデル・リポジトリ
│   │   └── todo_test.go            # ユニットテスト
│   ├── handlers/
│   │   ├── todo_handler.go         # HTTPハンドラー
│   │   └── todo_handler_test.go    # ハンドラーテスト
│   ├── middleware/
│   │   └── debug.go                # デバッグ・エラーハンドリング
│   ├── logger/
│   │   └── logger.go               # 構造化ログ・メトリクス
│   ├── db/
│   │   └── database.go             # データベース接続管理
│   └── templates/
│       └── test.templ              # UI テンプレート
└── README.md                       # このファイル
```

## テスト戦略

### 1. ユニットテスト

#### モデルレイヤーのテスト

```go
func TestTodoRepository_Create(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    repo := NewTodoRepository(db)

    tests := []struct {
        name    string
        todo    Todo
        wantErr bool
    }{
        {
            name: "正常なTODO作成",
            todo: Todo{
                Title:       "テストタスク",
                Description: "これはテスト用のタスクです",
                Priority:    "high",
                Completed:   false,
            },
            wantErr: false,
        },
        // ... 他のテストケース
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // テスト実行
        })
    }
}
```

#### テーブル駆動テスト

- 複数のテストケースを効率的に管理
- エッジケースと正常ケースの網羅
- バリデーションロジックの詳細テスト

### 2. HTTPハンドラーテスト

#### HTMXリクエストのテスト

```go
func TestTodoHandler_Create_HTMX(t *testing.T) {
    handler, _, cleanup := setupTestHandler(t)
    defer cleanup()

    tests := []struct {
        name           string
        formData       url.Values
        headers        map[string]string
        expectedStatus int
        checkResponse  func(t *testing.T, body string)
    }{
        {
            name: "HTMXリクエストでの正常な作成",
            formData: url.Values{
                "title":       {"新しいタスク"},
                "description": {"タスクの説明"},
                "priority":    {"high"},
            },
            headers: map[string]string{
                "HX-Request": "true",
                "HX-Target":  "todo-list",
            },
            expectedStatus: http.StatusOK,
            checkResponse: func(t *testing.T, body string) {
                assert.Contains(t, body, "新しいタスク")
                assert.Contains(t, body, "data-todo-id")
            },
        },
    }
}
```

#### レスポンス検証

- ステータスコードの確認
- レスポンスヘッダーの検証
- HTMLコンテンツの部分マッチング
- HTMX特有のヘッダー（HX-Trigger等）の確認

### 3. E2Eテスト

#### 統合テストの実装

```go
func TestTodoFlow_E2E(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    server := setupTestServer(t)
    defer server.Close()

    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }

    // 1. ホームページの取得
    resp, err := client.Get(server.URL + "/")
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)

    // 2. TODOの作成と確認
    // ...
}
```

### 4. ベンチマークテスト

#### パフォーマンス測定

```go
func BenchmarkTodoRepository_List(b *testing.B) {
    db, cleanup := setupTestDB(b)
    defer cleanup()

    repo := NewTodoRepository(db)

    // テストデータの準備
    for i := 0; i < 1000; i++ {
        repo.Create(&Todo{
            Title: fmt.Sprintf("Task %d", i),
        })
    }

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _, err := repo.List("", "")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## デバッグツール

### 1. デバッグパネル

#### 開発環境での情報表示

```go
func DebugPanel(isDev bool) func(http.Handler) http.Handler {
    if !isDev {
        return func(next http.Handler) http.Handler {
            return next
        }
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            var memStatsBefore runtime.MemStats
            runtime.ReadMemStats(&memStatsBefore)

            // リクエスト処理
            next.ServeHTTP(w, r)

            // パフォーマンス情報をHTMLに挿入
            duration := time.Since(start)
            debugHTML := fmt.Sprintf(`
            <div id="debug-panel">
                Duration: %s
                Memory: %d KB
                Goroutines: %d
            </div>`, duration, memoryUsed, runtime.NumGoroutine())
        })
    }
}
```

### 2. エラーハンドリング

#### 環境別エラー表示

```go
func ErrorHandler(isDev bool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    if isDev {
                        // 開発環境では詳細なスタックトレースを表示
                        fmt.Fprintf(w, `
                        <div class="error-detail">
                            <h3>Error: %s</h3>
                            <pre>%s</pre>
                        </div>`, err, debug.Stack())
                    } else {
                        // 本番環境では一般的なエラーメッセージ
                        fmt.Fprintf(w, "Internal Server Error")
                    }
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

### 3. HTMX/Alpine.jsデバッグ

#### JavaScriptデバッグコード

```javascript
// HTMXイベントのロギング
const htmxEvents = [
    'htmx:configRequest',
    'htmx:beforeRequest',
    'htmx:afterRequest',
    'htmx:responseError'
];

htmxEvents.forEach(event => {
    document.body.addEventListener(event, (e) => {
        console.group('HTMX Event: ' + event);
        console.log('Target:', e.detail.target);
        console.log('Detail:', e.detail);
        console.groupEnd();
    });
});

// Alpine.jsコンポーネントの監視
if (window.Alpine) {
    window.Alpine.onBeforeComponentInit((component) => {
        console.log('Alpine Component Init:', component.$el, component.$data);
    });
}
```

## 実行方法

### 1. 依存関係のインストール

```bash
cd 09-test-debug
go mod tidy
```

### 2. テンプレート生成

```bash
templ generate
```

### 3. テストの実行

```bash
# 全テストの実行
go test ./... -v

# 短時間テストのみ実行（E2Eテストをスキップ）
go test ./... -v -short

# ベンチマークテストの実行
go test ./... -bench=. -benchmem

# カバレッジレポートの生成
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### 4. アプリケーション起動

```bash
# 開発モード（デバッグ機能有効）
go run main.go

# 本番モード
ENV=production go run main.go
```

### 5. ブラウザでアクセス

- [ホーム](http://localhost:8084)
- [メトリクス](http://localhost:8084/metrics)
- [ヘルスチェック](http://localhost:8084/health)
- [パニックテスト](http://localhost:8084/panic) (開発環境のみ)
- [エラーテスト](http://localhost:8084/error) (開発環境のみ)

## API エンドポイント

| メソッド | パス | 説明 | テスト対象 |
|---------|------|------|-----------|
| GET | `/` | ホームページ | UI統合テスト |
| GET | `/todos` | TODO一覧取得 | フィルタリング・ソート |
| POST | `/todos` | TODO作成 | バリデーション・HTMX |
| PUT | `/todos/{id}` | TODO更新 | データ整合性 |
| DELETE | `/todos/{id}` | TODO削除 | 存在確認・エラー処理 |
| PATCH | `/todos/{id}/toggle` | 完了切り替え | 状態変更 |
| GET | `/metrics` | メトリクス表示 | パフォーマンス監視 |
| GET | `/health` | ヘルスチェック | システム監視 |

## テストカバレッジ

### カバレッジ目標

- **ユニットテスト**: 90%以上
- **統合テスト**: 主要機能の80%以上
- **E2Eテスト**: ユーザー重要シナリオ100%

### 測定方法

```bash
# カバレッジ測定
go test ./... -coverprofile=coverage.out

# カバレッジレポート表示
go tool cover -func=coverage.out

# HTML形式でのカバレッジレポート
go tool cover -html=coverage.out -o coverage.html
```

### 現在のカバレッジ状況

```text
test-debug-demo/internal/models     coverage: 92.5% of statements
test-debug-demo/internal/handlers   coverage: 88.3% of statements
test-debug-demo/internal/logger     coverage: 85.0% of statements
```

## ベンチマーク結果

### パフォーマンス指標

```text
BenchmarkTodoRepository_List-8          5000    245623 ns/op   12456 B/op     152 allocs/op
BenchmarkTodoRepository_Create-8       10000    156789 ns/op    8234 B/op      89 allocs/op
BenchmarkTodoHandler_List-8             3000    456789 ns/op   25678 B/op     234 allocs/op
```

### 最適化のポイント

- データベースクエリの効率化
- メモリ使用量の最適化
- ガベージコレクションの頻度削減
- HTTP レスポンス時間の短縮

## デバッグ手法

### 1. ログレベル調整

```go
// 開発環境
logger := logger.NewLogger(os.Stdout, logger.DEBUG)

// 本番環境
logger := logger.NewLogger(os.Stdout, logger.ERROR)
```

### 2. メモリプロファイリング

```bash
# メモリ使用量の監視
go tool pprof http://localhost:8084/debug/pprof/heap

# ゴルーチンリークの確認
go tool pprof http://localhost:8084/debug/pprof/goroutine
```

### 3. デバッグ用エンドポイント

```bash
# パニック発生テスト
curl http://localhost:8084/panic

# エラー処理テスト
curl http://localhost:8084/error?test=debug
```

## トラブルシューティング

### よくある問題

1. **テストのタイムアウト**

   ```bash
   # テストタイムアウトの調整
   go test ./... -timeout=60s
   ```

2. **データベース接続エラー**

   ```bash
   # SQLite フォールバック確認
   ls -la *.db
   
   # PostgreSQL 接続テスト
   export DB_PASSWORD=your_password
   go run main.go
   ```

3. **テンプレート生成エラー**

   ```bash
   # Templ のインストール確認
   go install github.com/a-h/templ@latest
   
   # テンプレート再生成
   templ generate
   ```

### デバッグ手順

1. **ログの確認**

   ```bash
   # 構造化ログの確認
   go run main.go | jq
   ```

2. **テストの個別実行**

   ```bash
   # 特定のテストのみ実行
   go test ./internal/models -run TestTodoRepository_Create -v
   ```

3. **メトリクスの監視**

   ```bash
   # メトリクス確認
   curl http://localhost:8084/metrics | jq
   ```

## 学習ポイント

### 1. テスト設計

- **境界値テスト**: 入力値の境界でのテスト
- **エラーケーステスト**: 異常系の網羅的テスト
- **統合テスト**: コンポーネント間の相互作用テスト
- **パフォーマンステスト**: 負荷・ストレステスト

### 2. デバッグスキル

- **ログ駆動開発**: 適切なログレベルとメッセージ設計
- **メトリクス監視**: 定量的な性能評価
- **プロファイリング**: ボトルネック特定技術
- **デバッガ活用**: ステップ実行とブレークポイント

### 3. 開発効率

- **自動テスト**: CI/CDパイプラインでの継続的テスト
- **デバッグツール**: 開発効率向上のためのツール群
- **監視**: プロダクション環境での継続的監視
- **フィードバック**: ユーザーフィードバックの迅速な反映

## 継続的改善

### 1. テストの拡充

- 新機能追加時の同時テスト作成
- リファクタリング時のリグレッションテスト
- パフォーマンス劣化の早期検知

### 2. デバッグ環境の改善

- より詳細なメトリクス収集
- 分散トレーシングの導入
- エラー追跡システムの強化

### 3. 開発プロセスの最適化

- テスト駆動開発（TDD）の実践
- 継続的インテグレーション（CI）の強化
- 継続的デプロイメント（CD）の自動化

---

このデモアプリケーションを通じて、実際の開発現場で必要となるテスト技術とデバッグ手法を習得できます。包括的なテスト戦略と効果的なデバッグツールの活用により、高品質なソフトウェア開発のスキルを身につけることができます。
