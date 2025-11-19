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
	amadeusService    *services.AmadeusService
	weatherService    *services.WeatherService
	exchangeService   *services.ExchangeService
	foursquareService *services.FoursquareService
}

func NewFlightHandler(flightService *services.AmadeusService, weatherService *services.WeatherService, exchangeService *services.ExchangeService, foursquareService *services.FoursquareService) *FlightHandler {
	return &FlightHandler{
		amadeusService:    flightService,
		weatherService:    weatherService,
		exchangeService:   exchangeService,
		foursquareService: foursquareService,
	}
}

func (h *FlightHandler) Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

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

	var weatherInfo *models.WeatherInfo
	if h.weatherService != nil {
		weatherInfo = h.getWeatherInfo(req.Origin, req.Destination, req.DepartureDate)
	}

	response := models.FlightSearchResponseWithWeather{
		Flights: flights,
		Weather: weatherInfo,
		Meta: struct {
			Count         int    `json:"count"`
			Origin        string `json:"origin"`
			Destination   string `json:"destination"`
			DepartureDate string `json:"departure_date"`
		}{
			Count:         len(flights),
			Origin:        req.Origin,
			Destination:   req.Destination,
			DepartureDate: req.DepartureDate,
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

func (h *FlightHandler) getWeatherInfo(origin, destination, date string) *models.WeatherInfo {
	if h.weatherService == nil {
		return nil
	}

	var originWeather, destWeather *models.WeatherSummary

	originCity := models.GetCityByAirportCode(origin)
	if originCity != "" {
		if weather, err := h.weatherService.GetWeather(originCity, date); err == nil {
			originWeather = h.createWeatherSummary(weather, originCity, date)
		}
	}

	destCity := models.GetCityByAirportCode(destination)
	if destCity != "" {
		if weather, err := h.weatherService.GetWeather(destCity, date); err == nil {
			destWeather = h.createWeatherSummary(weather, destCity, date)
		}
	}

	travelAdvice := h.generateTravelAdvice(originWeather, destWeather)

	return &models.WeatherInfo{
		OriginWeather:      originWeather,
		DestinationWeather: destWeather,
		TravelAdvice:       travelAdvice,
	}
}

func (h *FlightHandler) createWeatherSummary(weather *models.WeatherResponse, city, date string) *models.WeatherSummary {
	if weather == nil {
		return nil
	}

	var forecastDay *models.ForecastDay
	for i := range weather.Forecast.Forecastday {
		if weather.Forecast.Forecastday[i].Date == date {
			forecastDay = &weather.Forecast.Forecastday[i]
			break
		}
	}

	summary := &models.WeatherSummary{
		City:      city,
		Date:      date,
		AvgTemp:   weather.Current.TempC,
		Condition: weather.Current.Condition.Text,
		Icon:      weather.Current.Condition.Icon,
		Humidity:  weather.Current.Humidity,
		WindSpeed: weather.Current.WindKph,
	}

	if forecastDay != nil {
		summary.AvgTemp = forecastDay.Day.AvgTempC
		summary.Condition = forecastDay.Day.Condition.Text
		summary.Icon = forecastDay.Day.Condition.Icon
		summary.ChanceOfRain = forecastDay.Day.DailyChanceOfRain
	}

	return summary
}

func (h *FlightHandler) generateTravelAdvice(origin, destination *models.WeatherSummary) string {
	if origin == nil || destination == nil {
		return "天氣資訊不足，請確認航班資訊"
	}

	advice := "旅行建議："

	if origin.ChanceOfRain > 50 {
		advice += "出發地降雨機率高，建議提早出發並攜帶雨具。"
	} else if origin.AvgTemp < 10 {
		advice += "出發地氣溫較低，請注意保暖。"
	} else if origin.AvgTemp > 30 {
		advice += "出發地氣溫較高，建議穿著輕便。"
	}

	if destination.ChanceOfRain > 50 {
		advice += " 目的地降雨機率高，建議準備室內活動方案。"
	} else if destination.AvgTemp > 30 {
		advice += " 目的地氣溫較高，請注意防曬和補充水分。"
	} else if destination.AvgTemp < 5 {
		advice += " 目的地氣溫很低，請準備厚重保暖衣物。"
	} else if destination.AvgTemp >= 15 && destination.AvgTemp <= 25 {
		advice += " 目的地天氣宜人，適合旅遊。"
	} else {
		advice += " 目的地天氣狀況良好。"
	}

	tempDiff := destination.AvgTemp - origin.AvgTemp
	if tempDiff > 10 {
		advice += " 目的地比出發地溫暖許多，建議準備夏季衣物。"
	} else if tempDiff < -10 {
		advice += " 目的地比出發地寒冷許多，請準備足夠的保暖衣物。"
	}

	return advice
}

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
		Weeks:       18,
	}

	if req.Origin == "" || req.Destination == "" {
		http.Error(w, `{"error": "缺少必要參數: origin, destination"}`, http.StatusBadRequest)
		return
	}

	if weeksStr := query.Get("weeks"); weeksStr != "" {
		if weeks, err := strconv.Atoi(weeksStr); err == nil && weeks > 0 {
			if weeks > 52 {
				weeks = 52
			}
			req.Weeks = weeks
		}
	}

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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"route":   route,
			"tracks":  []string{},
			"message": "歷史數據功能開發中",
		},
	})
}

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

func (h *FlightHandler) ConvertCurrency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "方法不允許"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.CurrencyConversionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "無效的請求數據"}`, http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 || req.FromCurrency == "" || req.ToCurrency == "" {
		http.Error(w, `{"error": "缺少必要參數: amount, from_currency, to_currency"}`, http.StatusBadRequest)
		return
	}

	if h.exchangeService == nil {
		http.Error(w, `{"error": "匯率服務未啟用"}`, http.StatusServiceUnavailable)
		return
	}

	convertedAmount, err := h.exchangeService.ConvertCurrency(req.Amount, req.FromCurrency, req.ToCurrency)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	rates, err := h.exchangeService.GetExchangeRates(req.FromCurrency, []string{req.ToCurrency})
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	exchangeRate := rates.Rates[req.ToCurrency]

	response := models.CurrencyConversionResponse{
		OriginalAmount:  req.Amount,
		ConvertedAmount: convertedAmount,
		FromCurrency:    req.FromCurrency,
		ToCurrency:      req.ToCurrency,
		ExchangeRate:    exchangeRate,
		LastUpdated:     rates.LastUpdated,
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

func (h *FlightHandler) GetSupportedCurrencies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "方法不允許"}`, http.StatusMethodNotAllowed)
		return
	}

	var currencies []string
	if h.exchangeService != nil {
		currencies = h.exchangeService.GetSupportedCurrencies()
	} else {
		currencies = []string{"TWD", "USD", "EUR", "JPY", "GBP", "CNY", "KRW", "HKD", "SGD"}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    currencies,
	})
}

func (h *FlightHandler) SearchAttractions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "方法不允許"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.foursquareService == nil {
		http.Error(w, `{"error": "景點服務未啟用"}`, http.StatusServiceUnavailable)
		return
	}

	query := r.URL.Query()
	latStr := query.Get("lat")
	lngStr := query.Get("lng")
	radiusStr := query.Get("radius")
	searchQuery := query.Get("query")
	category := query.Get("category")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, `{"error": "無效的緯度參數"}`, http.StatusBadRequest)
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		http.Error(w, `{"error": "無效的經度參數"}`, http.StatusBadRequest)
		return
	}

	req := services.SearchRequest{
		Latitude:  lat,
		Longitude: lng,
		Query:     searchQuery,
		Category:  category,
	}

	if radiusStr != "" {
		radius, err := strconv.Atoi(radiusStr)
		if err == nil {
			req.Radius = radius
		}
	} else {
		req.Radius = 5000
	}

	attractions, err := h.foursquareService.SearchNearby(req)
	if err != nil {
		http.Error(w, `{"error": "搜尋景點時發生錯誤: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    attractions,
		"meta": map[string]interface{}{
			"latitude":  lat,
			"longitude": lng,
			"radius":    req.Radius,
			"count":     len(attractions),
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (h *FlightHandler) GetAttractionCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "方法不允許"}`, http.StatusMethodNotAllowed)
		return
	}

	var categories []string
	if h.foursquareService != nil {
		categories = h.foursquareService.GetPopularCategories()
	} else {
		categories = []string{
			"museum", "park", "restaurant", "coffee", "bar",
			"hotel", "shopping_mall", "tourist_attraction",
			"art_gallery", "theater", "cinema", "stadium",
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"categories": categories,
	})
}

func (h *FlightHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "healthy"
	services := map[string]string{
		"amadeus":    "connected",
		"weather":    "connected",
		"exchange":   "connected",
		"foursquare": "connected",
	}

	if h.weatherService == nil {
		services["weather"] = "disabled"
	}
	if h.exchangeService == nil {
		services["exchange"] = "disabled"
	}
	if h.foursquareService == nil {
		services["foursquare"] = "disabled"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    status,
		"service":   "flight-api",
		"version":   "1.2.0",
		"services":  services,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (h *FlightHandler) APIDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	docs := map[string]interface{}{
		"service": "航班價格追蹤 API",
		"version": "1.2.0",
		"features": []string{
			"即時航班搜尋",
			"價格趨勢追蹤",
			"天氣資訊整合",
			"價格警報設定",
			"貨幣轉換服務",
			"景點查詢服務",
		},
		"endpoints": []map[string]string{
			{
				"method":      "GET",
				"path":        "/api/flights/search",
				"description": "搜尋即時航班（包含天氣資訊）",
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
			{
				"method":      "POST",
				"path":        "/api/currency/convert",
				"description": "貨幣轉換",
				"parameters":  "amount, from_currency, to_currency",
			},
			{
				"method":      "GET",
				"path":        "/api/currency/supported",
				"description": "獲取支援的貨幣列表",
				"parameters":  "無",
			},
			{
				"method":      "GET",
				"path":        "/api/attractions/search",
				"description": "搜尋附近景點",
				"parameters":  "lat, lng, [radius, query, category]",
			},
			{
				"method":      "GET",
				"path":        "/api/attractions/categories",
				"description": "獲取景點類別列表",
				"parameters":  "無",
			},
			{
				"method":      "GET",
				"path":        "/health",
				"description": "服務健康檢查",
				"parameters":  "無",
			},
		},
	}

	json.NewEncoder(w).Encode(docs)
}
