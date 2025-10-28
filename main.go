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

	// 驗證配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ 配置驗證失敗: %v", err)
	}

	log.Printf("✅ 配置載入成功")
	log.Printf("🌍 環境: %s", cfg.Environment)
	log.Printf("🔧 Amadeus API: %s", cfg.AmadeusBaseURL)

	// 初始化服務
	amadeusService := services.NewAmadeusService(cfg)
	flightHandler := handlers.NewFlightHandler(amadeusService)

	// 設置路由
	setupRoutes(flightHandler)

	// 啟動伺服器
	serverAddress := cfg.GetServerAddress()
	log.Printf("🚀 伺服器啟動在 http://localhost%s", serverAddress)
	log.Printf("📊 航班搜尋服務已就緒")
	log.Printf("📈 價格追蹤服務已就緒")

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
