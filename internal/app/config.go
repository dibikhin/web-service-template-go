package app

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Port    string        `env:"PORT" envDefault:":8080"`
	Mode    string        `env:"MODE" envDefault:"debug"`
	Timeout time.Duration `env:"TIMEOUT"`

	Postgres PostgresConfig
	Redis    RedisConfig
	Mongo    MongoConfig
}

type PostgresConfig struct {
	Host     string        `env:"POSTGRES_HOST"`
	Port     uint16        `env:"POSTGRES_PORT"`
	User     string        `env:"POSTGRES_USER"`
	Password string        `env:"POSTGRES_PASSWORD,unset"`
	Database string        `env:"POSTGRES_DATABASE"`
	Timeout  time.Duration `env:"POSTGRES_TIMEOUT"`
}

type RedisConfig struct {
	Host     string        `env:"REDIS_HOST"`
	Port     uint16        `env:"REDIS_PORT"`
	Password string        `env:"REDIS_PASSWORD,unset"`
	Timeout  time.Duration `env:"REDIS_TIMEOUT"`
}

type MongoConfig struct {
	Host     string        `env:"MONGO_HOST"`
	Port     uint16        `env:"MONGO_PORT"`
	Database string        `env:"MONGO_DATABASE"`
	Timeout  time.Duration `env:"MONGO_TIMEOUT"`
}

func LoadConfig(filename string) (*Config, error) {
	if err := godotenv.Load(filename); err != nil {
		return nil, fmt.Errorf("loading file: %w", err)
	}
	var cfg Config
	opts := env.Options{RequiredIfNoDef: true}

	if err := env.Parse(&cfg, opts); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}
