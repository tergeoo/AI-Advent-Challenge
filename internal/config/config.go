package config

import (
	"fmt"
	"os"
)

// Config содержит конфигурацию приложения
type Config struct {
	OpenAIKey string
}

// Load загружает конфигурацию из .env файла
func Load() (*Config, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY не установлен в переменных окружения")
	}

	return &Config{
		OpenAIKey: apiKey,
	}, nil
}
