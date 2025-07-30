package common

import (
	"github.com/mattn/go-runewidth"
)

func StringWidth(s string) int {
	return runewidth.StringWidth(s)
}

func TruncateStringByWidth(s string, limit int) string {
	if runewidth.StringWidth(s) <= limit {
		return s
	}

	width := 0

	runes := []rune(s)
	for i, r := range runes {
		width += runewidth.RuneWidth(r)
		if width > limit {
			return string(runes[:i]) + "..."
		}
	}

	return s
}
