# API接口测试报告

**测试时间**: 2026-05-28 03:27:32 UTC
**测试人**: 小猪蹄儿(Hermes Agent)

## 测试结果汇总

| 接口 | 路由 | 状态 | 响应 |
|------|------|------|------|
| **Health** | `/health` | ✅ 正常 | healthy, uptime 58300秒 |
| **Login** | `/api/v1/auth/login` | ✅ 正常 | 错误信息正确 |
| **ROI stats** | `/api/v1/roi/stats` | ✅ 正常 | 需认证（正常） |
| **Devices** | `/api/v1/devices` | ✅ 正常 | 需认证（正常） |
| **Cache status** | `/cache/status` | ✅ 正常 | redis连接正常 |
| **WebSocket stats** | `/ws/stats` | ✅ 正常 | clients_count: 0 |

## Docker容器状态

| 容器 | 状态 | 运行时间 |
|------|------|----------|
| **backend** | ✅ healthy | 16 hours |
| **postgres** | ✅ healthy | 16 hours |
| **redis** | ✅ healthy | 16 hours |

## 结论

所有API接口测试正常，服务稳定运行16小时，无异常。
