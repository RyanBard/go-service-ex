package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	BaseURL     string `envconfig:"BASE_URL" default:"http://localhost:4000"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"debug"`
	JWTSecret   string `envconfig:"JWT_SECRET"`
	JWTAudience string `envconfig:"JWT_AUDIENCE" default:"gin-ex"`
	JWTIssuer   string `envconfig:"JWT_ISSUER" default:"something"`
}

func LoadConfig() (c Config, err error) {
	err = envconfig.Process("gin-ex-its", &c)
	return c, err
}
