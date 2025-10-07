package repositories

import (
	"context"
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
	type Row struct {
		Month int
		Count int
	}
	var rows []Row

	if err := r.db.WithContext(ctx).
		Raw(`
			SELECT EXTRACT(MONTH FROM occurred_at)::int AS month, COUNT(*) AS count
			FROM transactions
			WHERE user_email = ?
			GROUP BY 1
			ORDER BY 1
		`, userEmail).Scan(&rows).Error; err != nil {
		return domain.MonthlySummary{}, err
	}

	trxByMonth := make(map[time.Month]int)
	for _, rrow := range rows {
		trxByMonth[time.Month(rrow.Month)] = rrow.Count
	}

	type Agg struct {
		Balance float64
		AvgDeb  *float64
		AvgCred *float64
	}
	var agg Agg
	if err := r.db.WithContext(ctx).
		Raw(`
			WITH base AS (
			  SELECT amount FROM transactions WHERE user_email = ?
			),
			debits AS (
			  SELECT ABS(amount) AS val FROM base WHERE amount < 0
			),
			credits AS (
			  SELECT amount AS val FROM base WHERE amount > 0
			)
			SELECT
			  (SELECT COALESCE(SUM(amount),0) FROM base) AS balance,
			  (SELECT AVG(val) FROM debits) AS avg_deb,
			  (SELECT AVG(val) FROM credits) AS avg_cred
		`, userEmail).Scan(&agg).Error; err != nil {
		return domain.MonthlySummary{}, err
	}

	var avgDeb, avgCred float64
	if agg.AvgDeb != nil {
		avgDeb = *agg.AvgDeb
	}
	if agg.AvgCred != nil {
		avgCred = *agg.AvgCred
	}

	return domain.MonthlySummary{
		BalanceTotal:        agg.Balance,
		TransactionsByMonth: trxByMonth,
		AvgDebit:            avgDeb,
		AvgCredit:           avgCred,
	}, nil
}