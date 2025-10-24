package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Mongo    MongoConfig
	Logger   LoggerConfig
	Smtp     SmtpConfig
	Security SecurityConfig
	Stripe   StripeConfig
	OAuth    OAuthConfig
	Stations StationConfig
}

type AppConfig struct {
	AppEnv        string
	AppPort       string
	JWTIssuer     string
	JWTPrivPEM    string
	JWTPubPEM     string
	PublicBaseURL string
}

type MongoConfig struct {
	URI      string
	Database string
}

type LoggerConfig struct {
	Level       string
	Development bool
}

type SmtpConfig struct {
	Host      string
	Port      string
	Username  string
	Password  string
	From      string
	TLSPolicy string
}

type SecurityConfig struct {
	EnableHSTS     bool
	EnableCSP      bool
	TrustedOrigins []string
}

type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
}

type OAuthConfig struct {
	Google GoogleConfig
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string // optional; recommended for web client exchange
}

type StationConfig struct {
	QRSecret        string
	QRMaxAgeSeconds int
}

func Load() Config {
	// Load .env files by default outside Docker. In Docker, allow opt-in via ALLOW_DOTENV_IN_DOCKER=true
	allowDotenvInDocker := os.Getenv("ALLOW_DOTENV_IN_DOCKER") == "true"
	if !isDockerEnvironment() || allowDotenvInDocker {
		files := []string{".env"}

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
			JWTIssuer:     getEnv("JWT_ISSUER"),
			JWTPrivPEM:    getEnv("JWT_PRIV_PEM"),
			JWTPubPEM:     getEnv("JWT_PUB_PEM"),
			PublicBaseURL: getEnv("PUBLIC_BASE_URL"),
		},
		Mongo: MongoConfig{
			URI:      getEnv("MONGO_URI"),
			Database: getEnv("MONGO_DATABASE"),
		},
		Logger: LoggerConfig{
			Level:       getEnv("LOG_LEVEL"),
			Development: getEnvAsBool("LOG_DEVELOPMENT"),
		},
		Smtp: SmtpConfig{
			Host:      getEnv("SMTP_HOST"),
			Port:      getEnv("SMTP_PORT"),
			Username:  getEnv("SMTP_USERNAME"),
			Password:  getEnv("SMTP_PASSWORD"),
			From:      getEnv("SMTP_FROM"),
			TLSPolicy: getEnv("SMTP_TLS_POLICY"),
		},
		Security: SecurityConfig{
			EnableHSTS:     getEnvAsBool("SECURITY_ENABLE_HSTS"),
			EnableCSP:      getEnvAsBool("SECURITY_ENABLE_CSP"),
			TrustedOrigins: getTrustedOrigins("SECURITY_TRUSTED_ORIGINS"),
		},
		Stripe: StripeConfig{
			SecretKey:     getEnv("STRIPE_SECRET_KEY"),
			WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET"),
		},
		OAuth: OAuthConfig{
			Google: GoogleConfig{
				ClientID:     getEnv("GOOGLE_CLIENT_ID"),
				ClientSecret: getEnvOptional("GOOGLE_CLIENT_SECRET"),
			},
		},
		Stations: StationConfig{
			QRSecret:        getEnv("STATION_QR_SECRET"),
			QRMaxAgeSeconds: getEnvAsInt("STATION_QR_MAX_AGE_SECONDS", 600),
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

// getEnvAsBoolDefault gets a boolean from env or returns a default when unset/malformed
func getEnvAsBoolDefault(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	return def
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

// getEnvAsInt returns an int or default if empty/malformed
func getEnvAsInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if i, err := strconv.Atoi(v); err == nil {
		return i
	}
	return def
}
