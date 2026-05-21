# E2E 测试指南

## 概述

本目录包含 Industrial AI Agent Platform 的端到端 (E2E) 测试，使用 Playwright 框架实现。

## 测试覆盖

| 测试文件 | 场景 | 测试数量 |
|---------|------|---------|
| `auth/login.spec.ts` | 登录流程 | 8 |
| `devices/device-list.spec.ts` | 设备管理 | 12 |
| `telemetry/telemetry.spec.ts` | 遥测数据 | 10 |
| `alerts/alerts.spec.ts` | 告警触发 | 12 |
| `ai/ai-agent.spec.ts` | AI 对话 | 10 |
| `reports/reports.spec.ts` | 报告生成 | 10 |
| `i18n/i18n.spec.ts` | 多语言切换 | 12 |

**总计**: ~74 个测试场景

## 安装依赖

```bash
cd frontend

# 安装 Playwright
npm install

# 安装浏览器
npx playwright install
```

## 运行测试

### 全部测试

```bash
npm run test:e2e
```

### 单个测试文件

```bash
npx playwright test e2e/auth/login.spec.ts
```

### UI 模式 (推荐调试)

```bash
npm run test:e2e:ui
```

### Debug 模式

```bash
npx playwright test --debug
```

### 查看报告

```bash
npm run test:e2e:report
```

## 测试用户

| 角色 | 用户名 | 密码 | 权限 |
|------|--------|------|------|
| 管理员 | admin | admin123 | 全部 |
| 操作员 | operator | operator123 | 设备管理、告警处理 |
| 观察者 | viewer | viewer123 | 只读 |

## 测试设备

| 设备 ID | 名称 | 类型 |
|---------|------|------|
| test-cnc-001 | Test CNC Machine | CNC |
| test-inj-001 | Test Injection Molder | INJ |

## 环境变量

```bash
# 后端 API 地址
E2E_API_URL=http://localhost:8080

# 前端地址
E2E_BASE_URL=http://localhost:5173

# 管理员密码 (如果非默认)
E2E_ADMIN_PASSWORD=admin123
```

## 测试策略

### 1. 测试隔离

每个测试独立运行，使用 `beforeEach` 和 `afterEach` 清理数据。

### 2. 自定义 Fixtures

使用 Playwright fixtures 复用登录状态：

```typescript
// 使用已登录页面
test('测试', async ({ authenticatedPage }) => {
  await authenticatedPage.goto('/devices');
});
```

### 3. 多浏览器测试

默认测试 5 种浏览器/设备：
- Desktop Chrome
- Desktop Firefox  
- Desktop Safari
- Mobile Chrome (Pixel 5)
- Mobile Safari (iPhone 12)

### 4. 响应式测试

测试移动端布局：

```typescript
await page.setViewportSize({ width: 375, height: 667 });
```

## 编写新测试

### 模板

```typescript
import { test, expect } from '../fixtures/test-fixtures';

test.describe('新功能测试', () => {
  
  test.beforeEach(async ({ adminPage }) => {
    await adminPage.goto('/new-feature');
  });
  
  test('功能正常', async ({ adminPage }) => {
    // 验证元素存在
    await expect(adminPage.locator('[data-testid="element"]')).toBeVisible();
    
    // 执行操作
    await adminPage.click('button');
    
    // 验证结果
    await expect(adminPage.locator('text="成功"')).toBeVisible();
  });
});
```

### 最佳实践

1. **使用 data-testid**: 优先使用 `[data-testid="xxx"]` 选择器
2. **等待网络空闲**: `await page.waitForLoadState('networkidle')`
3. **超时设置**: 关键操作设置合理超时
4. **截图调试**: 失败时自动截图保存到 `test-results/`

## CI/CD 集成

### GitHub Actions

```yaml
name: E2E Tests

on:
  push:
    branches: [main]
  pull_request:

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: 20
          
      - name: Install dependencies
        run: |
          cd frontend
          npm ci
          npx playwright install --with-deps
          
      - name: Run E2E tests
        run: |
          cd frontend
          npm run test:e2e
          
      - name: Upload report
        uses: actions/upload-artifact@v3
        with:
          name: playwright-report
          path: frontend/playwright-report/
```

## 故障排查

### 测试超时

```bash
# 增加超时时间
npx playwright test --timeout=60000
```

### 浏览器问题

```bash
# 重新安装浏览器
npx playwright install --force
```

### 查看详细日志

```bash
npx playwright test --verbose
```

---

**测试框架**: Playwright 1.45+  
**更新日期**: 2026-05-13