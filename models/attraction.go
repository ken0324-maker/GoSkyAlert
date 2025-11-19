package models

type Attraction struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Address     string  `json:"address"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Category    string  `json:"category"`
	Rating      float64 `json:"rating"`
	Price       int     `json:"price"` // 1-4, $ - $$$$
	IsOpen      bool    `json:"is_open_now"`
	Phone       string  `json:"phone"`
	Website     string  `json:"website"`
	Description string  `json:"description"`
	Distance    float64 `json:"distance"` // 公尺
}

type AttractionSearchRequest struct {
	Latitude  float64 `json:"latitude" form:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" form:"longitude" binding:"required"`
	Radius    int     `json:"radius" form:"radius"`     // 公尺，預設 5000
	Query     string  `json:"query" form:"query"`       // 搜尋關鍵字
	Category  string  `json:"category" form:"category"` // 類別
}

type AttractionResponse struct {
	Success bool         `json:"success"`
	Data    []Attraction `json:"data"`
	Message string       `json:"message,omitempty"`
}
