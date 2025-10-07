package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Vasenti/stori_challenge/internal/application/ports"
	"github.com/Vasenti/stori_challenge/internal/application/services"
	"github.com/Vasenti/stori_challenge/internal/config"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/db"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/db/reader"
	"github.com/Vasenti/stori_challenge/internal/intrastructure/db/repositories"
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

	gdb, err := db.NewGorm(cfg)
	if err != nil {
		panic(err)
	}

	users := repositories.NewUserRepository(gdb)
	transactions := repositories.NewTransactionRepository(gdb)

	var rdr ports.Reader = reader.LocalFileReader{}
	if strings.HasPrefix(source, "s3://") {
		s3r, err := reader.NewS3Reader(cfg.S3Region, cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3ForcePathStyle)
		if err != nil { panic(err) }
		rdr = s3r
	}

	svc := services.NewTransactionReportService(
		rdr,
		users,
		transactions,
		parser.ParseTransactionsCSV,
	)

	if err := svc.Process(context.Background(), emailTo, source); err != nil {
		panic(err)
	}
}