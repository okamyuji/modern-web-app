---
title: "Alpine.jsによる状態管理とイベント処理"
---

# 第4章 Alpine.jsによる状態管理とイベント処理

ここでは以下のようなアプリケーションを作成します。

![画面1](/images/04-00.png)
![画面2](/images/04-01.png)
![画面3](/images/04-02.png)
![画面4](/images/04-03.png)
![画面5](/images/04-04.png)
![画面6](/images/04-05.png)
![画面7](/images/04-06.png)

この章のソースコードは[src/04-alpinejs](https://github.com/okamyuji/modern-web-app/tree/main/src/04-alpinejs)にあります。

## 1. Alpine.jsの本質的な理解

### なぜAlpine.jsを選ぶのか

Alpine.jsは「新しいjQuery」と呼ばれることがありますが、それは誤解を招く表現です。Alpine.jsは、HTMLに直接リアクティブな振る舞いを追加できる軽量フレームワークで、HTMXと完璧に補完し合います。

```html
<!-- Alpine.jsの宣言的な記述 -->
<div x-data="{ count: 0 }">
    <button @click="count++">クリック数: <span x-text="count"></span></button>
</div>
```

**💡 設計思想の理解:** Alpine.jsは「HTMLファースト」のアプローチを採用しています。これはHTMXの「HTMLドリブン」な設計と完全に調和し、JavaScriptの複雑さを最小限に抑えながら、必要な部分だけインタラクティブにできます。

### HTMXとの役割分担

```html
<!-- HTMXとAlpine.jsの理想的な使い分け -->
<div x-data="{ 
    showDetails: false,
    loading: false 
}">
    <!-- Alpine.js: UIの状態管理 -->
    <button @click="showDetails = !showDetails">
        詳細を<span x-text="showDetails ? '隠す' : '表示'"></span>
    </button>
    
    <!-- HTMX: サーバーとの通信 -->
    <div x-show="showDetails" 
         hx-get="/api/details" 
         hx-trigger="revealed"
         @htmx:before-request="loading = true"
         @htmx:after-settle="loading = false">
        <div x-show="loading">読み込み中...</div>
    </div>
</div>
```

**⚠️ よくある間違い:** HTMXでできることをAlpine.jsで実装してしまうケース。サーバー通信が必要な処理はHTMX、クライアント側の状態管理はAlpine.jsという明確な役割分担を意識しましょう。

## 2. 実践的な状態管理パターン

### コンポーネントの設計

```html
<!-- 実践的なコンポーネント設計 -->
<div x-data="todoComponent()" class="max-w-md mx-auto p-4">
    <h2 class="text-2xl font-bold mb-4">TODOリスト</h2>
    
    <!-- 入力フォーム -->
    <form @submit.prevent="addTodo" class="mb-4">
        <div class="flex gap-2">
            <input 
                type="text" 
                x-model="newTodo"
                @keyup.enter="addTodo"
                class="flex-1 px-3 py-2 border rounded"
                placeholder="新しいタスク..."
            >
            <button 
                type="submit"
                :disabled="!newTodo.trim()"
                class="px-4 py-2 bg-blue-500 text-white rounded disabled:opacity-50"
            >
                追加
            </button>
        </div>
    </form>
    
    <!-- フィルター -->
    <div class="mb-4 flex gap-2">
        <template x-for="filter in filters">
            <button 
                @click="currentFilter = filter.value"
                :class="{
                    'bg-blue-500 text-white': currentFilter === filter.value,
                    'bg-gray-200': currentFilter !== filter.value
                }"
                class="px-3 py-1 rounded text-sm"
                x-text="filter.label"
            ></button>
        </template>
    </div>
    
    <!-- TODOリスト -->
    <ul class="space-y-2">
        <template x-for="todo in filteredTodos" :key="todo.id">
            <li class="flex items-center gap-2 p-2 bg-white rounded shadow">
                <input 
                    type="checkbox"
                    :checked="todo.completed"
                    @change="toggleTodo(todo.id)"
                    class="w-4 h-4"
                >
                <span 
                    :class="{ 'line-through text-gray-500': todo.completed }"
                    x-text="todo.text"
                    class="flex-1"
                ></span>
                <button 
                    @click="removeTodo(todo.id)"
                    class="text-red-500 hover:text-red-700"
                >
                    削除
                </button>
            </li>
        </template>
    </ul>
    
    <!-- 統計情報 -->
    <div class="mt-4 text-sm text-gray-600">
        残りのタスク: <span x-text="incompleteTodos"></span>件
    </div>
</div>

<script>
function todoComponent() {
    return {
        todos: Alpine.$persist([]).as('todos'), // ローカルストレージに永続化
        newTodo: '',
        currentFilter: 'all',
        filters: [
            { value: 'all', label: 'すべて' },
            { value: 'active', label: '未完了' },
            { value: 'completed', label: '完了' }
        ],
        
        // 算出プロパティ
        get filteredTodos() {
            switch(this.currentFilter) {
                case 'active':
                    return this.todos.filter(t => !t.completed);
                case 'completed':
                    return this.todos.filter(t => t.completed);
                default:
                    return this.todos;
            }
        },
        
        get incompleteTodos() {
            return this.todos.filter(t => !t.completed).length;
        },
        
        // メソッド
        addTodo() {
            const text = this.newTodo.trim();
            if (!text) return;
            
            this.todos.push({
                id: Date.now(),
                text: text,
                completed: false
            });
            
            this.newTodo = '';
            this.$dispatch('todo-added', { count: this.todos.length });
        },
        
        toggleTodo(id) {
            const todo = this.todos.find(t => t.id === id);
            if (todo) {
                todo.completed = !todo.completed;
                this.$dispatch('todo-toggled', { id, completed: todo.completed });
            }
        },
        
        removeTodo(id) {
            this.todos = this.todos.filter(t => t.id !== id);
            this.$dispatch('todo-removed', { id });
        }
    }
}
</script>
```

**💡 ベストプラクティス:** Alpine.jsのコンポーネントは関数として定義し、明確な責務を持たせます。データ、算出プロパティ、メソッドを整理して配置することで、保守性が向上します。

### グローバル状態管理

```javascript
// stores.js
document.addEventListener('alpine:init', () => {
    Alpine.store('user', {
        name: '',
        email: '',
        isLoggedIn: false,
        
        login(userData) {
            this.name = userData.name;
            this.email = userData.email;
            this.isLoggedIn = true;
        },
        
        logout() {
            this.name = '';
            this.email = '';
            this.isLoggedIn = false;
        }
    });
    
    // 通知システム
    Alpine.store('notifications', {
        items: [],
        
        add(message, type = 'info') {
            const id = Date.now();
            this.items.push({ id, message, type });
            
            // 3秒後に自動削除
            setTimeout(() => this.remove(id), 3000);
        },
        
        remove(id) {
            this.items = this.items.filter(n => n.id !== id);
        }
    });
});
```

**⚠️ 重要な注意点:** グローバルストアは便利ですが、過度に使用するとコンポーネント間の依存関係が複雑になります。ローカルな状態はコンポーネント内に留め、本当に共有が必要なデータのみストアに置きましょう。

## 3. イベントシステムの活用

### カスタムイベントによる通信

```html
<!-- 親コンポーネント -->
<div x-data="{ totalPrice: 0 }" 
     @update-price="totalPrice = $event.detail.price">
    <h3>合計金額: ¥<span x-text="totalPrice.toLocaleString()"></span></h3>
    
    <!-- 子コンポーネント -->
    <div x-data="cartItem()" class="border p-4 rounded">
        <input 
            type="number" 
            x-model="quantity"
            @input="updatePrice"
            min="1"
            class="w-20 px-2 py-1 border rounded"
        >
        × ¥<span x-text="unitPrice.toLocaleString()"></span>
        = ¥<span x-text="(quantity * unitPrice).toLocaleString()"></span>
    </div>
</div>

<script>
function cartItem() {
    return {
        quantity: 1,
        unitPrice: 1000,
        
        updatePrice() {
            this.$dispatch('update-price', {
                price: this.quantity * this.unitPrice
            });
        },
        
        init() {
            // 初期値を親に通知
            this.updatePrice();
        }
    }
}
</script>
```

**💡 設計のコツ:** イベントは「下から上へ」（子から親へ）流れるように設計します。これにより、コンポーネントの独立性が保たれ、再利用性が向上します。

### HTMXイベントとの連携

```html
<div x-data="{ 
    items: [],
    loading: false,
    error: null 
}">
    <button 
        hx-get="/api/items"
        hx-target="#items-container"
        @htmx:before-request="loading = true; error = null"
        @htmx:after-settle="loading = false"
        @htmx:response-error="error = 'データの取得に失敗しました'"
        class="px-4 py-2 bg-blue-500 text-white rounded"
    >
        データを取得
    </button>
    
    <!-- ローディング表示 -->
    <div x-show="loading" class="mt-4">
        <div class="animate-spin h-8 w-8 border-4 border-blue-500 rounded-full border-t-transparent"></div>
    </div>
    
    <!-- エラー表示 -->
    <div x-show="error" x-text="error" class="mt-4 text-red-500"></div>
    
    <!-- コンテンツ -->
    <div id="items-container"></div>
</div>
```

**⚠️ エラーハンドリングの重要性:** HTMXとAlpine.jsを組み合わせる際は、必ずエラー状態も管理しましょう。ユーザーは何が起きているか常に把握できる必要があります。

## 4. パフォーマンスとベストプラクティス

### メモリリークの防止

```javascript
// 悪い例：メモリリークの可能性
<div x-data="{
    interval: null,
    count: 0,
    
    init() {
        this.interval = setInterval(() => {
            this.count++;
        }, 1000);
    }
}">
    カウント: <span x-text="count"></span>
</div>

// 良い例：適切なクリーンアップ
<div x-data="{
    interval: null,
    count: 0,
    
    init() {
        this.interval = setInterval(() => {
            this.count++;
        }, 1000);
    },
    
    destroy() {
        if (this.interval) {
            clearInterval(this.interval);
        }
    }
}">
    カウント: <span x-text="count"></span>
</div>
```

**⚠️ パフォーマンス注意:** タイマー、イベントリスナー、WebSocket接続などは必ず`destroy()`メソッドでクリーンアップしましょう。これを怠ると、深刻なメモリリークにつながります。

### 大量データの効率的な処理

```html
<div x-data="virtualList()">
    <div class="h-96 overflow-y-auto" @scroll="handleScroll">
        <div :style="`height: ${totalHeight}px; position: relative;`">
            <template x-for="item in visibleItems" :key="item.id">
                <div 
                    :style="`position: absolute; top: ${item.top}px; height: ${itemHeight}px;`"
                    class="w-full px-4 py-2 border-b"
                >
                    <span x-text="item.text"></span>
                </div>
            </template>
        </div>
    </div>
</div>

<script>
function virtualList() {
    return {
        items: Array.from({ length: 10000 }, (_, i) => ({
            id: i,
            text: `アイテム ${i + 1}`
        })),
        itemHeight: 40,
        containerHeight: 384, // h-96 = 24rem = 384px
        scrollTop: 0,
        
        get totalHeight() {
            return this.items.length * this.itemHeight;
        },
        
        get visibleItems() {
            const startIndex = Math.floor(this.scrollTop / this.itemHeight);
            const endIndex = Math.ceil((this.scrollTop + this.containerHeight) / this.itemHeight);
            
            return this.items
                .slice(startIndex, endIndex + 1)
                .map((item, i) => ({
                    ...item,
                    top: (startIndex + i) * this.itemHeight
                }));
        },
        
        handleScroll(event) {
            this.scrollTop = event.target.scrollTop;
        }
    }
}
</script>
```

**💡 最適化のポイント:** 大量のデータを扱う場合は、仮想スクロールなどのテクニックを使用してDOMノードの数を制限します。Alpine.jsは軽量ですが、数千のDOM要素を同時に管理することは避けるべきです。

## 復習問題

1. HTMXとAlpine.jsの使い分けについて、それぞれが得意とする処理を3つずつ挙げてください。

2. 以下のコードの問題点を指摘し、修正してください

    ```html
    <div x-data="{ users: [] }">
        <button @click="
            fetch('/api/users')
                .then(r => r.json())
                .then(data => users = data)
        ">
            ユーザー一覧を取得
        </button>
    </div>
    ```

3. Alpine.jsでグローバルストアを使うべき場面と、使うべきでない場面をそれぞれ説明してください。

## 模範解答

1. 使い分けの例
   - HTMX: サーバーからのHTML取得、フォーム送信、部分的なページ更新
   - Alpine.js: UIの状態管理、フォームバリデーション、アニメーション制御

2. 修正版

    ```html
    <div x-data="{ 
        users: [],
        loading: false,
        error: null,
        
        async fetchUsers() {
            this.loading = true;
            this.error = null;
            try {
                const response = await fetch('/api/users');
                if (!response.ok) throw new Error('取得失敗');
                this.users = await response.json();
            } catch (e) {
                this.error = e.message;
            } finally {
                this.loading = false;
            }
        }
    }">
        <button @click="fetchUsers" :disabled="loading">
            ユーザー一覧を取得
        </button>
        <div x-show="loading">読み込み中...</div>
        <div x-show="error" x-text="error" class="text-red-500"></div>
    </div>
    ```

3. グローバルストアの使い分け
   - 使うべき場面：認証情報、通知システム、テーマ設定など、複数のコンポーネントで共有する必要があるデータ
   - 使うべきでない場面：特定のコンポーネント内でのみ使用するフォームの状態、一時的なUIの状態など
