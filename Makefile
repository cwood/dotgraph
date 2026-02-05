.PHONY: test test-unit test-coverage test-html clean

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-unit:
	go test -v ./...

# Run tests with coverage
test-coverage:
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./...
	@echo ""
	@echo "Coverage report written to coverage/coverage.out"
	@go tool cover -func=coverage/coverage.out

# Generate HTML coverage report
test-html: test-coverage
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo ""
	@echo "HTML coverage report: coverage/coverage.html"
	@open coverage/coverage.html 2>/dev/null || true

# Clean coverage artifacts
clean:
	@rm -rf coverage/
