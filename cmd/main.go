package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/Vasenti/stori_challenge/internal/application/ports"
	"github.com/Vasenti/stori_challenge/internal/application/services"
	"github.com/Vasenti/stori_challenge/internal/config"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/db"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/db/reader"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/parser"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	var emailTo string
	var source string
	var templatePath string

	flag.StringVar(&emailTo, "email", "", "User email to send the report")
	flag.StringVar(&source, "src", "", "CSV Route (local or s3://bucket/key)")
	flag.StringVar(&templatePath, "template", "", "HTML Template Path (local path)")
	flag.Parse()

	if emailTo == "" || source == "" {
		fmt.Println("email and src flags are required")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	_, err = db.NewGorm(cfg)
	if err != nil {
		panic(err)
	}

	var reader ports.Reader = reader.LocalFileReader{}

	svc := services.NewTransactionReportService(
		reader,
		parser.ParseTransactionsCSV,
	)

	if err := svc.Process(context.Background(), emailTo, source); err != nil {
		panic(err)
	}

	fmt.Println("Report processed successfully")
}