package config

import "os"

type Config struct {
	APIKey    string
	LLMURL    string
	ModelName string
}

func New() *Config {
	return &Config{
		APIKey:    os.Getenv("LLM_API_KEY"),
		LLMURL:    os.Getenv("LLM_URL"),
		ModelName: os.Getenv("MODEL_NAME"),
	}
}
