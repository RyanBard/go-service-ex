package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port     int    `envconfig:"PORT" default:"4000"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"debug"`
	DB       DBConfig
}

type DBConfig struct {
	User         string        `envconfig:"DB_USER" default:"postgres"`
	Password     string        `envconfig:"DB_PASSWORD"`
	DBName       string        `envconfig:"DB_NAME" default:"postgres"`
	SSLMode      string        `envconfig:"DB_SSL_MODE" default:"disable"`
	QueryTimeout time.Duration `envconfig:"DB_QUERY_TIMEOUT" default:"30s"`
}

func LoadConfig() (c Config, err error) {
	err = envconfig.Process("gin-ex", &c)
	return c, err
}
