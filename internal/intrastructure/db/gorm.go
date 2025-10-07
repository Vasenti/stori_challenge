package db

import (
	"fmt"
	"time"

	"github.com/Vasenti/stori_challenge/internal/config"
	"github.com/Vasenti/stori_challenge/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewGorm(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	gormLogger := logger.Default.LogMode(logger.Warn)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpen)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBMaxLifetimeSecs) * time.Second)

	if err := db.AutoMigrate(&domain.User{}, &domain.Transaction{}); err != nil {
		return nil, err
	}

	return db, nil
}