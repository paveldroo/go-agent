package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	APIKey    string `validate:"required"`
	LLMURL    string `validate:"required"`
	ModelName string `validate:"required"`
}

func New() (*Config, error) {
	cfg := &Config{
		APIKey:    os.Getenv("LLM_API_KEY"),
		LLMURL:    os.Getenv("LLM_URL"),
		ModelName: os.Getenv("MODEL_NAME"),
	}

	err := validator.New().Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}
