package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Vasenti/stori_challenge/internal/application/ports"
	"github.com/Vasenti/stori_challenge/internal/domain"
)

type TransactionReportService struct {
	reader ports.Reader
	parseCSV   func(io.Reader, string, time.Time) ([]domain.Transaction, error)
}

func NewTransactionReportService(
	reader ports.Reader,
	parseCSV func(io.Reader, string, time.Time) ([]domain.Transaction, error),
) ports.TransactionReportService {
	return &TransactionReportService{
		reader:   reader,
		parseCSV: parseCSV,
	}
}

func (s *TransactionReportService) Process(ctx context.Context, userEmail string, csvSourcePath string) error {
	rc, err := s.reader.Open(csvSourcePath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer rc.Close()

	txs, err := s.parseCSV(rc, userEmail, time.Now())
	if err != nil {
		return fmt.Errorf("parse csv: %w", err)
	}

	fmt.Println("Parsed transactions:", txs)

	return nil
}