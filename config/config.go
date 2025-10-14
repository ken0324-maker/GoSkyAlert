package config

import (
	"os"
)

type Config struct {
	AmadeusAPIKey    string
	AmadeusAPISecret string
	AmadeusBaseURL   string
}

func LoadConfig() *Config {
	return &Config{
		AmadeusAPIKey:    getEnv("AMADEUS_API_KEY", ""),
		AmadeusAPISecret: getEnv("AMADEUS_API_SECRET", ""),
		AmadeusBaseURL:   getEnv("AMADEUS_BASE_URL", "https://test.api.amadeus.com/v2"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
