// Package font go generate ./internal/font
package font

import _ "embed"

//go:generate go run ../../scripts/font

//go:embed font.ttf
var Font []byte
