# 智慧航班搜尋與旅行工具 (Smart Flight Search & Travel Utility)

這是一個基於 Go 語言（Golang）開發的全功能旅行輔助應用程式，集成了航班即時搜尋、價格趨勢追蹤、目的地天氣預報、貨幣匯率計算、時區時差計算以及附近景點查詢等多項實用功能。

## 主要功能

* **即時航班搜尋**：使用 Amadeus API 查詢單程或往返航班，並提供價格、航線、停留站點等詳細資訊。
* **機票價格追蹤與分析**：追蹤特定航線在未來數週的價格趨勢，提供價格分析和最佳購買日期建議。
* **目的地天氣預報**：在航班搜尋結果中整合出發地和目的地的天氣資訊，幫助規劃行程。
* **貨幣匯率計算機**：提供即時貨幣轉換和匯率查詢功能。
* **時區時差計算**：使用 WorldTimeAPI 計算兩個指定時區（例如 `Asia/Taipei` 與 `Europe/London`）之間的時差。
* **附近景點查詢**：使用 Foursquare Places API 搜尋指定地點附近的景點、餐廳、商店等。
* **Telegram 通知（基礎）**：具備發送簡單航班通知的能力。

## API 依賴

本專案依賴於以下外部服務。您需要註冊並獲取相應的 API 金鑰。

| 服務 | 用途 | 配置文件中的環境變數 | 註冊/文件連結 |
| :--- | :--- | :--- | :--- |
| **Amadeus** | 航班搜尋、價格追蹤 | `AMADEUS_API_KEY`, `AMADEUS_API_SECRET` | [Amadeus Developers](https://developers.amadeus.com/self-service/apis-docs/guides/developer-guides/quick-start/) |
| **ExchangeRate-API** | 貨幣匯率計算 | `EXCHANGE_RATE_API_KEY` | [ExchangeRate-API](https://www.exchangerate-api.com/) |
| **WeatherAPI** | 目的地天氣資訊 | `WEATHER_API_KEY` | [WeatherAPI](https://www.weatherapi.com/) |
| **Foursquare Places** | 附近景點搜尋 | `FOURSQUARE_API_KEY` | [Foursquare Developers](https://foursquare.com/developers/) |
| **WorldTimeAPI** | 時區時差計算 | *無需金鑰* | [WorldTimeAPI](http://worldtimeapi.org/) |
| **OpenStreetMap Nominatim** | 地理編碼（地址轉經緯度） | *無需金鑰* | [Nominatim](https://nominatim.org/) |
| **Telegram** | 航班通知功能 | `TELEGRAM_BOT_TOKEN` | *可選* |
| **Discord** | BOT | *無需金鑰* | 加入群駔 : https://discord.gg/55VemjSU |

## 環境設置與運行

### 1. 環境變數配置

請在專案根目錄創建一個 `.env` 文件（或直接設置您的系統環境變數），填入從各服務商獲取的金鑰。

```bash
# Amadeus API (必填 - 航班搜尋核心功能)
AMADEUS_API_KEY="YOUR_AMADEUS_KEY"
AMADEUS_API_SECRET="YOUR_AMADEUS_SECRET"
AMADEUS_BASE_URL="https://test.api.amadeus.com/v2"

# Exchange Rate API (必填 - 匯率計算功能)
EXCHANGE_RATE_API_KEY="YOUR_EXCHANGE_RATE_KEY"

# Weather API (選填 - 如果不設定，天氣功能將禁用)
WEATHER_API_KEY="YOUR_WEATHER_API_KEY"

# Foursquare API (選填 - 如果不設定，景點搜尋功能將禁用)
FOURSQUARE_API_KEY="YOUR_FOURSQUARE_API_KEY"

# Telegram Bot (選填 - 如果不設定，通知功能將禁用)
TELEGRAM_BOT_TOKEN="YOUR_TELEGRAM_BOT_TOKEN"

# 服務器配置 (預設值)
PORT="8080"
ENVIRONMENT="development"
```

Discord 指令
邀請您的 Bot 進入伺服器後，可使用以下指令：
|指令|說明|範例|
|:---:|:---:|:---:|
|/help|顯示所有指令說明|/help|
|/price|查詢航班與價格分析|/price TPE NRT 2025-12-01|
|/weather|查詢城市天氣|/weather Tokyo|
|/rate|查詢即時匯率|/rate USD TWD|
|/spot|查詢附近景點|/spot 大阪|

(您也可以使用 ! 作為前綴，例如 !price)


|目錄/文件|說明|
|:---:|:---:|
|main.go|應用程式入口點，負責初始化配置、服務和路由。|
|config/config.go|載入和驗證環境變數配置。|
|handlers/|處理 HTTP 請求和響應的邏輯層。|
|handlers/flight.go|處理航班搜尋、價格追蹤、天氣和匯率相關的路由。|
|handlers/attraction.go|處理景點搜尋和類別查詢的路由。|
|handlers/timezone.go|處理時差計算的路由。|
|services/|處理業務邏輯和外部 API 交互的服務層。|
|services/amadeus.go|Amadeus API 相關邏輯（航班、價格趨勢）。|
|services/exchangeService.go|匯率 API 相關邏輯。|
|services/foursquare_service.go|Foursquare Places API 相關邏輯（景點搜尋）。|
|services/weather_service.go|天氣 API 相關邏輯。|
|services/timezone_service.go|時區 API 相關邏輯。|
|services/telegram.go|Telegram 通知發送邏輯。|
|models/|定義請求和響應的數據結構。|
|static/|存放靜態文件（CSS, JS 等）。|
|templates/|存放 HTML 模板文件，index.html 為前端單頁應用。|

