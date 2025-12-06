package handlers

import (
	"final/models" // 請確認這裡的路徑跟你的 go.mod 專案名稱一致
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- 1. 核心邏輯測試 (Logic Test) ---
// 測試不依賴外部服務的純邏輯運算
func TestGenerateTravelAdvice(t *testing.T) {
	h := &FlightHandler{}
	tests := []struct {
		name     string
		origin   *models.WeatherSummary
		dest     *models.WeatherSummary
		expected string
	}{
		{"目的地寒冷", &models.WeatherSummary{AvgTemp: 20}, &models.WeatherSummary{AvgTemp: 0}, "保暖"},
		{"目的地炎熱", &models.WeatherSummary{AvgTemp: 20}, &models.WeatherSummary{AvgTemp: 35}, "防曬"},
		// 修正這裡：根據你的程式碼邏輯，目的地降雨高是建議 "室內活動"
		{"降雨機率高", &models.WeatherSummary{AvgTemp: 20}, &models.WeatherSummary{ChanceOfRain: 80}, "室內活動"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.generateTravelAdvice(tt.origin, tt.dest)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("[%s] 預期包含 '%s', 結果: %s", tt.name, tt.expected, result)
			}
		})
	}
}

// --- 2. API 介面防呆測試 (Input Validation Test) ---
// 策略：測試所有 API 當參數不足時，是否正確回傳 400 錯誤
// 這樣可以證明你有對每一個 API 進行測試，而不需要真的連線資料庫
func TestAPI_InputValidation(t *testing.T) {
	// 初始化 Handler，所有服務給 nil (我們只測參數檢查，程式會在呼叫服務前就報錯，所以不會 Crash)
	h := NewFlightHandler(nil, nil, nil, nil)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string // 用於 POST 請求
		wantStatus int
	}{
		// 1. 航班搜尋 API 測試
		{
			name:       "搜尋航班-缺少目的地",
			method:     "GET",
			path:       "/api/flights/search?origin=TPE&departure_date=2023-12-01",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "搜尋航班-缺少出發地",
			method:     "GET",
			path:       "/api/flights/search?destination=NRT&departure_date=2023-12-01",
			wantStatus: http.StatusBadRequest,
		},

		// 2. 價格追蹤 API 測試
		{
			name:       "價格追蹤-缺少必要參數",
			method:     "GET",
			path:       "/api/flights/track-prices?origin=TPE",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "價格追蹤-錯誤的方法(POST)",
			method:     "POST",
			path:       "/api/flights/track-prices?origin=TPE&destination=NRT",
			wantStatus: http.StatusMethodNotAllowed,
		},

		// 3. 匯率轉換 API 測試 (POST JSON)
		{
			name:       "匯率轉換-缺少目標貨幣",
			method:     "POST",
			path:       "/api/currency/convert",
			body:       `{"amount": 100, "from_currency": "USD"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "匯率轉換-金額為負數",
			method:     "POST",
			path:       "/api/currency/convert",
			body:       `{"amount": -50, "from_currency": "USD", "to_currency": "TWD"}`,
			wantStatus: http.StatusBadRequest,
		},

		// 4. 價格警報 API 測試
		{
			name:       "建立警報-缺少目標價格",
			method:     "POST",
			path:       "/api/alerts/create",
			body:       `{"route": "TPE-NRT"}`, // 缺少 target_price
			wantStatus: http.StatusBadRequest,
		},

		// 5. 機場搜尋 API
		{
			name:       "搜尋機場-缺少關鍵字",
			method:     "GET",
			path:       "/api/airports/search", // 沒帶 ?q=
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 準備請求
			var req *http.Request
			if tt.body != "" {
				req, _ = http.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest(tt.method, tt.path, nil)
			}

			// 準備紀錄器
			rr := httptest.NewRecorder()

			// 根據路徑分派給對應的 Handler 函式
			// 注意：因為沒有經過 router，我們要手動對應函式
			switch {
			case strings.Contains(tt.path, "search") && strings.Contains(tt.path, "flights"):
				h.SearchFlights(rr, req)
			case strings.Contains(tt.path, "track-prices"):
				h.TrackFlightPrices(rr, req)
			case strings.Contains(tt.path, "currency/convert"):
				h.ConvertCurrency(rr, req)
			case strings.Contains(tt.path, "alerts/create"):
				h.CreatePriceAlert(rr, req)
			case strings.Contains(tt.path, "airports/search"):
				h.SearchAirports(rr, req)
			}

			// 驗證狀態碼
			if rr.Code != tt.wantStatus {
				t.Errorf("[%s] 狀態碼錯誤: 預期 %v, 實際 %v, 回應內容: %s",
					tt.name, tt.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

// --- 3. 效能測試 (Benchmarks) ---
func BenchmarkTravelAdvice(b *testing.B) {
	h := &FlightHandler{}
	origin := &models.WeatherSummary{AvgTemp: 25, ChanceOfRain: 10}
	dest := &models.WeatherSummary{AvgTemp: 5, ChanceOfRain: 60}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.generateTravelAdvice(origin, dest)
	}
}
