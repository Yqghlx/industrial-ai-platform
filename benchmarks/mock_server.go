package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Mock server for k6 performance testing
// Provides minimal endpoints without external dependencies

var (
	requestCounter sync.Map
	startTime      = time.Now()
)

// JWT Claims
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// JWT secret for mock
	jwtSecret := []byte("mock-secret-key-for-k6-testing-minimum-32-characters")

	// Public endpoints
	r.GET("/health", healthCheck)

	// Auth endpoints
	auth := r.Group("/api/v1/auth")
	auth.POST("/login", mockLogin(jwtSecret))

	// Protected endpoints
	api := r.Group("/api/v1")
	api.Use(mockAuthMiddleware(jwtSecret))
	api.GET("/devices", listDevices)
	api.GET("/devices/:id", getDevice)
	api.POST("/devices/telemetry", submitTelemetry)
	api.GET("/devices/latest", getLatestTelemetry)
	api.GET("/rules", listRules)
	api.POST("/agent/query", agentQuery)
	api.GET("/roi/stats", roiStats)

	log.Println("🚀 Mock server starting on :8080")
	log.Println("📊 Ready for k6 performance testing")
	r.Run(":8080")
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":   "healthy",
		"uptime":   time.Since(startTime).Seconds(),
		"version":  "mock-1.0.0",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func mockLogin(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		c.BindJSON(&req)

		// Mock: accept any credentials
		claims := Claims{
			UserID:   1,
			Username: req.Username,
			Role:     "admin",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(secret)

		c.JSON(200, gin.H{
			"code": 0,
			"message": "success",
			"data": gin.H{
				"token": tokenString,
				"expires_in": 900,
			},
		})
	}
}

func mockAuthMiddleware(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Allow public telemetry endpoint
			if c.Request.URL.Path == "/api/v1/devices/telemetry" {
				c.Next()
				return
			}
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:] // Remove "Bearer "
		_, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})

		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func listDevices(c *gin.Context) {
	// Mock device list
	devices := []gin.H{
		{"id": "device-001", "name": "CNC Machine 1", "type": "CNC", "status": "running"},
		{"id": "device-002", "name": "CNC Machine 2", "type": "CNC", "status": "idle"},
		{"id": "device-003", "name": "Robot Arm 1", "type": "Robot", "status": "running"},
	}
	c.JSON(200, gin.H{"code": 0, "data": devices, "total": 3})
}

func getDevice(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"id":     id,
			"name":   "Device " + id,
			"type":   "CNC",
			"status": "running",
			"metrics": gin.H{
				"temperature": 75.5,
				"vibration":   2.8,
			},
		},
	})
}

func submitTelemetry(c *gin.Context) {
	// Accept any telemetry
	var req map[string]interface{}
	c.BindJSON(&req)
	c.JSON(200, gin.H{"code": 0, "message": "telemetry received"})
}

func getLatestTelemetry(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 0,
		"data": []gin.H{
			{"device_id": "device-001", "temperature": 75.5, "timestamp": time.Now().Format(time.RFC3339)},
			{"device_id": "device-002", "temperature": 68.2, "timestamp": time.Now().Format(time.RFC3339)},
		},
	})
}

func listRules(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 0,
		"data": []gin.H{
			{"id": 1, "name": "High Temperature Alert", "condition": "temperature > 80"},
			{"id": 2, "name": "High Vibration Alert", "condition": "vibration > 4"},
		},
	})
}

func agentQuery(c *gin.Context) {
	// Mock AI response
	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"session_id": "session-mock",
			"response":   "建议检查设备温度传感器，必要时进行冷却系统维护。",
			"agent":      "maintenance-agent",
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	})
}

func roiStats(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"total_savings":       125000,
			"maintenance_reduction": 35,
			"uptime_improvement":    12.5,
			"energy_savings":        8500,
		},
	})
}