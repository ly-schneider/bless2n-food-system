# Test Makefile for E2E testing

# Use bash with safer defaults
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

# Test configuration
TEST_DB_NAME := bless2n_test_db
TEST_TIMEOUT := 300s
COVERAGE_THRESHOLD := 80

.PHONY: help test-setup test-teardown test-unit test-e2e test-all test-coverage test-clean test-station test-run

help:
	@echo "Available test commands:"
	@echo "  test-setup     - Start test infrastructure (MongoDB, SMTP)"
	@echo "  test-teardown  - Stop test infrastructure"
	@echo "  test-unit      - Run unit tests"
	@echo "  test-e2e       - Run end-to-end tests"
	@echo "  test-run       - Run specific test suite/method (use RUN=pattern)"
	@echo "  test-all       - Run all tests with coverage"
	@echo "  test-coverage  - Generate and view coverage report"
	@echo "  test-clean     - Clean test artifacts and data"
	@echo ""
	@echo "Examples:"
	@echo "  make test-run RUN=TestStationTestSuite"
	@echo "  make test-run RUN=TestAuthTestSuite"
	@echo "  make test-run RUN=TestStationTestSuite/TestStationRequestFlow"

test-setup:
	@echo "Starting test infrastructure..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for services to be ready..."
	@for i in {1..30}; do \
		if docker compose -f docker-compose.test.yml exec mongo-test mongosh --eval "db.adminCommand('ping')" --quiet >/dev/null 2>&1; then \
			echo "Test infrastructure is ready!"; \
			exit 0; \
		fi; \
		echo "Waiting for MongoDB... ($$i/30)"; \
		sleep 2; \
	done; \
	echo "❌ MongoDB failed to start within 60 seconds"; \
	exit 1

test-teardown:
	@echo "Stopping test infrastructure..."
	docker compose -f docker-compose.test.yml down --remove-orphans
	@echo "Test infrastructure stopped!"

test-unit:
	@echo "Running unit tests..."
	cd .. && go test -v -race -timeout $(TEST_TIMEOUT) ./internal/...

test-e2e: test-setup
	@echo "Running end-to-end tests..."
	@cd .. && bash -c 'set -a; \
		source test/config/.env.test; \
		export APP_ENV=test; \
		export MONGO_URI=$$MONGO_URI_LOCAL; \
		export SMTP_HOST=$$SMTP_HOST_LOCAL; \
		export SMTP_PORT=$$SMTP_PORT_LOCAL; \
		set +a; \
		go test -v -race -timeout $(TEST_TIMEOUT) -p 1 ./test/e2e/...'
	@$(MAKE) test-teardown

test-all: test-setup
	@echo "Running all tests with coverage..."
	@cd .. && bash -c 'set -a; \
		source test/config/.env.test; \
		export APP_ENV=test; \
		export MONGO_URI=$$MONGO_URI_LOCAL; \
		export SMTP_HOST=$$SMTP_HOST_LOCAL; \
		export SMTP_PORT=$$SMTP_PORT_LOCAL; \
		set +a; \
		go test -v -race -timeout $(TEST_TIMEOUT) -p 1 -coverprofile=coverage.out -covermode=atomic ./...'
	@$(MAKE) test-teardown

test-coverage:
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Checking coverage threshold..."
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'); \
	echo "Total coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE >= $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "✅ Coverage threshold of $(COVERAGE_THRESHOLD)% met!"; \
	else \
		echo "❌ Coverage threshold of $(COVERAGE_THRESHOLD)% not met. Current: $$COVERAGE%"; \
		exit 1; \
	fi

test-clean:
	@echo "Cleaning test artifacts..."
	rm -f coverage.out coverage.html
	docker compose -f docker-compose.test.yml down --volumes --remove-orphans
	@echo "Test artifacts cleaned!"

# Quick test commands
quick-test: test-unit
	@echo "Quick unit tests completed!"

test-run: test-setup
	@if [ -z "$(RUN)" ]; then \
		echo "❌ Error: Please specify a test pattern with RUN=pattern"; \
		echo "Examples:"; \
		echo "  make test-run RUN=TestStationTestSuite"; \
		echo "  make test-run RUN=TestAuthTestSuite"; \
		echo "  make test-run RUN=TestStationTestSuite/TestStationRequestFlow"; \
		$(MAKE) test-teardown; \
		exit 1; \
	fi
	@echo "Running e2e tests matching pattern: $(RUN)"
	@cd .. && bash -c 'set -a; \
		source test/config/.env.test; \
		export APP_ENV=test; \
		export MONGO_URI=$$MONGO_URI_LOCAL; \
		export SMTP_HOST=$$SMTP_HOST_LOCAL; \
		export SMTP_PORT=$$SMTP_PORT_LOCAL; \
		set +a; \
		go test -v -race -timeout $(TEST_TIMEOUT) -p 1 -run "$(RUN)" ./test/e2e/'
	@$(MAKE) test-teardown

integration-test: test-e2e
	@echo "Integration tests completed!"
