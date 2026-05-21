// Package middleware - DEPRECATED FILE
// This file has been merged into cors.go
// All CORS-related functionality has been consolidated in:
// - backend/internal/middleware/cors.go
//
// This file is kept for backward compatibility documentation.
// Please use the unified CORS middleware from cors.go instead.
//
// Migration guide:
// - CORSConfig -> cors.go CORSConfig (same structure)
// - DefaultCORSConfig() -> cors.go DefaultCORSConfig()
// - CORSSecurityWithConfig() -> cors.go CORSWithConfig()
// - isOriginAllowed() -> cors.go isOriginAllowed()
// - ParseOrigins() -> cors.go ParseOrigins()
// - itoa() -> use strconv.Itoa instead (FIX-055)
//
// For CSRF protection, see csrf.go file.
package middleware
