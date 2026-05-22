# Configuration Module

This module handles loading and validating environment variables for the Industrial AI Platform backend.

## Features

- **Type-safe configuration**: All configuration values are strongly typed
- **Validation**: Required fields are validated on startup with clear error messages
- **Environment-aware**: Different validation rules for development vs production
- **Warning system**: Non-fatal configuration issues are reported as warnings
- **Default values**: Sensible defaults for optional fields

## Required Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | **Yes** |
| `JWT_SECRET` | Secret key for JWT signing | Production only |

## Optional Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `LLM_API_KEY` | LLM API key for AI features | - |
| `LLM_BASE_URL` | LLM API base URL | - |
| `LLM_MODEL` | LLM model name | - |
|| `CORS_ORIGINS` | Comma-separated allowed origins | Empty (must be set in production) |
| `ADMIN_PASSWORD` | Admin user password | Randomly generated |
| `ENV` | Environment (development/production) | `development` |

## Usage

### Basic Usage (Recommended)

```go
package main

import (
    "github.com/industrial-ai/platform/internal/config"
)

func main() {
    // Load and validate configuration
    // Panics on validation error with detailed message
    cfg := config.MustLoad()
    
    // Print any configuration warnings
    cfg.PrintWarnings()
    
    // Use configuration
    fmt.Printf("Starting server on port %s\n", cfg.Port)
}
```

### With Error Handling

```go
func main() {
    cfg, err := config.LoadAndValidate()
    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
    // Use cfg...
}
```

### Manual Configuration

```go
func main() {
    cfg := config.LoadFromEnv()
    
    // Override values
    cfg.Port = "9090"
    
    // Validate manually
    if err := cfg.Validate(); err != nil {
        log.Fatal(err)
    }
    // Use cfg...
}
```

## Environment-Specific Behavior

### Development Environment

- `JWT_SECRET` is optional (will show warning)
- Default environment is development
- Warnings are printed for missing optional fields

### Production Environment

Set `ENV=production` to enable production mode:

- `JWT_SECRET` is **required** (validation fails if missing)
- Stricter validation rules apply

## Validation Errors

When required configuration is missing, the module provides clear error messages:

```
Configuration Error:
==================
  - DATABASE_URL: is required. Please set the DATABASE_URL environment variable 
    with your PostgreSQL connection string.
  - JWT_SECRET: is required in production environment. Please set the JWT_SECRET 
    environment variable.

Please check your environment variables and try again.
Refer to .env.example for required configuration.
```

## Configuration Warnings

Non-fatal issues are reported as warnings:

```
Configuration Warnings:
=======================
  ⚠ JWT_SECRET is not set. Using default secret (not recommended for production).
  ⚠ ADMIN_PASSWORD is not set. A random password will be generated on first startup.
  ⚠ LLM_API_KEY is not set. AI agent features may not work properly.
  ⚠ CORS_ORIGINS is not set or set to '*'. Consider restricting allowed origins in production.
```

## Helper Methods

```go
cfg.IsProduction()    // Returns true if ENV=production
cfg.IsDevelopment()   // Returns true if ENV=development or empty
cfg.GetCORSOrigins()  // Returns CORS origins as a slice
cfg.GetWarnings()     // Returns list of warning messages
```

## Testing

Run tests with:

```bash
go test ./internal/config/...
```

## Example .env File

See `.env.example` in the backend directory for a complete example:

```bash
# Required
DATABASE_URL=postgres://user:pass@localhost:5432/industrial_ai?sslmode=disable
JWT_SECRET=your-secret-key-here

# Optional
PORT=8080
CORS_ORIGINS=http://localhost:3000,http://localhost:8080
ADMIN_PASSWORD=your-admin-password
LLM_API_KEY=your-llm-api-key
LLM_BASE_URL=https://api.example.com/v1
LLM_MODEL=gpt-4
ENV=development
```