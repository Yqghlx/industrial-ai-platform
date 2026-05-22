.PHONY: all test build lint fmt clean help security-scan vulncheck

# 变量
GO := go
GOFMT := gofmt
GOLINT := golangci-lint
GOVULNCHECK := govulncheck
COVERAGE_FILE := coverage.out

# 默认目标
all: test build

# 运行所有测试
test:
	@echo "🧪 Running tests..."
	$(GO) test ./... -v -coverprofile=$(COVERAGE_FILE)
	@echo "📊 Coverage report:"
	$(GO) tool cover -func=$(COVERAGE_FILE) | tail -1

# 运行单元测试
test-unit:
	@echo "🧪 Running unit tests..."
	$(GO) test ./internal/handler/... ./internal/service/... ./pkg/... -v

# 运行集成测试（需要本地数据库）
test-integration:
	@echo "🧪 Running integration tests..."
	$(GO) test ./tests/integration/... -v

# 运行E2E测试
test-e2e:
	@echo "🧪 Running E2E tests..."
	$(GO) test ./tests/e2e/... -v

# 生成覆盖率HTML报告
coverage-html:
	@echo "📊 Generating coverage HTML report..."
	$(GO) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# 构建
build:
	@echo "🏗️ Building..."
	$(GO) build -v -o bin/industrial-ai-platform .
	@echo "✅ Build complete: bin/industrial-ai-platform"

# 运行linter
lint:
	@echo "🔍 Running linters..."
	$(GOLINT) run --timeout=5m

# 检查格式
fmt-check:
	@echo "🔍 Checking format..."
	@if [ -n "$(shell $(GOFMT) -l .)" ]; then \
		echo "❌ Files need formatting:"; \
		$(GOFMT) -l .; \
		exit 1; \
	fi
	@echo "✅ Format check passed"

# 格式化代码
fmt:
	@echo "📝 Formatting code..."
	$(GOFMT) -w .
	@echo "✅ Code formatted"

# 清理
clean:
	@echo "🧹 Cleaning..."
	$(GO) clean
	rm -f $(COVERAGE_FILE) coverage.html
	rm -rf bin/
	@echo "✅ Clean complete"

# 安装依赖
deps:
	@echo "📦 Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "✅ Dependencies installed"

# SEC-LOW-03: 漏洞扫描自动化
# 运行 Go 漏洞扫描 (govulncheck)
vulncheck:
	@echo "🔒 Running Go vulnerability scan..."
	@if command -v $(GOVULNCHECK) >/dev/null 2>&1; then \
		$(GOVULNCHECK) ./...; \
	else \
		@echo "⚠️  govulncheck not installed. Install with:"; \
		@echo "    go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# 完整安全扫描
security-scan:
	@echo "🔒 Running complete security scan..."
	@echo ""
	@echo "1. Running govulncheck..."
	@if command -v $(GOVULNCHECK) >/dev/null 2>&1; then \
		$(GOVULNCHECK) ./...; \
	else \
		@echo "⚠️  govulncheck not installed. Install with:"; \
		@echo "    go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi
	@echo ""
	@echo "2. Running golangci-lint security checks..."
	$(GOLINT) run --timeout=5m ./... --enable=gosec
	@echo ""
	@echo "✅ Security scan complete"

# 查看帮助
help:
	@echo "Industrial AI Platform - Available targets:"
	@echo ""
	@echo "  test            - Run all tests with coverage"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration- Run integration tests (requires PostgreSQL)"
	@echo "  test-e2e        - Run E2E tests"
	@echo "  coverage-html   - Generate HTML coverage report"
	@echo "  build           - Build the application"
	@echo "  lint            - Run golangci-lint"
	@echo "  fmt-check       - Check code formatting"
	@echo "  fmt             - Format code"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install dependencies"
	@echo "  vulncheck       - Run Go vulnerability scan (SEC-LOW-03)"
	@echo "  security-scan   - Run complete security scan"
	@echo "  help            - Show this help message"