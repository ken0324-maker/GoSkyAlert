package config

import (
	"os"
	"strings"
)

type Config struct {
	AmadeusAPIKey      string
	AmadeusAPISecret   string
	AmadeusBaseURL     string
	WeatherAPIKey      string // 新增 WeatherAPI 金鑰
	ExchangeRateAPIKey string // 新增匯率 API 金鑰
	FoursquareAPIKey   string // 新增 Foursquare API 金鑰
	ServerPort         string
	Environment        string
	LogLevel           string
}

func LoadConfig() *Config {
	return &Config{
		AmadeusAPIKey:      getEnv("AMADEUS_API_KEY", ""),
		AmadeusAPISecret:   getEnv("AMADEUS_API_SECRET", ""),
		AmadeusBaseURL:     getEnv("AMADEUS_BASE_URL", "https://test.api.amadeus.com/v2"),
		WeatherAPIKey:      getEnv("WEATHER_API_KEY", ""),       // 新增 WeatherAPI 配置
		ExchangeRateAPIKey: getEnv("EXCHANGE_RATE_API_KEY", ""), // 新增匯率 API 配置
		FoursquareAPIKey:   getEnv("FOURSQUARE_API_KEY", ""),    // 新增 Foursquare API 配置
		ServerPort:         getEnv("PORT", "8080"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

// 修改驗證配置，添加 WeatherAPI 驗證（可選）
func (c *Config) Validate() error {
	if c.AmadeusAPIKey == "" {
		return &ConfigError{Field: "AMADEUS_API_KEY", Message: "Amadeus API Key 不能為空"}
	}
	if c.AmadeusAPISecret == "" {
		return &ConfigError{Field: "AMADEUS_API_SECRET", Message: "Amadeus API Secret 不能為空"}
	}
	// WeatherAPI 可選，如果沒有設定就只顯示航班資訊
	if c.WeatherAPIKey == "" {
		return &ConfigError{Field: "WEATHER_API_KEY", Message: "WeatherAPI Key 為空，天氣功能將無法使用"}
	}
	return nil
}

// 配置錯誤類型
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Message
}

// 檢查是否為測試環境
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// 檢查是否為生產環境
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// 獲取伺服器地址
func (c *Config) GetServerAddress() string {
	return ":" + c.ServerPort
}

// 新增：檢查是否啟用天氣功能
func (c *Config) HasWeatherAPI() bool {
	return c.WeatherAPIKey != ""
}

// 新增：檢查是否啟用匯率功能
func (c *Config) HasExchangeRateAPI() bool {
	return c.ExchangeRateAPIKey != ""
}

// 新增：檢查是否啟用景點功能
func (c *Config) HasFoursquareAPI() bool {
	return c.FoursquareAPIKey != ""
}
