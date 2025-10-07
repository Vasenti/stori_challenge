package ports

import (
	"context"
	"github.com/Vasenti/stori_challenge/internal/domain"
)

type UserRepository interface {
	Ensure(ctx context.Context, email string) error
}

type TransactionRepository interface {
	BulkUpsert(ctx context.Context, txs []domain.Transaction) error
	GetMonthlySummary(ctx context.Context, userEmail string) (domain.MonthlySummary, error)
}