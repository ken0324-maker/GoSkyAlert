package services

import (
	"encoding/json"
	"final/config" // 改成 final
	"final/models" // 改成 final
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type AmadeusService struct {
	config      *config.Config
	client      *http.Client
	accessToken string
	tokenExpiry time.Time
}

func NewAmadeusService(cfg *config.Config) *AmadeusService {
	return &AmadeusService{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// 獲取訪問令牌
func (s *AmadeusService) getAccessToken() (string, error) {
	// 如果令牌還有至少5分鐘有效期，直接返回
	if s.accessToken != "" && time.Now().Before(s.tokenExpiry.Add(-5*time.Minute)) {
		return s.accessToken, nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", s.config.AmadeusAPIKey)
	data.Set("client_secret", s.config.AmadeusAPISecret)

	req, err := http.NewRequest("POST", "https://test.api.amadeus.com/v1/security/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("創建令牌請求失敗: %v", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("令牌請求失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("讀取令牌響應失敗: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("令牌獲取失敗: %s", string(body))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("解析令牌響應失敗: %v", err)
	}

	s.accessToken = tokenResponse.AccessToken
	s.tokenExpiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)

	log.Println("✅ Amadeus訪問令牌獲取成功")
	return s.accessToken, nil
}

// 搜尋航班報價
func (s *AmadeusService) SearchFlights(req models.SearchRequest) ([]models.Flight, error) {
	token, err := s.getAccessToken()
	if err != nil {
		return nil, err
	}

	// 構建API URL
	apiURL := fmt.Sprintf("%s/shopping/flight-offers", s.config.AmadeusBaseURL)

	// 構建查詢參數
	params := url.Values{}
	params.Add("originLocationCode", req.Origin)
	params.Add("destinationLocationCode", req.Destination)
	params.Add("departureDate", req.DepartureDate)

	if req.ReturnDate != "" {
		params.Add("returnDate", req.ReturnDate)
	}

	params.Add("adults", strconv.Itoa(req.Adults))
	params.Add("currencyCode", req.Currency)
	params.Add("max", "10") // 限制結果數量

	fullURL := apiURL + "?" + params.Encode()

	// 創建請求
	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("創建請求失敗: %v", err)
	}

	httpReq.Header.Add("Authorization", "Bearer "+token)

	log.Printf("🔍 搜尋航班: %s -> %s 日期: %s", req.Origin, req.Destination, req.DepartureDate)

	// 發送請求
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API請求失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取響應失敗: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ API錯誤響應: %s", string(body))
		return nil, fmt.Errorf("API錯誤: 狀態碼 %d", resp.StatusCode)
	}

	// 解析響應
	var apiResponse models.AmadeusFlightOffersResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("解析JSON失敗: %v", err)
	}

	log.Printf("✅ 找到 %d 個航班報價", len(apiResponse.Data))

	// 轉換為統一格式
	return s.transformResponse(apiResponse), nil
}

// 轉換Amadeus響應為統一格式
func (s *AmadeusService) transformResponse(response models.AmadeusFlightOffersResponse) []models.Flight {
	var flights []models.Flight

	for _, offer := range response.Data {
		if len(offer.Itineraries) == 0 {
			continue
		}

		itinerary := offer.Itineraries[0]
		if len(itinerary.Segments) == 0 {
			continue
		}

		firstSegment := itinerary.Segments[0]
		lastSegment := itinerary.Segments[len(itinerary.Segments)-1]

		// 解析價格
		price, _ := strconv.ParseFloat(offer.Price.Total, 64)

		// 解析時間
		departureTime, _ := time.Parse(time.RFC3339, firstSegment.Departure.At)
		arrivalTime, _ := time.Parse(time.RFC3339, lastSegment.Arrival.At)

		// 獲取航空公司名稱
		airline := s.getAirlineName(firstSegment.CarrierCode)

		flight := models.Flight{
			ID:           offer.ID,
			Price:        price,
			Currency:     offer.Price.Currency,
			Airline:      airline,
			FlightNumber: fmt.Sprintf("%s%s", firstSegment.CarrierCode, firstSegment.Number),
			From: models.Airport{
				Code:     firstSegment.Departure.IATACode,
				Terminal: firstSegment.Departure.Terminal,
			},
			To: models.Airport{
				Code:     lastSegment.Arrival.IATACode,
				Terminal: lastSegment.Arrival.Terminal,
			},
			Departure: departureTime,
			Arrival:   arrivalTime,
			Duration:  itinerary.Duration,
			Stops:     len(itinerary.Segments) - 1,
			Aircraft:  firstSegment.Aircraft.Code,
		}

		flights = append(flights, flight)
	}

	return flights
}

// 獲取航空公司名稱
func (s *AmadeusService) getAirlineName(code string) string {
	airlines := map[string]string{
		"CI": "中華航空",
		"BR": "長榮航空",
		"CX": "國泰航空",
		"JL": "日本航空",
		"NH": "全日空",
		"KE": "大韓航空",
		"SQ": "新加坡航空",
		"TG": "泰國航空",
		"UA": "聯合航空",
		"AA": "美國航空",
	}

	if name, exists := airlines[code]; exists {
		return name
	}
	return code
}

// 搜尋機場
func (s *AmadeusService) SearchAirports(keyword string) ([]models.Airport, error) {
	token, err := s.getAccessToken()
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s/reference-data/locations", s.config.AmadeusBaseURL)

	params := url.Values{}
	params.Add("subType", "AIRPORT")
	params.Add("keyword", keyword)
	params.Add("page[limit]", "10")

	fullURL := apiURL + "?" + params.Encode()

	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Add("Authorization", "Bearer "+token)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("機場搜尋失敗: 狀態碼 %d", resp.StatusCode)
	}

	var apiResponse models.AirportResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	var airports []models.Airport
	for _, airport := range apiResponse.Data {
		airports = append(airports, models.Airport{
			Code: airport.IATACode,
			Name: airport.Name,
			City: airport.Address.CityName,
		})
	}

	return airports, nil
}
