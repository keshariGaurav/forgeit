.PHONY: fmt lint lint-fix

# Format Go files
fmt:
	goimports -w .

# Run linter
lint:
	golangci-lint run

# Format and fix linting issues
lint-fix:
	goimports -w .
	golangci-lint run --fix