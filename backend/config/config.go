package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

var (
	once sync.Once
	cfg  *Config
)

type Config struct {
	APP struct {
		Env       string `env:"APP_ENV,required"`
		Version   string `env:"APP_VERSION,required"`
		JWTSecret string `env:"APP_JWT_SECRET,required"`
	}

	HTTP struct {
		Host string `env:"HTTP_HOST,required"`
		Port int    `env:"HTTP_PORT,required"`
	}

	DB struct {
		Host     string `env:"DB_HOST,required"`
		Port     int    `env:"DB_PORT,required"`
		User     string `env:"DB_USER,required"`
		Password string `env:"DB_PASSWORD,required"`
		Name     string `env:"DB_NAME,required"`
	}

	AUTH0 struct {
		ClientID     string        `env:"AUTH0_CLIENT_ID,required"`
		Domain       string        `env:"AUTH0_DOMAIN,required"`
		ClientSecret string        `env:"AUTH0_CLIENT_SECRET,required"`
		CallbackUrl  string        `env:"AUTH0_CALLBACK_URL,required"`
		Audience     string        `env:"AUTH0_AUDIENCE,required"`
		CacheTTL     time.Duration `env:"AUTH0_CACHE_TTL,default=5m"`
	}
}

func Load(ctx context.Context) (*Config, error) {
	var err error
	once.Do(func() {
		env := strings.ToLower(os.Getenv("APP_ENV"))
		if env == "" {
			env = "development"
		}
		fmt.Printf("Loading configuration for environment: %s\n", env)

		paths := []string{
			fmt.Sprintf(".env.%s", env),
			".env",
		}
		_ = godotenv.Load(paths...)

		var c Config
		err = envconfig.Process(ctx, &c)
		if err == nil {
			cfg = &c
		}
	})
	return cfg, err
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		c.DB.User, c.DB.Password, c.DB.Host, c.DB.Port, c.DB.Name)
}

// GetBaseURL returns the base URL for the application
func (c *Config) GetBaseURL() string {
	return fmt.Sprintf("http://%s:%d", c.HTTP.Host, c.HTTP.Port)
}
