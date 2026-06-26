package game

import (
	"regexp"
	"strings"
)

var bracketRegex = regexp.MustCompile(`\([^)]*\)|\[[^\]]*\]|\{[^}]*\}`)

func Matches(input, title string) bool {
	return strings.EqualFold(input, PrimaryTitle(title))
}

func PrimaryTitle(title string) string {
	clean := bracketRegex.ReplaceAllString(title, "")
	if idx := strings.Index(clean, " - "); idx != -1 {
		clean = clean[:idx]
	}
	return clean
}