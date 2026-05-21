# E2E Tests - 端到端测试框架

工业AI代理平台端到端测试框架，使用 Playwright 进行自动化测试。

## 目录结构

```
tests/e2e/
├── playwright.config.ts    # Playwright 配置文件
├── login_flow.spec.ts      # 登录流程测试
├── package.json            # 测试依赖配置
└── README.md               # 本文档
```

## 安装依赖

```bash
cd tests/e2e
npm install
npx playwright install  # 安装浏览器驱动
```

## 运行测试

### 运行所有测试

```bash
npm test
# 或
npx playwright test
```

### 运行特定浏览器测试

```bash
npm run test:chromium    # Chrome
npm run test:firefox     # Firefox
npm run test:webkit      # Safari
```

### 运行移动端测试

```bash
npm run test:mobile      # 移动设备测试
npm run test:tablet      # 平板设备测试
```

### 调试测试

```bash
npm run test:debug       # 调试模式
npm run test:ui          # UI 模式
```

### 运行特定测试

```bash
npx playwright test -g "should login successfully"
```

## 测试报告

```bash
npm run test:report      # 打开 HTML 报告
```

报告位置: `playwright-report/index.html`

## 测试覆盖范围

### login_flow.spec.ts - 登录流程测试

| 测试套件 | 测试内容 |
|----------|----------|
| Login Page Access | 登录页面元素显示、表单属性 |
| Login Form Validation | 表单验证（空字段、错误凭据） |
| Successful Login Flow | 成功登录、Token 存储、页面导航 |
| Token Refresh and Expiration | Token 刷新、过期处理 |
| Logout Flow | 登出流程、Token 清除 |
| Login Security | 速率限制、HTTPS、密码清除 |
| Login Responsive Design | 移动端、平板、桌面响应式 |
| Login API Response | API 响应结构验证 |

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `E2E_BASE_URL` | 前端 URL | `http://localhost:3000` |
| `E2E_API_URL` | 后端 API URL | `http://localhost:8080` |
| `E2E_TEST_USER` | 测试用户名 | `admin` |
| `E2E_TEST_PASSWORD` | 测试密码 | `Admin@123456` |
| `CI` | CI 环境（自动重试） | - |

## CI 集成

### GitHub Actions 配置示例

```yaml
name: E2E Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          
      - name: Install dependencies
        run: |
          cd tests/e2e
          npm ci
          npx playwright install --with-deps
          
      - name: Run tests
        run: cd tests/e2e && npm test
        env:
          CI: true
          E2E_BASE_URL: ${{ secrets.E2E_BASE_URL }}
          E2E_API_URL: ${{ secrets.E2E_API_URL }}
          
      - name: Upload report
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: playwright-report
          path: tests/e2e/playwright-report/
```

## 扩展测试

### 添加新测试文件

1. 在 `tests/e2e/` 目录创建新的 `.spec.ts` 文件
2. 使用标准 Playwright 测试结构：

```typescript
import { test, expect } from '@playwright/test';

test.describe('My Feature Tests', () => {
  test('should work correctly', async ({ page }) => {
    await page.goto('/my-feature');
    await expect(page.locator('.element')).toBeVisible();
  });
});
```

### 测试代码生成

```bash
npm run test:codegen    # 交互式生成测试代码
```

## 最佳实践

1. **使用 page objects**: 将页面操作封装为可复用的类
2. **独立测试**: 每个测试应独立运行，不依赖其他测试
3. **清理状态**: 在 `beforeEach` 中清理认证状态
4. **等待策略**: 使用 `expect` 自动等待，避免手动 `waitForTimeout`
5. **API 测试**: 使用 `page.request` 进行 API 层测试
6. **Mock 数据**: 在 CI 中使用 Mock 数据保持一致性