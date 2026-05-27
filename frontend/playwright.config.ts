import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E 测试配置
 * Industrial AI Agent Platform
 */
export default defineConfig({
  // 测试目录
  testDir: './e2e',
  
  // TypeScript 项目配置
  tsconfig: './tsconfig.e2e.json',
  
  // 全局测试超时
  timeout: 30000,
  
  // 每个测试的超时
  expect: {
    timeout: 5000,
  },
  
  // 完全并行运行测试
  fullyParallel: true,
  
  // CI 上失败时禁止 retry
  retries: process.env.CI ? 2 : 0,
  
  // 限制并行 worker 数量，避免触发后端速率限制
  workers: process.env.CI ? 1 : 2,
  
  // Reporter 配置
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'test-results/results.json' }],
    ['list'],
  ],
  
  // 全局设置
  use: {
    // 基础 URL
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:3000',
    
    // 收集失败测试的 trace
    trace: 'on-first-retry',
    
    // 截图
    screenshot: 'only-on-failure',
    
    // 视频录制
    video: 'retain-on-failure',
    
    // 浏览器上下文选项
    contextOptions: {
      ignoreHTTPSErrors: true,
    },
  },
  
  // 项目配置 - 多浏览器测试
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    // 移动端测试
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],
  
  // 使用现有服务器（前端 dev server 在 3000，后端在 8080）
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: true,
    timeout: 30000,
  },
});