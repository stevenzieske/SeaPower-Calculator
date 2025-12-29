package configs

import (
	"os"
	"path/filepath"
)

type Config struct {
	GAME_DIR string
}

func Load() *Config {
	return &Config{
		GAME_DIR: getEnv("GAME_DIR", getDefaultDataDir()),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDefaultDataDir() string {
	// Get the executable's path
	exePath, err := os.Executable()
	if err != nil {
		// Fallback to current directory if we can't get executable path
		return "./data"
	}

	// Get the directory containing the executable
	exeDir := filepath.Dir(exePath)
	exeDataDir := filepath.Join(exeDir, "data")

	// Check if data directory exists next to executable
	if _, err := os.Stat(exeDataDir); err == nil {
		return exeDataDir
	}

	// Fallback to current directory (for development with 'go run')
	return "./data"
}
