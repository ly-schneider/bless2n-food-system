package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App          AppConfig
	DB           DBConfig
	Redis        RedisConfig
	Logger       LoggerConfig
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"  envDefault:"10s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"15s"`
}

type AppConfig struct {
	AppEnv       string `env:"APP_ENV"       envDefault:"local"`
	AppPort      string `env:"APP_PORT"      envDefault:"8080"`
	JWTSecretKey string `env:"JWT_SECRET_KEY" required:"true"`
}

type DBConfig struct {
	User string `env:"POSTGRES_USER"     required:"true"`
	Pass string `env:"POSTGRES_PASSWORD" required:"true"`
	Name string `env:"POSTGRES_DB"       envDefault:"rentro"`
	Host string `env:"POSTGRES_HOST"     envDefault:"localhost"`
	Port string `env:"POSTGRES_PORT"     envDefault:"5432"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST"     envDefault:"localhost"`
	Port     string `env:"REDIS_PORT"     envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD" required:"true"`
	DB       int    `env:"REDIS_DB"       envDefault:"0"`
}

type LoggerConfig struct {
	Level       string `env:"LOG_LEVEL"       envDefault:"info"`
	Development bool   `env:"LOG_DEVELOPMENT" envDefault:"true"`
}

func Load() Config {
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

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("config: failed to parse environment variables: %v", err)
	}

	return cfg
}

func (d DBConfig) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		d.User, d.Pass, d.Host, d.Port, d.Name)
}

func (r RedisConfig) GetRedisAddr() string {
	return r.Host + ":" + r.Port
}
