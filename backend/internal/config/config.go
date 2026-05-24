package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DBPath     string
	JWTSecret  string
	DockerHost string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBPath:     getEnv("DB_PATH", "./data/nas-partner.db"),
		JWTSecret:  getEnv("JWT_SECRET", "change-me-in-production"),
		DockerHost: getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
