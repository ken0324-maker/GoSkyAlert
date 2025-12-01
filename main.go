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

	if err := cfg.Validate(); err != nil {
		log.Printf("⚠️ 配置警告: %v", err)
	}

	log.Printf("✅ 配置載入成功")
	log.Printf("🌍 環境: %s", cfg.Environment)

	// 初始化 Amadeus 服務
	amadeusService := services.NewAmadeusService(cfg)

	// 初始化其他服務 (天氣、匯率、Foursquare)
	var weatherService *services.WeatherService
	if cfg.HasWeatherAPI() {
		weatherService = services.NewWeatherService(cfg.WeatherAPIKey)
		log.Printf("🌤️ 天氣服務已初始化")
	}

	var exchangeService *services.ExchangeService
	if cfg.HasExchangeRateAPI() {
		exchangeService = services.NewExchangeService(cfg.ExchangeRateAPIKey)
		log.Printf("💱 匯率服務已初始化")
	}

	var foursquareService *services.FoursquareService
	if cfg.HasFoursquareAPI() {
		foursquareService = services.NewFoursquareService(cfg.FoursquareAPIKey)
		log.Printf("🏛️  景點服務已初始化")
	}

	// [新增] 初始化 Discord Bot
	if cfg.HasDiscordAPI() {
		discordService, err := services.NewDiscordService(cfg.DiscordBotToken, amadeusService)
		if err != nil {
			log.Printf("❌ Discord 服務初始化失敗: %v", err)
		} else {
			// 啟動 Discord 連線
			if err := discordService.Start(); err != nil {
				log.Printf("❌ Discord 連線失敗: %v", err)
			} else {
				// 程式結束時關閉連線
				defer discordService.Stop()
			}
		}
	} else {
		log.Printf("⚠️ 未設定 DISCORD_BOT_TOKEN，Bot 功能已禁用")
	}

	// 初始化 Handler
	flightHandler := handlers.NewFlightHandler(amadeusService, weatherService, exchangeService, foursquareService)

	// 設置路由
	setupRoutes(flightHandler)

	// 啟動伺服器
	serverAddress := cfg.GetServerAddress()
	log.Printf("🚀 伺服器啟動在 http://localhost%s", serverAddress)

	log.Fatal(http.ListenAndServe(serverAddress, nil))
}

func setupRoutes(flightHandler *handlers.FlightHandler) {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	templateFs := http.FileServer(http.Dir("./templates"))
	http.Handle("/templates/", http.StripPrefix("/templates/", templateFs))

	http.HandleFunc("/", flightHandler.Index)
	http.HandleFunc("/api/flights/search", flightHandler.SearchFlights)
	http.HandleFunc("/api/flights/track-prices", flightHandler.TrackFlightPrices)
	http.HandleFunc("/api/flights/price-trend", flightHandler.GetPriceTrend)
	http.HandleFunc("/api/airports/search", flightHandler.SearchAirports)
	http.HandleFunc("/api/alerts/create", flightHandler.CreatePriceAlert)
	http.HandleFunc("/api/currency/convert", flightHandler.ConvertCurrency)
	http.HandleFunc("/api/currency/supported", flightHandler.GetSupportedCurrencies)
	http.HandleFunc("/api/attractions/search", flightHandler.SearchAttractions)
	http.HandleFunc("/api/attractions/categories", flightHandler.GetAttractionCategories)
	http.HandleFunc("/api/docs", flightHandler.APIDocs)
	http.HandleFunc("/health", flightHandler.HealthCheck)
	http.HandleFunc("/timediff", handlers.TimeDiffHandler)
}
