package config

import (
	"os"
	"strings"
)

type Config struct {
	AmadeusAPIKey    string
	AmadeusAPISecret string
	AmadeusBaseURL   string
	ServerPort       string
	Environment      string
	LogLevel         string
}

func LoadConfig() *Config {
	return &Config{
		AmadeusAPIKey:    getEnv("AMADEUS_API_KEY", ""),
		AmadeusAPISecret: getEnv("AMADEUS_API_SECRET", ""),
		AmadeusBaseURL:   getEnv("AMADEUS_BASE_URL", "https://test.api.amadeus.com/v2"),
		ServerPort:       getEnv("PORT", "8080"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

// 新增：驗證配置是否完整
func (c *Config) Validate() error {
	if c.AmadeusAPIKey == "" {
		return &ConfigError{Field: "AMADEUS_API_KEY", Message: "Amadeus API Key 不能為空"}
	}
	if c.AmadeusAPISecret == "" {
		return &ConfigError{Field: "AMADEUS_API_SECRET", Message: "Amadeus API Secret 不能為空"}
	}
	return nil
}

// 新增：配置錯誤類型
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Message
}

// 新增：檢查是否為測試環境
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// 新增：檢查是否為生產環境
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// 新增：獲取伺服器地址
func (c *Config) GetServerAddress() string {
	return ":" + c.ServerPort
}
