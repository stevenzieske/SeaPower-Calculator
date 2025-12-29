package configs

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	GAME_DIR string
}

func Load() *Config {
	// Try to load .env file from executable directory
	loadEnvFile()

	return &Config{
		GAME_DIR: getEnv("GAME_DIR", getDefaultDataDir()),
	}
}

func loadEnvFile() {
	// Get the executable's directory
	exePath, err := os.Executable()
	if err != nil {
		// Try current directory as fallback
		exePath, _ = os.Getwd()
	}

	exeDir := filepath.Dir(exePath)
	envFilePath := filepath.Join(exeDir, ".env")

	// Try to open .env file
	file, err := os.Open(envFilePath)
	if err != nil {
		// .env file doesn't exist or can't be read, that's okay
		return
	}
	defer file.Close()

	// Parse .env file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Only set if not already set in system environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
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
