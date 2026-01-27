package handler

import (
	"strconv"
	"time"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// parseMonthYear
func parseMonthYear(s string) (time.Time, error) {
	return time.Parse("01-2006", s)
}

// parseQueryInt парсит целое число из строки запроса
func parseQueryInt(s string, target *int) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	*target = val
	return val, nil
}
