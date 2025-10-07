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
	urepo  ports.UserRepository
	trepo  ports.TransactionRepository
	email  ports.EmailSender
	renderHTML func(domain.MonthlySummary, string, string) (string, error)
	parseCSV   func(io.Reader, string, time.Time) ([]domain.Transaction, error)
}

func NewTransactionReportService(
	reader ports.Reader,
	urepo ports.UserRepository,
	trepo ports.TransactionRepository,
	email ports.EmailSender,
	renderHTML func(domain.MonthlySummary, string, string) (string, error),
	parseCSV func(io.Reader, string, time.Time) ([]domain.Transaction, error),
) ports.TransactionReportService {
	return &TransactionReportService{
		reader:   reader,
		urepo:   urepo,
		trepo:   trepo,
		email:    email,
		renderHTML: renderHTML,
		parseCSV: parseCSV,
	}
}

func (s *TransactionReportService) Process(ctx context.Context, userEmail string, csvSourcePath string, templateHtml string) error {
	// 1) Ensure user exists or create it
	if err := s.urepo.Ensure(ctx, userEmail); err != nil {
		return fmt.Errorf("ensure user: %w", err)
	}

	// 2) Read CSV from source (local FS or S3)
	rc, err := s.reader.Open(csvSourcePath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer rc.Close()

	// 3) Parse CSV
	transactions, err := s.parseCSV(rc, userEmail, time.Now())
	if err != nil {
		return fmt.Errorf("parse csv: %w", err)
	}

	// 4) Bulk upsert transactions
	if err := s.trepo.BulkUpsert(ctx, transactions); err != nil {
		return fmt.Errorf("bulk upsert: %w", err)
	}

	// 5) Get monthly summary
	summary, err := s.trepo.GetMonthlySummary(ctx, userEmail)
	if err != nil {
		return fmt.Errorf("get monthly summary: %w", err)
	}

	fmt.Printf("Get monthly summary - balance: %f, avgCredit: %f, avgDebit: %f", summary.BalanceTotal, summary.AvgCredit, summary.AvgDebit)

	// 6) Render HTML
	htmlBody, err := s.renderHTML(summary, userEmail, templateHtml)
	if err != nil {
		return fmt.Errorf("render html: %w", err)
	}

	// 7) Send email
	subject := fmt.Sprintf("Your transaction report - %s", time.Now().Format("January 2006"))
	if err := s.email.Send(userEmail, subject, htmlBody); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	fmt.Println("Email sent to", userEmail)
	return nil
}