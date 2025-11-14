<<<<<<< HEAD
AMADEUS_API_KEY 和 AMADEUS_API_SECRET => https://developers.amadeus.com/self-service/apis-docs/guides/developer-guides/quick-start/
10/28 新增機票趨勢(amadeus API)
11/3  新增天氣功能(www.weatherapi.com)
11/10 新增匯率功能(https://www.exchangerate-api.com)
=======
AMADEUS_API_KEY and AMADEUS_API_SECRET => https://developers.amadeus.com/self-service/apis-docs/guides/developer-guides/quick-start/
EXCHANGE_RATE_API_KEY => https://www.exchangerate-api.com/
WEATHER_API_KEY => https://www.weatherapi.com/
>>>>>>> dc58b5785a21de51743ac71e711ad207629129b6

|路徑/檔案|說明|
|:---:|:---:|
|main.go|專案入口點：載入配置、初始化服務、設置路由並啟動 HTTP 伺服器。|
|config.go|應用程式配置：從環境變數載入 API Key 和伺服器設定。|
|handlers/ (Implied)|處理 HTTP 請求的邏輯層 (例如 flightHandler.go)。|
|services/amadeus.go|處理 Amadeus API 的邏輯：航班搜尋、價格追蹤、機場搜尋。|
|services/weather_service.go|處理 WeatherAPI 的邏輯：獲取天氣預報、生成旅行建議。|
|services/exchangeService.go|處理匯率 API 的邏輯：貨幣轉換、獲取支援貨幣。|
|services/timezone_service.go|處理 WorldTimeAPI 的邏輯：計算時區差異。|
|models/flight.go|數據模型定義：包含搜尋請求、航班結果、價格分析、天氣/匯率資訊等結構。|
|static/index.html|應用程式的根 HTML 檔案和介面結構。|
|static/css/style.css|網頁的樣式文件。|
|static/js/app.js|網頁的客戶端 JavaScript 邏輯：處理表單提交、API 呼叫、結果渲染、標籤切換等。|

```python
# 請在專案根目錄下建立一個 .env 檔案（或直接使用環境變數），並填寫您的 API Key
# ----------------------------------------------------
# 核心配置 (Core Configuration)
# ----------------------------------------------------

# Amadeus API
AMADEUS_API_KEY="YOUR_AMADEUS_API_KEY"
AMADEUS_API_SECRET="YOUR_AMADEUS_API_SECRET"
AMADEUS_BASE_URL="https://test.api.amadeus.com/v2" # 預設為測試環境


# WeatherAPI.com API (啟用天氣功能)
WEATHER_API_KEY="YOUR_WEATHER_API_KEY"

# ExchangeRate-API.com API (啟用匯率計算功能)
EXCHANGE_RATE_API_KEY="YOUR_EXCHANGE_RATE_API_KEY"

# ----------------------------------------------------
# 伺服器配置 (Server Configuration)
# ----------------------------------------------------
PORT="8080"
ENVIRONMENT="development"
LOG_LEVEL="info"
```
