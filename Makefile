.PHONY: release font

release:
	@echo "Running release tool..."
	@go run ./scripts/release/main.go

font:
	@echo "Fetching Google Font at build time..."
	@go generate ./internal/font