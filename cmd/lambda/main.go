package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/Vasenti/stori_challenge/internal/application/ports"
	"github.com/Vasenti/stori_challenge/internal/application/services"
	"github.com/Vasenti/stori_challenge/internal/config"
	"github.com/Vasenti/stori_challenge/internal/domain"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/db"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/db/reader"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/db/repositories"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/email"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/parser"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/templating"
)

type Event struct {
	Email    string `json:"email"`
	Src      string `json:"src"`
	Template string `json:"template,omitempty"`
}

type Response struct {
	OK       bool    `json:"ok"`
	Message  string  `json:"message,omitempty"`
	Balance  float64 `json:"balance,omitempty"`
	AvgDebit float64 `json:"avg_debit,omitempty"`
	AvgCred  float64 `json:"avg_credit,omitempty"`
}

func isLikelyPath(s string) bool {
	if s == "" { return false }
	if strings.HasPrefix(s, "s3://") { return true }
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "./") { return true }
	if strings.HasSuffix(strings.ToLower(s), ".html") || strings.HasSuffix(strings.ToLower(s), ".tmpl") { return true }
	return false
}

func loadTemplate(ctx context.Context, tpl string, rdr ports.Reader) (string, error) {
	if tpl == "" { // usa default embebido
		return "", nil
	}
	// S3
	if strings.HasPrefix(tpl, "s3://") {
		rc, err := rdr.Open(tpl)
		if err != nil { return "", fmt.Errorf("open template s3: %w", err) }
		defer rc.Close()
		b, err := io.ReadAll(rc)
		if err != nil { return "", fmt.Errorf("read template s3: %w", err) }
		return string(b), nil
	}
	// Ruta local
	if isLikelyPath(tpl) {
		b, err := os.ReadFile(tpl)
		if err != nil { return "", fmt.Errorf("read template file: %w", err) }
		return string(b), nil
	}
	// Si no parece ruta, asumimos HTML inline
	return tpl, nil
}

func handler(ctx context.Context, e Event) (Response, error) {
	if e.Email == "" || e.Src == "" {
		return Response{OK: false, Message: "email and src are required"}, fmt.Errorf("missing email/src")
	}
	cfg, err := config.Load()
	if err != nil { return Response{OK: false, Message: "config error"}, err }

	gdb, err := db.NewGorm(cfg)
	if err != nil { return Response{OK: false, Message: "db error"}, err }

	users := repositories.NewUserRepository(gdb)
	trxs  := repositories.NewTransactionRepository(gdb)

	// reader para CSV y (si hace falta) para template en S3
	var rdr ports.Reader = reader.LocalFileReader{}
	if strings.HasPrefix(e.Src, "s3://") || strings.HasPrefix(e.Template, "s3://") {
		s3r, err := reader.NewS3Reader(cfg.S3Region, cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3ForcePathStyle)
		if err != nil { return Response{OK:false, Message:"s3 reader error"}, err }
		rdr = s3r
	}

	// Cargar contenido del template (de ruta local, S3 o HTML inline)
	tplContent, err := loadTemplate(ctx, e.Template, rdr)
	if err != nil { return Response{OK:false, Message:"template load error"}, err }

	mailer := email.NewSMTPSender(cfg)

	render := func(sum domain.MonthlySummary, u, t string) (string, error) {
		return templating.Render(sum, u, t, time.Now())
	}

	svc := services.NewTransactionReportService(
		rdr,
		users,
		trxs,
		mailer,
		render,
		parser.ParseTransactionsCSV,
	)

	if err := svc.Process(ctx, e.Email, e.Src, tplContent); err != nil {
		return Response{OK: false, Message: err.Error()}, err
	}

	sum, err := trxs.GetMonthlySummary(ctx, e.Email)
	if err != nil { return Response{OK: true, Message: "email sent (summary fetch failed)"}, nil }

	return Response{
		OK:       true,
		Message:  "email sent",
		Balance:  sum.BalanceTotal,
		AvgDebit: sum.AvgDebit,
		AvgCred:  sum.AvgCredit,
	}, nil
}

func main() { lambda.Start(handler) }
