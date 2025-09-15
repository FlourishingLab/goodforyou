package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Environment string   `json:"environment"`
	CorsOrigins []string `json:"cors_origins"`
}

func LoadConfig() (*Config, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev" // Default to development if APP_ENV is not set
	}

	configFile := fmt.Sprintf("config/%s.json", env)
	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	log.Printf("Loaded config with origins: %s", cfg.CorsOrigins)

	return &cfg, nil
}
