# GitOps Workflow - Industrial AI Platform

## 概述

本文档描述 Industrial AI Platform 的 GitOps 工作流程，包括：
- 开发流程
- 部署流程
- Rollback 策略
- 环境管理
- 最佳实践

---

## 🔄 GitOps 核心原则

| 原则 | 说明 |
|------|------|
| **Git 作为唯一真理源** | 所有配置存储在 Git，集群状态与 Git 一致 |
| **声明式配置** | 使用 YAML/Kustomize/Helm 定义期望状态 |
| **自动化同步** | ArgoCD 自动检测变更并同步 |
| **审计日志** | Git 提交历史作为变更审计记录 |
| **Rollback** | 通过 Git revert 或 ArgoCD rollback 快速回滚 |

---

## 📂 目录结构

```
industrial-ai-platform/
├── infra/
│   ├── k8s/                    # K8s manifests (ArgoCD 监控此目录)
│   │   ├── deployment.yaml
│   │   ├── deployment-health.yaml
│   │   ├── hpa.yaml
│   │   ├── ingress-tls.yaml
│   │   ├── secrets.yaml
│   │   └── prometheus-adapter.yaml
│   ├── gitops/                 # ArgoCD 配置
│   │   ├── argocd-install.yaml
│   │   ├── application.yaml
│   │   ├── project.yaml
│   │   └── sync-policy.yaml
│   └── kustomize/              # Kustomize overlays
│       ├── base/
│       ├── overlays/
│       │   ├── dev/
│       │   ├── staging/
│       │   └── production/
└── .github/
    └── workflows/              # CI/CD (与 ArgoCD 配合)
        ├── ci.yaml
        ├── preview.yaml
        ├── release.yaml
```

---

## 🚀 开发工作流

### 1. Feature Branch 开发

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ 1. 创建分支 │ ──▶ │ 2. 开发提交 │ ──▶ │ 3. PR 创建 │ ──▶ │ 4. 合并     │
│ feature/*   │     │ 代码+配置   │     │ Code Review │     │ main        │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                          │                   │                    │
                          ▼                   ▼                    ▼
                    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
                    │ 自动测试     │     │ Preview 环境│     │ 自动部署     │
                    │ CI Pipeline │     │ 自动创建    │     │ ArgoCD Sync │
                    └─────────────┘     └─────────────┘     └─────────────┘
```

### 2. PR Preview 环境

每个 Pull Request 自动创建独立 Preview 环境：

```yaml
# .github/workflows/preview.yaml
name: Create Preview Environment
on:
  pull_request:
    types: [opened, synchronize]

jobs:
  preview:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Create Preview Namespace
        run: |
          kubectl create namespace industrial-ai-preview-pr-${{ github.event.pull_request.number }}
      
      - name: Deploy Preview
        run: |
          kubectl apply -k infra/kustomize/overlays/dev \
            -n industrial-ai-preview-pr-${{ github.event.pull_request.number }}
      
      - name: Comment Preview URL
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '🚀 Preview environment ready!\nURL: https://pr-${{ github.event.pull_request.number }}.preview.industrial-ai.example.com'
            })
```

### 3. Preview 环境清理

PR 合成或关闭时自动清理：

```yaml
# .github/workflows/preview-cleanup.yaml
name: Cleanup Preview Environment
on:
  pull_request:
    types: [closed]

jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - name: Delete Preview Namespace
        run: |
          kubectl delete namespace industrial-ai-preview-pr-${{ github.event.pull_request.number }} --ignore-not-found=true
```

---

## 🎯 部署工作流

### 环境层级

| 环境 | 分支 | ArgoCD Project | Sync Policy |
|------|------|----------------|-------------|
| **Development** | `develop` | `industrial-ai-dev` | 自动同步 |
| **Staging** | `staging` | `industrial-ai-staging` | 自动同步 + 手动审批 |
| **Production** | `main` | `industrial-ai` | 手动审批 |

### Deployment Pipeline

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Deployment Pipeline                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Git Push ──▶ CI Build ──▶ Image Push ──▶ Git Tag ──▶ ArgoCD Sync       │
│                                                                          │
│  ┌───────┐    ┌───────┐    ┌───────┐    ┌───────┐    ┌───────────────┐ │
│  │ Push  │───▶│ Test  │───▶│ Build │───▶│ Push  │───▶│ ArgoCD Update │ │
│  │ main  │    │ Suite │    │ Image │    │ Reg.  │    │ Application   │ │
│  └───────┘    └───────┘    └───────┘    └───────┘    └───────────────┘ │
│       │           │           │           │              │              │
│       │           ▼           ▼           │              ▼              │
│       │      ┌─────────┐ ┌─────────┐      │      ┌───────────────┐      │
│       │      │ Unit    │ │ Docker  │      │      │ Auto/Manual   │      │
│       │      │ + E2E   │ │ Build   │      │      │ Sync Trigger  │      │
│       │      └─────────┘ └─────────┘      │      └───────────────┘      │
│       │                                  │              │              │
│       │                                  │              ▼              │
│       │                                  │      ┌───────────────┐      │
│       │                                  │      │ Health Check  │      │
│       │                                  │      │ + Validation  │      │
│       │                                  │      └───────────────┘      │
│       │                                  │              │              │
│       │                                  │              ▼              │
│       │                                  │      ┌───────────────┐      │
│       │                                  │      │ Slack Notify  │      │
│       │                                  │      └───────────────┘      │
│       ▼                                  ▼              ▼              │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                      Production Cluster                          │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐         │  │
│  │  │ Backend  │  │ Frontend │  │ Postgres │  │ Redis    │         │  │
│  │  │ (3 pods) │  │ (3 pods) │  │ (Primary │  │ (Sentinel│         │  │
│  │  │          │  │          │  │  + Rep)  │  │  HA)     │         │  │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘         │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### ArgoCD Sync Triggers

| 触发方式 | 环境 | 说明 |
|----------|------|------|
| **自动同步** | Dev, Staging | Git 变更自动触发 |
| **手动审批** | Production | ArgoCD UI/API 手动触发 |
| **Webhook** | All | GitHub webhook 触发 |
| **定时同步** | Production | 每日定时检查 |

---

## ⏪ Rollback 流程

### Rollback 策略

```
┌─────────────────────────────────────────────────────────────────┐
│                    Rollback Decision Tree                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Issue Detected ──▶ Assessment ──▶ Strategy Selection          │
│                                                                  │
│   ┌───────────┐     ┌───────────┐     ┌───────────────────────┐ │
│   │ Health    │───▶│ Severity? │───▶│ Choose Strategy        │ │
│   │ Check Fail│     │           │     │                       │ │
│   └───────────┘     └───────────┘     └───────────────────────┘ │
│                           │                       │              │
│                           ▼                       ▼              │
│                   ┌───────────────┐     ┌─────────────────────┐ │
│                   │ CRITICAL      │     │ IMMEDIATE ROLLBACK  │ │
│                   │ (< 2 min)     │     │ • Fast recovery     │ │
│                   └───────────────┘     │ • Skip validation   │ │
│                           │             │ • RTO < 90s         │ │
│                           │             └─────────────────────┘ │
│                           │                       │              │
│                           ▼                       ▼              │
│                   ┌───────────────┐     ┌─────────────────────┐ │
│                   │ HIGH          │     │ STAGED ROLLBACK     │ │
│                   │ (< 15 min)    │     │ • Create backup     │ │
│                   └───────────────┘     │ • Run validation    │ │
│                           │             │ • RTO < 8 min       │ │
│                           │             └─────────────────────┘ │
│                           │                       │              │
│                           ▼                       ▼              │
│                   ┌───────────────┐     ┌─────────────────────┐ │
│                   │ MEDIUM        │     │ EMERGENCY ROLLBACK  │ │
│                   │ (> 15 min)    │     │ • Bypass all checks │ │
│                   └───────────────┘     │ • Immediate action  │ │
│                           │             │ • RTO < 60s         │ │
│                           │             └─────────────────────┘ │
│                           ▼                       │              │
│                   ┌───────────────┐                 │              │
│                   │ Git Revert    │                 │              │
│                   │ + Manual Fix  │                 │              │
│                   └───────────────┘                 │              │
│                                                     │              │
│   ┌─────────────────────────────────────────────────┘              │
│   │                                                                │
│   ▼                                                                │
│   Rollback Complete ──▶ Post-Rollback Actions                      │
│                                                                  │
│   ┌───────────────┐    ┌───────────────────────────────────────┐ │
│   │ Notify Team   │───▶│ • Slack alert                        │ │
│   │               │    │ • Create incident ticket              │ │
│   │               │    │ • Update runbook                      │ │
│   │               │    │ • Root cause analysis                 │ │
│   └───────────────┘    └───────────────────────────────────────┘ │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Rollback 执行方式

#### 1. ArgoCD UI Rollback

```bash
# 在 ArgoCD UI 中：
# 1. 选择 Application
# 2. 点击 "History and rollback"
# 3. 选择目标 Revision
# 4. 点击 "Rollback"
```

#### 2. CLI Rollback

```bash
# 查看历史版本
argocd app history industrial-ai-platform

# 回滚到指定 revision
argocd app rollback industrial-ai-platform --revision <revision-id>

# 紧急回滚（跳过验证）
argocd app rollback industrial-ai-platform --revision <revision-id> --force
```

#### 3. Git Revert

```bash
# Git revert 方式回滚
git revert <commit-hash>
git push origin main

# ArgoCD 自动检测并同步
```

---

## 🔔 告警与通知

### ArgoCD 通知配置

| 事件 | 通知渠道 | 内容 |
|------|----------|------|
| **Sync Started** | Slack | "Application sync started" |
| **Sync Completed** | Slack | "Application synced successfully" |
| **Sync Failed** | Slack + Email | "Sync failed - requires intervention" |
| **Health Degraded** | Slack + PagerDuty | "Application health degraded" |
| **Rollback** | Slack + Email | "Rollback executed to revision X" |

### Slack Integration

```yaml
# argocd-notifications-cm.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
  namespace: argocd
data:
  service.slack: |
    {
      "token": "xoxb-..."
    }
  
  template.app-sync-status: |
    message: Application {{ .app.metadata.name }} sync status: {{ .app.status.sync.status }}
    slack:
      attachments: |
        [{
          "title": "{{ .app.metadata.name }}",
          "title_link": "https://argocd.example.com/applications/{{ .app.metadata.name }}",
          "text": "Sync Status: {{ .app.status.sync.status }}",
          "color": "{{ if eq .app.status.sync.status \"Synced\" }}#36a64f{{ else }}#f2c744{{ end }}"
        }]
  
  trigger.on-sync-completed: |
    - description: Application sync completed
      send: [app-sync-status]
      when: app.status.sync.status == 'Synced'
```

---

## 🛡️ 最佳实践

### 1. Git 提交规范

```
<type>(<scope>): <subject>

<body>

<footer>

类型:
- feat: 新功能
- fix: 修复
- docs: 文档
- style: 格式
- refactor: 重构
- test: 测试
- ci: CI/CD
- ops: 运维
- k8s: K8s 配置变更

示例:
feat(backend): add multi-tenant support

- Add tenant isolation middleware
- Update JWT claims with tenant_id
- Create tenant repository

Closes #123
```

### 2. PR 审核清单

```markdown
## PR Checklist

- [ ] 代码通过 lint 检查
- [ ] 单元测试通过
- [ ] E2E 测试通过（如适用）
- [ ] 文档更新（如有 API 变化）
- [ ] K8s 配置验证（kubectl apply --dry-run）
- [ ] Reviewer approval
- [ ] Preview 环境测试通过
```

### 3. 环境隔离

| 规则 | 说明 |
|------|------|
| **命名空间隔离** | dev/staging/production 使用不同 namespace |
| **资源配额** | 每个环境设置 ResourceQuota |
| **网络策略** | NetworkPolicy 防止跨环境通信 |
| **Secret 管理** | 每个环境独立 Secret |

### 4. 安全策略

- ✅ 生产环境禁用自动 sync
- ✅ 所有变更必须通过 PR
- ✅ Secrets 使用 Sealed Secrets 或 Vault
- ✅ RBAC 限制 ArgoCD 访问权限
- ✅ Audit log 记录所有操作

---

## 📊 监控 Dashboard

### ArgoCD 监控指标

| 指标 | 说明 |
|------|------|
| `argocd_app_sync_status` | 应用同步状态 |
| `argocd_app_health_status` | 应用健康状态 |
| `argocd_app_operation_running` | 正在运行的操作数 |
| `argocd_cluster_api_resources` | 集群资源统计 |

### Grafana Dashboard

```yaml
# infra/grafana/dashboards/argocd-dashboard.json
{
  "title": "ArgoCD Overview",
  "panels": [
    {
      "title": "Application Sync Status",
      "type": "stat",
      "targets": [{"expr": "count(argocd_app_sync_status == 'Synced')"}]
    },
    {
      "title": "Application Health",
      "type": "stat",
      "targets": [{"expr": "count(argocd_app_health_status == 'Healthy')"}]
    },
    {
      "title": "Sync Operations",
      "type": "graph",
      "targets": [{"expr": "rate(argocd_app_operation_total[5m])"}]
    }
  ]
}
```

---

## 📝 Runbook

### 常见问题处理

#### 1. Sync Failed

```bash
# 检查 sync 状态
argocd app get industrial-ai-platform

# 查看 sync 日志
argocd app logs industrial-ai-platform

# 强制重新 sync
argocd app sync industrial-ai-platform --force
```

#### 2. Application OutOfSync

```bash
# 检查差异
argocd app diff industrial-ai-platform

# 手动 sync
argocd app sync industrial-ai-platform
```

#### 3. Health Check Failed

```bash
# 检查 pod 状态
kubectl get pods -n industrial-ai

# 检查 deployment events
kubectl describe deployment backend -n industrial-ai

# Rollback if needed
argocd app rollback industrial-ai-platform --revision <id>
```

---

## 📚 参考资料

- [ArgoCD 官方文档](https://argo-cd.readthedocs.io/)
- [GitOps 最佳实践](https://opengitops.dev/)
- [Kustomize 官方文档](https://kustomize.io/)
- [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)