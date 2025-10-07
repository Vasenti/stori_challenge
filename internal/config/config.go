package config

import "github.com/caarlos0/env/v10"

type Config struct {
	AppEnv   string `env:"APP_ENV,notEmpty" envDefault:"dev"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	return cfg, env.Parse(cfg)
}