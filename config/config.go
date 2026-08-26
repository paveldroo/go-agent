package config

import "os"

type Config struct {
	ApiKey    string
	LLMURL    string
	ModelName string
}

func New() *Config {
	return &Config{
		ApiKey:    os.Getenv("LLM_API_KEY"),
		LLMURL:    os.Getenv("LLM_URL"),
		ModelName: os.Getenv("MODEL_NAME"),
	}
}
