package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Vasenti/stori_challenge/internal/domain"
)

func ParseTransactionsCSV(r io.Reader, userEmail string, now time.Time) ([]domain.Transaction, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	cr.FieldsPerRecord = -1

	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("csv is empty")
	}

	start := 0
	if len(rows[0]) >= 3 && strings.EqualFold(strings.TrimSpace(rows[0][0]), "id") {
		start = 1
	}

	out := make([]domain.Transaction, 0, len(rows)-start)
	for i := start; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 3 {
			continue
		}
		idStr := strings.TrimSpace(row[0])
		rawDate := strings.TrimSpace(row[1])
		rawAmt := strings.TrimSpace(row[2])

		idU64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("row %d: id invalid (%s): %w", i+1, idStr, err)
		}
		t, err := domain.ParseDate(rawDate, now)
		if err != nil {
			return nil, fmt.Errorf("row %d: date invalid: %w", i+1, err)
		}
		amt, err := domain.ParseAmount(rawAmt)
		if err != nil {
			return nil, fmt.Errorf("row %d: amount invalid: %w", i+1, err)
		}

		out = append(out, domain.Transaction{
			ID:         uint(idU64),
			UserEmail:  userEmail,
			OccurredAt: t,
			Amount:     amt,
			RawDate:    rawDate,
			RawAmount:  rawAmt,
		})
	}
	return out, nil
}