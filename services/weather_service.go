package services

import (
	"encoding/json"
	"final/models"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type WeatherService struct {
	APIKey  string
	BaseURL string
}

func NewWeatherService(apiKey string) *WeatherService {
	return &WeatherService{
		APIKey:  apiKey,
		BaseURL: "http://api.weatherapi.com/v1",
	}
}

// GetWeather 獲取指定城市和日期的天氣資訊
func (s *WeatherService) GetWeather(city, date string) (*models.WeatherResponse, error) {
	// 構建請求 URL
	endpoint := "/forecast.json"
	params := url.Values{}
	params.Add("key", s.APIKey)
	params.Add("q", city)
	params.Add("days", "3") // 獲取3天預報
	params.Add("aqi", "no")
	params.Add("alerts", "no")

	requestURL := fmt.Sprintf("%s%s?%s", s.BaseURL, endpoint, params.Encode())

	// 創建 HTTP 請求
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("創建請求失敗: %v", err)
	}

	// 發送請求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("天氣API請求失敗: %v", err)
	}
	defer resp.Body.Close()

	// 檢查響應狀態
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("天氣API錯誤: %s - %s", resp.Status, string(body))
	}

	// 解析響應
	var weatherResp models.WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("解析天氣響應失敗: %v", err)
	}

	return &weatherResp, nil
}

// GetCurrentWeather 獲取當前天氣
func (s *WeatherService) GetCurrentWeather(city string) (*models.WeatherResponse, error) {
	endpoint := "/current.json"
	params := url.Values{}
	params.Add("key", s.APIKey)
	params.Add("q", city)
	params.Add("aqi", "no")

	requestURL := fmt.Sprintf("%s%s?%s", s.BaseURL, endpoint, params.Encode())

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("創建請求失敗: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("天氣API請求失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("天氣API錯誤: %s - %s", resp.Status, string(body))
	}

	var weatherResp models.WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("解析天氣響應失敗: %v", err)
	}

	return &weatherResp, nil
}

// GetWeatherByAirport 根據機場代碼獲取天氣
func (s *WeatherService) GetWeatherByAirport(airportCode, date string) (*models.WeatherResponse, error) {
	city := models.GetCityByAirportCode(airportCode)
	if city == "" {
		return nil, fmt.Errorf("找不到機場代碼對應的城市: %s", airportCode)
	}
	return s.GetWeather(city, date)
}

// ValidateAPIKey 驗證 WeatherAPI 金鑰是否有效
func (s *WeatherService) ValidateAPIKey() error {
	// 使用一個已知的城市進行測試請求
	testCity := "London"
	_, err := s.GetCurrentWeather(testCity)
	if err != nil {
		return fmt.Errorf("WeatherAPI 金鑰驗證失敗: %v", err)
	}
	return nil
}

// GetWeatherSummary 獲取天氣摘要（用於航班顯示）
// GetWeatherSummary 獲取天氣摘要（用於航班顯示）
func (s *WeatherService) GetWeatherSummary(city, date string) (*models.WeatherSummary, error) {
	weather, err := s.GetWeather(city, date)
	if err != nil {
		return nil, err
	}

	// 找到指定日期的預報
	var targetForecast *models.ForecastDay
	for i := range weather.Forecast.Forecastday {
		if weather.Forecast.Forecastday[i].Date == date {
			targetForecast = &weather.Forecast.Forecastday[i]
			break
		}
	}

	summary := &models.WeatherSummary{
		City:      city, // 這裡會顯示城市名稱
		Date:      date,
		AvgTemp:   weather.Current.TempC,
		Condition: weather.Current.Condition.Text,
		Icon:      weather.Current.Condition.Icon,
		Humidity:  weather.Current.Humidity,
		WindSpeed: weather.Current.WindKph,
	}

	// 如果有預報數據，使用預報數據
	if targetForecast != nil {
		summary.AvgTemp = targetForecast.Day.AvgTempC
		summary.Condition = targetForecast.Day.Condition.Text
		summary.Icon = targetForecast.Day.Condition.Icon
		summary.ChanceOfRain = targetForecast.Day.DailyChanceOfRain
	}

	return summary, nil
}

// generateWeatherDescription 生成天氣描述
func (s *WeatherService) generateWeatherDescription(condition string, tempC float64, chanceOfRain int) string {
	description := condition

	if tempC < 10 {
		description += "，天氣寒冷"
	} else if tempC > 30 {
		description += "，天氣炎熱"
	} else if tempC >= 10 && tempC <= 25 {
		description += "，天氣舒適"
	}

	if chanceOfRain > 50 {
		description += "，降雨機率高"
	} else if chanceOfRain > 20 {
		description += "，可能有降雨"
	}

	return description
}

// IsWeatherSuitableForTravel 判斷天氣是否適合旅行
func (s *WeatherService) IsWeatherSuitableForTravel(weather *models.WeatherResponse) bool {
	if weather == nil {
		return true // 如果沒有天氣數據，默認適合旅行
	}

	// 檢查當前天氣條件
	current := weather.Current
	if current.Condition.Text == "Heavy rain" ||
		current.Condition.Text == "Thunderstorm" ||
		current.Condition.Text == "Snow" {
		return false
	}

	// 檢查風速是否過大
	if current.WindKph > 50 {
		return false
	}

	return true
}
