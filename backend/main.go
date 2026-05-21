package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/industrial-ai/platform/internal/config"
	"github.com/industrial-ai/platform/internal/handler"
	"github.com/industrial-ai/platform/internal/middleware" // FIX-001: 用于 SetJWTSecret
	"github.com/industrial-ai/platform/internal/service"    // FIX-001: 用于 SetJWTSecret
)

func main() {
	// Load and validate configuration
	appCfg := config.MustLoad()

	// Print any configuration warnings
	appCfg.PrintWarnings()

	// FIX-001: 初始化 JWT 密钥 - 强制要求设置
	if appCfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required. Please set the JWT_SECRET environment variable with at least 32 characters.")
	}
	if len(appCfg.JWTSecret) < 32 {
		log.Fatalf("JWT_SECRET must be at least 32 characters for security. Current length: %d", len(appCfg.JWTSecret))
	}
	// 设置 JWT 密钥到 middleware 和 service
	middleware.SetJWTSecret(appCfg.JWTSecret)
	service.SetJWTSecret(appCfg.JWTSecret)

	// Create server configuration
	serverCfg := handler.ServerConfig{
		DatabaseURL:          appCfg.DatabaseURL,
		Port:                 appCfg.Port,
		JWTSecret:            appCfg.JWTSecret,
		CORSOrigins:          appCfg.CORSOrigins,
		AdminPassword:        appCfg.AdminPassword,
		RedisURL:             appCfg.RedisURL,
		CacheEnabled:         appCfg.CacheEnabled,
		CachePrefix:          appCfg.CachePrefix,
		Environment:          appCfg.Environment, // FIX-016: Pass environment for WebSocket security
		WSCompressionEnabled: appCfg.WSCompressionEnabled,
		WSCompressionLevel:   appCfg.WSCompressionLevel,
		WSCompressionMinSize: appCfg.WSCompressionMinSize,
	}

	// Create server (using new architecture)
	server, err := handler.NewHTTPServerNew(serverCfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// FIX-008: 正确使用 context 控制 graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %s", serverCfg.Port)
		if err := server.Run(serverCfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
		// Server stopped, signal main thread
		cancel()
	}()

	log.Printf("Server running on http://localhost:%s", serverCfg.Port)
	log.Printf("Health check: http://localhost:%s/health", serverCfg.Port)
	log.Printf("API docs: http://localhost:%s/docs/", serverCfg.Port)
	log.Printf("WebSocket: ws://localhost:%s/ws", serverCfg.Port)
	log.Println("Press Ctrl+C to stop")

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either signal or server error
	select {
	case <-sigCh:
		log.Println("Shutdown signal received...")
	case <-ctx.Done():
		log.Println("Server stopped unexpectedly...")
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// FIX: 清理 middleware 中的后台 goroutine，防止 goroutine 泄漏
	middleware.CleanupMiddleware()

	if err := server.Close(); err != nil {
		log.Printf("Error closing server: %v", err)
	}

	// Wait for shutdown completion or timeout
	select {
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout reached")
	case <-time.After(100 * time.Millisecond):
		log.Println("Server shutdown complete")
	}
}

func init() {
	// Print startup banner
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║       Industrial AI Agent Platform - Backend              ║")
	fmt.Println("║       Version 1.0.0                                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
}
