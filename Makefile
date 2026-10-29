BINARY_NAME ?= afdian-sponsor

.PHONY: lint lint-fix test build clean security ci release dry-release font

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run

lint-fix:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run --fix

test:
	go test $(if $(filter 1,$(shell go env CGO_ENABLED)),-race) ./...

build:
	go build -trimpath -ldflags "-s -w" -o $(BINARY_NAME) ./cmd

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

security:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

ci: lint test security

release:
	goreleaser release --clean

dry-release:
	goreleaser release --snapshot --clean

font:
	@echo "Fetching Google Font at build time..."
	@go generate ./internal/font
