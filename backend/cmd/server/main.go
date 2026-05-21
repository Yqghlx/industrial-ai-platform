package main

import (
	"log"

	"github.com/industrial-ai/platform/internal/config"
	"github.com/industrial-ai/platform/internal/handler"
)

func main() {
	// Load configuration from environment
	cfg := config.MustLoad()

	// Create server config
	serverCfg := handler.ServerConfig{
		Port:        cfg.Port,
		DatabaseURL: cfg.DatabaseURL,
		JWTSecret:   cfg.JWTSecret,
	}

	// Create and start server
	server, err := handler.NewHTTPServerNew(serverCfg)
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := server.Run(cfg.Port); err != nil {
		log.Fatal("Server failed:", err)
	}
}
