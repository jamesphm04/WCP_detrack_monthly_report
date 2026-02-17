package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL        string
	APIKey         string
	FetchLimit     int
	SMTPHost       string
	SMTPPort       string
	EmailSender    string
	EmailPassword  string
	EmailReceivers string
}

func LoadConfig() (*Config, error) {
	// Only load .env in local development (not in AWS ECS)
	if os.Getenv("AWS_EXECUTION_ENV") == "" {
		_ = godotenv.Load()
	}

	fetchLimit, err := strconv.Atoi(getEnv("FETCH_LIMIT", "1000"))
	if err != nil {
		return nil, errors.New("Cannot convert FETCH_LIMIT to Int")
	}

	config := &Config{
		BaseURL:        getEnv("BASE_URL", "https://app.detrack.com/api/v2"),
		APIKey:         getEnv("API_KEY", ""),
		FetchLimit:     fetchLimit,
		SMTPHost:       getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:       getEnv("SMTP_PORT", "587"),
		EmailSender:    getEnv("EMAIL_SENDER", ""),
		EmailPassword:  getEnv("EMAIL_PASSWORD", ""),
		EmailReceivers: getEnv("EMAIL_RECEIVERS", ""), //comma separated for multiple receivers
	}

	// Validate required fields
	if config.APIKey == "" {
		return nil, errors.New("ENV: API_KEY not found")
	}

	if config.EmailPassword == "" {
		return nil, errors.New("ENV: EMAIL_PASSWORD not found")
	}

	if config.EmailSender == "" {
		return nil, errors.New("ENV: EMAIL_SENDER not found")
	}

	if config.EmailReceivers == "" {
		return nil, errors.New("ENV: EMAIL_RECEIVERS not found")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}