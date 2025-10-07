package templating

import (
	"bytes"
	"html/template"
	"sort"
	"time"
	_ "embed"
	"github.com/Vasenti/stori_challenge/internal/domain"
)

//go:embed default_report.html.tmpl
var defaultHTML []byte

type MonthCount struct {
	MonthName string
	Count     int
}
type Model struct {
	UserEmail    string
	Now          time.Time
	BalanceTotal float64
	AvgDebit     float64
	AvgCredit    float64
	ByMonth      []MonthCount
}

func Render(summary domain.MonthlySummary, userEmail, tpl string, now time.Time) (string, error) {
	src := defaultHTML
	if tpl != "" {
		src = []byte(tpl)
	}
	byMonth := make([]MonthCount, 0, 12)
	for m, c := range summary.TransactionsByMonth {
		byMonth = append(byMonth, MonthCount{MonthName: m.String(), Count: c})
	}
	sort.Slice(byMonth, func(i, j int) bool { return byMonth[i].MonthName < byMonth[j].MonthName })

	model := Model{
		UserEmail:    userEmail,
		Now:          now,
		BalanceTotal: summary.BalanceTotal,
		AvgDebit:     summary.AvgDebit,
		AvgCredit:    summary.AvgCredit,
		ByMonth:      byMonth,
	}
	t, err := template.New("report").Parse(string(src))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, model); err != nil {
		return "", err
	}
	return buf.String(), nil
}