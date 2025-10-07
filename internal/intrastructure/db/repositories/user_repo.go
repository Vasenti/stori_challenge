package repositories

import (
	"context"

	"github.com/Vasenti/stori_challenge/internal/application/ports"
	"github.com/Vasenti/stori_challenge/internal/domain"
	"gorm.io/gorm"
)

type userRepo struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) ports.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Ensure(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).FirstOrCreate(&domain.User{Email: email}).Error
}