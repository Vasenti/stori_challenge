package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Vasenti/stori_challenge/internal/config"
)

func main() {
	var emailTo string
	var source string
	var templatePath string

	flag.StringVar(&emailTo, "email", "", "User email to send the report")
	flag.StringVar(&source, "src", "", "CSV Route (local r s3://bucket/key)")
	flag.StringVar(&templatePath, "template", "", "HTML Template Path (local path)")
	flag.Parse()

	if emailTo == "" || source == "" {
		fmt.Println("email and src flags are required")
		os.Exit(1)
	}

	_, err := config.Load()
	if err != nil {
		panic(err)
	}
}