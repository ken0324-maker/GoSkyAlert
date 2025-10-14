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
