# 监控告警配置优化报告

## P2-002: 监控告警调优

### 项目路径
`/Users/yqgmac/yqg/project/industrial-ai-platform`

---

## 一、优化内容总结

### 1. 告警规则调优

#### 1.1 阈值调整
文件: `backend/pkg/constants/constants.go`

| 指标 | 原阈值 | 新阈值 | 说明 |
|------|--------|--------|------|
| **温度警告** | 100°C | 80°C | 提前预警，降低风险 |
| **温度严重** | 120°C | 100°C | 更合理的严重阈值 |
| **温度预警** | - | 70°C | 新增分层预警 |
| **振动警告** | 3.0mm/s | 2.8mm/s | 符合ISO 10816标准 |
| **振动严重** | 5.0mm/s | 4.5mm/s | 更早发现异常 |
| **振动预警** | - | 1.8mm/s | 新增分层预警 |
| **压力警告** | 150kPa | 120kPa | 更合理的工业阈值 |
| **压力严重** | - | 150kPa | 新增严重级别 |
| **压力预警** | - | 100kPa | 新增分层预警 |
| **湿度警告** | - | 85% | 新增湿度监控 |
| **湿度严重** | - | 95% | 新增湿度严重阈值 |
| **功率警告** | - | 5000W | 新增功率监控 |
| **功率严重** | - | 8000W | 新增功率严重阈值 |

#### 1.2 告警级别分层
- **P1 紧急 (critical)**: 立即处理，多渠道通知
- **P2 警告 (high)**: 标准处理，飞书+邮件通知
- **P3 提醒 (medium/low)**: 低优先级，钉钉通知

#### 1.3 新增告警规则
文件: `infra/prometheus/industrial_device_alert_rules.yml`

- 温度分层告警（预警/警告/严重）
- 振动分层告警（符合ISO 10816）
- 压力分层告警
- 湿度告警（高/低湿度）
- 功率告警（高/低功率）
- 综合健康告警（多指标异常）
- 维护周期提醒

---

### 2. 告警聚合策略

#### 2.1 Alertmanager配置
文件: `infra/prometheus/alertmanager_enhanced.yml`

**分组策略:**
```yaml
group_by: ['severity', 'category', 'device_id', 'priority']
group_wait: 30s
group_interval: 5m
repeat_interval: 4h
```

#### 2.2 抑制规则
- 同设备critical抑制warning/info
- 设备离线抑制其他设备告警
- 设备故障抑制所有设备告警
- 多指标异常抑制单独指标告警
- P1抑制同设备P2/P3告警
- 压力骤降抑制压力异常
- 功率突变抑制功率告警

#### 2.3 业务层聚合
- 已有cooldown机制（批量查询优化）
- 新增AlertArchiver自动归档调度

---

### 3. 通知渠道配置

#### 3.1 飞书通知
文件: `backend/pkg/notify/feishu.go`

- 已有完整实现
- 支持卡片消息格式
- 支持告警解决通知

#### 3.2 钉钉通知 (新增)
文件: `backend/pkg/notify/dingtalk.go`

- Markdown消息格式
- 支持@所有人（critical告警）
- 支持维护提醒
- 支持告警解决通知

#### 3.3 NotifyManager增强
```go
// 新增带钉钉的构造函数
NewNotifyManagerWithDingTalk(feishuWebhook, dingtalkWebhook, enabled bool)
```

#### 3.4 邮件通知配置
```yaml
smtp_smarthost: '${SMTP_HOST}:${SMTP_PORT}'
smtp_from: '${SMTP_FROM}'
smtp_auth_username: '${SMTP_USERNAME}'
smtp_auth_password_file: /etc/alertmanager/secrets/smtp-password
```

---

### 4. 告警历史管理

#### 4.1 Repository接口扩展
文件: `backend/internal/repository/alert_interface.go`

新增方法:
- `ArchiveOldAlerts(ctx, daysOld)` - 归档旧告警
- `GetArchivedAlerts(ctx, deviceID, page, pageSize)` - 查询归档告警
- `DeleteArchivedAlerts(ctx, daysOld)` - 删除过期归档
- `GetAlertStatistics(ctx)` - 获取告警统计

#### 4.2 AlertStatistics结构
```go
type AlertStatistics struct {
    TotalActive     int
    TotalResolved   int
    TotalArchived   int
    TodayTriggered  int
    TodayResolved   int
    WeekTriggered   int
    WeekResolved    int
    AvgResolveTime  int  // 秒
    CriticalCount   int
    WarningCount    int
}
```

#### 4.3 AlertArchiver调度器
文件: `backend/internal/service/alert_archiver.go`

配置:
- 默认归档: 30天后归档已解决告警
- 默认删除: 90天后删除归档告警
- 调度间隔: 每24小时执行一次

---

## 二、创建/修改文件列表

| 文件 | 操作 | 说明 |
|------|------|------|
| `backend/pkg/constants/constants.go` | 修改 | 阈值常量优化 |
| `infra/prometheus/industrial_device_alert_rules.yml` | 新建 | 工业设备告警规则 |
| `infra/prometheus/alertmanager_enhanced.yml` | 新建 | 增强Alertmanager配置 |
| `backend/pkg/notify/dingtalk.go` | 新建 | 钉钉通知实现 |
| `backend/pkg/notify/feishu.go` | 修改 | NotifyManager增强 |
| `backend/internal/repository/alert_interface.go` | 修改 | 归档接口扩展 |
| `backend/internal/service/alert_archiver.go` | 新建 | 告警归档调度器 |
| `backend/internal/service/alert_service.go` | 修改 | 默认规则扩展 |

---

## 三、配置说明

### 3.1 环境变量配置

需要在Alertmanager中配置以下环境变量:

```bash
# SMTP邮件配置
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_FROM=alerts@industrial-ai.example.com
SMTP_USERNAME=alerts@industrial-ai.example.com

# Webhook配置
FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxx
DINGTALK_WEBHOOK_URL=https://oapi.dingtalk.com/robot/send?access_token=xxx
ALERT_WEBHOOK_TOKEN=your-webhook-token

# 团队邮箱
OPS_TEAM_EMAIL=ops-team@example.com
MAINTENANCE_TEAM_EMAIL=maintenance@example.com
```

### 3.2 Secrets配置

需要创建以下secrets文件:
- `/etc/alertmanager/secrets/smtp-password`
- `/etc/alertmanager/secrets/pagerduty-key` (可选)
- `/etc/alertmanager/secrets/slack-api-url` (可选)

---

## 四、验证步骤

### 4.1 告警触发测试

```bash
# 测试温度告警触发
curl -X POST http://backend:8080/api/v1/telemetry \
  -H "Content-Type: application/json" \
  -d '{"device_id":"test-001","temperature":85,"pressure":100,"vibration":2.0}'

# 测试严重告警触发
curl -X POST http://backend:8080/api/v1/telemetry \
  -H "Content-Type: application/json" \
  -d '{"device_id":"test-001","temperature":105,"pressure":160,"vibration":5.0}'
```

### 4.2 告警通知测试

```bash
# 测试飞书通知
curl -X POST ${FEISHU_WEBHOOK_URL} \
  -H "Content-Type: application/json" \
  -d '{"msg_type":"text","content":{"text":"测试通知"}}'

# 测试钉钉通知
curl -X POST ${DINGTALK_WEBHOOK_URL} \
  -H "Content-Type: application/json" \
  -d '{"msgtype":"text","text":{"content":"测试通知"}}'
```

### 4.3 告警归档测试

```bash
# 手动触发归档（需要实现API）
curl -X POST http://backend:8080/api/v1/alerts/archive

# 查看告警统计
curl http://backend:8080/api/v1/alerts/statistics
```

---

## 五、向后兼容性

### 5.1 保持兼容
- 现有告警规则未删除，仅添加新规则
- 现有阈值常量名称保持不变
- AlertService接口未改变

### 5.2 新增功能
- 新阈值常量（Warning/Critical级别）
- 钉钉通知（可选启用）
- 告警归档（可选启用）

---

## 六、后续建议

### 6.1 待实现
1. Repository层实现归档SQL操作
2. 添加告警归档API接口
3. 集成测试告警流程
4. 配置生产环境Secrets

### 6.2 可选优化
1. 添加工作时间告警策略
2. 实现告警静默API
3. 添加告警趋势分析
4. 实现预测性告警