package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App          AppConfig
	DB           DBConfig
	Redis        RedisConfig
	Logger       LoggerConfig
	Mailgun      MailgunConfig
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type AppConfig struct {
	AppEnv       string
	AppPort      string
	JWTSecretKey string
}

type DBConfig struct {
	User string
	Pass string
	Name string
	Host string
	Port string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type LoggerConfig struct {
	Level       string
	Development bool
}

type MailgunConfig struct {
	Domain    string
	APIKey    string
	FromEmail string
	FromName  string
}

func Load() Config {
	// Load .env files only if not in Docker environment
	if !isDockerEnvironment() {
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
			AppEnv:       getEnvOrDefault("APP_ENV", "local"),
			AppPort:      getEnvOrDefault("APP_PORT", "8080"),
			JWTSecretKey: getEnvRequired("JWT_SECRET_KEY"),
		},
		DB: DBConfig{
			User: getEnvRequired("POSTGRES_USER"),
			Pass: getEnvRequired("POSTGRES_PASSWORD"),
			Name: getEnvOrDefault("POSTGRES_DB", "rentro"),
			Host: getDBHost(),
			Port: getEnvOrDefault("POSTGRES_PORT", "5432"),
		},
		Redis: RedisConfig{
			Host:     getRedisHost(),
			Port:     getEnvOrDefault("REDIS_PORT", "6379"),
			Password: getEnvRequired("REDIS_PASSWORD"),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Logger: LoggerConfig{
			Level:       getEnvOrDefault("LOG_LEVEL", "info"),
			Development: getEnvAsBool("LOG_DEVELOPMENT", true),
		},
		Mailgun: MailgunConfig{
			Domain:    getEnvRequired("MAILGUN_DOMAIN"),
			APIKey:    getEnvRequired("MAILGUN_API_KEY"),
			FromEmail: getEnvRequired("MAILGUN_FROM_EMAIL"),
			FromName:  getEnvOrDefault("MAILGUN_FROM_NAME", "Rentro"),
		},
		ReadTimeout:  parseDuration("READ_TIMEOUT", "10s"),
		WriteTimeout: parseDuration("WRITE_TIMEOUT", "15s"),
	}

	return cfg
}

// isDockerEnvironment checks if we're running inside a Docker container
func isDockerEnvironment() bool {
	// Check for Docker-specific environment variables or files
	if os.Getenv("DOCKER_CONTAINER") == "true" {
		return true
	}
	
	// Check for /.dockerenv file (standard Docker indicator)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	
	return false
}

// getDBHost returns the appropriate database host based on environment
func getDBHost() string {
	if isDockerEnvironment() {
		return getEnvOrDefault("POSTGRES_HOST", "postgres")
	}
	return getEnvOrDefault("POSTGRES_HOST", "localhost")
}

// getRedisHost returns the appropriate Redis host based on environment
func getRedisHost() string {
	if isDockerEnvironment() {
		return getEnvOrDefault("REDIS_HOST", "redis")
	}
	return getEnvOrDefault("REDIS_HOST", "localhost")
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvRequired gets a required environment variable or panics
func getEnvRequired(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Fatalf("config: required environment variable %s is not set", key)
	return ""
}

// getEnvAsInt gets an environment variable as an integer or returns default
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as a boolean or returns default
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// parseDuration parses a duration string or returns default
func parseDuration(key, defaultValue string) time.Duration {
	value := getEnvOrDefault(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	log.Printf("Warning: invalid duration for %s: %s, using default: %s", key, value, defaultValue)
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 10 * time.Second
}

func (d DBConfig) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		d.User, d.Pass, d.Host, d.Port, d.Name)
}

func (r RedisConfig) GetRedisAddr() string {
	return r.Host + ":" + r.Port
}
