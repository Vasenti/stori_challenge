package domain

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func ParseAmount(s string) (float64, error) {
	ss := strings.TrimSpace(s)
	ss = strings.ReplaceAll(ss, ",", "")
	ss = strings.TrimPrefix(ss, "+")
	return strconv.ParseFloat(ss, 64)
}

func ParseDate(raw string, now time.Time) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, errors.New("empty date")
	}
	withYear := []string{
		"2006/1/2", "2006-1-2", "2006/01/02", "2006-01-02",
		"01/02/2006", "1/2/2006", "01-02-2006", "1-2-2006",
	}
	for _, ly := range withYear {
		if t, err := time.ParseInLocation(ly, raw, now.Location()); err == nil {
			return t, nil
		}
	}
	withoutYear := []string{"1/2", "01/02", "1-2", "01-02"}
	for _, ly := range withoutYear {
		if t, err := time.ParseInLocation(ly, raw, now.Location()); err == nil {
			return time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location()), nil
		}
	}
	return time.Time{}, errors.New("unrecognized date format: " + raw)
}