package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Environment string
	ServerPort  string
	ServerHost  string

	// SuperAdminEmail email cuya cuenta recibe rol super_admin automáticamente en el primer login.
	// Solo debe haber 1-2 cuentas. Definir en variable de entorno SUPER_ADMIN_EMAIL.
	SuperAdminEmail string

	// CDNProvider selecciona el proveedor de video: "bunny" (default) | "cloudflare"
	CDNProvider string

	Database      DatabaseConfig
	Firebase      FirebaseConfig
	Bunny         BunnyConfig
	Cloudflare    CloudflareConfig
	VideoUpload   VideoUploadConfig
	RevenueCat    RevenueCatConfig
	AdReward      AdRewardConfig
	AdTier        AdTierConfig
	EpisodeCliff  EpisodeCliffConfig
	JWT           JWTConfig
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

// CloudflareConfig es para Cloudflare Stream (proveedor alternativo).
type CloudflareConfig struct {
	AccountID string
	APIToken  string
}

// VideoUploadConfig define los límites de validación en el upload de videos.
type VideoUploadConfig struct {
	// MaxFileSizeMB límite duro en MB (default 150 MB)
	MaxFileSizeMB int64
	// MaxDurationSeconds duración máxima del video en segundos (default 180)
	MaxDurationSeconds int
	// WarnFileSizeMB aviso suave si supera este tamaño (default 50 MB - objetivo 5-8 MB producción)
	WarnFileSizeMB int64
}

type RevenueCatConfig struct {
	APIKey        string
	WebhookSecret string
}

type AdRewardConfig struct {
	CoinsPerAd      int
	DailyLimit      int
	HourlyLimit     int
	CooldownMinutes int
}

// AdTierConfig mapea países de Tier-A (high eCPM) a un límite diario mayor.
// Los países listados en TierACountries reciben TierADailyLimit ads/día;
// el resto recibe el límite estándar de AdReward.DailyLimit.
type AdTierConfig struct {
	// TierACountries lista de country codes ISO-3166-1 alpha-2 (ej. "US,GB,AU,CA,DE")
	TierACountries []string
	TierADailyLimit int
}

// EpisodeCliffConfig configura el precio automático de episodios según su número.
// Episodios 1..CliffStart-1 usan BasePrice; episodios >= CliffStart usan CliffPrice.
// Si IsFree == true o el admin envía price_coins > 0, esta lógica se salta.
type EpisodeCliffConfig struct {
	// CliffStart primer episodio que entra en el precio alto (default 8)
	CliffStart int
	// BasePrice precio en monedas para ep 1..(CliffStart-1) (default 10)
	BasePrice int
	// CliffPrice precio para ep >= CliffStart (default 20)
	CliffPrice int
}

func Load() *Config {
	return &Config{
		Environment:     getEnv("ENVIRONMENT", "development"),
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		ServerHost:      getEnv("SERVER_HOST", "localhost"),
		SuperAdminEmail: getEnv("SUPER_ADMIN_EMAIL", ""),

		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "qenti"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		CDNProvider: getEnv("CDN_PROVIDER", "bunny"),

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

		Cloudflare: CloudflareConfig{
			AccountID: getEnv("CLOUDFLARE_ACCOUNT_ID", ""),
			APIToken:  getEnv("CLOUDFLARE_API_TOKEN", ""),
		},

		VideoUpload: VideoUploadConfig{
			MaxFileSizeMB:      getEnvInt64("VIDEO_MAX_FILE_SIZE_MB", 150),
			MaxDurationSeconds: getEnvInt("VIDEO_MAX_DURATION_SECONDS", 180),
			WarnFileSizeMB:     getEnvInt64("VIDEO_WARN_FILE_SIZE_MB", 50),
		},

		RevenueCat: RevenueCatConfig{
			APIKey:        getEnv("REVENUECAT_API_KEY", ""),
			WebhookSecret: getEnv("REVENUECAT_WEBHOOK_SECRET", ""),
		},

		AdReward: AdRewardConfig{
			CoinsPerAd:      getEnvInt("AD_REWARD_COINS_PER_AD", 10),
			DailyLimit:      getEnvInt("AD_REWARD_DAILY_LIMIT", 10),
			HourlyLimit:     getEnvInt("AD_REWARD_HOURLY_LIMIT", 3),
			CooldownMinutes: getEnvInt("AD_REWARD_COOLDOWN_MINUTES", 5),
		},

		AdTier: AdTierConfig{
			// Tier-A: US, UK, Canada, Australia, Germany, France, Japan, Netherlands, Switzerland, Sweden, Norway, Denmark
			TierACountries:  getEnvStringSlice("AD_TIER_A_COUNTRIES", "US,GB,CA,AU,DE,FR,JP,NL,CH,SE,NO,DK"),
			TierADailyLimit: getEnvInt("AD_TIER_A_DAILY_LIMIT", 20),
		},

		EpisodeCliff: EpisodeCliffConfig{
			CliffStart: getEnvInt("EPISODE_CLIFF_START", 8),
			BasePrice:  getEnvInt("EPISODE_CLIFF_BASE_PRICE", 10),
			CliffPrice: getEnvInt("EPISODE_CLIFF_PRICE", 20),
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

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		var result int64
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

// getEnvStringSlice lee una variable de entorno y la convierte en []string
// dividiendo por comas. Si la variable no está definida, parsea defaultValue.
func getEnvStringSlice(key, defaultValue string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		raw = defaultValue
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}