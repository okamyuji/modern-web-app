---
title: "第1章 環境構築とプロジェクトセットアップ"
emoji: "😸" 
type: "tech" 
topics: ["golang","go","alpinejs","htmx"] 
published: false
---

# 第1章 環境構築とプロジェクトセットアップ

## 1.1 技術スタックの概要

本書では、以下の技術を組み合わせてモダンなWebアプリケーションを構築します。

- Golang: シンプルで高性能なバックエンド言語
- HTMX: JavaScriptを最小限に抑えたインタラクティブなフロントエンド
- Alpine.js: 軽量な状態管理とイベント処理
- Tailwind CSS: ユーティリティファーストのスタイリング

この組み合わせにより、複雑なJavaScriptフレームワークを使わずに、メンテナブルで高性能なWebアプリケーションを構築できます。

## 1.2 開発環境の準備

### 1.2.1 必要なツールのインストール

#### Golangのインストール

[公式サイト](https://go.dev/dl/)から最新版をダウンロードしてインストールしましょう。Go 1.18以上を推奨します。

```bash
# インストール確認
go version
```

⚠️ **よくある躓きポイント**: GOPATHとGOROOTの設定で混乱することがありますが、Go 1.11以降のモジュールシステムを使用するため、これらの環境変数の設定は不要です。

#### Node.jsとnpmのインストール

Tailwind CSSのビルドに必要です。[公式サイト](https://nodejs.org/)からLTS版をインストールしてください。

```bash
# インストール確認
node --version
npm --version
```

#### Air（ホットリロード用）のインストール

開発効率を大幅に向上させるツールです。コードの変更を検知して自動的にアプリケーションを再起動します。

```bash
go install github.com/air-verse/air@latest
```

💡 **アドバイス**: Airがパスに追加されない場合は、`$GOPATH/bin`または`$HOME/go/bin`をPATH環境変数に追加してください。

#### Templ（テンプレートエンジン）のインストール

型安全なHTMLテンプレートを生成するツールです。

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

### 1.2.2 プロジェクト構造の設計

効率的な開発のため、以下の構造を採用します：

```text
myapp/
├── cmd/server/          # アプリケーションのエントリーポイント
├── internal/            # 内部パッケージ（外部からアクセス不可）
│   ├── handlers/        # HTTPハンドラー
│   ├── models/          # データモデル
│   ├── repositories/    # データアクセス層
│   └── templates/       # Templテンプレート
├── ui/
│   └── static/         # 静的ファイル（CSS、JS）
├── .air.toml           # Airの設定
├── go.mod              # Goモジュール定義
├── package.json        # npm依存関係
├── tailwind.config.js  # Tailwind設定
└── Makefile           # ビルドコマンド
```

この構造はCleanArchitectureの原則に基づいており、関心の分離を促進します。

## 1.3 プロジェクトのセットアップ

### 1.3.1 基本的な初期化

新しいプロジェクトを作成し、必要なディレクトリを準備します。

```bash
mkdir myapp && cd myapp
go mod init myapp

# ディレクトリ構造を作成
mkdir -p cmd/server internal/{handlers,models,repositories,templates}
mkdir -p ui/static/{css,js}
```

### 1.3.2 Tailwind CSSの設定

#### npm初期化とTailwind CSSのインストール

```bash
npm init -y
npm install tailwindcss postcss autoprefixer
npx tailwindcss init
```

#### tailwind.config.jsの設定

Tailwindが適切にファイルをスキャンできるよう設定します。

```javascript
module.exports = {
  content: [
    "./ui/**/*.{html,js}",
    "./internal/templates/**/*.{templ,go}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

⚠️ **重要**: `content`配列のパスは実際のファイル配置に合わせて調整してください。パスが間違っていると、使用しているクラスが最終CSSに含まれません。

#### CSSファイルの作成

`ui/static/css/input.css`を作成します。

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

#### package.jsonにビルドスクリプトを追加

```json
{
  "scripts": {
    "dev": "tailwindcss -i ./ui/static/css/input.css -o ./ui/static/css/main.css --watch",
    "build": "tailwindcss -i ./ui/static/css/input.css -o ./ui/static/css/main.css --minify"
  }
}
```

### 1.3.3 Airの設定

プロジェクトルートに`.air.toml`を作成し、ホットリロード環境を構築します。

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "templ generate && go build -o ./tmp/main ./cmd/server"
  bin = "./tmp/main"
  delay = 1000
  exclude_dir = ["tmp", "vendor", "node_modules"]
  include_ext = ["go", "templ", "html"]
  exclude_regex = ["_test.go"]

[log]
  time = false
```

💡 **アドバイス**: `delay`は変更検知後の待機時間です。連続した変更を無視するため、1000ms程度が適切です。

### 1.3.4 Makefileの作成

複雑なコマンドを簡単に実行するため、`Makefile`を作成します：

```makefile
.PHONY: dev build clean

dev:
    @echo "Starting development server..."
    @npm run dev &
    @air

build:
    @echo "Building application..."
    @templ generate
    @npm run build
    @go build -o ./bin/server ./cmd/server

clean:
    @rm -rf ./tmp ./bin
```

## 1.4 基本的なアプリケーションの実装

### 1.4.1 Templテンプレートの作成

`internal/templates/base.templ`でベーステンプレートを作成します。

```go
package templates

templ Base(title string) {
    <!DOCTYPE html>
    <html lang="ja">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            <title>{ title }</title>
            <link rel="stylesheet" href="/static/css/main.css"/>
            <script src="https://unpkg.com/htmx.org@1.9.5"></script>
            <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
        </head>
        <body class="bg-gray-100 min-h-screen">
            { children... }
        </body>
    </html>
}
```

⚠️ **注意**: `defer`属性を忘れないでください。Alpine.jsはDOMが完全に読み込まれた後に初期化される必要があります。

### 1.4.2 ホームページテンプレートの作成

`internal/templates/home.templ`でホームページを作成：

```go
package templates

templ Home() {
    @Base("ホーム") {
        <div class="container mx-auto px-4 py-8">
        <h1 class="text-3xl font-bold text-center mb-8">Golang + HTMX + Alpine.js + Tailwind CSS</h1>

        <!-- Alpine.jsのサンプル -->
            <div x-data="{ count: 0 }" class="bg-white p-6 rounded-lg shadow-md mb-6">
                <h2 class="text-xl font-semibold mb-4">Alpine.js カウンター</h2>
                <p class="mb-4">カウント: <span x-text="count" class="font-bold text-blue-600"></span></p>
                <button @click="count++" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">
                    増加
                </button>
            </div>

            <!-- HTMXのサンプル -->
            <div class="bg-white p-6 rounded-lg shadow-md">
                <h2 class="text-xl font-semibold mb-4">HTMX サンプル</h2>
                <button 
                    hx-get="/api/greeting" 
                    hx-target="#greeting-result"
                    class="bg-green-500 hover:bg-green-600 text-white px-4 py-2 rounded"
                >
                挨拶を取得
                </button>
                <div id="greeting-result" class="mt-4 p-4 bg-gray-100 rounded"></div>
            </div>
        </div>
    }
}
```

### 1.4.3 サーバーの実装

`cmd/server/main.go`でサーバーを実装：

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "myapp/internal/templates"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // 静的ファイルの提供
    fs := http.FileServer(http.Dir("./ui/static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // ルートハンドラー
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/api/greeting", greetingHandler)

    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
    
    err := templates.Home().Render(context.Background(), w)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func greetingHandler(w http.ResponseWriter, r *http.Request) {
    // 少し遅延を入れてHTMXの動作を確認
    time.Sleep(500 * time.Millisecond)
    fmt.Fprintf(w, "こんにちは！現在の時刻は %s です", time.Now().Format("15:04:05"))
}
```

## 1.5 アプリケーションの実行

### 1.5.1 開発環境での実行

以下の手順でアプリケーションを起動します。

```bash
# 1. Templテンプレートの生成
templ generate

# 2. Tailwind CSSのビルド（別ターミナル）
npm run dev

# 3. アプリケーションの起動（メインターミナル）
make dev
```

💡 **効率的な開発のコツ**: 3つのターミナルを開いて、それぞれでTempl、Tailwind、Airを実行すると、すべての変更がリアルタイムで反映されます。

### 1.5.2 動作確認

ブラウザで`http://localhost:8080`にアクセスし、以下を確認してください。

1. **Tailwind CSS**: スタイルが適用されている
2. **Alpine.js**: カウンターボタンが動作する
3. **HTMX**: 挨拶ボタンをクリックすると非同期でメッセージが表示される
4. **ホットリロード**: ファイルを変更すると自動的に再読み込みされる

⚠️ **トラブルシューティング**

- CSSが適用されない場合：`npm run dev`が実行されているか確認
- テンプレートエラー：`templ generate`を再実行
- ホットリロードが動作しない：`.air.toml`のパス設定を確認

## 1.6 復習問題

1. このスタックを選択する利点は何ですか？従来のReact/Vue.jsスタックとの違いを説明してください。

2. Airを使用する目的と、設定ファイルの重要なオプションを3つ挙げてください。

3. Tailwind CSSの`content`設定が重要な理由を説明してください。

4. 以下のTemplコードが生成するHTMLの動作を説明してください：

    ```go
    <div x-data="{ show: false }">
     <button @click="show = !show">切替</button>
     <p x-show="show">表示されました</p>
   </div>
   ```

5. HTMXの`hx-target`と`hx-swap`属性の役割を説明してください。

6. `internal`ディレクトリを使用する理由は何ですか？

7. Makefileを使用する利点を3つ挙げてください。

8. 本番環境でのビルドコマンドと開発環境のコマンドの違いを説明してください。

9. ホットリロードが正常に動作しない場合のチェックポイントを5つ挙げてください。

10. このセットアップをDocker化する場合の主な考慮点を説明してください。

## 1.7 復習問題の解答

1. スタックの利点
   - シンプルさ: 複雑なビルドツールやJavaScriptフレームワークが不要
   - パフォーマンス: サーバーサイドレンダリングによる高速な初期表示
   - メンテナビリティ: TypeScript、JSX、バンドラーなどの複雑な技術スタックを避けられる
   - プログレッシブエンハンスメント: JavaScriptが無効でも基本機能が動作
   従来のSPAスタックと比べ、学習コストが低く、長期的なメンテナンスが容易です。

2. Airの目的と重要オプション
   - 目的: コード変更の自動検知とアプリケーションの自動再起動
   - 重要オプション:
     - `cmd`: 実行するビルドコマンド
     - `include_ext`: 監視するファイル拡張子
     - `exclude_dir`: 監視から除外するディレクトリ（node_modules等）

3. content設定の重要性
   Tailwindは使用されているクラスのみを最終CSSに含めるため、正確なファイルパスの指定が必要です。パスが間違っていると、実際に使用しているクラスが削除され、スタイルが適用されません。

4. Templコードの動作
   Alpine.jsの機能を使用しています。`x-data`で`show`変数を初期化（false）、ボタンクリック時に`show`の真偽値を切り替え、`x-show`ディレクティブで`p`要素の表示/非表示を制御します。

5. HTMXの属性の役割*
   - `hx-target`: AJAXレスポンスを挿入する対象要素をCSSセレクタで指定
   - `hx-swap`: レスポンスの挿入方法を指定（innerHTML、outerHTML、beforeend等）

6. internalディレクトリの理由
   Goの言語仕様により、`internal`パッケージは他のモジュールからインポートできません。これにより、パッケージの内部実装を隠蔽し、外部に公開したくないコードを保護できます。

7. Makefileの利点
   - 複雑なコマンドの簡素化
   - 開発者間での統一されたワークフロー
   - 依存関係の管理と自動化

8. ビルドコマンドの違い
   - 開発環境: `--watch`フラグで変更を監視、圧縮なし
   - 本番環境: `--minify`フラグでCSS圧縮、一回限りのビルド

9. ホットリロードのチェックポイント
   - Airが正しくインストールされているか
   - `.air.toml`の`include_ext`に必要な拡張子が含まれているか
   - `exclude_dir`に重要なディレクトリが含まれていないか
   - ポートが他のプロセスに使用されていないか
   - ファイルパーミッションの問題がないか

10. Docker化の考慮点
    - マルチステージビルド: Node.js環境でTailwindをビルド後、軽量なGoイメージで実行
    - 静的ファイルの配置: ビルド済みCSS/JSファイルの適切なコピー
    - ポート設定: 環境変数によるポート設定の柔軟性
    - ホットリロードの無効化: 本番環境では開発用ツールを除外
