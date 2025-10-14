package main

import (
	"final/config"
	"final/handlers"
	"final/services"
	"log"
	"net/http"
	"os"
)

func main() {
	// 檢查環境變量
	if os.Getenv("AMADEUS_API_KEY") == "" || os.Getenv("AMADEUS_API_SECRET") == "" {
		log.Fatal("❌ 請設置 AMADEUS_API_KEY 和 AMADEUS_API_SECRET 環境變量")
	}

	// 加載配置
	cfg := config.LoadConfig()

	// 初始化服務
	amadeusService := services.NewAmadeusService(cfg)
	flightHandler := handlers.NewFlightHandler(amadeusService)

	// 設置靜態檔案服務
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 設置路由
	http.HandleFunc("/api/flights/search", flightHandler.SearchFlights)
	http.HandleFunc("/api/airports/search", flightHandler.SearchAirports)
	http.HandleFunc("/health", flightHandler.HealthCheck)
	http.HandleFunc("/", flightHandler.Index) // 新增首頁處理

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Amadeus航班API服務啟動於 http://localhost:%s", port)
	log.Printf("🎨 網頁UI已啟用: http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
