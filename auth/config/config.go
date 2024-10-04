package config

import (
	"log"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Configuration struct {
	LoggerLevel string `env:"LOGGER_LEVEL" envDefault:"info"`
	GRPC        GRPC
	Auth        Auth
	DB          DB
	Redis       Redis
}

type GRPC struct {
	Port    int           `env:"GRPC_PORT" envDefault:"4000"`
	Timeout time.Duration `env:"GRPC_TIMEOUT" envDefault:"10h"`
}

type Auth struct {
	AccessSecret  string `env:"ACCESS_SECRET"`
	RefreshSecret string `env:"REFRESH_SECRET"`
	Pepper        string `env:"PEPPER"`
}

type DB struct {
	Address  string `env:"DB_ADDRESS,required"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	Username string `env:"DB_USERNAME,required"`
	Password string `env:"DB_PASSWORD,required"`
	Name     string `env:"DB_NAME,required"`
}

type Redis struct {
	Address  string `env:"REDIS_ADDRESS,required"`
	Port     string `env:"REDIS_PORT" envDefault:"6379"`
	Username string `env:"REDIS_USERNAME,required"`
	Password string `env:"REDIS_PASSWORD,required"`
	ID       int    `env:"REDIS_ID,required"`
}

func NewConfig(files ...string) (*Configuration, error) {
	err := godotenv.Load(files...)
	if err != nil {
		log.Fatalf("Файл .env не найден: %s", err)
	}

	cfg := Configuration{}
	err = env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	err = env.Parse(&cfg.Redis)
	if err != nil {
		return nil, err
	}
	err = env.Parse(&cfg.DB)
	if err != nil {
		return nil, err
	}
	err = env.Parse(&cfg.GRPC)
	if err != nil {
		return nil, err
	}
	err = env.Parse(&cfg.Auth)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
