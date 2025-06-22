package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// === Progressive Enhancement APIs ===

var (
	counter = 0
	notificationActive = false
	notificationLogs = []string{}
)

// LoadMore handles infinite scroll loading
func (h *APIHandler) LoadMore(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 2
	}
	
	time.Sleep(800 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	// Generate items for this page
	startItem := (page-1)*5 + 1
	endItem := page * 5
	
	html := `<div class="space-y-2">`
	for i := startItem; i <= endItem; i++ {
		html += fmt.Sprintf(`<div class='p-3 bg-blue-50 border border-blue-200 rounded'>アイテム %d</div>`, i)
	}
	html += `</div>`
	
	// Add next loader if not the last page
	if page < 5 {
		html += fmt.Sprintf(`
			<div 
				hx-get="/api/load-more?page=%d"
				hx-target="this"
				hx-swap="outerHTML"
				hx-trigger="revealed"
				class="text-center py-4">
				<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-500 mx-auto"></div>
				<p class="text-sm text-gray-600 mt-2">さらに読み込み中...</p>
			</div>`, page+1)
	} else {
		html += `<div class="text-center py-4 text-gray-500">
					<p class="text-sm">すべてのアイテムを読み込みました</p>
				</div>`
	}
	
	fmt.Fprint(w, html)
}

// LiveCounter handles real-time counter updates
func (h *APIHandler) LiveCounter(w http.ResponseWriter, r *http.Request) {
	counter++
	
	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`<div class="text-2xl font-bold text-green-600">%d</div>
						<p class="text-sm text-green-600">リアルタイムカウンター</p>`, counter)
	
	fmt.Fprint(w, html)
}

// LiveStats handles real-time statistics updates
func (h *APIHandler) LiveStats(w http.ResponseWriter, r *http.Request) {
	users := rand.Intn(50) + 950    // 950-999
	sessions := rand.Intn(200) + 300 // 300-499
	
	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`<div class="grid grid-cols-2 gap-2 text-center">
							<div>
								<div class="text-lg font-bold text-purple-600">%d</div>
								<p class="text-xs text-purple-600">ユーザー</p>
							</div>
							<div>
								<div class="text-lg font-bold text-purple-600">%d</div>
								<p class="text-xs text-purple-600">セッション</p>
							</div>
						</div>`, users, sessions)
	
	fmt.Fprint(w, html)
}

// ProgressiveLoad handles step-by-step loading
func (h *APIHandler) ProgressiveLoad(w http.ResponseWriter, r *http.Request) {
	stepStr := r.URL.Query().Get("step")
	step, err := strconv.Atoi(stepStr)
	if err != nil {
		step = 1
	}
	
	w.Header().Set("Content-Type", "text/html")
	
	switch step {
	case 1:
		time.Sleep(500 * time.Millisecond)
		html := `<div class="space-y-4">
					<div class="p-3 bg-green-100 border border-green-300 rounded">
						<h4 class="font-bold text-green-700">ステップ 1: 基本データ読み込み完了</h4>
						<p class="text-sm mt-1">基本的なデータの読み込みが完了しました。</p>
					</div>
					<div 
						hx-get="/api/progressive-load?step=2"
						hx-target="this"
						hx-swap="beforeend"
						hx-trigger="load delay:1s"
						class="text-center py-2">
						<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-indigo-500 mx-auto"></div>
						<p class="text-sm text-gray-600 mt-1">ステップ 2 を読み込み中...</p>
					</div>
				</div>`
		fmt.Fprint(w, html)
		
	case 2:
		time.Sleep(800 * time.Millisecond)
		html := `<div class="p-3 bg-blue-100 border border-blue-300 rounded">
					<h4 class="font-bold text-blue-700">ステップ 2: 詳細データ読み込み完了</h4>
					<p class="text-sm mt-1">詳細なデータの読み込みが完了しました。</p>
				</div>
				<div 
					hx-get="/api/progressive-load?step=3"
					hx-target="this"
					hx-swap="beforeend"
					hx-trigger="load delay:1s"
					class="text-center py-2">
					<div class="animate-pulse bg-purple-500 rounded-full h-4 w-4 mx-auto"></div>
					<p class="text-sm text-gray-600 mt-1">ステップ 3 を読み込み中...</p>
				</div>`
		fmt.Fprint(w, html)
		
	case 3:
		time.Sleep(1000 * time.Millisecond)
		html := `<div class="p-3 bg-purple-100 border border-purple-300 rounded">
					<h4 class="font-bold text-purple-700">ステップ 3: 最終処理完了</h4>
					<p class="text-sm mt-1">すべての処理が正常に完了しました。</p>
				</div>
				<div class="mt-4 p-4 bg-yellow-100 border border-yellow-300 rounded">
					<h4 class="font-bold text-yellow-700">🎉 段階的読み込み完了！</h4>
					<p class="text-sm mt-1">3つのステップすべてが正常に完了しました。</p>
					<div class="mt-2 text-xs text-yellow-600">
						完了時刻: ` + time.Now().Format("2006-01-02 15:04:05") + `
					</div>
				</div>`
		fmt.Fprint(w, html)
	}
}

// LazyContent handles lazy loading content
func (h *APIHandler) LazyContent(w http.ResponseWriter, r *http.Request) {
	contentType := r.URL.Query().Get("type")
	
	// Simulate different loading times
	switch contentType {
	case "image":
		time.Sleep(800 * time.Millisecond)
	case "chart":
		time.Sleep(1200 * time.Millisecond)
	case "heavy":
		time.Sleep(2000 * time.Millisecond)
	}
	
	w.Header().Set("Content-Type", "text/html")
	
	var html string
	switch contentType {
	case "image":
		html = `<div class="h-32 bg-gradient-to-r from-blue-400 to-purple-500 rounded flex items-center justify-center">
					<div class="text-white text-center">
						<div class="text-2xl">🖼️</div>
						<p class="text-sm">画像コンテンツ</p>
					</div>
				</div>`
	case "chart":
		html = `<div class="h-32 bg-gradient-to-r from-green-400 to-blue-500 rounded flex items-center justify-center">
					<div class="text-white text-center">
						<div class="text-2xl">📊</div>
						<p class="text-sm">チャートデータ</p>
						<p class="text-xs">売上: ↗️ 15%</p>
					</div>
				</div>`
	case "heavy":
		html = `<div class="h-32 bg-gradient-to-r from-red-400 to-pink-500 rounded flex items-center justify-center">
					<div class="text-white text-center">
						<div class="text-2xl">⚡</div>
						<p class="text-sm">重いコンテンツ</p>
						<p class="text-xs">処理完了</p>
					</div>
				</div>`
	}
	
	fmt.Fprint(w, html)
}

// Autocomplete handles autocomplete search
func (h *APIHandler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("city")
	if query == "" {
		query = r.FormValue("city")
	}
	
	time.Sleep(200 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	if len(query) < 2 {
		fmt.Fprint(w, "")
		return
	}
	
	cities := []string{
		"東京", "大阪", "名古屋", "横浜", "札幌", "神戸", "京都", "福岡", "川崎", "さいたま",
		"広島", "仙台", "北九州", "千葉", "世田谷", "堺", "新潟", "浜松", "熊本", "相模原",
	}
	
	var matches []string
	queryLower := strings.ToLower(query)
	
	for _, city := range cities {
		if strings.Contains(strings.ToLower(city), queryLower) {
			matches = append(matches, city)
		}
	}
	
	if len(matches) == 0 {
		html := `<div class="mt-2 p-2 text-sm text-gray-500">該当する都市が見つかりません</div>`
		fmt.Fprint(w, html)
		return
	}
	
	html := `<div class="mt-2 border border-gray-300 rounded bg-white shadow-lg max-h-40 overflow-y-auto">`
	for _, city := range matches {
		if len(html) > 200 && len(matches) > 5 { // Limit results
			break
		}
		html += fmt.Sprintf(`
			<div class="p-2 hover:bg-gray-100 cursor-pointer border-b border-gray-100 last:border-b-0"
				 hx-get="/api/select-city?city=%s"
				 hx-target="#selected-city">
				%s
			</div>`, city, city)
	}
	html += `</div>`
	
	fmt.Fprint(w, html)
}

// SelectCity handles city selection
func (h *APIHandler) SelectCity(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	
	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`<div class="p-3 border border-gray-300 rounded bg-green-50">
							<h4 class="font-bold text-green-700">選択された都市</h4>
							<p class="text-sm mt-1">🏙️ %s</p>
							<p class="text-xs text-green-600 mt-1">選択時刻: %s</p>
						</div>`, city, time.Now().Format("15:04:05"))
	
	fmt.Fprint(w, html)
}

// SaveOrder handles drag & drop order saving
func (h *APIHandler) SaveOrder(w http.ResponseWriter, r *http.Request) {
	time.Sleep(300 * time.Millisecond)
	
	w.Header().Set("Content-Type", "text/html")
	
	// In a real application, you would parse the order from the form data
	html := fmt.Sprintf(`<div class="p-4 border border-gray-300 rounded bg-green-50">
							<h4 class="font-bold text-green-700">✅ 順序保存完了</h4>
							<p class="text-sm mt-1">タスクの順序が正常に保存されました。</p>
							<p class="text-xs text-green-600 mt-1">保存時刻: %s</p>
						</div>`, time.Now().Format("2006-01-02 15:04:05"))
	
	fmt.Fprint(w, html)
}

// StartNotifications handles notification system start
func (h *APIHandler) StartNotifications(w http.ResponseWriter, r *http.Request) {
	notificationActive = true
	notificationLogs = append(notificationLogs, fmt.Sprintf("[%s] 通知システムを開始しました", time.Now().Format("15:04:05")))
	
	w.Header().Set("Content-Type", "text/html")
	html := `<div class="p-3 border border-gray-300 rounded bg-green-50">
				<p class="text-green-700">✅ 通知システムが開始されました</p>
			</div>`
	
	fmt.Fprint(w, html)
}

// StopNotifications handles notification system stop
func (h *APIHandler) StopNotifications(w http.ResponseWriter, r *http.Request) {
	notificationActive = false
	notificationLogs = append(notificationLogs, fmt.Sprintf("[%s] 通知システムを停止しました", time.Now().Format("15:04:05")))
	
	w.Header().Set("Content-Type", "text/html")
	html := `<div class="p-3 border border-gray-300 rounded bg-red-50">
				<p class="text-red-700">⏹️ 通知システムが停止されました</p>
			</div>`
	
	fmt.Fprint(w, html)
}

// NotificationUpdates handles real-time notification updates
func (h *APIHandler) NotificationUpdates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	// Add random notifications when active
	if notificationActive && rand.Intn(3) == 0 { // 33% chance
		notifications := []string{
			"新しいユーザーが登録しました",
			"システムの健全性チェック完了",
			"データベースバックアップ完了",
			"新しいメッセージが届きました",
			"セキュリティスキャン完了",
		}
		
		notification := notifications[rand.Intn(len(notifications))]
		notificationLogs = append(notificationLogs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), notification))
		
		// Keep only last 10 logs
		if len(notificationLogs) > 10 {
			notificationLogs = notificationLogs[len(notificationLogs)-10:]
		}
	}
	
	html := `<div class="space-y-1">`
	for i := len(notificationLogs) - 1; i >= 0; i-- {
		html += fmt.Sprintf(`<div class="text-xs text-gray-600">%s</div>`, notificationLogs[i])
	}
	if len(notificationLogs) == 0 {
		html += `<p class="text-xs text-gray-500">通知ログがここに表示されます</p>`
	}
	html += `</div>`
	
	fmt.Fprint(w, html)
}