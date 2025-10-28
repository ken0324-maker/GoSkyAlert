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
	ID           string    `json:"id"`
	Price        float64   `json:"price"`
	Currency     string    `json:"currency"`
	Airline      string    `json:"airline"`
	FlightNumber string    `json:"flight_number"`
	From         Airport   `json:"from"`
	To           Airport   `json:"to"`
	Departure    time.Time `json:"departure"`
	Arrival      time.Time `json:"arrival"`
	Duration     string    `json:"duration"`
	Stops        int       `json:"stops"`
	Aircraft     string    `json:"aircraft"`
	DeepLink     string    `json:"deep_link,omitempty"`
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
