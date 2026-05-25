# Industrial AI Platform - 修复计划

**创建日期**: 2026-05-25
**基于审计**: docs/CODE_AUDIT_REPORT.md

---

## Phase 1: P0/CRITICAL 修复（预计12h）

### Loop 1.1: Security CRITICAL (6h)
| ID | 问题 | 工时 | 状态 |
|----|------|------|------|
| SEC-CRITICAL-01 | Kubernetes硬编码密钥 | 4h | pending |
| SEC-CRITICAL-02 | Docker Compose硬编码密码 | 2h | pending |

### Loop 1.2: Backend P0 (11.5h)
| ID | 问题 | 工时 | 状态 |
|----|------|------|------|
| BE-P0-01 | Handler Factory返回nil | 4h | pending |
| BE-P0-02 | ID解析忽略错误 | 2h | pending |
| BE-P0-03 | 内存过滤性能问题 | 4h | pending |
| BE-P0-04 | CSRF panic | 1h | pending |
| BE-P0-05 | 示例代码panic | 0.5h | pending |

### Loop 1.3: Frontend P0 (6h)
| ID | 问题 | 工时 | 状态 |
|----|------|------|------|
| FE-P0-01 | AlertReportPage国际化缺失 | 2h | pending |
| FE-P0-02 | PerformancePanel国际化缺失 | 1h | pending |
| FE-P0-03 | ROIStatsPage国际化缺失 | 1h | pending |
| FE-P0-04 | BlackBoxPage国际化缺失 | 1h | pending |
| FE-P0-05 | SystemStatusPage国际化缺失 | 1h | pending |

---

## Phase 2: P1/HIGH 修复（预计15h）

| 维度 | 数量 | 工时 |
|------|------|------|
| Backend P1 | 9项 | 8h |
| Frontend P1 | 8项 | 4h |
| Security HIGH | 4项 | 3h |

---

## Phase 3: P2/MEDIUM 修复（预计10h）

| 维度 | 数量 | 工时 |
|------|------|------|
| Backend P2 | 8项 | 4h |
| Frontend P2 | 12项 | 4h |
| Security MEDIUM | 3项 | 2h |

---

## Phase 4: P3/LOW 修复（预计5h）

| 维度 | 数量 | 工时 |
|------|------|------|
| Frontend P3 | 15项 | 4h |
| Security LOW | 5项 | 1h |

---

## 执行策略

1. **并行修复**: 使用 delegate_task 并行执行最多3个修复任务
2. **按Loop提交**: 每个 Loop 完成后 commit + push
3. **验证优先**: 修复后立即验证（go build, npm test）
4. **增量汇报**: 每个 Phase 完成后发送飞书卡片

---

## 优先级矩阵

| 优先级 | 数量 | 总工时 | 建议完成时间 |
|--------|------|--------|--------------|
| P0/CRITICAL | 12项 | 23.5h | 本周 |
| P1/HIGH | 21项 | 15h | 下周 |
| P2/MEDIUM | 23项 | 10h | 两周内 |
| P3/LOW | 20项 | 5h | 可选 |

---

**总修复项**: 76项
**总预估工时**: 53.5h