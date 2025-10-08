package config

import "github.com/caarlos0/env/v10"

type Config struct {
	AppEnv   string `env:"APP_ENV,notEmpty" envDefault:"dev"`

	// DB
	DBHost            string `env:"DB_HOST,notEmpty"`
	DBPort            int    `env:"DB_PORT" envDefault:"5432"`
	DBUser            string `env:"DB_USER,notEmpty" envDefault:"postgres"`
	DBPassword        string `env:"DB_PASSWORD,notEmpty" envDefault:"postgres"`
	DBName            string `env:"DB_NAME,notEmpty"`
	DBSSLMode         string `env:"DB_SSLMODE" envDefault:"disable"`
	DBMaxOpen         int    `env:"DB_MAX_OPEN" envDefault:"10"`
	DBMaxIdle         int    `env:"DB_MAX_IDLE" envDefault:"5"`
	DBMaxLifetimeSecs int    `env:"DB_MAX_LIFETIME_SECS" envDefault:"600"`

	// S3 (not required, only if using S3 as source)
	S3Region         string `env:"S3_REGION" envDefault:"us-east-1"`
	S3Endpoint       string `env:"S3_ENDPOINT"`
	S3AccessKey      string `env:"S3_ACCESS_KEY"`
	S3SecretKey      string `env:"S3_SECRET_KEY"`
	S3ForcePathStyle bool   `env:"S3_FORCE_PATH_STYLE" envDefault:"false"`

	// SMTP
	SMTPHost     string `env:"SMTP_HOST" envDefault:"localhost"`
	SMTPPort     int    `env:"SMTP_PORT" envDefault:"1025"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	SMTPFrom     string `env:"SMTP_FROM,notEmpty" envDefault:"no-reply@example.com"`

	ReportTemplatePath string `env:"REPORT_TEMPLATE_PATH"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	return cfg, env.Parse(cfg)
}