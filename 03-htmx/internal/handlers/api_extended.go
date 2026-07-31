package handlers

import (
	"fmt"
	htmlpkg "html"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// === Indicators APIs ===

// SlowResponse handles slow response for loading indicators
func (h *APIHandler) SlowResponse(w http.ResponseWriter, r *http.Request) {
	delayStr := r.URL.Query().Get("delay")
	responseType := r.URL.Query().Get("type")
	
	delay, err := strconv.Atoi(delayStr)
	if err != nil {
		delay = 1000
	}
	
	time.Sleep(time.Duration(delay) * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	var html string
	switch responseType {
	case "users":
		html = `<div class='p-3 bg-green-100 border border-green-300 rounded'>
					<h4 class='font-bold text-green-700'>👥 ユーザー一覧</h4>
					<ul class='mt-2 space-y-1'>
						<li>• 田中太郎 (tanaka@example.com)</li>
						<li>• 佐藤花子 (sato@example.com)</li>
						<li>• 鈴木一郎 (suzuki@example.com)</li>
					</ul>
				</div>`
	case "stats":
		html = `<div class='p-3 bg-purple-100 border border-purple-300 rounded'>
					<h4 class='font-bold text-purple-700'>📊 統計情報</h4>
					<div class='mt-2 grid grid-cols-2 gap-2 text-sm'>
						<div>総ユーザー: 1,234</div>
						<div>アクティブ: 987</div>
						<div>今月の登録: 45</div>
						<div>平均セッション: 12分</div>
					</div>
				</div>`
	case "reports":
		html = `<div class='p-3 bg-red-100 border border-red-300 rounded'>
					<h4 class='font-bold text-red-700'>📈 レポート</h4>
					<div class='mt-2 text-sm'>
						<p>• 月次売上レポート生成完了</p>
						<p>• ユーザー行動分析完了</p>
						<p>• パフォーマンス指標更新</p>
					</div>
				</div>`
	case "custom1":
		html = `<div class='p-3 bg-yellow-100 border border-yellow-300 rounded'>
					<h4 class='font-bold text-yellow-700'>⭐ カスタム処理1</h4>
					<p class='text-sm mt-1'>ドットアニメーション付きで処理が完了しました。</p>
				</div>`
	case "custom2":
		html = `<div class='p-3 bg-pink-100 border border-pink-300 rounded'>
					<h4 class='font-bold text-pink-700'>🌊 カスタム処理2</h4>
					<p class='text-sm mt-1'>波アニメーション付きで処理が完了しました。</p>
				</div>`
	default:
		html = fmt.Sprintf(`<div class='p-4 bg-blue-100 border border-blue-300 rounded'>
					<h4 class='font-bold text-blue-700'>✅ 処理完了</h4>
					<p class='text-sm mt-1'>%dms の遅延後に応答しました。</p>
					<p class='text-xs text-blue-600'>時刻: %s</p>
				</div>`, delay, time.Now().Format("15:04:05"))
	}
	
	fmt.Fprint(w, html)
}

// ProgressResponse handles progress bar response
func (h *APIHandler) ProgressResponse(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	
	w.Header().Set("Content-Type", "text/html")
	html := `<div class='p-4 bg-indigo-100 border border-indigo-300 rounded'>
				<h4 class='font-bold text-indigo-700'>🎉 プログレス処理完了！</h4>
				<p class='text-sm mt-1'>プログレスバー付きの処理が正常に完了しました。</p>
				<div class='mt-2 text-xs text-indigo-600'>
					処理時間: 2秒 | 完了率: 100%
				</div>
			</div>`
	
	fmt.Fprint(w, html)
}

// SkeletonResponse handles skeleton loading response
func (h *APIHandler) SkeletonResponse(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1500 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	html := `<div class='p-4 border border-teal-300 rounded bg-teal-50'>
				<h4 class='font-bold text-teal-700'>📄 実際のコンテンツ</h4>
				<p class='text-sm mt-2'>スケルトンローディングの後に表示される実際のコンテンツです。</p>
				<p class='text-sm mt-1'>このコンテンツは1.5秒の遅延後に読み込まれました。</p>
				<div class='mt-3 flex space-x-2'>
					<span class='px-2 py-1 bg-teal-200 text-teal-800 rounded text-xs'>タグ1</span>
					<span class='px-2 py-1 bg-teal-200 text-teal-800 rounded text-xs'>タグ2</span>
				</div>
			</div>`
	
	fmt.Fprint(w, html)
}

// === Forms APIs ===

// FormSubmit handles basic form submission
func (h *APIHandler) FormSubmit(w http.ResponseWriter, r *http.Request) {
	time.Sleep(500 * time.Millisecond)
	
	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")
	
	w.Header().Set("Content-Type", "text/html")
	
	if name == "" || email == "" {
		html := `<div class='p-4 bg-red-100 border border-red-300 rounded'>
					<h4 class='font-bold text-red-700'>❌ エラー</h4>
					<p class='text-sm mt-1'>名前とメールアドレスは必須です。</p>
				</div>`
		fmt.Fprint(w, html)
		return
	}
	
	html := fmt.Sprintf(`<div class='p-4 bg-green-100 border border-green-300 rounded'>
				<h4 class='font-bold text-green-700'>✅ 送信完了</h4>
				<div class='text-sm mt-2'>
					<p><strong>名前:</strong> %s</p>
					<p><strong>メール:</strong> %s</p>
					<p><strong>メッセージ:</strong> %s</p>
				</div>
				<p class='text-xs text-green-600 mt-2'>送信時刻: %s</p>
			</div>`, htmlpkg.EscapeString(name), htmlpkg.EscapeString(email), htmlpkg.EscapeString(message), time.Now().Format("2006-01-02 15:04:05"))
	
	fmt.Fprint(w, html)
}

// ValidateUsername handles username validation
func (h *APIHandler) ValidateUsername(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	time.Sleep(200 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	if username == "" {
		return
	}
	
	if len(username) < 3 {
		html := `<div class='text-sm text-red-600'>❌ ユーザー名は3文字以上である必要があります</div>`
		fmt.Fprint(w, html)
		return
	}
	
	// Simulate checking if username exists
	existingUsers := []string{"admin", "test", "user", "demo"}
	for _, existing := range existingUsers {
		if strings.ToLower(username) == existing {
			html := `<div class='text-sm text-red-600'>❌ このユーザー名は既に使用されています</div>`
			fmt.Fprint(w, html)
			return
		}
	}
	
	html := `<div class='text-sm text-green-600'>✅ 使用可能なユーザー名です</div>`
	fmt.Fprint(w, html)
}

// ValidatePassword handles password validation
func (h *APIHandler) ValidatePassword(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	time.Sleep(150 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	if password == "" {
		return
	}
	
	var messages []string
	if len(password) < 8 {
		messages = append(messages, "8文字以上")
	}
	if !strings.ContainsAny(password, "0123456789") {
		messages = append(messages, "数字を含む")
	}
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		messages = append(messages, "大文字を含む")
	}
	
	if len(messages) > 0 {
		html := fmt.Sprintf(`<div class='text-sm text-red-600'>❌ 必要な条件: %s</div>`, strings.Join(messages, ", "))
		fmt.Fprint(w, html)
		return
	}
	
	html := `<div class='text-sm text-green-600'>✅ 強力なパスワードです</div>`
	fmt.Fprint(w, html)
}

// ValidatePasswordConfirm handles password confirmation validation
func (h *APIHandler) ValidatePasswordConfirm(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")
	time.Sleep(100 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	if passwordConfirm == "" {
		return
	}
	
	if password != passwordConfirm {
		html := `<div class='text-sm text-red-600'>❌ パスワードが一致しません</div>`
		fmt.Fprint(w, html)
		return
	}
	
	html := `<div class='text-sm text-green-600'>✅ パスワードが一致しています</div>`
	fmt.Fprint(w, html)
}

// ValidateForm handles complete form validation
func (h *APIHandler) ValidateForm(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")
	
	time.Sleep(300 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	if username == "" || password == "" || passwordConfirm == "" {
		html := `<div class='p-4 bg-red-100 border border-red-300 rounded'>
					<h4 class='font-bold text-red-700'>❌ バリデーションエラー</h4>
					<p class='text-sm mt-1'>すべてのフィールドを入力してください。</p>
				</div>`
		fmt.Fprint(w, html)
		return
	}
	
	if password != passwordConfirm {
		html := `<div class='p-4 bg-red-100 border border-red-300 rounded'>
					<h4 class='font-bold text-red-700'>❌ バリデーションエラー</h4>
					<p class='text-sm mt-1'>パスワードが一致しません。</p>
				</div>`
		fmt.Fprint(w, html)
		return
	}
	
	html := fmt.Sprintf(`<div class='p-4 bg-green-100 border border-green-300 rounded'>
				<h4 class='font-bold text-green-700'>✅ 登録完了</h4>
				<p class='text-sm mt-1'>ユーザー「%s」が正常に登録されました。</p>
				<p class='text-xs text-green-600 mt-2'>登録時刻: %s</p>
			</div>`, htmlpkg.EscapeString(username), time.Now().Format("2006-01-02 15:04:05"))
	
	fmt.Fprint(w, html)
}

// Subcategories handles dynamic subcategory loading
func (h *APIHandler) Subcategories(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	time.Sleep(200 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	var options []string
	switch category {
	case "technology":
		options = []string{"プログラミング", "AI・機械学習", "クラウド", "セキュリティ"}
	case "business":
		options = []string{"マーケティング", "営業", "経営戦略", "財務"}
	case "design":
		options = []string{"UI/UX", "グラフィック", "Web デザイン", "ブランディング"}
	default:
		fmt.Fprint(w, "")
		return
	}
	
	html := `<div>
				<label class="block text-sm font-medium text-gray-700">サブカテゴリ</label>
				<select name="subcategory" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500">
					<option value="">サブカテゴリを選択</option>`
	
	for _, option := range options {
		html += fmt.Sprintf(`<option value="%s">%s</option>`, option, option)
	}
	
	html += `</select></div>`
	
	fmt.Fprint(w, html)
}

// DynamicForm handles dynamic form submission
func (h *APIHandler) DynamicForm(w http.ResponseWriter, r *http.Request) {
	category := r.FormValue("category")
	subcategory := r.FormValue("subcategory")
	title := r.FormValue("title")
	
	time.Sleep(400 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	html := fmt.Sprintf(`<div class='p-4 bg-purple-100 border border-purple-300 rounded'>
				<h4 class='font-bold text-purple-700'>✅ 作成完了</h4>
				<div class='text-sm mt-2'>
					<p><strong>カテゴリ:</strong> %s</p>
					<p><strong>サブカテゴリ:</strong> %s</p>
					<p><strong>タイトル:</strong> %s</p>
				</div>
				<p class='text-xs text-purple-600 mt-2'>作成時刻: %s</p>
			</div>`, htmlpkg.EscapeString(category), htmlpkg.EscapeString(subcategory), htmlpkg.EscapeString(title), time.Now().Format("2006-01-02 15:04:05"))
	
	fmt.Fprint(w, html)
}

// FileUpload handles file upload simulation
func (h *APIHandler) FileUpload(w http.ResponseWriter, r *http.Request) {
	description := r.FormValue("description")
	
	// Simulate file processing time
	time.Sleep(1500 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	// Simulate file info
	fileSize := rand.Intn(5000) + 1000 // 1KB - 5KB
	fileName := "uploaded_file_" + time.Now().Format("20060102_150405") + ".txt"
	
	html := fmt.Sprintf(`<div class='p-4 bg-red-100 border border-red-300 rounded'>
				<h4 class='font-bold text-red-700'>📁 アップロード完了</h4>
				<div class='text-sm mt-2'>
					<p><strong>ファイル名:</strong> %s</p>
					<p><strong>サイズ:</strong> %d KB</p>
					<p><strong>説明:</strong> %s</p>
				</div>
				<p class='text-xs text-red-600 mt-2'>アップロード時刻: %s</p>
			</div>`, fileName, fileSize, htmlpkg.EscapeString(description), time.Now().Format("2006-01-02 15:04:05"))
	
	fmt.Fprint(w, html)
}

// EditField handles inline editing
func (h *APIHandler) EditField(w http.ResponseWriter, r *http.Request) {
	field := r.URL.Query().Get("field")
	
	w.Header().Set("Content-Type", "text/html")
	
	var html string
	switch field {
	case "name":
		html = `<form hx-post="/api/save-field" hx-target="this" hx-swap="outerHTML">
					<input type="hidden" name="field" value="name">
					<input type="text" name="value" value="田中太郎" 
						   class="w-full p-1 border border-gray-300 rounded focus:border-indigo-500">
					<div class="mt-1 space-x-2">
						<button type="submit" class="text-xs bg-green-500 text-white px-2 py-1 rounded">保存</button>
						<button type="button" 
								hx-get="/api/cancel-edit?field=name" 
								hx-target="this" 
								hx-swap="outerHTML"
								class="text-xs bg-gray-500 text-white px-2 py-1 rounded">キャンセル</button>
					</div>
				</form>`
	case "job":
		html = `<form hx-post="/api/save-field" hx-target="this" hx-swap="outerHTML">
					<input type="hidden" name="field" value="job">
					<input type="text" name="value" value="ソフトウェアエンジニア" 
						   class="w-full p-1 border border-gray-300 rounded focus:border-indigo-500">
					<div class="mt-1 space-x-2">
						<button type="submit" class="text-xs bg-green-500 text-white px-2 py-1 rounded">保存</button>
						<button type="button" 
								hx-get="/api/cancel-edit?field=job" 
								hx-target="this" 
								hx-swap="outerHTML"
								class="text-xs bg-gray-500 text-white px-2 py-1 rounded">キャンセル</button>
					</div>
				</form>`
	}
	
	fmt.Fprint(w, html)
}

// SaveField handles saving inline edits
func (h *APIHandler) SaveField(w http.ResponseWriter, r *http.Request) {
	field := r.FormValue("field")
	value := r.FormValue("value")
	
	time.Sleep(200 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	html := fmt.Sprintf(`<div 
				hx-get="/api/edit-field?field=%s"
				hx-target="this"
				hx-trigger="click"
				class="mt-1 p-2 border border-transparent rounded cursor-pointer hover:border-gray-300 hover:bg-gray-50">
				%s <span class="text-xs text-gray-500">(クリックして編集)</span>
			</div>`, htmlpkg.EscapeString(url.QueryEscape(field)), htmlpkg.EscapeString(value))
	
	fmt.Fprint(w, html)
}

// CancelEdit handles canceling inline edits
func (h *APIHandler) CancelEdit(w http.ResponseWriter, r *http.Request) {
	field := r.URL.Query().Get("field")
	
	w.Header().Set("Content-Type", "text/html")
	
	var originalValue string
	switch field {
	case "name":
		originalValue = "田中太郎"
	case "job":
		originalValue = "ソフトウェアエンジニア"
	}
	
	html := fmt.Sprintf(`<div 
				hx-get="/api/edit-field?field=%s"
				hx-target="this"
				hx-trigger="click"
				class="mt-1 p-2 border border-transparent rounded cursor-pointer hover:border-gray-300 hover:bg-gray-50">
				%s <span class="text-xs text-gray-500">(クリックして編集)</span>
			</div>`, htmlpkg.EscapeString(url.QueryEscape(field)), originalValue)
	
	fmt.Fprint(w, html)
}

// BulkAction handles bulk operations
func (h *APIHandler) BulkAction(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	
	w.Header().Set("Content-Type", "text/html")
	
	switch action {
	case "select-all":
		html := `<div class="space-y-2">
					<label class="flex items-center space-x-2">
						<input type="checkbox" name="items" value="1" class="rounded" checked>
						<span>アイテム 1</span>
					</label>
					<label class="flex items-center space-x-2">
						<input type="checkbox" name="items" value="2" class="rounded" checked>
						<span>アイテム 2</span>
					</label>
					<label class="flex items-center space-x-2">
						<input type="checkbox" name="items" value="3" class="rounded" checked>
						<span>アイテム 3</span>
					</label>
					<label class="flex items-center space-x-2">
						<input type="checkbox" name="items" value="4" class="rounded" checked>
						<span>アイテム 4</span>
					</label>
				</div>`
		fmt.Fprint(w, html)
	case "deselect-all":
		html := `<div class="space-y-2">
					<label class="flex items-center space-x-2">
						<input type="checkbox" name="items" value="1" class="rounded">
						<span>アイテム 1</span>
					</label>
					<label class="flex items-center space-x-2">
						<input type="checkbox" name="items" value="2" class="rounded">
						<span>アイテム 2</span>
					</label>
					<label class="flex items-center space-x-2">
						<input type="checkbox" name="items" value="3" class="rounded">
						<span>アイテム 3</span>
					</label>
					<label class="flex items-center space-x-2">
						<input type="checkbox" name="items" value="4" class="rounded">
						<span>アイテム 4</span>
					</label>
				</div>`
		fmt.Fprint(w, html)
	case "delete-selected":
		selectedItems := r.Form["items"]
		if len(selectedItems) == 0 {
			html := `<div class='p-4 bg-yellow-100 border border-yellow-300 rounded'>
						<p class='text-yellow-700'>削除するアイテムが選択されていません。</p>
					</div>`
			fmt.Fprint(w, html)
			return
		}
		
		html := fmt.Sprintf(`<div class='p-4 bg-green-100 border border-green-300 rounded'>
					<h4 class='font-bold text-green-700'>✅ 削除完了</h4>
					<p class='text-sm mt-1'>%d個のアイテムが削除されました。</p>
					<p class='text-xs text-green-600 mt-1'>削除されたアイテム: %s</p>
				</div>`, len(selectedItems), htmlpkg.EscapeString(strings.Join(selectedItems, ", ")))
		fmt.Fprint(w, html)
	}
}