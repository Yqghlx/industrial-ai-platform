package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/industrial-ai/platform/internal/database"
	"github.com/industrial-ai/platform/internal/middleware"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/cache"
	"github.com/industrial-ai/platform/pkg/cache_service"
	"github.com/industrial-ai/platform/pkg/logger"
	"github.com/industrial-ai/platform/pkg/wscompression"
	"go.uber.org/zap"

	_ "github.com/lib/pq" // PostgreSQL driver registration

	// Swagger documentation
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// ============================================
// HTTPServerNew - 新架构 HTTP 服务器（完整版）
// ============================================

// ServerConfig holds server configuration
type ServerConfig struct {
	DatabaseURL          string
	Port                 string
	JWTSecret            string
	CORSOrigins          string
	AdminPassword        string
	RedisURL             string
	CacheEnabled         bool
	CachePrefix          string
	Environment          string
	WSCompressionEnabled bool
	WSCompressionLevel   int
	WSCompressionMinSize int
	// AllowedOrigins specifies allowed origins for WebSocket connections
	// In production: must be explicitly configured
	// In development: defaults to localhost if not set
	AllowedOrigins []string

	// 数据库连接池配置（对应 config.Config 中的 DB 字段）
	DBMaxOpenConns    int           // 数据库最大打开连接数（默认 25）
	DBMaxIdleConns    int           // 数据库最大空闲连接数（默认 10）
	DBConnMaxLifetime time.Duration // 数据库连接最大生命周期（默认 15 分钟）
	DBConnMaxIdleTime time.Duration // 数据库空闲连接最大存活时间（默认 10 分钟）

	// 全局限流配置（对应 config.Config 中的 RateLimit 字段）
	RateLimitCapacity   int     // 全局限流桶容量（默认 60）
	RateLimitRefillRate float64 // 全局限流令牌补充速率，每秒补充令牌数（默认 1）
}

// Server is alias for HTTPServerNew for backward compatibility
type Server = HTTPServerNew

// NewServer creates a new server (alias for NewHTTPServerNew)
func NewServer(cfg ServerConfig) (*Server, error) {
	return NewHTTPServerNew(cfg)
}

// HTTPServerNew HTTP服务器（新架构）
type HTTPServerNew struct {
	db            *sql.DB
	router        *gin.Engine
	wsUpgrader    websocket.Upgrader
	broadcastFn   func(msg model.WSMessage)
	jwtSecret     string
	adminPassword string
	startTime     time.Time

	// Services
	authSvc      service.AuthServiceInterface
	userSvc      service.UserServiceInterface
	deviceSvc    service.DeviceServiceInterface
	alertSvc     service.AlertServiceInterface
	telemetrySvc service.TelemetryServiceInterface
	tenantSvc    service.TenantServiceInterface
	rbacSvc      service.RBACServiceInterface
	exportSvc    service.ExportServiceInterface
	reportSvc    service.ReportServiceInterface
	workOrderSvc      service.WorkOrderServiceInterface
	notificationSvc  service.NotificationServiceInterface
	blackBoxSvc      service.BlackBoxServiceInterface
	cacheSvc         *cache_service.CacheServiceIntegration
	agentSvc         service.AgentServiceInterface

	// Handlers (new architecture)
	alertHandler     *AlertHandler
	deviceHandler    *DeviceHandlerNew
	businessHandler  *BusinessHandlerNew
	telemetryHandler *TelemetryHandlerNew
	authHandler      *AuthHandlerNew
	tenantHandler    *TenantHandler
	rbacHandler      *RBACHandler
	adminHandler     *AdminHandlerNew
	healthHandler    *HealthHandlerNew
	exportHandler    *ExportHandler

	// WebSocket
	wsClients     map[*websocket.Conn]bool
	wsClientsMu   sync.RWMutex
	broadcastChan chan model.WSMessage
	heartbeatChan chan struct{}
	wsCompressor  *wscompression.Compressor
	stopTicker    chan struct{} // P0: Stop channel for heartbeat ticker goroutine

	// Cache
	cache cache.CacheService

	// 限流配置
	rateLimitCapacity   int
	rateLimitRefillRate float64
}

// NewHTTPServerNew creates a new HTTP server (new architecture)
// 使用 ServerConfig 定义从 server.go
func NewHTTPServerNew(cfg ServerConfig) (*HTTPServerNew, error) {
	// SEC-HIGH-01: 设置 JWT 密钥到 middleware 和 service
	if cfg.JWTSecret != "" {
		middleware.SetJWTSecret(cfg.JWTSecret)
		if err := service.SetJWTSecret(cfg.JWTSecret); err != nil {
			return nil, fmt.Errorf("failed to initialize JWT: %w", err)
		}
	}

	// Connect to database
	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		logger.L().Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// 配置数据库连接池 - 使用 ServerConfig 中的值，提供安全的默认值
	// 默认值：MaxOpen=25, MaxIdle=10, Lifetime=15min, IdleTime=10min
	maxOpenConns := cfg.DBMaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 25
	}
	maxIdleConns := cfg.DBMaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 10
	}
	connMaxLifetime := cfg.DBConnMaxLifetime
	if connMaxLifetime <= 0 {
		connMaxLifetime = 15 * time.Minute
	}
	connMaxIdleTime := cfg.DBConnMaxIdleTime
	if connMaxIdleTime <= 0 {
		connMaxIdleTime = 10 * time.Minute
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	logger.L().Info("[DB] 连接池配置",
		zap.Int("maxOpenConns", maxOpenConns),
		zap.Int("maxIdleConns", maxIdleConns),
		zap.Duration("connMaxLifetime", connMaxLifetime),
		zap.Duration("connMaxIdleTime", connMaxIdleTime),
	)

	// Initialize cache
	cacheConfig := &cache.Config{
		RedisURL:      cfg.RedisURL,
		Enabled:       cfg.CacheEnabled,
		MaxMemorySize: 100 * 1024 * 1024,
		Prefix:        cfg.CachePrefix,
	}
	if cacheConfig.Prefix == "" {
		cacheConfig.Prefix = "iai:"
	}
	cacheSvc := cache_service.NewCacheServiceIntegration(cacheConfig)
	logger.L().Info("[Cache] Initialized", zap.String("backend", cacheSvc.GetCache().GetStats().BackendType))

	// Phase 3: Handler 层完全通过 ServiceFactory 获取服务，不直接依赖 Repository
	serviceFactory := service.NewServiceFactoryFromDB(db)
	serviceFactory.InitializeAgentService(cacheSvc.GetCache())

	// 从 ServiceFactory 获取服务实例（不再直接引用 Repository）
	authSvc := serviceFactory.GetAuthService()
	userSvc := serviceFactory.GetUserService()
	alertSvc := serviceFactory.GetAlertService()
	deviceSvc := serviceFactory.GetDeviceService()
	telemetrySvc := serviceFactory.GetTelemetryService()
	tenantSvc := serviceFactory.GetTenantService()
	rbacSvc := serviceFactory.GetRBACService()
	exportSvc := serviceFactory.GetExportService()
	reportSvc := serviceFactory.GetReportService()
	workOrderSvc := serviceFactory.GetWorkOrderService()
	notificationSvc := serviceFactory.GetNotificationService()
	blackBoxSvc := serviceFactory.GetBlackBoxService()
	agentSvc := serviceFactory.GetAgentService()

	// Initialize Gin router
	router := gin.New()

	// Parse CORS origins
	corsOrigins := []string{}
	if cfg.CORSOrigins != "" {
		corsOrigins = strings.Split(cfg.CORSOrigins, ",")
	}

	// WebSocket upgrader with secure origin checking
	// FIX-015: WebSocket Origin environment-based restriction
	isProduction := strings.ToLower(cfg.Environment) == "production"

	// Build allowed origins map for WebSocket
	wsAllowedOrigins := make(map[string]bool)

	// Add configured allowed origins
	for _, o := range cfg.AllowedOrigins {
		wsAllowedOrigins[strings.TrimSpace(o)] = true
	}

	// Also add CORS origins as allowed for WebSocket
	for _, o := range corsOrigins {
		wsAllowedOrigins[strings.TrimSpace(o)] = true
	}

	// Development environment: add default localhost origins
	if !isProduction {
		// 支持通过 WS_DEV_ORIGINS 环境变量自定义开发环境的 WebSocket 源
		// 格式: WS_DEV_ORIGINS=http://localhost:3000,http://localhost:8080
		if devOrigins := os.Getenv("WS_DEV_ORIGINS"); devOrigins != "" {
			for _, o := range strings.Split(devOrigins, ",") {
				o = strings.TrimSpace(o)
				if o != "" {
					wsAllowedOrigins[o] = true
				}
			}
		} else {
			// 默认开发环境源
			wsAllowedOrigins["http://localhost"] = true
			wsAllowedOrigins["http://localhost:3000"] = true
			wsAllowedOrigins["http://localhost:8080"] = true
			wsAllowedOrigins["http://localhost:5173"] = true
			wsAllowedOrigins["http://127.0.0.1"] = true
			wsAllowedOrigins["http://127.0.0.1:3000"] = true
			wsAllowedOrigins["http://127.0.0.1:8080"] = true
			wsAllowedOrigins["http://127.0.0.1:5173"] = true
		}
	}

	wsUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			// No origin header (e.g., same-origin requests, mobile apps, curl)
			// Allow in development, deny in production unless explicitly configured
			if origin == "" {
				if !isProduction {
					return true
				}
				return false
			}

			// Check against allowed origins map
			if wsAllowedOrigins[origin] {
				return true
			}

			// Check for wildcard (not recommended for production)
			if wsAllowedOrigins["*"] {
				return true
			}

			// Development: allow localhost origins even if not explicitly configured
			if !isProduction {
				return isLocalhostOrigin(origin)
			}

			return false
		},
	}

	// WebSocket compressor
	wsCompressorConfig := &wscompression.CompressionConfig{
		Enabled: cfg.WSCompressionEnabled,
		Level:   cfg.WSCompressionLevel,
		MinSize: cfg.WSCompressionMinSize,
	}
	if wsCompressorConfig.MinSize == 0 {
		wsCompressorConfig.MinSize = 1024
	}
	wsCompressor := wscompression.NewCompressor(wsCompressorConfig)

	// 创建 Server 实例（broadcastChan 在此处初始化）
	s := &HTTPServerNew{
		db:            db,
		router:        router,
		wsUpgrader:    wsUpgrader,
		broadcastFn:   nil, // 后面设置，因为依赖 s.broadcastChan
		jwtSecret:     cfg.JWTSecret,
		adminPassword: cfg.AdminPassword,
		startTime:     time.Now(),

		authSvc:       authSvc,
		userSvc:       userSvc,
		deviceSvc:     deviceSvc,
		alertSvc:      alertSvc,
		telemetrySvc:  telemetrySvc,
		tenantSvc:     tenantSvc,
		rbacSvc:       rbacSvc,
		exportSvc:     exportSvc,
		reportSvc:     reportSvc,
		workOrderSvc:     workOrderSvc,
		notificationSvc: notificationSvc,
		blackBoxSvc:     blackBoxSvc,
		cacheSvc:        cacheSvc,
		agentSvc:      agentSvc,
		cache:         cacheSvc.GetCache(),
		wsClients:     make(map[*websocket.Conn]bool),
		broadcastChan: make(chan model.WSMessage, 100),
		heartbeatChan: make(chan struct{}),
		wsCompressor:  wsCompressor,
		stopTicker:         make(chan struct{}),
		rateLimitCapacity:  cfg.RateLimitCapacity,
		rateLimitRefillRate: cfg.RateLimitRefillRate,
	}

	// 设置 broadcastFn：将消息发送到 Server 的 broadcastChan
	// 必须在 s 创建之后设置，因为依赖 s.broadcastChan
	s.broadcastFn = func(msg model.WSMessage) {
		select {
		case s.broadcastChan <- msg:
		default:
			logger.L().Warn("[WS] Broadcast channel full, dropping message", zap.String("type", msg.Type))
		}
	}

	// 将 broadcastFn 注入到 service 层，使业务层的广播消息能到达 WebSocket 客户端
	if ts, ok := telemetrySvc.(interface{ SetBroadcastFn(func(model.WSMessage)) }); ok {
		ts.SetBroadcastFn(s.broadcastFn)
	}
	if as, ok := alertSvc.(interface{ SetBroadcastFn(func(model.WSMessage)) }); ok {
		as.SetBroadcastFn(s.broadcastFn)
	}

	// Setup middleware
	s.setupMiddleware(corsOrigins)

	// Setup all handlers
	s.setupHandlers()

	// Initialize database
	s.initDatabase()

	// MINOR-04: WebSocket broadcaster 启动时机明确
	// 在服务器初始化时显式调用 startBroadcaster()，确保 WebSocket 服务可用
	// 这替代了之前在 init() 中隐式启动的做法，更加可控和安全
	s.startBroadcaster()

	// Warmup cache
	cacheSvc.WarmupAsync()

	return s, nil
}

// setupMiddleware sets up middleware
// SEC-MEDIUM-02: 中间件顺序已正确配置
// SEC-MEDIUM-04: 添加全局速率限制
func (s *HTTPServerNew) setupMiddleware(corsOrigins []string) {
	middleware.InitPrometheus()
	s.router.Use(middleware.Logger())
	s.router.Use(middleware.Recovery())
	s.router.Use(middleware.RequestID())
	s.router.Use(middleware.SecurityHeaders())
	s.router.Use(middleware.PrometheusMiddleware())
	s.router.Use(middleware.CORS(corsOrigins))
	// WAF 中间件 - 检测 SQL 注入、XSS、路径遍历、命令注入等攻击
	s.router.Use(middleware.WAFMiddleware(middleware.LoadWAFConfigFromEnv(), logger.L().Logger))
	// SEC-MEDIUM-04: 全局速率限制 - 作为最后一层，避免影响其他中间件
	s.router.Use(middleware.DefaultRateLimit(s.rateLimitCapacity, s.rateLimitRefillRate))
}

// setupHandlers sets up all handlers
func (s *HTTPServerNew) setupHandlers() {
	// Initialize handlers
	s.alertHandler = NewAlertHandler(s.alertSvc, s.broadcastFn)
	s.deviceHandler = NewDeviceHandlerNew(s.deviceSvc, s.alertSvc, s.authSvc, s.telemetrySvc, s.broadcastFn)
	s.businessHandler = NewBusinessHandlerNew(s.workOrderSvc, s.notificationSvc, s.blackBoxSvc, s.reportSvc, s.alertSvc, s.broadcastFn, s.cache)
	s.telemetryHandler = NewTelemetryHandlerNew(s.telemetrySvc, s.agentSvc)
	s.authHandler = NewAuthHandlerNew(s.authSvc, s.userSvc)
	s.tenantHandler = NewTenantHandler(s.tenantSvc)
	s.rbacHandler = NewRBACHandler(NewRBACServiceAdapter(s.rbacSvc))
	s.adminHandler = NewAdminHandlerNew(s.authSvc, s.telemetrySvc)
	s.healthHandler = NewHealthHandlerNew(s.startTime)
	s.exportHandler = NewExportHandler(s.exportSvc)

	// Setup public routes
	s.router.GET("/health", s.healthCheck)

	// Swagger API Documentation
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	s.router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	middleware.SetupPrometheusEndpoint(s.router)

	// SEC-MED-02: Public telemetry endpoint - intentionally public for edge device ingestion
	// See docs/SECURITY_TELEMETRY.md for security justification and measures
	authPublic := s.router.Group("/api/v1")
	authPublic.POST("/auth/login", middleware.LoginRateLimit(), s.authHandler.Login)
	authPublic.POST("/auth/register", middleware.RegisterRateLimit(), s.authHandler.Register)
	authPublic.POST("/auth/refresh", s.authHandler.RefreshToken)
	// SEC-HIGH-02: CSRF Token endpoint for optional additional protection
	// Note: JWT via Authorization header already provides CSRF-safe authentication
	authPublic.GET("/auth/csrf-token", s.authHandler.GetCSRFToken)

	// SEC-MED-02: Telemetry endpoint with rate limiting and input validation
	// SEC-HIGH-02: 添加设备 API Key 认证机制
	// Device authentication is required for telemetry data ingestion
	// Devices must provide valid API key via X-Device-Key header or device_key query parameter
	s.router.POST("/api/v1/devices/telemetry",
		middleware.TelemetryRateLimit(),
		middleware.DeviceAuthRequired(nil), // SEC-HIGH-02: 添加设备认证
		s.telemetryHandler.IngestTelemetry)

	// SEC-MED-01: WebSocket endpoint - public with rate limiting
	// WebSocket authentication is available via ws_auth.go middleware
	// For authenticated WebSocket, use WSAuthRequired middleware before this route
	// See docs/SECURITY_CSRF.md and internal/middleware/ws_auth.go for details
	s.router.GET("/ws", middleware.WebSocketRateLimit(), s.handleWebSocket)

	// Setup authenticated routes
	auth := s.router.Group("/api/v1")
	auth.Use(middleware.AuthRequired(s.jwtSecret))
	{
		// Alerts
		auth.GET("/alerts", s.alertHandler.ListAlerts)
		auth.GET("/alerts/:id", s.alertHandler.GetAlert)
		auth.PUT("/alerts/:id/resolve", s.alertHandler.ResolveAlert)
		auth.PUT("/alerts/:id/acknowledge", s.alertHandler.AcknowledgeAlert)
		auth.GET("/alerts/stats", s.businessHandler.GetAlertStats)
		auth.GET("/alerts/report/trend", s.alertHandler.GetTrend)
		auth.GET("/alerts/report/ranking", s.alertHandler.GetRanking)
		auth.GET("/alerts/report/efficiency", s.alertHandler.GetEfficiency)

		// Devices
		auth.GET("/devices", s.deviceHandler.ListDevices)
		auth.GET("/devices/latest", s.deviceHandler.GetLatestTelemetry)
		auth.GET("/devices/graph", s.deviceHandler.GetDeviceGraph)
		auth.GET("/devices/:id", s.deviceHandler.GetDevice)
		auth.GET("/devices/:id/telemetry", s.deviceHandler.GetDeviceTelemetry)
		auth.GET("/devices/:id/stats", s.deviceHandler.GetDeviceStats)
		auth.POST("/devices", s.deviceHandler.CreateDevice)
		auth.PUT("/devices/:id", s.deviceHandler.UpdateDevice)
		auth.DELETE("/devices/:id", s.deviceHandler.DeleteDevice)

		// Rules
		auth.GET("/rules", s.deviceHandler.ListRules)
		auth.POST("/rules", s.deviceHandler.CreateRule)
		auth.GET("/rules/:id", s.deviceHandler.GetRule)
		auth.PUT("/rules/:id", s.deviceHandler.UpdateRule)
		auth.PUT("/rules/:id/toggle", s.deviceHandler.ToggleRule)

		// Auth
		auth.POST("/auth/logout", s.authHandler.Logout)
		auth.PUT("/auth/password", s.authHandler.ChangePassword)
		auth.GET("/auth/validate", s.authHandler.ValidateToken)

		// Work Orders
		auth.GET("/workorders", s.businessHandler.ListWorkOrders)
		auth.POST("/workorders", s.businessHandler.CreateWorkOrder)
		auth.GET("/workorders/:id", s.businessHandler.GetWorkOrder)
		auth.PUT("/workorders/:id/status", s.businessHandler.UpdateWorkOrderStatus)

		// Notifications
		auth.GET("/notifications", s.businessHandler.ListNotifications)
		auth.POST("/notifications/:id/read", s.businessHandler.MarkNotificationRead)

		// Reports
		auth.GET("/reports", s.businessHandler.ListReports)
		auth.POST("/reports/generate", s.businessHandler.GenerateReport)
		auth.GET("/roi/stats", middleware.ROIStatsRateLimit(), s.businessHandler.GetROIStats)

		// Export
		auth.GET("/reports/devices/export", s.exportHandler.ExportDevices)
		auth.GET("/reports/alerts/export", s.exportHandler.ExportAlerts)
		auth.GET("/reports/roi/export", s.exportHandler.ExportROI)

		// Telemetry
		auth.GET("/telemetry/latest", s.telemetryHandler.GetLatestTelemetry)
		auth.GET("/telemetry/device/:id", s.telemetryHandler.GetDeviceTelemetry)
		auth.GET("/telemetry/status", s.telemetryHandler.GetSystemStatus)
		auth.GET("/ai/status", s.telemetryHandler.GetAIStatus)
		auth.POST("/agent/query", middleware.AgentQueryRateLimit(), s.telemetryHandler.AgentQuery)

		// Blackbox
		auth.GET("/blackbox", s.businessHandler.ListBlackBox)
		auth.GET("/blackbox/:id/data", s.businessHandler.GetBlackBoxData)

		// Tenants
		auth.GET("/tenants/:id", s.tenantHandler.GetTenant)
		auth.PUT("/tenants/:id", s.tenantHandler.UpdateTenant)

		// RBAC
		auth.GET("/roles", s.rbacHandler.ListRoles)
		auth.GET("/roles/:id", s.rbacHandler.GetRole)
		auth.GET("/permissions", s.rbacHandler.ListPermissions)
		auth.GET("/users/:id/roles", s.rbacHandler.GetUserRoles)
	}

	// Admin routes
	admin := s.router.Group("/api/v1")
	admin.Use(middleware.AuthRequired(s.jwtSecret))
	admin.Use(middleware.AdminRequired())
	{
		admin.GET("/admin/users", s.adminHandler.ListUsers)
		admin.POST("/admin/users", s.adminHandler.CreateUser)
		admin.DELETE("/admin/users/:id", s.adminHandler.DeleteUser)
		admin.GET("/system/status", s.adminHandler.GetSystemStatus)

		admin.POST("/tenants", s.tenantHandler.CreateTenant)
		admin.GET("/tenants", s.tenantHandler.ListTenants)
		admin.DELETE("/tenants/:id", s.tenantHandler.DeleteTenant)

		admin.DELETE("/rules/:id", s.deviceHandler.DeleteRule)

		admin.POST("/roles", s.rbacHandler.CreateRole)
		admin.PUT("/roles/:id", s.rbacHandler.UpdateRole)
		admin.DELETE("/roles/:id", s.rbacHandler.DeleteRole)
		admin.POST("/users/:id/roles", s.rbacHandler.AssignRole)
		admin.DELETE("/users/:id/roles/:role_id", s.rbacHandler.RemoveRole)
		admin.POST("/roles/:id/permissions", s.rbacHandler.AssignPermission)
		admin.DELETE("/roles/:id/permissions/:perm_id", s.rbacHandler.RemovePermission)
	}

	// Public monitoring endpoints
	s.router.GET("/cache/status", s.healthHandler.GetCacheStatus)
	s.router.GET("/ws/stats", s.healthHandler.GetWSStats)
}

// initDatabase initializes the database
func (s *HTTPServerNew) initDatabase() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Run migrations
	migrator := database.NewMigrator(s.db)
	if err := migrator.Up(ctx); err != nil {
		logger.L().Error("Database migration failed", zap.Error(err))
	}

	// Initialize default rules
	if s.alertSvc != nil {
		s.alertSvc.InitializeDefaultRules(ctx)
	}

	// Create default admin
	s.createDefaultAdmin(ctx)
}

// createDefaultAdmin 创建默认管理员（委托给 AuthService）
func (s *HTTPServerNew) createDefaultAdmin(ctx context.Context) {
	if err := s.authSvc.EnsureDefaultAdmin(ctx, s.adminPassword); err != nil {
		logger.L().Error("创建默认管理员失败", zap.Error(err))
	}
}

// healthCheck handles health check
func (s *HTTPServerNew) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    int64(time.Since(s.startTime).Seconds()),
	})
}

// Run starts the server
func (s *HTTPServerNew) Run(port string) error {
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}
	return s.router.Run(":" + port)
}

// Close closes the server
func (s *HTTPServerNew) Close() error {
	if s.cacheSvc != nil {
		s.cacheSvc.Close()
	}
	return s.db.Close()
}

// GetRouter returns the router
func (s *HTTPServerNew) GetRouter() *gin.Engine {
	return s.router
}

// ============================================
// Helper Functions
// ============================================

// getRequestContext creates a context with timeout
func getRequestContext(c *gin.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	return ctx, cancel
}

// Pagination holds pagination params (defined in validation.go)
// type Pagination struct {
// 	Page     int
// 	PageSize int
// }


// ============================================
// Backward Compatibility Methods
// ============================================

// NewAuthHandler creates old-style auth handler (backward compat)
func NewAuthHandler(userSvc service.UserServiceInterface, jwtSecret string) *AuthHandlerNew {
	return NewAuthHandlerNew(&compatAuthSvc{userSvc: userSvc}, userSvc)
}

// compatAuthSvc wraps UserServiceInterface to implement AuthServiceInterface
// 用于向后兼容旧的 API 签名
//
// 注意：这是一个部分实现，仅用于兼容性目的：
// - Register: 返回 nil（不支持直接注册）
// - RefreshToken/ValidateToken: 返回错误（不支持 JWT 功能）
// - 这些方法不应在生产环境中使用，仅用于过渡期兼容
type compatAuthSvc struct {
	userSvc service.UserServiceInterface
}

func (c *compatAuthSvc) Login(ctx context.Context, username, password string) (*model.User, string, error) {
	user, err := c.userSvc.Authenticate(username, password)
	if err != nil {
		return nil, "", err
	}
	return user, "token", nil
}

func (c *compatAuthSvc) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, string, error) {
	return nil, "", nil
}

func (c *compatAuthSvc) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	return c.userSvc.GetByID(id)
}

// FIX-016/017: 实现新增的 AuthServiceInterface 方法
// compatAuthSvc 不支持完整的 JWT 功能
// 注意：调用此方法将返回错误，生产环境应使用完整的 AuthService 实现
func (c *compatAuthSvc) RefreshToken(ctx context.Context, refreshToken string) (*service.TokenPair, error) {
	return nil, fmt.Errorf("refresh token not supported in compat mode")
}

func (c *compatAuthSvc) ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	// 验证旧密码
	user, err := c.userSvc.Authenticate("", oldPassword)
	if err != nil {
		return err
	}
	_ = user // 验证通过，忽略用户信息

	// Hash new password
	newHash, err := service.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	return c.userSvc.UpdatePassword(userID, newHash)
}

func (c *compatAuthSvc) ValidateToken(ctx context.Context, token string) (*service.Claims, error) {
	// 注意：compatAuthSvc 不支持完整的 JWT 功能
	// 生产环境应使用完整的 AuthService 实现
	return nil, fmt.Errorf("validate token not supported in compat mode")
}

// ListUsers 用户列表 - compat模式不支持
func (c *compatAuthSvc) ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int, error) {
	return nil, 0, fmt.Errorf("list users not supported in compat mode")
}

// DeleteUser 删除用户 - compat模式不支持
// SEC-HIGH-03: 新增删除用户方法
func (c *compatAuthSvc) DeleteUser(ctx context.Context, userID int) error {
	return fmt.Errorf("delete user not supported in compat mode")
}

// EnsureDefaultAdmin - compat模式不支持
func (c *compatAuthSvc) EnsureDefaultAdmin(ctx context.Context, password string) error {
	return fmt.Errorf("ensure default admin not supported in compat mode")
}

// getCacheStatus wrapper for backward compat
func (s *HTTPServerNew) getCacheStatus(c *gin.Context) {
	s.healthHandler.GetCacheStatus(c)
}

// exportDevices wrapper for backward compat
func (s *HTTPServerNew) exportDevices(c *gin.Context) {
	s.exportHandler.ExportDevices(c)
}

// exportAlerts wrapper for backward compat
func (s *HTTPServerNew) exportAlerts(c *gin.Context) {
	s.exportHandler.ExportAlerts(c)
}

// exportROI wrapper for backward compat
func (s *HTTPServerNew) exportROI(c *gin.Context) {
	s.exportHandler.ExportROI(c)
}

// Stop gracefully stops the server and cleans up resources
// P0: Closes stopTicker channel to stop heartbeat ticker goroutine
func (s *HTTPServerNew) Stop() {
	close(s.stopTicker)
}

// isLocalhostOrigin checks if the origin is a localhost address
// FIX-015: Helper for WebSocket origin validation in development
func isLocalhostOrigin(origin string) bool {
	// Patterns to match - must be followed by port (:), path (/), or end of string
	localhostPatterns := []string{
		"http://localhost",
		"https://localhost",
		"http://127.0.0.1",
		"https://127.0.0.1",
		"http://[::1]",
		"https://[::1]",
	}
	for _, pattern := range localhostPatterns {
		if strings.HasPrefix(origin, pattern) {
			// Check that the next character is valid (port, path, or end)
			rest := origin[len(pattern):]
			if len(rest) == 0 || rest[0] == ':' || rest[0] == '/' {
				return true
			}
		}
	}
	return false
}
