package config

// Example usage of the config module:
//
// Basic usage in main.go:
//
//	func main() {
//		// Load and validate configuration (panics on error)
//		cfg := config.MustLoad()
//
//		// Use configuration
//		db, err := sql.Open("postgres", cfg.DatabaseURL)
//		// ...
//	}
//
// Alternative usage with error handling:
//
//	func main() {
//		cfg, err := config.LoadAndValidate()
//		if err != nil {
//			log.Fatalf("Configuration error: %v", err)
//		}
//		// Use configuration...
//	}
//
// Manual configuration loading:
//
//	func main() {
//		cfg := config.LoadFromEnv()
//		cfg.Port = "9090" // Override
//		if err := cfg.Validate(); err != nil {
//			log.Fatal(err)
//		}
//		// Use configuration...
//	}
//
// Checking environment:
//
//	func main() {
//		cfg := config.MustLoad()
//
//		if cfg.IsProduction() {
//			// Production-specific setup
//			gin.SetMode(gin.ReleaseMode)
//		}
//
//		if cfg.IsDevelopment() {
//			// Development-specific setup
//			log.Println("Running in development mode")
//		}
//	}
