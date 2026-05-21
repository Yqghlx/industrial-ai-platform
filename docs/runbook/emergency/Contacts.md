# 联系人列表 (Contacts)

> **Industrial AI Platform 运维联系人目录**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📞 核心运维团队

### Ops Team (运维团队)

| 姓名 | 角色 | 主要职责 | 电话 | Slack | 邮箱 | 工作时间 |
|------|------|---------|------|-------|------|---------|
| **Ops A** | 值班 Ops | 主值班 Week 1 | +86-138-xxxx-0001 | @ops-a | ops-a@industrial-ai.example.com | 24h on-call |
| **Ops B** | 备值班 Ops | 备值班 Week 1 | +86-138-xxxx-0002 | @ops-b | ops-b@industrial-ai.example.com | 24h on-call |
| **Ops C** | 值班 Ops | 主值班 Week 2 | +86-138-xxxx-0003 | @ops-c | ops-c@industrial-ai.example.com | 24h on-call |
| **Ops D** | 备值班 Ops | 备值班 Week 2 | +86-138-xxxx-0004 | @ops-d | ops-d@industrial-ai.example.com | 24h on-call |
| **Ops Lead** | 运维主管 | P0/P1 事件负责人 | +86-138-xxxx-0005 | @ops-lead | ops-lead@industrial-ai.example.com | 工作时间 + 紧急 |

---

## 🔧 开发团队

### Backend Team (后端开发)

| 姓名 | 角色 | 主要职责 | 电话 | Slack | 邮箱 | 工作时间 |
|------|------|---------|------|-------|------|---------|
| **Dev Lead** | 技术主管 | Panic/崩溃分析 | +86-139-xxxx-1001 | @dev-lead | dev-lead@industrial-ai.example.com | 工作时间 + 紧急 |
| **Dev A** | 高级开发 | 核心服务开发 | +86-139-xxxx-1002 | @dev-a | dev-a@industrial-ai.example.com | 工作时间 |
| **Dev B** | 高级开发 | AI Agent 开发 | +86-139-xxxx-1003 | @dev-b | dev-b@industrial-ai.example.com | 工作时间 |

### Frontend Team (前端开发)

| 姓名 | 角色 | 主要职责 | 电话 | Slack | 箱 | 工作时间 |
|------|------|---------|------|-------|------|---------|
| **FE Lead** | 前端主管 | 前端故障处理 | +86-139-xxxx-2001 | @fe-lead | fe-lead@industrial-ai.example.com | 工作时间 |
| **FE A** | 前端开发 | React 组件开发 | +86-139-xxxx-2002 | @fe-a | fe-a@industrial-ai.example.com | 工作时间 |

---

## 🗄️ 数据库团队

### DBA Team

| 姓名 | 角色 | 主要职责 | 电话 | Slack | 邮箱 | 工作时间 |
|------|------|---------|------|-------|------|---------|
| **DBA Lead** | 数据库主管 | 数据库紧急事件 | +86-137-xxxx-3001 | @dba-lead | dba-lead@industrial-ai.example.com | 工作时间 + 紧急 |
| **DBA A** | DBA | PostgreSQL 维护 | +86-137-xxxx-3002 | @dba-a | dba-a@industrial-ai.example.com | 工作时间 |

---

## 🤖 AI 团队

### AI Team

| 姓名 | 角色 | 主要职责 | 电话 | Slack | 邮箱 | 工作时间 |
|------|------|---------|------|-------|------|---------|
| **AI Lead** | AI 主管 | AI 服务故障处理 | +86-136-xxxx-4001 | @ai-lead | ai-lead@industrial-ai.example.com | 工作时间 + 紧急 |
| **AI A** | AI 工程师 | 模型优化 | +86-136-xxxx-4002 | @ai-a | ai-a@industrial-ai.example.com | 工作时间 |

---

## 🏢 管理层

### Management

| 姓名 | 角色 | 主要职责 | 电话 | Slack | 邮箱 | 工作时间 |
|------|------|---------|------|-------|------|---------|
| **Manager** | 运维经理 | P0 事件决策 | +86-135-xxxx-5001 | @manager | manager@industrial-ai.example.com | 工作时间 + 紧急 |
| **Director** | 技术总监 | 重大事件决策 | +86-135-xxxx-5002 | @director | director@industrial-ai.example.com | 紧急事件 |
| **CEO** | CEO | 业务连续性决策 | +86-135-xxxx-5003 | @ceo | ceo@industrial-ai.example.com | 灾难级事件 |

---

## 📨 通用联系方式

### 团队邮箱
| 团队 | 邮箱 | 用途 |
|------|------|------|
| Ops Team | ops@industrial-ai.example.com | 运维通知 |
| Dev Team | dev@industrial-ai.example.com | 开发协调 |
| DBA Team | dba@industrial-ai.example.com | 数据库事务 |
| AI Team | ai-team@industrial-ai.example.com | AI 服务 |
| All Teams | all@industrial-ai.example.com | 全员通知 |

### Slack 频道
| 频道 | 用途 | 成员 |
|------|------|------|
| #alerts-critical | Critical 告警通知 | Ops + Lead |
| #alerts-warning | Warning 告警通知 | Ops Team |
| #ops-alerts | HTTP/WebSocket 告警 | Ops Team |
| #dba-alerts | Database 告警 | DBA Team |
| #ai-alerts | AI 服务告警 | AI Team |
| #incident-response | 紧急事件协调 | 全团队 |
| #ops-team | Ops 内部沟通 | Ops Team |

---

## 📱 PagerDuty 配置

### 服务映射
| 服务 | PD 服务 | 通知策略 | 响应时间 |
|------|---------|---------|---------|
| Backend | Backend-Critical | 立即电话 + Slack | 5 分钟 |
| PostgreSQL | Database-Critical | 立即电话 + Slack | 5 分钟 |
| Redis | Cache-Critical | 立即电话 + Slack | 5 分钟 |
| AI Service | AI-Critical | 立即电话 + Slack | 5 分钟 |

### Escalation Policy
```
Level 1: 值班 Ops (5 分钟)
Level 2: Ops Lead + 相关开发 (15 分钟)
Level 3: Manager (30 分钟)
Level 4: Director (60 分钟)
```

---

## 🔗 外部服务商

### 云服务支持
| 服务商 | 服务类型 | 联系方式 | SLA |
|--------|---------|---------|------|
| **阿里云** | ECS/RDS | 客服热线 400-xxx-xxxx | P1 < 30min |
| **腾讯云** | COS/CDN | 客服热线 400-xxx-xxxx | P1 < 30min |

### 技术支持
| 服务 | 类型 | 联系方式 | 费用 |
|------|------|---------|------|
| **PostgreSQL 专家** | 数据库咨询 | pg-support@example.com | 合同内 |
| **Redis 专家** | 缓存咨询 | redis-support@example.com | 合同内 |

---

## 📋 值班轮换规则

### 轮换周期
- **主值班**: 每周轮换，周一 09:00 开始
- **备值班**: 同周备岗，主值班不可用时接管
- **交接时间**: 周一 09:00 Slack #ops-team

### 值班职责
- **主值班**:
  - 24 小时响应告警
  - 处理 P2/P3 告警
  - 协调 P0/P1 事件
  - 更新值班日志

- **备值班**:
  - 主值班休息时接管
  - 紧急事件协助处理
  - 复杂问题技术支持

### 值班交接清单
```bash
# 交接时必须确认:
1. 当前告警状态 (Grafana)
2. 未处理工单 (Jira)
3. 最近变更记录
4. 待跟进事项
5. 特殊注意事项
```

---

## 📝 联系人更新流程

1. **新增联系人**: Ops Lead 审核后添加
2. **修改联系方式**: 本人提交变更申请
3. **删除联系人**: Manager 审核后删除
4. **紧急变更**: 立即生效，事后补记录

**更新频率**: 每月审核一次  
**责任人**: Ops Lead  
**审核人**: Manager

---

## 🔗 相关链接

- **值班表**: https://wiki.industrial-ai.example.com/ops/rotation
- **PagerDuty**: https://industrial-ai.pagerduty.com
- **Slack**: https://industrial-ai.slack.com
- **工单系统**: https://jira.industrial-ai.example.com

---

**最后更新**: 2026-05-13  
**维护人**: Ops Lead  
**审核人**: Manager