package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	MongoDBURI    string
	DBName        string
	JWTSecret     string
	EncryptionKey string
	FrontendURL   string
	Environment   string
}

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Port:          getEnv("PORT", "8080"),
		MongoDBURI:    getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DBName:        getEnv("DB_NAME", "evote"),
		JWTSecret:     getEnv("JWT_SECRET", "default-secret-change-in-production"),
		EncryptionKey: getEnv("ENCRYPTION_KEY", "default-32-byte-key-change-this!"),
		FrontendURL:   getEnv("FRONTEND_URL", "http://localhost:3000"),
		Environment:   getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
