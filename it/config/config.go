package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	BaseURL   string `envconfig:"BASE_URL" default:"http://localhost:4000"`
	LogLevel  string `envconfig:"LOG_LEVEL" default:"debug"`
	AuthToken string `envconfig:"AUTH_TOKEN"`
}

func LoadConfig() (c Config, err error) {
	err = envconfig.Process("gin-ex-its", &c)
	return c, err
}
