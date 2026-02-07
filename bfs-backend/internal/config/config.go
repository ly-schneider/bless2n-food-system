package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App         AppConfig
	Postgres    PostgresConfig
	Logger      LoggerConfig
	Security    SecurityConfig
	Payrexx     PayrexxConfig
	Plunk       PlunkConfig
	BlobStorage BlobStorageConfig
	Elvanto     ElvantoConfig
}

type ElvantoConfig struct {
	APIKey  string
	GroupID string
}

type AppConfig struct {
	AppEnv        string
	AppPort       string
	PublicBaseURL string
}

type PostgresConfig struct {
	DSN             string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type LoggerConfig struct {
	Level       string
	Development bool
}

type SecurityConfig struct {
	EnableHSTS     bool
	EnableCSP      bool
	TrustedOrigins []string
}

type PayrexxConfig struct {
	InstanceName  string // PAYREXX_INSTANCE - Payrexx instance name
	APISecret     string // PAYREXX_API_SECRET - API secret for HMAC signature
	WebhookSecret string // PAYREXX_WEBHOOK_SECRET - Secret for webhook verification
}

type PlunkConfig struct {
	APIKey    string
	FromName  string
	FromEmail string
	ReplyTo   string
}

type BlobStorageConfig struct {
	AccountName  string
	AccountKey   string
	Container    string
	BlobEndpoint string
}

func Load() Config {
	// Load .env files by default outside Docker. In Docker, allow opt-in via ALLOW_DOTENV_IN_DOCKER=true
	allowDotenvInDocker := os.Getenv("ALLOW_DOTENV_IN_DOCKER") == "true"
	if !isDockerEnvironment() || allowDotenvInDocker {
		files := []string{".env.local"}

		if appEnv := os.Getenv("APP_ENV"); appEnv != "" && appEnv != "local" {
			envFile := fmt.Sprintf(".env.%s", appEnv)
			if _, err := os.Stat(envFile); err == nil {
				files = append(files, envFile)
			}
		}

		if err := godotenv.Overload(files...); err != nil {
			log.Printf("Warning: could not load env files %v: %v", files, err)
		}
	}

	cfg := Config{
		App: AppConfig{
			AppEnv:        getEnv("APP_ENV"),
			AppPort:       getEnv("APP_PORT"),
			PublicBaseURL: getEnv("PUBLIC_BASE_URL"),
		},
		Postgres: PostgresConfig{
			DSN:             getEnvOptional("DATABASE_URL"),
			MaxConns:        25,
			MinConns:        5,
			MaxConnLifetime: 1 * time.Hour,
			MaxConnIdleTime: 30 * time.Minute,
		},
		Logger: LoggerConfig{
			Level:       getEnv("LOG_LEVEL"),
			Development: getEnvAsBool("LOG_DEVELOPMENT"),
		},
		Security: SecurityConfig{
			EnableHSTS:     getEnvAsBool("SECURITY_ENABLE_HSTS"),
			EnableCSP:      getEnvAsBool("SECURITY_ENABLE_CSP"),
			TrustedOrigins: getTrustedOrigins("SECURITY_TRUSTED_ORIGINS"),
		},
		Payrexx: PayrexxConfig{
			InstanceName:  getEnvOptional("PAYREXX_INSTANCE"),
			APISecret:     getEnvOptional("PAYREXX_API_SECRET"),
			WebhookSecret: getEnvOptional("PAYREXX_WEBHOOK_SECRET"),
		},
		Plunk: PlunkConfig{
			APIKey:    getEnvOptional("PLUNK_API_KEY"),
			FromName:  getEnvOptional("PLUNK_FROM_NAME"),
			FromEmail: getEnvOptional("PLUNK_FROM_EMAIL"),
			ReplyTo:   getEnvOptional("PLUNK_REPLY_TO"),
		},
		BlobStorage: BlobStorageConfig{
			AccountName:  getEnvOptional("AZURE_STORAGE_ACCOUNT_NAME"),
			AccountKey:   getEnvOptional("AZURE_STORAGE_ACCOUNT_KEY"),
			Container:    getEnvWithDefault("AZURE_STORAGE_CONTAINER", "product-images"),
			BlobEndpoint: getEnvOptional("AZURE_STORAGE_BLOB_ENDPOINT"),
		},
		Elvanto: ElvantoConfig{
			APIKey:  getEnvOptional("ELVANTO_API_KEY"),
			GroupID: getEnvWithDefault("ELVANTO_GROUP_ID", "fc939b75-cda0-4e37-b728-a61e943d66ad"),
		},
	}

	return cfg
}

// isDockerEnvironment checks if we're running inside a Docker container
func isDockerEnvironment() bool {
	if os.Getenv("DOCKER_CONTAINER") == "true" {
		return true
	}

	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	return false
}

// getEnv gets a environment variable or panics
func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Fatalf("config: environment variable %s is not set", key)
	return ""
}

// getEnvAsBool gets an environment variable as a boolean or panics
func getEnvAsBool(key string) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	log.Fatalf("config: environment variable %s is not set", key)
	return false
}

// getTrustedOrigins parses comma-separated list of trusted origins from environment
func getTrustedOrigins(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("config: environment variable %s is not set", key)
		return make([]string, 0)
	}

	origins := strings.Split(value, ",")
	var trimmedOrigins []string
	for _, origin := range origins {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			trimmedOrigins = append(trimmedOrigins, trimmed)
		}
	}
	return trimmedOrigins
}

// getEnvOptional gets an environment variable or returns empty string
func getEnvOptional(key string) string {
	return os.Getenv(key)
}

// getEnvWithDefault gets an environment variable or returns the default value
func getEnvWithDefault(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return def
}

