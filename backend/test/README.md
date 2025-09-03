# E2E Testing Architecture

This directory contains the comprehensive end-to-end testing architecture for the Bless2n Food System backend.

## Architecture Overview

**Key Principle**: Use real services with minimal mocking for maximum confidence.

The testing architecture follows Go best practices and includes:

- **Real service integration** with only external dependencies mocked (EmailService)
- **Complete test isolation** with dedicated test database and services
- **Docker-based test infrastructure** for consistent testing environments
- **Comprehensive auth flow testing** using actual OTP generation and verification
- **Minimal custom logic** - tests use real HTTP handlers and business logic
- **Mock verification** - tests can verify emails were sent without actual SMTP
- **Coverage reporting** with 80% threshold enforcement
- **CI/CD integration** with GitHub Actions

## Directory Structure

```
test/
├── README.md                     # This documentation
├── Makefile                      # Test commands and automation
├── docker-compose.test.yml       # Test infrastructure (MongoDB, SMTP)
├── config/
│   └── test.env                  # Test environment configuration
├── keys/                         # Generated JWT test keys
│   ├── ed25519.pem              # Private key for JWT signing
│   └── ed25519.pub.pem          # Public key for JWT verification
├── fixtures/
│   └── test_data.go             # Test data fixtures and payloads
├── helpers/
│   ├── test_server.go           # Real server with mocked EmailService
│   ├── http_client.go           # HTTP client for making test requests
│   └── database.go              # Database helper for seeding/cleaning
├── mocks/
│   ├── email_service.go         # Mock EmailService implementation
│   └── email_service_test.go    # Tests for mock service
└── e2e/
    └── auth_test.go             # Real auth flow e2e tests
```

## Quick Start

### 1. Run All Tests
```bash
# From project root
make test-coverage
```

### 2. Run Only E2E Tests
```bash
make test-e2e
```

### 3. Run Only Unit Tests
```bash
make test
```

## Test Infrastructure

### Services

The test infrastructure includes:

- **MongoDB Test Instance**: Port 27018, isolated test database
- **Mailpit SMTP Server**: Port 1026 (SMTP), 8026 (Web UI) - not used in tests, only for manual testing
- **Mock EmailService**: Captures sent emails/OTPs for verification in tests

### Configuration

Test configuration is managed through:
- `test/config/test.env`: Environment variables for testing
- `test/docker-compose.test.yml`: Docker services configuration

### Database Management

The test database is automatically:
- Set up before each test suite
- Cleaned between individual tests (real MongoDB operations)
- Torn down after test completion

### Mocking Strategy

**What's mocked**: Only external dependencies that would be problematic in tests:
- EmailService (to avoid sending real emails)

**What's real**: Everything else uses actual implementations:
- HTTP server and all handlers
- All business logic services (AuthService, OTPService, TokenService)
- Database operations (with test database)
- JWT token generation and validation
- Password hashing and OTP generation

## Auth E2E Test Coverage

The auth test suite covers:

### Happy Path Flows
- ✅ **Complete Registration Flow**: Register → Verify OTP → Get Tokens
- ✅ **Complete Login Flow**: Request OTP → Login → Get Tokens
- ✅ **Token Refresh Flow**: Use Refresh Token → Get New Tokens
- ✅ **Logout Flow**: Logout → Invalidate Refresh Token
- ✅ **OTP Resend**: Request New OTP for existing user
- ✅ **End-to-End Journey**: Register → Verify → Login → Refresh → Logout

### Error Scenarios
- ✅ **Invalid Email Formats**: Malformed email addresses
- ✅ **Missing Required Fields**: Email, OTP, Client ID validation
- ✅ **Invalid OTP Format**: Wrong length, invalid characters
- ✅ **Non-existent Users**: Login attempts for unregistered users
- ✅ **Invalid OTPs**: Real OTP verification with wrong codes
- ✅ **Invalid Refresh Tokens**: Reuse after logout

### Security Testing
- ✅ **JWT Token Validation**: Real JWT signing and verification
- ✅ **Client ID Validation**: Required for all auth operations
- ✅ **OTP Security**: Real OTP generation, hashing, and single-use verification
- ✅ **Password Hashing**: Real Argon2id hashing for OTP storage
- ✅ **Rate Limiting**: OTP attempt limits and usage tracking

## Coverage Requirements

- **Target Coverage**: 80% minimum
- **Coverage Scope**: All packages under `./...`
- **Coverage Enforcement**: Automated in CI/CD pipeline
- **Coverage Reports**: Generated in HTML format (`coverage.html`)

## Running Tests

### Prerequisites

1. **Docker**: For test infrastructure
2. **Go 1.24+**: For running tests
3. **Make**: For test automation

### Local Development

```bash
# Setup test infrastructure
make test-setup

# Run e2e tests (manual)
ENV_FILE=test/config/test.env go test -v ./test/e2e/...

# Teardown test infrastructure
make test-teardown
```

### Available Commands

From project root:
```bash
make test                   # Unit tests only
make test-e2e              # E2E tests only
make test-coverage         # All tests with coverage
make test-setup            # Start test infrastructure
make test-teardown         # Stop test infrastructure
```

From test directory:
```bash
cd test/
make help                  # Show all test commands
make test-all              # Run all tests
make test-coverage         # Generate coverage report
```

## CI/CD Integration

Tests run automatically on:
- **Push to main/develop branches**
- **Pull requests to main/develop**

The CI pipeline:
1. Sets up test infrastructure (MongoDB, SMTP)
2. Runs unit tests
3. Runs E2E tests with coverage
4. Validates 80% coverage threshold
5. Uploads coverage to Codecov

## Extending Tests

### Adding New E2E Tests

1. **Create test file** in `test/e2e/` (e.g., `product_test.go`)
2. **Follow the pattern** from `auth_test.go`:
   - Use testify/suite for test organization
   - Set up/tear down database state
   - Use test helpers for common operations
3. **Add test data** to `test/fixtures/test_data.go`
4. **Update documentation** with new test coverage

### Adding Test Helpers

Add reusable test utilities to:
- `test/helpers/http_client.go`: HTTP request helpers
- `test/helpers/database.go`: Database seeding/cleaning
- `test/helpers/test_server.go`: Server management

### Adding Test Data

Add fixtures to `test/fixtures/test_data.go`:
- Valid request payloads
- Invalid request variations
- Domain objects for seeding

## Best Practices

1. **Test Isolation**: Each test should be independent
2. **Database Cleanup**: Always clean state between tests
3. **Meaningful Assertions**: Use descriptive error messages
4. **Test Naming**: Use descriptive test names that explain the scenario
5. **Coverage Goals**: Aim for high coverage with meaningful tests
6. **Performance**: Keep tests fast and reliable

## Troubleshooting

### Common Issues

**Tests hanging or timing out**:
- Check if test infrastructure is running (`make test-setup`)
- Verify ports are not in use (27018, 1026, 8026)

**Database connection errors**:
- Ensure MongoDB test container is healthy
- Check test environment configuration

**Coverage threshold failures**:
- Add tests for uncovered code paths
- Review coverage report (`coverage.html`)

### Debug Mode

Enable verbose test output:
```bash
go test -v -race ./test/e2e/... -timeout 300s
```

View test infrastructure logs:
```bash
docker compose -f test/docker-compose.test.yml logs -f
```

## Monitoring

- **Coverage Reports**: Generated after each test run
- **Test Metrics**: Duration, pass/fail rates
- **Infrastructure Health**: Docker service status
- **CI/CD Status**: GitHub Actions workflow results

This testing architecture ensures reliable, comprehensive testing of the authentication system and provides a solid foundation for testing additional features as the application grows.