package main

import (
	"final/config"
	"final/handlers"
	"final/services"
	"log"
	"net/http"
)

func main() {
	// 載入配置
	cfg := config.LoadConfig()

	// 驗證配置（改為警告而非致命錯誤）
	if err := cfg.Validate(); err != nil {
		log.Printf("⚠️ 配置警告: %v", err)
		// 不直接退出，讓服務繼續運行
	}

	log.Printf("✅ 配置載入成功")
	log.Printf("🌍 環境: %s", cfg.Environment)
	log.Printf("🔧 Amadeus API: %s", cfg.AmadeusBaseURL)

	// 初始化服務
	amadeusService := services.NewAmadeusService(cfg)

	// 新增：初始化天氣服務
	var weatherService *services.WeatherService
	if cfg.HasWeatherAPI() {
		weatherService = services.NewWeatherService(cfg.WeatherAPIKey)
		log.Printf("🌤️ 天氣服務已初始化")

		// 驗證天氣 API 金鑰
		if err := weatherService.ValidateAPIKey(); err != nil {
			log.Printf("❌ 天氣API金鑰驗證失敗: %v", err)
			log.Printf("⚠️ 天氣功能將被禁用")
			weatherService = nil
		} else {
			log.Printf("✅ 天氣API金鑰驗證成功")
		}
	} else {
		log.Printf("⚠️ 未設定WeatherAPI金鑰，天氣功能已禁用")
		log.Printf("💡 請設定 WEATHER_API_KEY 環境變數以啟用天氣功能")
	}

	// 修正：傳入天氣服務（第二個參數）
	flightHandler := handlers.NewFlightHandler(amadeusService, weatherService)

	// 設置路由
	setupRoutes(flightHandler)

	// 啟動伺服器
	serverAddress := cfg.GetServerAddress()
	log.Printf("🚀 伺服器啟動在 http://localhost%s", serverAddress)
	log.Printf("📊 航班搜尋服務已就緒")
	log.Printf("📈 價格追蹤服務已就緒")

	// 新增：顯示天氣服務狀態
	if weatherService != nil {
		log.Printf("🌤️ 天氣服務已就緒")
	} else {
		log.Printf("🌤️ 天氣服務已禁用")
	}

	log.Printf("===========================================")
	log.Printf("📋 可用端點:")
	log.Printf("   GET  /                         - 首頁")
	log.Printf("   GET  /api/flights/search       - 搜尋航班（含天氣）")
	log.Printf("   GET  /api/flights/track-prices - 追蹤價格")
	log.Printf("   GET  /api/flights/price-trend  - 價格趨勢")
	log.Printf("   GET  /api/airports/search      - 搜尋機場")
	log.Printf("   POST /api/alerts/create        - 創建警報")
	log.Printf("   GET  /health                   - 健康檢查")
	log.Printf("   GET  /api/docs                 - API文檔")
	log.Printf("===========================================")

	log.Fatal(http.ListenAndServe(serverAddress, nil))
}

func setupRoutes(flightHandler *handlers.FlightHandler) {
	// 靜態文件服務
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 模板文件服務
	templateFs := http.FileServer(http.Dir("./templates"))
	http.Handle("/templates/", http.StripPrefix("/templates/", templateFs))

	// API 路由
	http.HandleFunc("/", flightHandler.Index)
	http.HandleFunc("/api/flights/search", flightHandler.SearchFlights)
	http.HandleFunc("/api/flights/track-prices", flightHandler.TrackFlightPrices)
	http.HandleFunc("/api/flights/price-trend", flightHandler.GetPriceTrend)
	http.HandleFunc("/api/flights/tracking-history", flightHandler.GetTrackingHistory)
	http.HandleFunc("/api/airports/search", flightHandler.SearchAirports)
	http.HandleFunc("/api/alerts/create", flightHandler.CreatePriceAlert)
	http.HandleFunc("/api/docs", flightHandler.APIDocs)
	http.HandleFunc("/health", flightHandler.HealthCheck)
}
