package config

import (
	"fmt"
	"os"
)

type Config struct {
	Environment string
	ServerPort  string
	ServerHost  string

	Database DatabaseConfig
	Firebase FirebaseConfig
	Bunny    BunnyConfig
	RevenueCat RevenueCatConfig
	AdReward AdRewardConfig
	JWT        JWTConfig
}

type JWTConfig struct {
	SecretKey string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type FirebaseConfig struct {
	ProjectID      string
	CredentialsPath string
}

type BunnyConfig struct {
	StreamAPIKey    string
	StreamLibraryID string
	StorageAPIKey   string
	StorageZone     string
	CDNHostname     string
	SecurityKey     string
}

type RevenueCatConfig struct {
	APIKey       string
	WebhookSecret string
}

type AdRewardConfig struct {
	CoinsPerAd      int
	DailyLimit      int
	HourlyLimit     int
	CooldownMinutes int
}

func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		ServerHost:  getEnv("SERVER_HOST", "localhost"),

		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "qenti"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		Firebase: FirebaseConfig{
			ProjectID:       getEnv("FIREBASE_PROJECT_ID", ""),
			CredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", "./firebase-credentials.json"),
		},

		Bunny: BunnyConfig{
			StreamAPIKey:    getEnv("BUNNY_STREAM_API_KEY", ""),
			StreamLibraryID: getEnv("BUNNY_STREAM_LIBRARY_ID", ""),
			StorageAPIKey:   getEnv("BUNNY_STORAGE_API_KEY", ""),
			StorageZone:     getEnv("BUNNY_STORAGE_ZONE", ""),
			CDNHostname:     getEnv("BUNNY_CDN_HOSTNAME", ""),
			SecurityKey:     getEnv("BUNNY_SECURITY_KEY", ""),
		},

		RevenueCat: RevenueCatConfig{
			APIKey:       getEnv("REVENUECAT_API_KEY", ""),
			WebhookSecret: getEnv("REVENUECAT_WEBHOOK_SECRET", ""),
		},

		AdReward: AdRewardConfig{
			CoinsPerAd:      getEnvInt("AD_REWARD_COINS_PER_AD", 10),
			DailyLimit:      getEnvInt("AD_REWARD_DAILY_LIMIT", 10),
			HourlyLimit:     getEnvInt("AD_REWARD_HOURLY_LIMIT", 3),
			CooldownMinutes: getEnvInt("AD_REWARD_COOLDOWN_MINUTES", 5),
		},

		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET", "change-this-secret-key-in-production"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
