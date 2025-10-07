package ports

import "context"

type TransactionReportService interface {
	Process(ctx context.Context, userEmail string, csvSourcePath string) error
}