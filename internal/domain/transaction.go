package domain

import "time"

type Transaction struct {
	ID         uint    `gorm:"primaryKey"`  
	UserEmail  string    `gorm:"index;not null"`
	OccurredAt time.Time `gorm:"index;not null"`
	Amount     float64   `gorm:"not null"`
	RawDate    string    `gorm:"not null"`
	RawAmount  string    `gorm:"not null"`
}

func (Transaction) TableName() string { return "transactions" }