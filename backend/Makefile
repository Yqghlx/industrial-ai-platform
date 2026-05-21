.PHONY: all test build lint fmt clean help

# 变量
GO := go
GOFMT := gofmt
GOLINT := golangci-lint
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
	@echo "  help            - Show this help message"