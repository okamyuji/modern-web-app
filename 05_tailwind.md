# 第5章 Tailwind CSSによる効率的なスタイリング

## 1. Tailwind CSSの本質的な理解

### なぜTailwind CSSなのか

Tailwind CSSは単なる「クラス名の羅列」ではありません。これは、デザインシステムをコードに落とし込む効率的な方法論です。特にGolang/HTMX/Alpine.jsスタックにおいて、Tailwind CSSは理想的な選択です。

```html
<!-- 従来のCSS approach -->
<div class="card">
    <h2 class="card-title">タイトル</h2>
    <p class="card-content">内容</p>
</div>
<style>
.card { /* 10行以上のCSS */ }
.card-title { /* さらに5行 */ }
.card-content { /* さらに3行 */ }
</style>

<!-- Tailwind CSS approach -->
<div class="bg-white rounded-lg shadow-md p-6">
    <h2 class="text-xl font-bold mb-2">タイトル</h2>
    <p class="text-gray-600">内容</p>
</div>
```

**💡 根本的な利点:** Tailwind CSSはHTMLとスタイルの距離を最小化します。HTMXがHTMLドリブンなアプローチを取るのと同様に、Tailwind CSSはスタイルをHTMLに近づけることで、コンポーネントの見通しを劇的に改善します。

### 設計システムとしてのTailwind

```javascript
// tailwind.config.js - プロジェクト固有のデザインシステム
module.exports = {
    content: ['./templates/**/*.html'],
    theme: {
        extend: {
            colors: {
                // ブランドカラーの定義
                'primary': {
                    50: '#eff6ff',
                    500: '#3b82f6',
                    600: '#2563eb',
                    700: '#1d4ed8',
                },
                'secondary': {
                    500: '#10b981',
                    600: '#059669',
                }
            },
            spacing: {
                // 独自の間隔システム
                '18': '4.5rem',
                '88': '22rem',
            },
            animation: {
                // カスタムアニメーション
                'fade-in': 'fadeIn 0.5s ease-in-out',
                'slide-up': 'slideUp 0.3s ease-out',
            }
        }
    },
    plugins: [
        require('@tailwindcss/forms'),
        require('@tailwindcss/typography'),
    ]
}
```

**⚠️ よくある間違い:** Tailwindの設定を「デフォルトのまま」使用すること。プロジェクト固有のデザインシステムを構築することで、一貫性のあるUIが実現できます。

## 2. コンポーネントパターンの実装

### 再利用可能なコンポーネント設計

```html
<!-- Goテンプレートでのコンポーネント化 -->
{{define "button"}}
<button 
    class="
        {{if eq .Variant "primary"}}
            bg-primary-600 hover:bg-primary-700 text-white
        {{else if eq .Variant "secondary"}}
            bg-secondary-500 hover:bg-secondary-600 text-white
        {{else if eq .Variant "outline"}}
            border-2 border-gray-300 hover:border-gray-400 text-gray-700
        {{else}}
            bg-gray-500 hover:bg-gray-600 text-white
        {{end}}
        
        {{if eq .Size "sm"}}
            px-3 py-1 text-sm
        {{else if eq .Size "lg"}}
            px-6 py-3 text-lg
        {{else}}
            px-4 py-2
        {{end}}
        
        rounded-md font-medium transition-colors duration-200
        focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500
        disabled:opacity-50 disabled:cursor-not-allowed
        
        {{.ExtraClasses}}
    "
    {{if .HtmxAttrs}}{{.HtmxAttrs | safe}}{{end}}
    {{if .AlpineAttrs}}{{.AlpineAttrs | safe}}{{end}}
    {{if .Disabled}}disabled{{end}}
>
    {{if .Icon}}
        <span class="{{if .Text}}mr-2{{end}}">{{.Icon | safe}}</span>
    {{end}}
    {{.Text}}
</button>
{{end}}

<!-- 使用例 -->
{{template "button" map 
    "Variant" "primary" 
    "Size" "lg" 
    "Text" "送信する"
    "Icon" `<svg>...</svg>`
    "HtmxAttrs" `hx-post="/submit" hx-target="#result"`
    "AlpineAttrs" `@click="submitting = true"`
}}
```

**💡 設計のコツ:** Goテンプレートの機能を活用して、Tailwindクラスを動的に組み立てます。これにより、一貫性のあるコンポーネントライブラリが構築できます。

### レスポンシブデザインの実践

```html
<!-- モバイルファーストの実装例 -->
<div class="container mx-auto px-4">
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {{range .Products}}
        <article 
            class="bg-white rounded-lg shadow-sm hover:shadow-lg transition-shadow duration-300"
            x-data="{ imageLoaded: false }"
        >
            <!-- 画像コンテナ（アスペクト比を維持） -->
            <div class="relative aspect-w-16 aspect-h-9 overflow-hidden rounded-t-lg">
                <img 
                    src="{{.ImageURL}}"
                    alt="{{.Name}}"
                    class="object-cover w-full h-full"
                    :class="{ 'opacity-0': !imageLoaded, 'opacity-100': imageLoaded }"
                    @load="imageLoaded = true"
                    loading="lazy"
                >
                <!-- ローディングスケルトン -->
                <div 
                    x-show="!imageLoaded"
                    class="absolute inset-0 bg-gray-200 animate-pulse"
                ></div>
            </div>
            
            <div class="p-4 sm:p-6">
                <h3 class="text-lg sm:text-xl font-semibold mb-2 line-clamp-2">
                    {{.Name}}
                </h3>
                <p class="text-gray-600 text-sm sm:text-base mb-4 line-clamp-3">
                    {{.Description}}
                </p>
                
                <!-- 価格とアクション -->
                <div class="flex items-center justify-between">
                    <span class="text-2xl font-bold text-primary-600">
                        ¥{{.Price | formatNumber}}
                    </span>
                    <button 
                        hx-post="/cart/add/{{.ID}}"
                        hx-target="#cart-count"
                        hx-swap="innerHTML"
                        class="
                            bg-primary-600 hover:bg-primary-700 
                            text-white px-4 py-2 rounded-md
                            text-sm sm:text-base
                            transition-colors duration-200
                        "
                    >
                        カートに追加
                    </button>
                </div>
            </div>
        </article>
        {{end}}
    </div>
</div>
```

**⚠️ レスポンシブの落とし穴:** `sm:`、`md:`、`lg:`プレフィックスは「最小幅」を意味します。モバイルファーストで設計し、大きな画面向けに拡張していくことが重要です。

## 3. パフォーマンス最適化

### 未使用CSSの削除とビルド最適化

```javascript
// postcss.config.js
module.exports = {
    plugins: {
        tailwindcss: {},
        autoprefixer: {},
        ...(process.env.NODE_ENV === 'production' ? {
            cssnano: {
                preset: 'default',
            }
        } : {})
    }
}

// package.json scripts
{
    "scripts": {
        "css:dev": "tailwindcss -i ./assets/input.css -o ./static/output.css --watch",
        "css:build": "NODE_ENV=production tailwindcss -i ./assets/input.css -o ./static/output.css --minify"
    }
}
```

**💡 最適化のポイント:** 本番環境では必ずPurgeCSSが動作することを確認しましょう。数MBあるTailwindのCSSファイルが、実際に使用している部分だけの数十KBまで削減されます。

### 動的クラスの扱い方

```go
// Goサイドでの動的クラス生成の注意点
type AlertConfig struct {
    Type    string
    Message string
}

// 悪い例：PurgeCSSに検出されない
func (a AlertConfig) GetClasses() string {
    return fmt.Sprintf("bg-%s-100 border-%s-400 text-%s-700", a.Type, a.Type, a.Type)
}

// 良い例：完全なクラス名を使用
func (a AlertConfig) GetClasses() string {
    classes := map[string]string{
        "success": "bg-green-100 border-green-400 text-green-700",
        "error":   "bg-red-100 border-red-400 text-red-700",
        "warning": "bg-yellow-100 border-yellow-400 text-yellow-700",
        "info":    "bg-blue-100 border-blue-400 text-blue-700",
    }
    if class, ok := classes[a.Type]; ok {
        return class
    }
    return classes["info"] // デフォルト
}
```

**⚠️ 重要な注意:** Tailwindのクラス名は完全な形で記述する必要があります。動的に組み立てると、PurgeCSSがクラスを検出できず、本番環境でスタイルが適用されません。

## 4. Alpine.jsとの統合パターン

### 状態に応じたスタイリング

```html
<!-- Alpine.jsの状態とTailwindの統合 -->
<div x-data="{ 
    tab: 'overview',
    loading: false,
    error: null 
}" class="w-full max-w-4xl mx-auto">
    <!-- タブナビゲーション -->
    <div class="border-b border-gray-200">
        <nav class="flex -mb-px">
            <button 
                @click="tab = 'overview'"
                :class="{
                    'border-primary-500 text-primary-600': tab === 'overview',
                    'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300': tab !== 'overview'
                }"
                class="
                    py-2 px-4 border-b-2 font-medium text-sm
                    transition-colors duration-200
                "
            >
                概要
            </button>
            <button 
                @click="tab = 'details'"
                :class="{
                    'border-primary-500 text-primary-600': tab === 'details',
                    'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300': tab !== 'details'
                }"
                class="
                    py-2 px-4 border-b-2 font-medium text-sm
                    transition-colors duration-200
                "
            >
                詳細
            </button>
        </nav>
    </div>
    
    <!-- タブコンテンツ -->
    <div class="py-4">
        <div x-show="tab === 'overview'" x-transition>
            <!-- 概要コンテンツ -->
        </div>
        <div x-show="tab === 'details'" x-transition>
            <!-- 詳細コンテンツ -->
        </div>
    </div>
</div>
```

### アニメーションとトランジション

```html
<!-- Tailwindのアニメーションユーティリティ拡張 -->
<style>
@keyframes slideUp {
    from {
        transform: translateY(20px);
        opacity: 0;
    }
    to {
        transform: translateY(0);
        opacity: 1;
    }
}

@keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
}
</style>

<!-- 通知コンポーネント -->
<div 
    x-data="{ 
        notifications: [],
        add(message, type = 'info') {
            const id = Date.now();
            this.notifications.push({ id, message, type });
            setTimeout(() => this.remove(id), 5000);
        },
        remove(id) {
            this.notifications = this.notifications.filter(n => n.id !== id);
        }
    }"
    @notify="add($event.detail.message, $event.detail.type)"
    class="fixed top-4 right-4 z-50 space-y-2"
>
    <template x-for="notification in notifications" :key="notification.id">
        <div 
            x-transition:enter="transition ease-out duration-300"
            x-transition:enter-start="opacity-0 transform translate-x-full"
            x-transition:enter-end="opacity-100 transform translate-x-0"
            x-transition:leave="transition ease-in duration-200"
            x-transition:leave-start="opacity-100"
            x-transition:leave-end="opacity-0"
            :class="{
                'bg-green-50 border-green-200 text-green-800': notification.type === 'success',
                'bg-red-50 border-red-200 text-red-800': notification.type === 'error',
                'bg-blue-50 border-blue-200 text-blue-800': notification.type === 'info'
            }"
            class="
                p-4 rounded-lg border shadow-lg 
                min-w-[300px] max-w-md
                animate-slide-up
            "
        >
            <div class="flex items-start">
                <div class="flex-1" x-text="notification.message"></div>
                <button 
                    @click="remove(notification.id)"
                    class="ml-4 text-gray-400 hover:text-gray-600"
                >
                    ✕
                </button>
            </div>
        </div>
    </template>
</div>
```

**💡 アニメーションのベストプラクティス:** Alpine.jsのトランジションディレクティブとTailwindのアニメーションクラスを組み合わせることで、スムーズで自然なUIインタラクションが実現できます。

## 5. 実践的なデバッグとメンテナンス

### 開発時のユーティリティ

```html
<!-- 開発時のレスポンシブインジケーター -->
{{if .IsDevelopment}}
<div class="fixed bottom-4 left-4 bg-black text-white px-2 py-1 rounded text-xs z-50">
    <span class="sm:hidden">XS</span>
    <span class="hidden sm:inline md:hidden">SM</span>
    <span class="hidden md:inline lg:hidden">MD</span>
    <span class="hidden lg:inline xl:hidden">LG</span>
    <span class="hidden xl:inline">XL</span>
</div>
{{end}}
```

**⚠️ デバッグのコツ:** ブレークポイントインジケーターを表示することで、レスポンシブデザインの動作を視覚的に確認できます。本番環境では必ず削除しましょう。

## 復習問題

1. Tailwindの動的クラス生成で避けるべきパターンと、その理由を説明してください。

2. 以下のコードを、Tailwindのベストプラクティスに従って改善してください。

    ```html
    <div style="margin-top: 20px; padding: 10px; background-color: #f3f4f6;">
        <h2 style="font-size: 24px; font-weight: bold;">タイトル</h2>
        <p style="color: #6b7280;">説明文</p>
    </div>
    ```

3. レスポンシブデザインにおいて、モバイルファーストアプローチが重要な理由を説明してください。

## 模範解答

1. 避けるべきパターン
   - 文字列連結での動的生成：`bg-${color}-500`
   - 理由：PurgeCSSがクラスを検出できず、本番ビルドで削除される
   - 解決策：完全なクラス名を条件分岐で使用するか、safelistに追加

2. 改善版

    ```html
    <div class="mt-5 p-2.5 bg-gray-100">
        <h2 class="text-2xl font-bold">タイトル</h2>
        <p class="text-gray-500">説明文</p>
    </div>
    ```

3. モバイルファーストの重要性
   - モバイルユーザーが大多数を占める現代において基本となる体験を最初に設計
   - プログレッシブエンハンスメントにより、画面サイズに応じて機能を追加
   - CSSの記述量が減り、オーバーライドの複雑さが軽減される
