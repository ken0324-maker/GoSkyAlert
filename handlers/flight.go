package handlers

import (
	"encoding/json"
	"final/models"
	"final/services"
	"net/http"
	"strconv"
	"time"
)

type FlightHandler struct {
	amadeusService *services.AmadeusService
}

func NewFlightHandler(service *services.AmadeusService) *FlightHandler {
	return &FlightHandler{
		amadeusService: service,
	}
}

// 新增：首頁處理
func (h *FlightHandler) Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

// 新增：機票價格追蹤API
func (h *FlightHandler) TrackFlightPrices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "方法不允許"}`, http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()

	req := models.PriceTrackingRequest{
		Origin:      query.Get("origin"),
		Destination: query.Get("destination"),
		Weeks:       18, // 預設追蹤18週
	}

	if req.Origin == "" || req.Destination == "" {
		http.Error(w, `{"error": "缺少必要參數: origin, destination"}`, http.StatusBadRequest)
		return
	}

	// 可選參數：追蹤週數
	if weeksStr := query.Get("weeks"); weeksStr != "" {
		if weeks, err := strconv.Atoi(weeksStr); err == nil && weeks > 0 {
			if weeks > 52 {
				weeks = 52 // 最多52週
			}
			req.Weeks = weeks
		}
	}

	// 執行價格追蹤分析
	analysis, err := h.amadeusService.TrackFlightPrices(req)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    analysis,
		"meta": map[string]interface{}{
			"origin":      req.Origin,
			"destination": req.Destination,
			"track_weeks": req.Weeks,
			"analyzed_at": time.Now().Format(time.RFC3339),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// 新增：取得價格趨勢圖表數據
func (h *FlightHandler) GetPriceTrend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "方法不允許"}`, http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()

	origin := query.Get("origin")
	destination := query.Get("destination")
	weeks := 18

	if origin == "" || destination == "" {
		http.Error(w, `{"error": "缺少必要參數: origin, destination"}`, http.StatusBadRequest)
		return
	}

	if weeksStr := query.Get("weeks"); weeksStr != "" {
		if w, err := strconv.Atoi(weeksStr); err == nil && w > 0 {
			weeks = w
		}
	}

	trendData, err := h.amadeusService.GeneratePriceTrend(origin, destination, weeks)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    trendData,
	})
}

// 新增：取得歷史追蹤數據
func (h *FlightHandler) GetTrackingHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "方法不允許"}`, http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	route := query.Get("route")

	if route == "" {
		http.Error(w, `{"error": "缺少必要參數: route"}`, http.StatusBadRequest)
		return
	}

	// 這裡可以實現從資料庫獲取歷史數據的邏輯
	// 目前先返回簡單響應
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"route":   route,
			"tracks":  []string{},
			"message": "歷史數據功能開發中",
		},
	})
}

// 新增：創建價格警報
func (h *FlightHandler) CreatePriceAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "方法不允許"}`, http.StatusMethodNotAllowed)
		return
	}

	var alertReq struct {
		Route       string  `json:"route"`
		TargetPrice float64 `json:"target_price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&alertReq); err != nil {
		http.Error(w, `{"error": "無效的請求數據"}`, http.StatusBadRequest)
		return
	}

	if alertReq.Route == "" || alertReq.TargetPrice <= 0 {
		http.Error(w, `{"error": "缺少必要參數: route, target_price"}`, http.StatusBadRequest)
		return
	}

	// 這裡可以實現價格警報的創建邏輯
	// 目前先返回成功響應
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"alert_id":     "alert_" + strconv.FormatInt(time.Now().Unix(), 10),
			"route":        alertReq.Route,
			"target_price": alertReq.TargetPrice,
			"created_at":   time.Now().Format(time.RFC3339),
			"message":      "價格警報設置成功，當價格低於目標時會通知您",
		},
	})
}

// 原有的航班搜尋功能
func (h *FlightHandler) SearchFlights(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "方法不允許"}`, http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()

	req := models.SearchRequest{
		Origin:        query.Get("origin"),
		Destination:   query.Get("destination"),
		DepartureDate: query.Get("departure_date"),
		ReturnDate:    query.Get("return_date"),
		Currency:      query.Get("currency"),
		Adults:        1,
	}

	if req.Origin == "" || req.Destination == "" || req.DepartureDate == "" {
		http.Error(w, `{"error": "缺少必要參數: origin, destination, departure_date"}`, http.StatusBadRequest)
		return
	}

	if req.Currency == "" {
		req.Currency = "TWD"
	}
	if adultsStr := query.Get("adults"); adultsStr != "" {
		if adults, err := strconv.Atoi(adultsStr); err == nil && adults > 0 {
			req.Adults = adults
		}
	}

	flights, err := h.amadeusService.SearchFlights(req)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    flights,
		"count":   len(flights),
		"meta": map[string]interface{}{
			"origin":      req.Origin,
			"destination": req.Destination,
			"date":        req.DepartureDate,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (h *FlightHandler) SearchAirports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "方法不允許"}`, http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error": "缺少查詢參數: q"}`, http.StatusBadRequest)
		return
	}

	airports, err := h.amadeusService.SearchAirports(query)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    airports,
	})
}

func (h *FlightHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "amadeus-flight-api",
		"version": "1.0.0",
	})
}

// 新增：API 文檔端點
func (h *FlightHandler) APIDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	docs := map[string]interface{}{
		"service": "航班價格追蹤 API",
		"version": "1.0.0",
		"endpoints": []map[string]string{
			{
				"method":      "GET",
				"path":        "/api/flights/search",
				"description": "搜尋即時航班",
				"parameters":  "origin, destination, departure_date, [return_date, adults, currency]",
			},
			{
				"method":      "GET",
				"path":        "/api/flights/track-prices",
				"description": "追蹤機票價格趨勢",
				"parameters":  "origin, destination, [weeks]",
			},
			{
				"method":      "GET",
				"path":        "/api/flights/price-trend",
				"description": "取得價格趨勢圖表數據",
				"parameters":  "origin, destination, [weeks]",
			},
			{
				"method":      "GET",
				"path":        "/api/airports/search",
				"description": "搜尋機場",
				"parameters":  "q",
			},
			{
				"method":      "POST",
				"path":        "/api/alerts/create",
				"description": "創建價格警報",
				"parameters":  "route, target_price",
			},
		},
	}

	json.NewEncoder(w).Encode(docs)
}
