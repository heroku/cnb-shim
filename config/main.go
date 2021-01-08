package config

import (
	"github.com/joeshaw/envdecode"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	RollbarAccessToken string `env:"ROLLBAR_ACCESS_TOKEN"`
	RollbarEnvironment string `env:"ROLLBAR_ENVIRONMENT,default=development"`
	Port               string `env:"PORT,default=:5000"`
}

func LoadConfig(conf *Config) {
	if err := godotenv.Load(); err != nil {
		log.Info("env load error:", err)
	}

	if err := envdecode.Decode(conf); err != nil {
		log.Info("env decode error:", err)
	}
}
