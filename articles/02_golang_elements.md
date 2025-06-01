---
title: "第2章 Golangの基礎と実践パターン"
emoji: "😸" 
type: "tech" 
topics: ["golang","go","alpinejs","htmx"] 
published: false
---

# 第2章 Golangの基礎と実践パターン

## 2.1 Goの特徴とWeb開発における利点

Goは2009年にGoogleによって開発された言語で、シンプルさとパフォーマンスを両立した設計が特徴です。Web開発において特に以下の利点があります。

### 2.1.1 シンプルな言語設計

Goは意図的に機能を絞り込んでいます。複雑な継承や例外処理の代わりに、コンポジションとエラー値を使用します。

```go
// シンプルなエラーハンドリング
func getUserByID(id string) (*User, error) {
    user, err := database.FindUser(id)
    if err != nil {
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    return user, nil
}
```

💡 **学習のコツ**: Goのエラーハンドリングは最初は冗長に感じますが、この明示的なアプローチにより、エラーが見過ごされることを防げます。

### 2.1.2 優れた並行処理

ゴルーチンとチャネルにより、軽量な並行処理が簡単に実装できます。

```go
func handleRequests() {
    results := make(chan string, 3)
    
    // 3つの処理を並行実行
    go fetchData("API1", results)
    go fetchData("API2", results)
    go fetchData("API3", results)
    
    // 結果を収集
    for i := 0; i < 3; i++ {
        result := <-results
        fmt.Println("Got:", result)
    }
}
```

⚠️ **注意点**: チャネルのバッファサイズを適切に設定しないと、ゴルーチンがブロックされる可能性があります。

### 2.1.3 標準ライブラリの充実

GoのHTTPサーバーは標準ライブラリだけで実装でき、外部依存を最小限に抑えられます。

```go
func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    })
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## 2.2 クリーンアーキテクチャの実装

### 2.2.1 レイヤー分離の理念

クリーンアーキテクチャでは、以下の4つのレイヤーに分離します。

1. Entities（エンティティ）: ビジネスルールの核となる部分
2. Use Cases（ユースケース）: アプリケーション固有のビジネス論理
3. Interface Adapters（インターフェースアダプター）: 外部システムとの変換層
4. Frameworks & Drivers（フレームワーク・ドライバー）: 外部技術の詳細

💡 **重要な原則**: 依存関係は常に内側（ビジネスロジック）に向かって流れるようにします。

### 2.2.2 Goでのディレクトリ構造

```text
internal/
├── domain/          # エンティティとビジネスルール
│   ├── models/      # ドメインモデル
│   └── services/    # ドメインサービス
├── usecases/        # アプリケーションのユースケース
├── interfaces/      # インターフェースアダプター
│   ├── handlers/    # HTTPハンドラー
│   ├── repositories/ # データアクセス実装
│   └── presenters/  # プレゼンテーション層
└── infrastructure/  # 外部技術の詳細
    ├── database/    # データベース設定
    └── server/      # サーバー設定
```

### 2.2.3 インターフェースを活用した設計

依存性の逆転を実現するため、インターフェースを積極的に活用します。

```go
// domain/repositories/user_repository.go
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*models.User, error)
    Save(ctx context.Context, user *models.User) error
}

// usecases/user_usecase.go
type UserUseCase struct {
    userRepo UserRepository
    logger   Logger // これもインターフェース
}

func NewUserUseCase(userRepo UserRepository, logger Logger) *UserUseCase {
    return &UserUseCase{
        userRepo: userRepo,
        logger:   logger,
    }
}

func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*models.User, error) {
    uc.logger.Info("Getting user", "id", id)
    
    user, err := uc.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return user, nil
}
```

⚠️ **よくある間違い**: 具体的な実装に依存してしまうこと。常にインターフェースに依存するよう心がけてください。

## 2.3 HTTPハンドラーの実装パターン

### 2.3.1 構造体ベースのハンドラー

ハンドラーを構造体として実装することで、依存関係を明確に管理できます。

```go
type UserHandler struct {
    userUseCase *usecases.UserUseCase
    validator   *validator.Validate
}

func NewUserHandler(userUseCase *usecases.UserUseCase) *UserHandler {
    return &UserHandler{
        userUseCase: userUseCase,
        validator:   validator.New(),
    }
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    // パスパラメータの取得（URLルーティングライブラリを使用）
    userID := mux.Vars(r)["id"]
    
    // バリデーション
    if userID == "" {
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }
    
    // ユースケースの実行
    user, err := h.userUseCase.GetUser(r.Context(), userID)
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        
        log.Printf("Error getting user: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    
    // レスポンスの送信
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

💡 **ベストプラクティス**: エラーハンドリングでは、適切なHTTPステータスコードを返すことが重要です。

### 2.3.2 ミドルウェアパターン

横断的関心事（ログ、認証、CORS等）はミドルウェアで実装します。

```go
// ログミドルウェア
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // ログ用のwriter（レスポンスコードを取得するため）
        lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        next.ServeHTTP(lw, r)
        
        log.Printf("%s %s %d %v", 
            r.Method, 
            r.URL.Path, 
            lw.statusCode, 
            time.Since(start))
    })
}

type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
    lw.statusCode = code
    lw.ResponseWriter.WriteHeader(code)
}
```

⚠️ **注意点**: ミドルウェアの順序は重要です。認証→ログ→CORS の順番が一般的です。

### 2.3.3 認証ミドルウェアの実装

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Authorizationヘッダーから認証情報を取得
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }
        
        // Bearer トークンの検証
        token := strings.TrimPrefix(authHeader, "Bearer ")
        if !isValidToken(token) {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // コンテキストにユーザー情報を追加
        ctx := context.WithValue(r.Context(), "userID", getUserIDFromToken(token))
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## 2.4 エラーハンドリングのベストプラクティス

### 2.4.1 カスタムエラー型の定義

意味のあるエラー処理のために、カスタムエラー型を定義します。

```go
// domain/errors.go
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

type NotFoundError struct {
    Resource string
    ID       string
}

func (e NotFoundError) Error() string {
    return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}
```

### 2.4.2 エラーラッピングの活用

Go 1.13以降のエラーラッピングを活用して、エラーの文脈を保持します。

```go
func (r *userRepository) GetByID(ctx context.Context, id string) (*User, error) {
    query := "SELECT * FROM users WHERE id = $1"
    
    row := r.db.QueryRowContext(ctx, query, id)
    
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, NotFoundError{Resource: "user", ID: id}
        }
        return nil, fmt.Errorf("failed to scan user: %w", err)
    }
    
    return &user, nil
}
```

💡 **アドバイス**: `fmt.Errorf`と`%w`を使用してエラーをラップすることで、元のエラー情報を保持できます。

## 2.5 JSONレスポンスの設計

### 2.5.1 一貫性のあるAPIレスポンス

APIレスポンスの形式を統一することで、フロントエンドでの処理が簡単になります。

```go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
    response := APIResponse{
        Success: statusCode < 400,
        Data:    data,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}

func WriteJSONError(w http.ResponseWriter, statusCode int, code, message string) {
    response := APIResponse{
        Success: false,
        Error: &APIError{
            Code:    code,
            Message: message,
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}
```

### 2.5.2 実際の使用例

```go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        WriteJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
        return
    }
    
    // バリデーション
    if err := h.validator.Struct(req); err != nil {
        WriteJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
        return
    }
    
    // ユーザー作成
    user, err := h.userUseCase.CreateUser(r.Context(), req)
    if err != nil {
        WriteJSONError(w, http.StatusInternalServerError, "CREATION_FAILED", "Failed to create user")
        return
    }
    
    WriteJSONResponse(w, http.StatusCreated, user)
}
```

⚠️ **セキュリティ注意**: 本番環境では、内部エラーの詳細をクライアントに返すべきではありません。ログに記録し、一般的なエラーメッセージを返しましょう。

## 2.6 パフォーマンス考慮事項

### 2.6.1 コンテキストの活用

リクエストのタイムアウト管理には`context.Context`を活用します。

```go
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    // 3秒のタイムアウトを設定
    ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
    defer cancel()
    
    user, err := h.userUseCase.GetUser(ctx, userID)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            WriteJSONError(w, http.StatusRequestTimeout, "TIMEOUT", "Request timed out")
            return
        }
        // その他のエラー処理
    }
    
    WriteJSONResponse(w, http.StatusOK, user)
}
```

### 2.6.2 データベース接続の最適化

```go
func NewDB() (*sql.DB, error) {
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }
    
    // 接続プールの設定
    db.SetMaxOpenConns(25)                 // 最大接続数
    db.SetMaxIdleConns(25)                 // アイドル接続数
    db.SetConnMaxLifetime(5 * time.Minute) // 接続の最大生存時間
    
    return db, nil
}
```

💡 **パフォーマンスTips**: データベース接続プールの設定は、アプリケーションのパフォーマンスに大きく影響します。本番環境では必ず適切な値を設定してください。

## 2.7 復習問題

1. Goのエラーハンドリング方式の利点を、他言語の例外処理と比較して説明してください。

2. クリーンアーキテクチャにおける依存関係の方向について、具体例を交えて説明してください。

3. 以下のコードの問題点を指摘し、改善案を提示してください。

   ```go
   func GetUser(id string) *User {
       db := sql.Open("postgres", "connection_string")
       row := db.QueryRow("SELECT * FROM users WHERE id = ?", id)
       var user User
       row.Scan(&user.ID, &user.Name)
       return &user
   }
   ```

4. ミドルウェアパターンの利点と、実装時の注意点を説明してください。

5. Go言語でHTTPサーバーを実装する際の、パフォーマンス最適化のポイントを5つ挙げてください。

6. `context.Context`の主な用途と、使用時の注意点を説明してください。

7. JSONレスポンスの設計において、一貫性を保つために重要な要素を3つ挙げてください。

8. カスタムエラー型を定義する利点と、実装時のベストプラクティスを説明してください。

9. データベース接続プールの設定パラメータとその意味を説明してください。

10. 本番環境でのエラーハンドリングにおいて、セキュリティ面で注意すべき点を説明してください。

## 2.8 復習問題の解答

1. Goのエラーハンドリングの利点  
   - 明示性: エラーが戻り値として明示されるため、見過ごしにくい
   - 制御フロー: try-catchのような複雑な制御フローがなく、線形的で理解しやすい  
   - パフォーマンス: 例外処理のオーバーヘッドがない
   - デバッグの容易さ: エラーが発生した箇所と処理する箇所が明確

2. 依存関係の方向  
   内側のレイヤーは外側のレイヤーを知らない設計にします。

   例：
   - ✅ `UserUseCase` → `UserRepository`インターフェース ← `MySQLUserRepository`
   - ❌ `UserUseCase` → `MySQLUserRepository`（具体実装に依存）
   これにより、テストが容易になり、実装の変更に柔軟に対応できます。

3. コードの問題点と改善案  
   問題点: エラーハンドリング不足、リソースリーク、コンテキスト未使用

   改善案

   ```go
   func GetUser(ctx context.Context, db *sql.DB, id string) (*User, error) {
       query := "SELECT id, name FROM users WHERE id = $1"
       row := db.QueryRowContext(ctx, query, id)
       
       var user User
       err := row.Scan(&user.ID, &user.Name)
       if err != nil {
           if err == sql.ErrNoRows {
               return nil, ErrUserNotFound
           }
           return nil, fmt.Errorf("failed to scan user: %w", err)
       }
       
       return &user, nil
   }
   ```

4. ミドルウェアの利点と注意点  
   利点: 横断的関心事の分離、再利用性、テストの容易さ

   注意点

   - 実行順序が重要（認証→認可→ログ等）
   - パフォーマンスへの影響を考慮
   - 各ミドルウェアの責任を明確に分離

5. HTTPサーバーのパフォーマンス最適化
   - データベース接続プールの適切な設定
   - `context.Context`によるタイムアウト管理
   - 適切なHTTPヘッダー設定（Keep-Alive等）
   - ゴルーチンプールの活用
   - メモリプールの使用（sync.Pool）

6. context.Contextの用途と注意点  
   用途: タイムアウト管理、キャンセル処理、値の受け渡し

   注意点

   - 構造体のフィールドに保存しない
   - nilコンテキストを渡さない
   - 値の受け渡しは最小限に

7. JSONレスポンス設計の重要要素
   - 統一されたレスポンス構造（success、data、errorフィールド等）
   - 一貫性のあるエラーコード体系
   - 適切なHTTPステータスコードの使用

8. カスタムエラー型の利点  
   利点: 型安全なエラー処理、詳細な情報の保持、エラー種別の判定

   ベストプラクティス

   - `Error()`メソッドの実装
   - エラーラッピングの活用
   - エラー種別ごとの適切な分類

9. データベース接続プール設定
   - `MaxOpenConns`: 同時接続数上限（デフォルト無制限）
   - `MaxIdleConns`: アイドル接続数（デフォルト2）  
   - `ConnMaxLifetime`: 接続の最大生存時間
   - 適切な設定により、リソース使用量とパフォーマンスのバランスを取る

10. 本番環境のエラーハンドリング
    - 内部エラーの詳細をクライアントに返さない
    - エラーは詳細にログに記録し、監視システムで追跡
    - ユーザーには一般的なエラーメッセージを表示
    - センシティブな情報（データベーススキーマ等）の漏洩を防ぐ
