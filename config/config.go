package config

import (
	"os"
	"strings"
)

type Config struct {
	AmadeusAPIKey      string
	AmadeusAPISecret   string
	AmadeusBaseURL     string
	WeatherAPIKey      string
	ExchangeRateAPIKey string
	FoursquareAPIKey   string
	DiscordBotToken    string // [修改] 改用 Discord Token
	ServerPort         string
	Environment        string
	LogLevel           string
}

func LoadConfig() *Config {
	return &Config{
		AmadeusAPIKey:      getEnv("AMADEUS_API_KEY", ""),
		AmadeusAPISecret:   getEnv("AMADEUS_API_SECRET", ""),
		AmadeusBaseURL:     getEnv("AMADEUS_BASE_URL", "https://test.api.amadeus.com/v2"),
		WeatherAPIKey:      getEnv("WEATHER_API_KEY", ""),
		ExchangeRateAPIKey: getEnv("EXCHANGE_RATE_API_KEY", ""),
		FoursquareAPIKey:   getEnv("FOURSQUARE_API_KEY", ""),
		DiscordBotToken:    getEnv("DISCORD_BOT_TOKEN", ""), // [修改] 讀取 Discord 環境變數
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

func (c *Config) Validate() error {
	if c.AmadeusAPIKey == "" {
		return &ConfigError{Field: "AMADEUS_API_KEY", Message: "Amadeus API Key 不能為空"}
	}
	if c.AmadeusAPISecret == "" {
		return &ConfigError{Field: "AMADEUS_API_SECRET", Message: "Amadeus API Secret 不能為空"}
	}
	// WeatherAPI 可選
	if c.WeatherAPIKey == "" {
		return &ConfigError{Field: "WEATHER_API_KEY", Message: "WeatherAPI Key 為空，天氣功能將無法使用"}
	}
	return nil
}

type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Message
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) GetServerAddress() string {
	return ":" + c.ServerPort
}

func (c *Config) HasWeatherAPI() bool {
	return c.WeatherAPIKey != ""
}

func (c *Config) HasExchangeRateAPI() bool {
	return c.ExchangeRateAPIKey != ""
}

func (c *Config) HasFoursquareAPI() bool {
	return c.FoursquareAPIKey != ""
}

// [新增] 檢查是否啟用 Discord
func (c *Config) HasDiscordAPI() bool {
	return c.DiscordBotToken != ""
}
