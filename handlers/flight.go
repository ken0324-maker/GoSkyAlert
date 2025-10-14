package handlers

import (
	"encoding/json"
	"final/models"
	"final/services"
	"net/http"
	"strconv"
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

// 其餘程式碼保持不變...
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
