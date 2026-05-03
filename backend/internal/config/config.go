package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                   string    `yaml:"env" env-default:"local"`
	DB                    DBConfig  `yaml:"-"`
	JWT                   JWTConfig `yaml:"-"`
	HTTPServer            `yaml:"http_server"`
	CORSAllowedOriginsRaw string `yaml:"-" env:"CORS_ALLOWED_ORIGINS" env-default:"http://localhost:5173"`
}

type DBConfig struct {
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	User     string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"password"`
	DBName   string `env:"POSTGRES_DB" env-default:"notree"`
}

type JWTConfig struct {
	Secret string `env:"JWT_SECRET" env-required:"true"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"0.0.0.0:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func (c *Config) CORSAllowedOrigins() []string {
	if c.CORSAllowedOriginsRaw == "" {
		return nil
	}

	origins := strings.Split(c.CORSAllowedOriginsRaw, ",")

	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	return origins
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
