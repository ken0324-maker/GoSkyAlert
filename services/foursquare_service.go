package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type FoursquareService struct {
	apiKey string
	client *http.Client
}

func NewFoursquareService(apiKey string) *FoursquareService {
	return &FoursquareService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

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
	Price       int     `json:"price"`
	IsOpen      bool    `json:"is_open_now"`
	Phone       string  `json:"phone"`
	Website     string  `json:"website"`
	Description string  `json:"description"`
	Distance    float64 `json:"distance"`
}

type SearchRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    int     `json:"radius"`
	Query     string  `json:"query"`
	Category  string  `json:"category"`
}

// 搜索附近景點 - 使用新的端點和認證
func (fs *FoursquareService) SearchNearby(req SearchRequest) ([]Attraction, error) {
	// 使用新的端點
	baseURL := "https://places-api.foursquare.com/places/search"

	// 構建查詢參數
	params := url.Values{}
	params.Add("ll", fmt.Sprintf("%f,%f", req.Latitude, req.Longitude))

	if req.Radius > 0 {
		params.Add("radius", fmt.Sprintf("%d", req.Radius))
	} else {
		params.Add("radius", "5000")
	}

	params.Add("limit", "20")
	params.Add("sort", "DISTANCE")

	if req.Query != "" {
		params.Add("query", req.Query)
	}
	if req.Category != "" {
		params.Add("categories", req.Category)
	}

	// 構建完整 URL
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 創建請求
	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// 設置新的 headers - 按照遷移指南
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", fs.apiKey)) // 新的認證格式
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("X-Places-Api-Version", "2025-06-17") // 新的版本號

	log.Printf("🔍 發送 Foursquare 新 API 請求: %s", fullURL)
	log.Printf("🔑 Headers: Authorization=Bearer %s...", fs.apiKey[:20])
	log.Printf("🔑 Headers: X-Places-Api-Version=2025-06-17")

	// 發送請求
	resp, err := fs.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("📡 Foursquare 新 API 回應狀態: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		// 讀取錯誤回應主體
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			log.Printf("❌ Foursquare API 錯誤詳情: %v", errorResp)
		} else {
			bodyBytes := make([]byte, 1024)
			n, _ := resp.Body.Read(bodyBytes)
			log.Printf("❌ Foursquare API 原始錯誤: %s", string(bodyBytes[:n]))
		}
		return nil, fmt.Errorf("Foursquare API error: %s", resp.Status)
	}

	// 解析回應 - 適應新的回應格式
	var apiResponse struct {
		Results []struct {
			FSQPlaceID string `json:"fsq_place_id"` // 新的欄位名稱
			Name       string `json:"name"`
			Categories []struct {
				ID   string `json:"id"` // 現在是 BSON ID
				Name string `json:"name"`
			} `json:"categories"`
			Location struct {
				FormattedAddress string `json:"formatted_address"`
				Locality         string `json:"locality"`
				Region           string `json:"region"`
				Country          string `json:"country"`
			} `json:"location"`
			Latitude  float64 `json:"latitude"`  // 新的位置格式
			Longitude float64 `json:"longitude"` // 新的位置格式
			Distance  int     `json:"distance"`
			Rating    float64 `json:"rating"`
			Hours     struct {
				OpenNow bool `json:"open_now"`
			} `json:"hours"`
			Price       int    `json:"price"`
			Tel         string `json:"tel"`
			Website     string `json:"website"`
			Description string `json:"description"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		log.Printf("❌ 解析 Foursquare 回應錯誤: %v", err)
		return nil, err
	}

	// 轉換為我們的模型
	var attractions []Attraction
	for _, result := range apiResponse.Results {
		category := ""
		if len(result.Categories) > 0 {
			category = result.Categories[0].Name
		}

		attraction := Attraction{
			ID:          result.FSQPlaceID,
			Name:        result.Name,
			Address:     result.Location.FormattedAddress,
			City:        result.Location.Locality,
			Country:     result.Location.Country,
			Latitude:    result.Latitude,
			Longitude:   result.Longitude,
			Category:    category,
			Rating:      result.Rating,
			Price:       result.Price,
			IsOpen:      result.Hours.OpenNow,
			Phone:       result.Tel,
			Website:     result.Website,
			Description: result.Description,
			Distance:    float64(result.Distance),
		}
		attractions = append(attractions, attraction)
	}

	log.Printf("✅ 找到 %d 個景點", len(attractions))
	return attractions, nil
}

// 驗證 API Key - 使用新端點
func (fs *FoursquareService) ValidateAPIKey() error {
	testURL := "https://places-api.foursquare.com/places/search?ll=25.0330,121.5654&limit=1"

	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", fs.apiKey))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Places-Api-Version", "2025-06-17")

	resp, err := fs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("🔑 Foursquare 新 API 測試回應狀態: %s", resp.Status)

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("API test failed with status: %s", resp.Status)
}

// 獲取熱門景點類別
func (fs *FoursquareService) GetPopularCategories() []string {
	return []string{
		"13000", // Arts & Entertainment
		"16000", // Landmarks & Outdoors
		"10000", // Professional & Other Places
	}
}
