package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	MongoDBURI          string
	DBName              string
	JWTSecret           string
	EncryptionKey       string
	FrontendURL         string
	Environment         string
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	BrevoAPIKey         string
	SenderEmail         string
	SenderName          string
}

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	frontendURL := getEnv("FRONTEND_URL", "https://hallmark-evoting.vercel.app")
	frontendURL = strings.TrimSuffix(frontendURL, "/")

	return &Config{
		Port:                getEnv("PORT", "8080"),
		MongoDBURI:          getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DBName:              getEnv("DB_NAME", "evote"),
		JWTSecret:           getEnv("JWT_SECRET", "default-secret-change-in-production"),
		EncryptionKey:       getEnv("ENCRYPTION_KEY", "default-32-byte-key-change-this!"),
		FrontendURL:         frontendURL,
		Environment:         getEnv("ENVIRONMENT", "development"),
		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),
		BrevoAPIKey:         getEnv("BREVO_API_KEY", ""),
		SenderEmail:         getEnv("SENDER_EMAIL", ""),
		SenderName:          getEnv("SENDER_NAME", "Hallmark E-Voting"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
