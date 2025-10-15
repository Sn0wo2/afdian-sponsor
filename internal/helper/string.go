package helper

import (
	"unsafe"

	"github.com/mattn/go-runewidth"
)

func BytesToString(v []byte) string {
	return *(*string)(unsafe.Pointer(&v)) //nolint:gosec
}

func StringToBytes(v string) []byte {
	return unsafe.Slice(unsafe.StringData(v), len(v)) //nolint:gosec
}

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
