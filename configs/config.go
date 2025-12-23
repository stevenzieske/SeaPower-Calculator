package configs

import "os"

type Config struct {
	GAME_DIR string
}

func Load() *Config {
	return &Config{
		GAME_DIR: getEnv("GAME_DIR", "./data"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
