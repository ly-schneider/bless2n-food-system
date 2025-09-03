package helpers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"backend/internal/app"
	"backend/internal/config"
	"backend/internal/service"
	"backend/test/mocks"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// TestServer represents a test HTTP server instance
type TestServer struct {
    Server      *http.Server
    Port        int
    BaseURL     string
    Config      config.Config
    Logger      *zap.Logger
    EmailMock   *mocks.MockEmailService
    fxApp       *fx.App
    stopFunc    func()
}

// NewTestServer creates a new test server instance
func NewTestServer() (*TestServer, error) {
	// Load test environment variables - detect working directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

    var testEnvPath string
    if filepath.Base(wd) == "e2e" {
        // Running from test/e2e directory
        testEnvPath = "../config/.env.test"
    } else if filepath.Base(wd) == "test" {
        // Running from test directory
        testEnvPath = "config/.env.test"
    } else {
        // Running from project root or other directory
        testEnvPath = filepath.Join("test", "config", ".env.test")
    }

	if err := godotenv.Load(testEnvPath); err != nil {
		return nil, fmt.Errorf("failed to load test environment from %s (wd: %s): %w", testEnvPath, wd, err)
	}

	// Generate test keys if they don't exist
	if err := generateTestKeys(); err != nil {
		return nil, fmt.Errorf("failed to generate test keys: %w", err)
	}

	// Find available port
	port, err := getFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to get free port: %w", err)
	}

	// Override port in environment
	os.Setenv("APP_PORT", fmt.Sprintf("%d", port))

	// Create mock email service
	emailMock := mocks.NewMockEmailService()

	// Create FX application with real dependencies but mocked email
    var cfg config.Config
    var logger *zap.Logger
    var router http.Handler

    fxApp := fx.New(
        // Provide all dependencies needed for the application
        fx.Provide(
            app.NewConfig,
            app.ProvideAppConfig,
            app.NewLogger,
            app.NewDB,
            app.NewRouter,
        ),
        app.NewRepositories(),
        app.NewServices(),
        // Override EmailService with mock by decorating the provided instance
        fx.Decorate(func(service.EmailService) service.EmailService { return emailMock }),
        app.NewHandlers(),
        fx.Populate(&cfg, &logger, &router),
        fx.NopLogger, // Disable FX logs during testing
    )

	// Create HTTP server manually
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	// Create test server instance
    ts := &TestServer{
        Server:    server,
        Port:      port,
        BaseURL:   fmt.Sprintf("http://localhost:%d", port),
        Config:    cfg,
        Logger:    logger,
        EmailMock: emailMock,
        fxApp:     fxApp,
    }

	return ts, nil
}

// Start starts the test server
func (ts *TestServer) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Start the FX application
	if err := ts.fxApp.Start(ctx); err != nil {
		return fmt.Errorf("failed to start test application: %w", err)
	}

	// Start HTTP server in goroutine
	go func() {
		if err := ts.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ts.Logger.Error("test server failed to start", zap.Error(err))
		}
	}()

	// Wait for server to be ready
	return ts.waitForServer()
}

// Stop stops the test server and cleans up resources
func (ts *TestServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := ts.Server.Shutdown(ctx); err != nil {
		ts.Logger.Error("failed to shutdown test server", zap.Error(err))
	}

	// Stop FX application
	if err := ts.fxApp.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop test application: %w", err)
	}

	return nil
}

// waitForServer waits for the server to be ready to accept connections
func (ts *TestServer) waitForServer() error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("test server failed to start within timeout")
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", ts.Port), time.Second)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}

// getFreePort returns an available port on the local machine
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}

// generateTestKeys creates test JWT keys if they don't exist
func generateTestKeys() error {
	// Use the secrets directory based on working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

    // Prefer explicit paths from environment if provided
    privKeyPath := os.Getenv("JWT_PRIV_PEM_PATH")
    pubKeyPath := os.Getenv("JWT_PUB_PEM_PATH")

    if privKeyPath == "" || pubKeyPath == "" {
        var keyDir string
        switch filepath.Base(wd) {
        case "e2e":
            // Running from test/e2e directory
            keyDir = filepath.Join("..", "config", "secrets", "test")
        case "test":
            // Running from test directory
            keyDir = filepath.Join("config", "secrets", "test")
        default:
            // Running from project root or elsewhere
            keyDir = filepath.Join("test", "config", "secrets", "test")
        }
        privKeyPath = filepath.Join(keyDir, "ed25519.pem")
        pubKeyPath = filepath.Join(keyDir, "ed25519.pub.pem")
    }

    // Ensure directory exists for private key path
    if err := os.MkdirAll(filepath.Dir(privKeyPath), 0755); err != nil {
        return fmt.Errorf("failed to create key directory: %w", err)
    }

	// Check if keys already exist
    if _, err := os.Stat(privKeyPath); err == nil {
        if _, err := os.Stat(pubKeyPath); err == nil {
            return nil // Keys already exist
        }
    }

	// Generate test keys (simplified version for testing)
	// In a real implementation, you'd use crypto/ed25519 to generate proper keys
	privKey := `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIGWGZKg7/WBDrYdnN6p4ZhGPbWJbGK4OpmKgB3gStGzS
-----END PRIVATE KEY-----`

	pubKey := `-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEATtJdNZLNrztSpLaB4lBFqFx+E1fIJ8TQZ4QZZh6gVMA=
-----END PUBLIC KEY-----`

	if err := os.WriteFile(privKeyPath, []byte(privKey), 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	if err := os.WriteFile(pubKeyPath, []byte(pubKey), 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}
