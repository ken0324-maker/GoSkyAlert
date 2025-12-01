package models

import "time"

// 搜尋請求
type SearchRequest struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
	ReturnDate    string `json:"return_date,omitempty"`
	Adults        int    `json:"adults"`
	Currency      string `json:"currency"`
}

// 新增：價格追蹤請求
type PriceTrackingRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Weeks       int    `json:"weeks,omitempty"`
}

// 新增：價格數據點
type PricePoint struct {
	Week     int       `json:"week"`
	Date     time.Time `json:"date"`
	Price    float64   `json:"price"`
	Currency string    `json:"currency"`
	MinPrice float64   `json:"min_price,omitempty"`
	MaxPrice float64   `json:"max_price,omitempty"`
}

// 新增：價格分析結果
type PriceAnalysis struct {
	Route          string       `json:"route"`
	TrackWeeks     int          `json:"track_weeks"`
	DataPoints     []PricePoint `json:"data_points"`
	MinPrice       float64      `json:"min_price"`
	MaxPrice       float64      `json:"max_price"`
	AvgPrice       float64      `json:"avg_price"`
	BestDate       time.Time    `json:"best_date"`
	Recommendation string       `json:"recommendation"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// 新增：價格趨勢圖表數據
type PriceTrend struct {
	Route    string         `json:"route"`
	Weeks    int            `json:"weeks"`
	Labels   []string       `json:"labels"`    // 日期標籤
	Prices   []float64      `json:"prices"`    // 價格數據
	WeekNums []int          `json:"week_nums"` // 週數（重命名避免衝突）
	Summary  *PriceAnalysis `json:"summary"`   // 分析摘要
}

// 新增：追蹤任務狀態
type TrackingTask struct {
	ID          string               `json:"id"`
	Request     PriceTrackingRequest `json:"request"`
	Status      string               `json:"status"`   // running, completed, error
	Progress    int                  `json:"progress"` // 0-100
	CurrentWeek int                  `json:"current_week"`
	Analysis    *PriceAnalysis       `json:"analysis,omitempty"`
	StartedAt   time.Time            `json:"started_at"`
	CompletedAt *time.Time           `json:"completed_at,omitempty"`
}

// 新增：價格警報設定
type PriceAlert struct {
	ID          string     `json:"id"`
	Route       string     `json:"route"`
	TargetPrice float64    `json:"target_price"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	TriggeredAt *time.Time `json:"triggered_at,omitempty"`
}

// 新增：歷史價格記錄
type HistoricalPrice struct {
	ID         string    `json:"id"`
	Route      string    `json:"route"`
	SearchDate time.Time `json:"search_date"`
	TravelDate time.Time `json:"travel_date"`
	Price      float64   `json:"price"`
	Currency   string    `json:"currency"`
	Airline    string    `json:"airline,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// 航班報價
type FlightOffer struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Price struct {
		Total    string `json:"total"`
		Currency string `json:"currency"`
	} `json:"price"`
	Itineraries      []Itinerary       `json:"itineraries"`
	TravelerPricings []TravelerPricing `json:"travelerPricings"`
}

type Itinerary struct {
	Duration string    `json:"duration"`
	Segments []Segment `json:"segments"`
}

type Segment struct {
	Departure struct {
		IATACode string `json:"iataCode"` // ✅ 修正：IATACode (不能有空格)
		Terminal string `json:"terminal"`
		At       string `json:"at"`
	} `json:"departure"`
	Arrival struct {
		IATACode string `json:"iataCode"` // ✅ 修正：IATACode (不能有空格)
		Terminal string `json:"terminal"`
		At       string `json:"at"`
	} `json:"arrival"`
	CarrierCode string `json:"carrierCode"`
	Number      string `json:"number"`
	Aircraft    struct {
		Code string `json:"code"`
	} `json:"aircraft"`
	Operating struct {
		CarrierCode string `json:"carrierCode"`
	} `json:"operating"`
	Duration string `json:"duration"`
}

type TravelerPricing struct {
	TravelerID string `json:"travelerId"`
	Price      struct {
		Total    string `json:"total"`
		Currency string `json:"currency"`
	} `json:"price"`
}

// 統一的航班響應格式
type Flight struct {
	ID           string  `json:"id"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	Airline      string  `json:"airline"`
	FlightNumber string  `json:"flight_number"`
	From         Airport `json:"from"`
	To           Airport `json:"to"`
	// [修改] 改用 string 以保持原始當地時間 (解決 08:06 問題)
	Departure string `json:"departure"`
	Arrival   string `json:"arrival"`
	Duration  string `json:"duration"`
	Stops     int    `json:"stops"`
	Aircraft  string `json:"aircraft"`
	DeepLink  string `json:"deep_link,omitempty"`
}

type Airport struct {
	Code     string `json:"code"`
	Name     string `json:"name,omitempty"`
	City     string `json:"city,omitempty"`
	Terminal string `json:"terminal,omitempty"`
}

// Amadeus API 響應
type AmadeusFlightOffersResponse struct {
	Data []FlightOffer `json:"data"`
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
}

type AirportResponse struct {
	Data []struct {
		IATACode string `json:"iataCode"`
		Name     string `json:"name"`
		Address  struct {
			CityName string `json:"cityName"`
		} `json:"address"`
	} `json:"data"`
}

// 新增：API 響應格式
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// 新增：追蹤進度響應
type TrackingProgressResponse struct {
	TaskID      string `json:"task_id"`
	Status      string `json:"status"`
	Progress    int    `json:"progress"`
	CurrentWeek int    `json:"current_week"`
	TotalWeeks  int    `json:"total_weeks"`
	Message     string `json:"message"`
}

// 新增：價格比較結果
type PriceComparison struct {
	CurrentPrice   float64   `json:"current_price"`
	HistoricalLow  float64   `json:"historical_low"`
	AveragePrice   float64   `json:"average_price"`
	Savings        float64   `json:"savings"`
	SavingsPercent float64   `json:"savings_percent"`
	IsGoodDeal     bool      `json:"is_good_deal"`
	Recommendation string    `json:"recommendation"`
	ComparedDate   time.Time `json:"compared_date"`
}

// 新增：季節性價格模式
type SeasonalPattern struct {
	Month       int     `json:"month"`
	MonthName   string  `json:"month_name"`
	Multiplier  float64 `json:"multiplier"`
	Description string  `json:"description"`
}

// 新增：航線基礎資訊
type RouteInfo struct {
	Origin           string            `json:"origin"`
	Destination      string            `json:"destination"`
	BasePrice        float64           `json:"base_price"`
	Distance         int               `json:"distance"`   // 公里
	Popularity       int               `json:"popularity"` // 1-10
	SeasonalPatterns []SeasonalPattern `json:"seasonal_patterns,omitempty"`
}

// 新增：天氣相關模型

// 天氣請求
type WeatherRequest struct {
	City    string `json:"city"`
	Date    string `json:"date"` // 格式: YYYY-MM-DD
	Airport string `json:"airport,omitempty"`
}

// 天氣響應
type WeatherResponse struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		Region  string `json:"region"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		TempF     float64 `json:"temp_f"`
		Condition struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
		Humidity    int     `json:"humidity"`
		WindKph     float64 `json:"wind_kph"`
		WindDir     string  `json:"wind_dir"`
		FeelsLikeC  float64 `json:"feelslike_c"`
		FeelsLikeF  float64 `json:"feelslike_f"`
		UV          float64 `json:"uv"`
		VisKM       float64 `json:"vis_km"`
		PrecipMM    float64 `json:"precip_mm"`
		Cloud       int     `json:"cloud"`
		LastUpdated string  `json:"last_updated"`
	} `json:"current"`
	Forecast struct {
		Forecastday []ForecastDay `json:"forecastday"`
	} `json:"forecast"`
}

// 預報天數
type ForecastDay struct {
	Date      string `json:"date"`
	DateEpoch int64  `json:"date_epoch"`
	Day       struct {
		MaxTempC          float64 `json:"maxtemp_c"`
		MaxTempF          float64 `json:"maxtemp_f"`
		MinTempC          float64 `json:"mintemp_c"`
		MinTempF          float64 `json:"mintemp_f"`
		AvgTempC          float64 `json:"avgtemp_c"`
		AvgTempF          float64 `json:"avgtemp_f"`
		MaxWindKph        float64 `json:"maxwind_kph"`
		TotalPrecipMM     float64 `json:"totalprecip_mm"`
		AvgVisKM          float64 `json:"avgvis_km"`
		AvgHumidity       float64 `json:"avghumidity"`
		DailyWillItRain   int     `json:"daily_will_it_rain"`
		DailyChanceOfRain int     `json:"daily_chance_of_rain"`
		DailyWillItSnow   int     `json:"daily_will_it_snow"`
		DailyChanceOfSnow int     `json:"daily_chance_of_snow"`
		Condition         struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
		UV float64 `json:"uv"`
	} `json:"day"`
	Hour []HourlyForecast `json:"hour"`
}

// 小時預報
type HourlyForecast struct {
	TimeEpoch int64   `json:"time_epoch"`
	Time      string  `json:"time"`
	TempC     float64 `json:"temp_c"`
	TempF     float64 `json:"temp_f"`
	Condition struct {
		Text string `json:"text"`
		Icon string `json:"icon"`
		Code int    `json:"code"`
	} `json:"condition"`
	WindKph      float64 `json:"wind_kph"`
	WindDir      string  `json:"wind_dir"`
	Humidity     int     `json:"humidity"`
	Cloud        int     `json:"cloud"`
	FeelsLikeC   float64 `json:"feelslike_c"`
	FeelsLikeF   float64 `json:"feelslike_f"`
	WindchillC   float64 `json:"windchill_c"`
	WindchillF   float64 `json:"windchill_f"`
	HeatindexC   float64 `json:"heatindex_c"`
	HeatindexF   float64 `json:"heatindex_f"`
	DewpointC    float64 `json:"dewpoint_c"`
	DewpointF    float64 `json:"dewpoint_f"`
	WillItRain   int     `json:"will_it_rain"`
	ChanceOfRain int     `json:"chance_of_rain"`
	WillItSnow   int     `json:"will_it_snow"`
	ChanceOfSnow int     `json:"chance_of_snow"`
	VisKM        float64 `json:"vis_km"`
	VisMiles     float64 `json:"vis_miles"`
	GustKph      float64 `json:"gust_kph"`
	UV           float64 `json:"uv"`
	ShortRad     float64 `json:"short_rad"`
	DiffRad      float64 `json:"diff_rad"`
}

// 航班搜尋響應（包含天氣）
type FlightSearchResponseWithWeather struct {
	Flights []Flight     `json:"flights"`
	Weather *WeatherInfo `json:"weather,omitempty"`
	Meta    struct {
		Count         int    `json:"count"`
		Origin        string `json:"origin"`
		Destination   string `json:"destination"`
		DepartureDate string `json:"departure_date"`
	} `json:"meta"`
}

// 天氣資訊摘要
type WeatherInfo struct {
	OriginWeather      *WeatherSummary `json:"origin_weather,omitempty"`
	DestinationWeather *WeatherSummary `json:"destination_weather,omitempty"`
	TravelAdvice       string          `json:"travel_advice,omitempty"`
}

// 天氣摘要（用於前端顯示）
type WeatherSummary struct {
	City         string  `json:"city"`
	Date         string  `json:"date"`
	AvgTemp      float64 `json:"avg_temp"` // 平均溫度
	Condition    string  `json:"condition"`
	Icon         string  `json:"icon"`
	Humidity     int     `json:"humidity"`
	WindSpeed    float64 `json:"wind_speed"` // km/h
	ChanceOfRain int     `json:"chance_of_rain"`
	Description  string  `json:"description"`
}

// 新增：機場代碼到城市名稱的映射
var AirportCityMap = map[string]string{
	// 台灣機場
	"TPE": "Taipei",
	"TSA": "Taipei",
	"KHH": "Kaohsiung",
	"RMQ": "Taichung",
	"TNN": "Tainan",
	"KNH": "Kinmen",
	"LZN": "Matsu",

	// 日本機場
	"NRT": "Tokyo",
	"HND": "Tokyo",
	"KIX": "Osaka",
	"ITM": "Osaka",
	"FUK": "Fukuoka",
	"CTS": "Sapporo",
	"OKA": "Okinawa",

	// 韓國機場
	"ICN": "Seoul",
	"GMP": "Seoul",
	"PUS": "Busan",
	"CJU": "Jeju",

	// 中國機場
	"PEK": "Beijing",
	"PVG": "Shanghai",
	"SHA": "Shanghai",
	"CAN": "Guangzhou",
	"SZX": "Shenzhen",

	// 香港、澳門
	"HKG": "Hong Kong",
	"MFM": "Macau",

	// 東南亞
	"SIN": "Singapore",
	"BKK": "Bangkok",
	"DMK": "Bangkok",
	"KUL": "Kuala Lumpur",
	"CGK": "Jakarta",
	"DPS": "Denpasar",
	"MNL": "Manila",
	"CRK": "Manila",

	// 美洲
	"LAX": "Los Angeles",
	"SFO": "San Francisco",
	"JFK": "New York",
	"ORD": "Chicago",
	"YYZ": "Toronto",
	"YVR": "Vancouver",

	// 歐洲
	"LHR": "London",
	"CDG": "Paris",
	"FRA": "Frankfurt",
	"AMS": "Amsterdam",
	"FCO": "Rome",
	"MAD": "Madrid",

	// 大洋洲
	"SYD": "Sydney",
	"MEL": "Melbourne",
	"BNE": "Brisbane",
	"AKL": "Auckland",
}

// 新增：機場資訊查詢函數
func GetCityByAirportCode(airportCode string) string {
	if city, exists := AirportCityMap[airportCode]; exists {
		return city
	}
	return airportCode // 如果找不到，回傳機場代碼本身
}

// 新增：機場詳細資訊結構
type AirportInfo struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Timezone  string  `json:"timezone"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

// 新增：航班搜尋請求擴展（包含城市資訊）
type EnhancedSearchRequest struct {
	SearchRequest
	OriginCity      string `json:"origin_city,omitempty"`
	DestinationCity string `json:"destination_city,omitempty"`
}

// 新增：創建增強搜尋請求
func CreateEnhancedSearchRequest(req SearchRequest) EnhancedSearchRequest {
	return EnhancedSearchRequest{
		SearchRequest:   req,
		OriginCity:      GetCityByAirportCode(req.Origin),
		DestinationCity: GetCityByAirportCode(req.Destination),
	}
}

// 新增：匯率相關模型

// 匯率資訊
type ExchangeRateInfo struct {
	BaseCurrency string             `json:"base_currency"`
	Rates        map[string]float64 `json:"rates"`
	LastUpdated  time.Time          `json:"last_updated"`
	NextUpdate   time.Time          `json:"next_update"`
}

// 航班搜尋響應（包含天氣和匯率）
type FlightSearchResponseWithWeatherAndExchange struct {
	Flights  []Flight          `json:"flights"`
	Weather  *WeatherInfo      `json:"weather,omitempty"`
	Exchange *ExchangeRateInfo `json:"exchange,omitempty"`
	Meta     struct {
		Count         int    `json:"count"`
		Origin        string `json:"origin"`
		Destination   string `json:"destination"`
		DepartureDate string `json:"departure_date"`
		Currency      string `json:"currency"`
	} `json:"meta"`
}

// 貨幣轉換請求
type CurrencyConversionRequest struct {
	Amount       float64 `json:"amount"`
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
}

// 貨幣轉換響應
type CurrencyConversionResponse struct {
	OriginalAmount  float64   `json:"original_amount"`
	ConvertedAmount float64   `json:"converted_amount"`
	FromCurrency    string    `json:"from_currency"`
	ToCurrency      string    `json:"to_currency"`
	ExchangeRate    float64   `json:"exchange_rate"`
	LastUpdated     time.Time `json:"last_updated"`
}
