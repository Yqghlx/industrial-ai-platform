# Kubernetes 配置已迁移

本目录中的 K8s 配置已合并到更完整的位置：

**请使用：`infra/k8s/`**

该目录包含：

- `deployment.yaml` - Backend + Frontend 部署（含安全加固）
- `deployment-health.yaml` - 健康探针专用配置（Liveness/Readiness/Startup）
- `hpa.yaml` - 水平自动扩缩容（Backend/Frontend/AI Worker）
- `ingress-tls.yaml` - Ingress + TLS（含 cert-manager）
- `secrets.yaml` - Secrets + RBAC + 密钥轮换 CronJob
- `prometheus-adapter.yaml` - Prometheus 自定义指标适配器
- `monitoring/` - Prometheus + Grafana 监控栈
- `scripts/` - 运维脚本（密钥轮换、配置验证）
