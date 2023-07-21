package config

import (
	"github.com/caarlos0/env/v9"
)

type Config struct {
	ListenAddr         string `env:"LISTEN_ADDR" envDefault:":8000"`
	RedisUrl           string `env:"REDIS_URL" envDefault:"redis://localhost:6379/0"`
	PostgresqlUrl      string `env:"POSTGRESQL_URL" envDefault:"postgresql://postgres:postgres@postgresql:5432/postgres"`
	ProductServiceUrl  string `env:"PRODUCT_SERVICE_URL" envDefault:"http://localhost:8001"`
	ExchangeServiceUrl string `env:"EXCHANGE_SERVICE_URL" envDefault:"http://localhost:8002"`
	PaymentServiceUrl  string `env:"PAYMENT_SERVICE_URL" envDefault:"http://localhost:8003"`
}

func LoadConfig() (*Config, error) {
	conf := &Config{}
	err := env.Parse(conf)
	return conf, err
}
