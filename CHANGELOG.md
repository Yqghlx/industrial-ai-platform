# Changelog

All notable changes to the Industrial AI Platform project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Test coverage enhancement: handler/service packages 75.9% (2026-05-28)
- Comprehensive test coverage for backend services (77%+ average)
- AgentOptimizer test suite (GetCachedAnswer, CacheAnswer, AcquireSlot, ReleaseSlot)
- GetDeviceContext test coverage
- GetCSRFToken test coverage
- Code audit report (CODE_AUDIT_REPORT.md)

### Security
- P0 bug fix: ReleaseSlot test error causing panic (已修复)
- P2 bug fix: WebSocket broadcaster confirmed working (已确认)
- P2 bug identified: JWT blacklist eviction strategy (需决策)
- P2 bug identified: Tenant isolation missing in GetByID (需决策)
- WebSocket real-time telemetry streaming
- AI Agent integration with GLM-4-flash
- Device fleet dashboard with live metrics
- Alert rule configuration system
- Work order management system
- Report generation and export functionality

### Changed
- Migration execution optimized with independent transactions per migration
- TimescaleDB hypertable creation made optional for standard PostgreSQL
- System status API now returns real telemetry service data

### Fixed
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