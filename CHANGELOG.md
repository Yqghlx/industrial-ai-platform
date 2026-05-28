# Changelog

All notable changes to the Industrial AI Platform project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Test coverage enhancement: Handler层覆盖率74.9% (2026-05-28)
- Comprehensive test coverage for backend services (平均80%+)
- AgentOptimizer test suite (GetCachedAnswer, CacheAnswer, AcquireSlot, ReleaseSlot)
- GetDeviceContext test coverage
- GetCSRFToken test coverage
- Code audit report (CODE_AUDIT_REPORT.md)
- Monitoring system verification report (MONITORING_SYSTEM_CHECK_REPORT.md)

### Security
- SEC-CRITICAL-01: 删除.secrets.tmp明文密钥文件，验证git历史无提交 (✅ 已修复)
- SEC-CRITICAL-02: 敏感文件权限改为0600 + O_EXCL防符号链接攻击 (✅ 已修复)
- SEC-HIGH-01: 数据库SSL连接配置，默认sslmode=require (✅ 已修复)
- SEC-HIGH-02: 遥测端点设备认证机制DeviceAuthRequired middleware (✅ 已修复)
- SEC-HIGH-03: 管理员接口完整实现，密码强度验证+角色验证 (✅ 已修复)
- SEC-HIGH-04: CORS通配符改为环境变量配置，自动过滤* (✅ 已修复)
- MAJOR-02: GetUsername/GetUserRole安全类型断言模式（带ok检查） (✅ 已修复)
- MAJOR-03: Token黑名单淘汰策略优化（检查条目过期时间） (✅ 已修复)
- WebSocket real-time telemetry streaming
- AI Agent integration with GLM-5
- Device fleet dashboard with live metrics
- Alert rule configuration system
- Work order management system
- Report generation and export functionality

### Changed
- Migration execution optimized with independent transactions per migration
- TimescaleDB hypertable creation made optional for standard PostgreSQL
- System status API now returns real telemetry service data
- Redis配置改为环境变量REDIS_URL (Phase 1)
- 正则表达式预编译优化性能 (Phase 1)
- 后端硬编码URL/端口改为环境变量 (Phase 3)
- 魔法数字提取为常量（6个ROI常量） (Phase 3)
- 前端React.memo优化（5个组件） (Phase 3)
- i18n硬编码文本国际化 (Phase 3)

### Fixed
- **Phase 1 P0/CRITICAL修复（9项）**
  - P0-01: Redis硬编码地址改为环境变量 (✅)
  - P0-02: 正则表达式移到包级别预编译 (✅)
  - P0-03: 正确处理json.Marshal错误返回值 (✅)
  - P0-04: 移除panic，使用正常错误处理流程 (✅)
  - P0-05: 检查初始化错误，添加fallback处理 (✅)
  - P0-06: 表名白名单修正 + 列名验证增强 (✅)
  - P0-07: 前端事件监听器未清理，添加removeEventListener (✅)
- **Phase 2 P1/HIGH修复（21项）**
  - P1后端错误处理缺失（17项） (✅)
  - P1前端eslint-disable修复（15处移除） (✅)
  - P1类型断言安全模式 (✅)
- **Phase 3 P2/MEDIUM修复（17项）**
  - P2后端硬编码URL/端口修复 (✅)
  - P2魔法数字提取为常量 (✅)
  - P2 Goroutine泄漏修复（添加ctx/WG管理） (✅)
  - P2前端React.memo优化 (✅)
  - P2 i18n硬编码文本修复 (✅)
- **MINOR级别修复（7项）**
  - MINOR-01: KnowledgeGraph innerHTML清空改为textContent (✅)
  - MINOR-02: 占位API标记为TODO (✅)
  - MINOR-03: Circuit Breaker滑动窗口统计 (✅)
  - MINOR-04: WebSocket broadcaster显式启动 (✅)
  - MINOR-05: Repository租户隔离迁移文档 (✅)
  - MINOR-06: 测试panic添加recover机制 (✅)
  - MINOR-07: useEffect依赖完整性检查 (✅)
- **测试修复**
  - TestAdminHandlerNew_CreateUser_Success (✅ PASS)
  - TestAdminHandlerNew_DeleteUser (✅ PASS)
  - TestBusinessHandlerNew_GetROIStats_CacheUnavailable (✅ PASS)
- AdminHandlerNew test compilation errors (missing TelemetryServiceInterface parameter)
- SystemStatus.tsx useEffect dependency warnings (added useCallback wrapper)
- Database schema missing columns (tenant_id, token_version)
- Admin login authentication with correct password hashing
- Migration blocking subsequent migrations on TimescaleDB failure

## [1.0.0] - 2026-05-26

### Added
- Initial release of Industrial AI Platform
- Go backend with microservices architecture
- React frontend with TypeScript
- PostgreSQL database with migrations
- Redis caching layer
- Docker Compose deployment
- Gateway service (port 80)
- Auth service with JWT tokens
- Telemetry service for device data
- Alert service for rule evaluation
- AI service integration endpoints

### Security
- Password hashing with bcrypt
- JWT token authentication
- Role-based access control (RBAC)
- Tenant isolation support
- Audit logging for sensitive operations

### Performance
- N+1 query optimization with batch queries
- Memory filtering replaced with SQL WHERE clauses
- WebSocket + polling fallback for real-time data
- useMemo/useCallback for React performance
- Array size limits to prevent memory leaks

---

## Version History Summary

| Version | Date | Key Changes |
|---------|------|-------------|
| 1.0.0 | 2026-05-26 | Initial release |
| Unreleased | - | Bug fixes, test coverage, optimization |

---

**Note**: For detailed commit history, see [GitHub Releases](https://github.com/Yqghlx/industrial-ai-platform/releases).