package handlers

import (
	"encoding/json"
	"final/services"
	"net/http"
	"strconv"
)

type AttractionHandler struct {
	foursquareService *services.FoursquareService
}

func NewAttractionHandler(foursquareService *services.FoursquareService) *AttractionHandler {
	return &AttractionHandler{
		foursquareService: foursquareService,
	}
}

// 搜尋附近景點
func (h *AttractionHandler) SearchAttractions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析查詢參數
	query := r.URL.Query()
	latStr := query.Get("lat")
	lngStr := query.Get("lng")
	radiusStr := query.Get("radius")
	searchQuery := query.Get("query")
	category := query.Get("category")

	// 轉換座標
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "無效的緯度參數", http.StatusBadRequest)
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		http.Error(w, "無效的經度參數", http.StatusBadRequest)
		return
	}

	req := services.SearchRequest{
		Latitude:  lat,
		Longitude: lng,
		Query:     searchQuery,
		Category:  category,
	}

	// 設定半徑
	if radiusStr != "" {
		radius, err := strconv.Atoi(radiusStr)
		if err == nil {
			req.Radius = radius
		}
	} else {
		req.Radius = 5000 // 預設 5公里
	}

	// 呼叫 Foursquare 服務
	attractions, err := h.foursquareService.SearchNearby(req)
	if err != nil {
		http.Error(w, "搜尋景點時發生錯誤: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回 JSON 回應
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"data":    attractions,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "編碼回應時發生錯誤", http.StatusInternalServerError)
	}
}

// 獲取景點類別列表
func (h *AttractionHandler) GetAttractionCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories := h.foursquareService.GetPopularCategories()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success":    true,
		"categories": categories,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "編碼回應時發生錯誤", http.StatusInternalServerError)
	}
}
