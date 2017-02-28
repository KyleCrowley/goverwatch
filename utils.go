package main

import (
	"strconv"
	"strings"
)

func TrimToInt(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}

func TrimToString(s string) string {
	return string(strings.TrimSpace(s))
}

func TrimToFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 32)
	return f
}
