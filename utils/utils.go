package utils

import (
	"regexp"
	"strings"
)

func CamelToSnake(s string) string {
	re1 := regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
	s = re1.ReplaceAllString(s, "${1}_${2}")
	re2 := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	s = re2.ReplaceAllString(s, "${1}_${2}")

	return strings.ToLower(s)
}
