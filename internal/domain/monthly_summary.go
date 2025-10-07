package domain

import "time"

type MonthlySummary struct {
	BalanceTotal        float64
	TransactionsByMonth map[time.Month]int
	AvgDebit            float64 
	AvgCredit           float64
}