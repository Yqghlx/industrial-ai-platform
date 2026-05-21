# 数据库分区指南

> **Industrial AI Platform 遥测数据时间分区最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 分区概述

Phase 4 P2 数据库分区目标：

| 指标 | 当前状态 | 目标 |
|------|---------|------|
| **查询性能** | 10s+ | <1s |
| **数据管理** | 手动 | 自动 |
| **存储效率** | 全表扫描 | 分区扫描 |
| **维护成本** | 高 | 低 |

---

## 🔄 分区架构

### 遥测数据分区策略

```
┌─────────────────────────────────────────┐
│  telemetry_data 表                       │
│  - 设备遥测数据                           │
│  - 每秒 10000+ 条数据                     │
│  - 每月 2.6B 条数据                       │
└─────────────────────────────────────────┘
          ↓ 时间分区
┌─────────────────────────────────────────┐
│  telemetry_data_2026_01                  │
│  - 2026年1月数据                          │
│  - 约 86M 条数据                          │
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│  telemetry_data_2026_02                  │
│  - 2026年2月数据                          │
│  - 约 78M 条数据                          │
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│  telemetry_data_2026_03                  │
│  - 2026年3月数据                          │
│  - 约 86M 条数据                          │
└─────────────────────────────────────────┘
          ↓ 自动管理
┌─────────────────────────────────────────┐
│  分区管理脚本                             │
│  - 自动创建新分区                         │
│  - 自动删除旧分区                         │
│  - 自动归档数据                           │
└─────────────────────────────────────────┘
```

---

## 🔧 PostgreSQL 分区实现

### 分区表创建

```sql
-- 创建分区父表
CREATE TABLE telemetry_data (
    telemetry_id VARCHAR(64) NOT NULL,
    device_id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    metric_type VARCHAR(32) NOT NULL,
    metric_value FLOAT NOT NULL,
    metric_unit VARCHAR(16),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY RANGE (timestamp);

-- 创建分区索引
CREATE INDEX idx_telemetry_device ON telemetry_data(device_id);
CREATE INDEX idx_telemetry_tenant ON telemetry_data(tenant_id);
CREATE INDEX idx_telemetry_type ON telemetry_data(metric_type);

-- 创建月度分区
CREATE TABLE telemetry_data_2026_01 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE telemetry_data_2026_02 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
```

---

## 📊 分区查询优化

### 分区剪枝效果

| 查询条件 | 全表扫描 | 分区扫描 | 性能提升 |
|---------|---------|---------|---------|
| **单日查询** | 10s | 0.1s | 100x |
| **单周查询** | 70s | 0.7s | 100x |
| **单月查询** | 300s | 3s | 100x |

---

## 📝 分区管理策略

### 自动分区管理

```bash
# 分区管理脚本
# 每月自动创建新分区
# 每季度自动归档旧分区
# 每年自动删除超期分区

# 保留策略
# - 热数据: 当前月 + 前 3 个月
# - 温数据: 前 6 个月 (归档到压缩表)
# - 冷数据: 超过 12 个月 (删除或导出)
```

---

## ✅ 分区验收标准

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **分区创建** | 按月分区 | 检查分区表 |
| **查询性能** | <1s | 性能测试 |
| **自动管理** | 脚本运行 | 检查脚本 |
| **数据归档** | 定期归档 | 检查归档表 |
| **存储节省** | 50%+ | 空间对比 |

---

**最后更新**: 2026-05-13  
**审核人**: Database Team