---
title: "HTMXによるインタラクティブなUI"
---

# 第3章 HTMXによるインタラクティブなUI

ここでは以下のようなアプリケーションを作成します。

![画面1](/images/03-00.png)
![画面2](/images/03-01.png)
![画面3](/images/03-02.png)
![画面4](/images/03-03.png)
![画面5](/images/03-04.png)
![画面6](/images/03-05.png)
![画面7](/images/03-06.png)

## 3.1 HTMXの思想とアプローチ

HTMXは「HTML本来の力を拡張する」という思想で設計されたライブラリです。従来のSPA（Single Page Application）が複雑なJavaScriptフレームワークを要求するのに対し、HTMXはHTML属性だけでモダンなUIを実現します。

### 3.1.1 HATEOAS（Hypermedia as the Engine of Application State）

HTMXの核となる概念は**HATEOAS**です。これは、サーバーがクライアントに対して「次に何ができるか」をHTMLの形で提供する設計原則です。

```html
<!-- 従来のSPAアプローチ -->
<button onclick="deleteUser(123)">削除</button>
<script>
function deleteUser(id) {
  fetch(`/api/users/${id}`, {method: 'DELETE'})
    .then(response => response.json())
    .then(data => updateUI(data));
}
</script>

<!-- HTMXアプローチ -->
<button hx-delete="/users/123" hx-target="#user-list" hx-confirm="本当に削除しますか？">
  削除
</button>
```

💡 **設計の利点**: HTMXでは、サーバーが返すHTMLにユーザーが次に実行できるアクションが含まれているため、フロントエンドとバックエンドの結合度が下がります。

### 3.1.2 プログレッシブエンハンスメント

HTMXの重要な特徴は**プログレッシブエンハンスメント**です。JavaScriptが無効でも基本機能が動作し、有効な場合はより良い体験を提供します。

```html
<!-- JavaScriptが無効でも動作する基本フォーム -->
<form action="/search" method="get">
  <input type="text" name="q" placeholder="検索...">
  <button type="submit">検索</button>
</form>

<!-- HTMXによる拡張（JavaScript有効時） -->
<form hx-get="/search" hx-target="#results" hx-trigger="keyup changed delay:500ms">
  <input type="text" name="q" placeholder="検索...">
  <div id="results"></div>
</form>
```

⚠️ **重要**: プログレッシブエンハンスメントを実現するには、サーバー側で通常のHTTPリクエストとHTMXリクエストの両方に対応する必要があります。

## 3.2 基本的な属性と使い方

### 3.2.1 HTTP動詞属性

HTMXでは、あらゆるHTML要素からHTTPリクエストを発行できます。

```html
<!-- 基本的なHTTP動詞 -->
<button hx-get="/api/users">ユーザー一覧</button>
<form hx-post="/api/users">
  <input type="text" name="name">
  <button type="submit">作成</button>
</form>
<button hx-put="/api/users/123">更新</button>
<button hx-delete="/api/users/123">削除</button>
```

💡 **使い分けのコツ**

- GET: データの取得、検索、フィルタリング
- POST: 新規作成、非冪等な操作
- PUT: 全体更新、冪等な操作  
- DELETE: 削除操作

### 3.2.2 トリガーの詳細制御

`hx-trigger`属性で、いつリクエストを発行するかを細かく制御できます。

```html
<!-- 基本的なトリガー -->
<input hx-get="/search" hx-trigger="keyup">

<!-- 修飾子付きトリガー -->
<input 
  hx-get="/search" 
  hx-trigger="keyup changed delay:500ms"
  hx-target="#search-results"
>

<!-- 複数のトリガー -->
<div hx-get="/status" hx-trigger="load, every 5s">
  ステータス表示
</div>

<!-- 条件付きトリガー -->
<button hx-post="/submit" hx-trigger="click[ctrlKey]">
  Ctrl+クリックで送信
</button>
```

⚠️ **パフォーマンス注意**: `delay`修飾子なしでkeyupイベントを使用すると、大量のリクエストが発生する可能性があります。必ず適切な遅延を設定してください。

### 3.2.3 ターゲットとスワップ戦略

レスポンスをどこに、どのように挿入するかを制御します。

```html
<!-- 基本的なターゲット指定 -->
<button hx-get="/content" hx-target="#main-content">
  コンテンツ読み込み
</button>
<div id="main-content"></div>

<!-- 相対的なターゲット指定 -->
<div class="card">
  <button hx-get="/details" hx-target="closest .card">詳細表示</button>
</div>

<!-- スワップ戦略の指定 -->
<button hx-get="/items" hx-target="#list" hx-swap="beforeend">
  アイテム追加
</button>

<ul id="list">
  <!-- 新しいアイテムがここに追加される -->
</ul>
```

💡 **スワップ戦略の使い分け**

- `innerHTML`（デフォルト）: 内容を置換
- `outerHTML`: 要素ごと置換
- `beforeend`: 末尾に追加（リスト項目の追加に最適）
- `afterbegin`: 先頭に追加

## 3.3 実践的なUIパターン

### 3.3.1 インクリメンタル検索の実装

ユーザーが入力するたびに検索結果を更新する機能

```html
<div class="search-container">
  <input 
    type="text" 
    name="search"
    placeholder="商品を検索..."
    hx-get="/api/search"
    hx-trigger="keyup changed delay:300ms"
    hx-target="#search-results"
    hx-indicator="#search-spinner"
    class="w-full px-4 py-2 border rounded-lg"
  >
  
  <div id="search-spinner" class="htmx-indicator">
    🔄 検索中...
  </div>
  
  <div id="search-results" class="mt-4">
    <!-- 検索結果がここに表示される -->
  </div>
</div>
```

**サーバー側の実装**:

```go
func searchHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("search")
    
    // 検索処理
    results := searchProducts(query)
    
    // HTMX用のパーシャルHTMLを返す
    if isHTMXRequest(r) {
        renderSearchResults(w, results)
        return
    }
    
    // 通常のHTTPリクエスト用の完全なページを返す
    renderFullSearchPage(w, query, results)
}

func isHTMXRequest(r *http.Request) bool {
    return r.Header.Get("HX-Request") == "true"
}
```

⚠️ **UX配慮**: 検索結果が0件の場合も、適切なメッセージを表示することが重要です。

### 3.3.2 モーダルダイアログの実装

HTMXでモーダルを実装する場合、CSSアニメーションと組み合わせると自然な動作になります。

```html
<!-- モーダルを開くボタン -->
<button 
  hx-get="/modal/user-form"
  hx-target="#modal-container"
  hx-swap="innerHTML"
  class="bg-blue-500 text-white px-4 py-2 rounded"
>
  ユーザー追加
</button>

<!-- モーダルコンテナ -->
<div id="modal-container"></div>
```

**サーバーから返されるモーダルHTML**:

```html
<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center" 
     id="modal-backdrop">
  <div class="bg-white p-6 rounded-lg max-w-md w-full mx-4">
    <h2 class="text-xl font-bold mb-4">新規ユーザー</h2>
    
    <form hx-post="/api/users" hx-target="body" hx-swap="outerHTML">
      <input type="text" name="name" placeholder="名前" class="w-full mb-4 px-3 py-2 border rounded">
      <input type="email" name="email" placeholder="メールアドレス" class="w-full mb-4 px-3 py-2 border rounded">
      
      <div class="flex justify-end space-x-2">
        <button type="button" 
                onclick="document.getElementById('modal-container').innerHTML=''"
                class="px-4 py-2 bg-gray-300 rounded">
          キャンセル
        </button>
        <button type="submit" class="px-4 py-2 bg-blue-500 text-white rounded">
          作成
        </button>
      </div>
    </form>
  </div>
</div>
```

💡 **アクセシビリティ**: モーダルには適切なARIA属性を追加し、Escapeキーでの閉じる機能も実装することを推奨します。

### 3.3.3 無限スクロールの実装

大量データの表示に適したパターンです：

```html
<div id="content-list">
  <!-- 初期コンテンツ -->
  <div class="item">アイテム1</div>
  <div class="item">アイテム2</div>
  <!-- ... -->
  
  <!-- 無限スクロールのトリガー -->
  <div 
    hx-get="/api/items?page=2"
    hx-trigger="revealed"
    hx-target="#content-list"
    hx-swap="beforeend"
    hx-indicator="#loading-indicator"
  >
    <div id="loading-indicator" class="htmx-indicator text-center py-4">
      読み込み中...
    </div>
  </div>
</div>
```

**サーバー側での次ページトリガーの生成**:

```go
func itemsHandler(w http.ResponseWriter, r *http.Request) {
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    
    items := getItemsByPage(page, 10) // 10件ずつ取得
    
    // アイテムのHTMLを出力
    for _, item := range items {
        fmt.Fprintf(w, `<div class="item">%s</div>`, item.Name)
    }
    
    // 次のページがある場合、新しいトリガーを追加
    if hasMoreItems(page) {
        nextPage := page + 1
        fmt.Fprintf(w, `
        <div hx-get="/api/items?page=%d" 
             hx-trigger="revealed"
             hx-target="#content-list" 
             hx-swap="beforeend">
          <div class="htmx-indicator text-center py-4">読み込み中...</div>
        </div>`, nextPage)
    }
}
```

⚠️ **パフォーマンス注意**: 大量のDOMが蓄積されるため、必要に応じて古いアイテムを削除する仕組みも検討してください。

## 3.4 フォーム処理とバリデーション

### 3.4.1 リアルタイムバリデーション

入力中にサーバーサイドバリデーションを実行

```html
<form hx-post="/api/register" hx-target="#form-messages">
  <div class="mb-4">
    <label class="block text-gray-700">ユーザー名</label>
    <input 
      type="text" 
      name="username"
      hx-post="/api/validate-username"
      hx-trigger="blur"
      hx-target="#username-feedback"
      class="w-full px-3 py-2 border rounded"
    >
    <div id="username-feedback" class="text-sm mt-1"></div>
  </div>
  
  <div class="mb-4">
    <label class="block text-gray-700">メールアドレス</label>
    <input 
      type="email" 
      name="email"
      hx-post="/api/validate-email"
      hx-trigger="blur"
      hx-target="#email-feedback"
      class="w-full px-3 py-2 border rounded"
    >
    <div id="email-feedback" class="text-sm mt-1"></div>
  </div>
  
  <button type="submit" class="bg-blue-500 text-white px-4 py-2 rounded">
    登録
  </button>
  
  <div id="form-messages" class="mt-4"></div>
</form>
```

### 3.4.2 楽観的UI更新

ユーザー体験向上のため、サーバーレスポンスを待たずにUIを更新

```html
<div class="like-button-container">
  <button 
    hx-post="/api/posts/123/like"
    hx-target="this"
    hx-swap="outerHTML"
    onclick="this.innerHTML='❤️ いいね済み'; this.disabled=true;"
    class="bg-red-500 text-white px-3 py-1 rounded"
  >
    🤍 いいね
  </button>
</div>
```

💡 **重要**: 楽観的更新では、サーバーエラー時の復旧処理も実装する必要があります。

## 3.5 HTMXのエラーハンドリング

### 3.5.1 HTTPステータスコードの処理

HTMXは4xx、5xxエラーに対してデフォルトでは何もしませんが、カスタムハンドリングが可能です。

```html
<script>
document.body.addEventListener('htmx:responseError', function(evt) {
    // 4xx, 5xx エラーの処理
    if (evt.detail.xhr.status === 422) {
        // バリデーションエラーの場合は内容を表示
        evt.detail.shouldSwap = true;
    } else if (evt.detail.xhr.status === 401) {
        // 認証エラーの場合はログインページにリダイレクト
        window.location.href = '/login';
    } else {
        // その他のエラーは一般的なメッセージを表示
        alert('エラーが発生しました。しばらく後に再試行してください。');
    }
});
</script>
```

### 3.5.2 ネットワークエラーの処理

```html
<script>
document.body.addEventListener('htmx:sendError', function(evt) {
    // ネットワークエラーやタイムアウト
    console.error('リクエストエラー:', evt.detail);
    alert('接続エラーが発生しました。インターネット接続を確認してください。');
});
</script>
```

⚠️ **ユーザビリティ**: エラーメッセージは具体的で、ユーザーが次に何をすべきかを明確にすることが重要です。

## 3.6 パフォーマンス最適化

### 3.6.1 リクエストの最適化

不要なリクエストを避けるテクニック

```html
<!-- debounce（連続したイベントの制御） -->
<input 
  hx-get="/api/search"
  hx-trigger="keyup changed delay:500ms"
  hx-target="#results"
>

<!-- throttle（一定間隔でのリクエスト制限） -->
<div 
  hx-get="/api/status"
  hx-trigger="every 10s"
  hx-target="this"
>
  ステータス: 確認中...
</div>

<!-- 条件付きリクエスト -->
<input 
  hx-get="/api/search"
  hx-trigger="keyup[target.value.length > 2] changed delay:300ms"
  hx-target="#results"
>
```

### 3.6.2 キャッシングの活用

```html
<!-- ブラウザキャッシュを活用 -->
<div 
  hx-get="/api/static-content"
  hx-trigger="load once"
  hx-target="this"
>
  読み込み中...
</div>
```

**サーバー側でのキャッシュヘッダー設定**:

```go
func staticContentHandler(w http.ResponseWriter, r *http.Request) {
    // 1時間のキャッシュを設定
    w.Header().Set("Cache-Control", "public, max-age=3600")
    w.Header().Set("ETag", `"static-content-v1"`)
    
    // ETagによる条件付きリクエストの確認
    if r.Header.Get("If-None-Match") == `"static-content-v1"` {
        w.WriteHeader(http.StatusNotModified)
        return
    }
    
    // コンテンツを返す
    fmt.Fprintf(w, "<div>静的コンテンツ</div>")
}
```

💡 **キャッシュ戦略**: 静的なコンテンツには長期キャッシュを、動的なコンテンツには短期キャッシュまたはキャッシュ無効化を設定します。

## 3.7 復習問題

1. HTMXのHATEOAS思想について、従来のSPAアプローチとの違いを具体例で説明してください。

2. プログレッシブエンハンスメントを実現するために、サーバー側で考慮すべき点を3つ挙げてください。

3. 以下のHTMXコードの動作を詳しく説明してください：

   ```html
   <input hx-get="/search" hx-trigger="keyup changed delay:500ms" hx-target="#results">
   ```

4. 無限スクロールを実装する際の、パフォーマンス上の注意点を説明してください。

5. HTMXでのエラーハンドリングにおいて、UX向上のために実装すべき機能を3つ挙げてください。

6. `hx-swap`属性の各オプション（innerHTML、outerHTML、beforeend、afterbegin）の使い分けを説明してください。

7. インクリメンタル検索を実装する際に、サーバー負荷を軽減するためのテクニックを説明してください。

8. HTMXとブラウザキャッシュを組み合わせる際の利点と実装方法を説明してください。

9. フォームのリアルタイムバリデーションを実装する際の、サーバー側とクライアント側の役割分担を説明してください。

10. HTMXアプリケーションでSEOを考慮する場合の対策を説明してください。

## 3.8 復習問題の解答

1. HATEOASと従来SPAの違い  
   HTMX/HATEOAS: サーバーがHTMLと共に可能なアクションも提供

   ```html
   <!-- サーバーが返すHTML -->
   <div class="user-card">
     <h3>田中太郎</h3>
     <button hx-put="/users/123">編集</button>
     <button hx-delete="/users/123">削除</button>
   </div>
   ```

   従来SPA: クライアントがAPIスキーマを事前に知っている必要がある
   - フロントエンドとバックエンドが密結合
   - APIの変更時に両方の修正が必要

2. プログレッシブエンハンスメントの考慮点
   - HTMXリクエストと通常リクエストの両方に対応
   - `HX-Request`ヘッダーで判定し、適切なレスポンス（パーシャルHTML vs 完全ページ）を返す
   - JavaScript無効時も基本機能が動作するフォールバック実装

3. HTMXコードの動作説明

   ```html
   <input hx-get="/search" hx-trigger="keyup changed delay:500ms" hx-target="#results">
   ```

   - `keyup`: キーを離したときにトリガー
   - `changed`: 前回リクエスト時から値が変更された場合のみ
   - `delay:500ms`: イベント発生から500ms後にリクエスト実行
   - `hx-target="#results"`: idが"results"の要素にレスポンスを挿入
   結果：入力停止後500msで、値が変更されている場合のみ検索実行

4. 無限スクロールのパフォーマンス注意点

   - DOM要素の蓄積によるメモリ使用量増加
   - 古いアイテムの削除機能の実装検討
   - `revealed`トリガーの複数発火防止
   - サーバー側でのページ制限設定
   - 適切なローディング表示とエラーハンドリング

5. エラーハンドリングでのUX向上機能

   - 明確なエラーメッセージ表示（何が問題で、どう解決するか）
   - ネットワークエラー時の再試行機能
   - ローディング状態の適切な表示と、エラー時の状態復帰

6. hx-swapオプションの使い分け

   - `innerHTML`（デフォルト）: 要素の内容のみ置換、レイアウト保持
   - `outerHTML`: 要素全体を置換、完全なコンポーネント更新
   - `beforeend`: リストの末尾に追加、チャットメッセージやフィード
   - `afterbegin`: リストの先頭に追加、新着通知や最新投稿

7. インクリメンタル検索のサーバー負荷軽減

   - `delay`修飾子による検索頻度制御（300-500ms推奨）
   - `changed`修飾子による重複リクエスト防止
   - 最小文字数制限（例：2文字以上で検索開始）
   - サーバー側でのレスポンスキャッシュ
   - データベースクエリの最適化（インデックス設定等）

8. HTMXとブラウザキャッシュの組み合わせ
   利点: レスポンス時間短縮、サーバー負荷軽減

   実装

   ```go
   w.Header().Set("Cache-Control", "public, max-age=3600")
   w.Header().Set("ETag", `"content-v1"`)
   ```

   HTMXは標準的なHTTPキャッシングと連携し、条件付きリクエストをサポート

9. リアルタイムバリデーションの役割分担

   - クライアント側（HTMX）
        - 適切なタイミングでのバリデーションリクエスト送信
        - バリデーション結果の表示
        - ユーザー入力の一時的な状態管理
   - サーバー側
        - ビジネスルールに基づく検証ロジック
        - データベース制約の確認（重複チェック等）
        - セキュリティ関連の検証

10. HTMXアプリケーションのSEO対策

    - サーバーサイドレンダリングによる初期HTMLの提供
    - 重要なコンテンツのJavaScript無効時の表示保証
    - 適切なHTMLセマンティクス（h1-h6、nav、main等）の使用
    - meta要素による適切なページ情報設定
    - プログレッシブエンハンスメントによる検索エンジン対応
