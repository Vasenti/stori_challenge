package repositories

import (
	"context"
	"math"
	"time"

	"github.com/Vasenti/stori_challenge/internal/application/ports"
	"github.com/Vasenti/stori_challenge/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type transactionRepo struct{ db *gorm.DB }

func NewTransactionRepository(db *gorm.DB) ports.TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) BulkUpsert(ctx context.Context, txs []domain.Transaction) error {
	if len(txs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}}, // PK
			DoNothing: true,
		}).
		Create(&txs).Error
}

func (r *transactionRepo) GetMonthlySummary(ctx context.Context, userEmail string) (domain.MonthlySummary, error) {
	var txs []domain.Transaction

	if err := r.db.WithContext(ctx).
		Select("occurred_at", "amount").
		Where("user_email = ?", userEmail).
		Find(&txs).Error; err != nil {
		return domain.MonthlySummary{}, err
	}

	trxByMonth := make(map[time.Month]int, 12)

	var balance float64
	var sumCredits float64
	var cntCredits int
	var sumDebitsAbs float64
	var cntDebits int

	for _, t := range txs {
		balance += t.Amount

		trxByMonth[t.OccurredAt.Month()]++

		if t.Amount > 0 {
			sumCredits += t.Amount
			cntCredits++
		} else if t.Amount < 0 {
			sumDebitsAbs += math.Abs(t.Amount)
			cntDebits++
		}
	}

	var avgCredit, avgDebit float64
	if cntCredits > 0 {
		avgCredit = sumCredits / float64(cntCredits)
	}
	if cntDebits > 0 {
		avgDebit = sumDebitsAbs / float64(cntDebits)
	}

	return domain.MonthlySummary{
		BalanceTotal:        balance,
		TransactionsByMonth: trxByMonth,
		AvgDebit:            avgDebit,
		AvgCredit:           avgCredit,
	}, nil
}