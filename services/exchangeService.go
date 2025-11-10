package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ExchangeService struct {
	APIKey  string
	BaseURL string
}

func NewExchangeService(apiKey string) *ExchangeService {
	return &ExchangeService{
		APIKey:  apiKey,
		BaseURL: "https://v6.exchangerate-api.com/v6",
	}
}

// 匯率響應結構
type ExchangeRateResponse struct {
	Result             string             `json:"result"`
	Documentation      string             `json:"documentation"`
	TermsOfUse         string             `json:"terms_of_use"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
	TimeLastUpdateUTC  string             `json:"time_last_update_utc"`
	TimeNextUpdateUnix int64              `json:"time_next_update_unix"`
	TimeNextUpdateUTC  string             `json:"time_next_update_utc"`
	BaseCode           string             `json:"base_code"`
	ConversionRates    map[string]float64 `json:"conversion_rates"`
}

// 匯率請求
type ExchangeRateRequest struct {
	BaseCurrency string   `json:"base_currency"`
	Currencies   []string `json:"currencies,omitempty"`
}

// 匯率結果
type ExchangeRateResult struct {
	BaseCurrency string             `json:"base_currency"`
	Rates        map[string]float64 `json:"rates"`
	LastUpdated  time.Time          `json:"last_updated"`
	NextUpdate   time.Time          `json:"next_update"`
}

// 獲取匯率
func (s *ExchangeService) GetExchangeRates(baseCurrency string, targetCurrencies []string) (*ExchangeRateResult, error) {
	url := fmt.Sprintf("%s/%s/latest/%s", s.BaseURL, s.APIKey, baseCurrency)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("創建匯率請求失敗: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("匯率API請求失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("匯率API錯誤: %s - %s", resp.Status, string(body))
	}

	var apiResponse ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("解析匯率響應失敗: %v", err)
	}

	if apiResponse.Result != "success" {
		return nil, fmt.Errorf("匯率API返回錯誤: %s", apiResponse.Result)
	}

	// 過濾需要的貨幣
	rates := make(map[string]float64)
	if len(targetCurrencies) > 0 {
		for _, currency := range targetCurrencies {
			if rate, exists := apiResponse.ConversionRates[currency]; exists {
				rates[currency] = rate
			}
		}
	} else {
		// 如果沒有指定貨幣，返回所有
		rates = apiResponse.ConversionRates
	}

	return &ExchangeRateResult{
		BaseCurrency: baseCurrency,
		Rates:        rates,
		LastUpdated:  time.Unix(apiResponse.TimeLastUpdateUnix, 0),
		NextUpdate:   time.Unix(apiResponse.TimeNextUpdateUnix, 0),
	}, nil
}

// 貨幣轉換
func (s *ExchangeService) ConvertCurrency(amount float64, fromCurrency, toCurrency string) (float64, error) {
	rates, err := s.GetExchangeRates(fromCurrency, []string{toCurrency})
	if err != nil {
		return 0, err
	}

	rate, exists := rates.Rates[toCurrency]
	if !exists {
		return 0, fmt.Errorf("不支援的貨幣轉換: %s -> %s", fromCurrency, toCurrency)
	}

	return amount * rate, nil
}

// 獲取支援的貨幣列表
func (s *ExchangeService) GetSupportedCurrencies() []string {
	return []string{
		"TWD", "USD", "EUR", "JPY", "GBP", "AUD", "CAD", "CHF", "CNY", "HKD",
		"KRW", "SGD", "THB", "VND", "MYR", "IDR", "PHP", "INR", "BRL", "RUB",
	}
}

// 驗證 API 金鑰
func (s *ExchangeService) ValidateAPIKey() error {
	_, err := s.GetExchangeRates("USD", []string{"TWD"})
	if err != nil {
		return fmt.Errorf("ExchangeRate API 金鑰驗證失敗: %v", err)
	}
	return nil
}
