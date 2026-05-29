package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Mock server for k6 performance testing - minimal implementation

var startTime = time.Now()

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

	jwtSecret := []byte("mock-secret-key-for-k6-testing-minimum-32-characters")

	// Public endpoints
	r.GET("/health", healthCheck)

	// Auth
	r.POST("/api/v1/auth/login", mockLogin(jwtSecret))

	// Protected endpoints
	api := r.Group("/api/v1")
	api.Use(mockAuthMiddleware(jwtSecret))
	api.GET("/devices", listDevices)
	api.GET("/devices/:id", getDevice)
	// telemetry is public - no auth needed
	api.GET("/devices/latest", getLatestTelemetry)
	api.GET("/rules", listRules)
	api.POST("/agent/query", agentQuery)
	api.GET("/roi/stats", roiStats)

	// Public telemetry endpoint (no auth)
	r.POST("/api/v1/devices/telemetry", submitTelemetry)

	log.Println("🚀 Mock server for k6 testing starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "healthy", "uptime": time.Since(startTime).Seconds()})
}

func mockLogin(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		_ = c.BindJSON(&req) // Ignore error in mock server

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

		c.JSON(200, gin.H{"code": 0, "data": gin.H{"token": tokenString, "expires_in": 900}})
	}
}

func mockAuthMiddleware(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
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
	c.JSON(200, gin.H{"code": 0, "data": []gin.H{
		{"id": "device-001", "name": "CNC 1", "type": "CNC", "status": "running"},
		{"id": "device-002", "name": "CNC 2", "type": "CNC", "status": "idle"},
	}, "total": 2})
}

func getDevice(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{"code": 0, "data": gin.H{"id": id, "name": "Device " + id, "status": "running"}})
}

func submitTelemetry(c *gin.Context) {
	var req map[string]interface{}
	_ = c.BindJSON(&req) // Ignore error in mock server
	c.JSON(200, gin.H{"code": 0, "message": "received"})
}

func getLatestTelemetry(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "data": []gin.H{
		{"device_id": "device-001", "temperature": 75.5, "timestamp": time.Now().Format(time.RFC3339)},
	}})
}

func listRules(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "data": []gin.H{
		{"id": 1, "name": "High Temp", "condition": "temp > 80"},
	}})
}

func agentQuery(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "data": gin.H{
		"session_id": "mock-session",
		"response":   "建议检查设备温度传感器",
		"agent":      "maintenance-agent",
	}})
}

func roiStats(c *gin.Context) {
	c.JSON(200, gin.H{"code": 0, "data": gin.H{
		"total_savings":      125000,
		"uptime_improvement": 12.5,
	}})
}
