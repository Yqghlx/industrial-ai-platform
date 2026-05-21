import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E 测试配置 - 使用已启动的服务器
 */
export default defineConfig({
  testDir: './e2e',
  tsconfig: './tsconfig.e2e.json',
  timeout: 30000,
  expect: {
    timeout: 5000,
  },
  fullyParallel: true,
  retries: 0,
  workers: 1,
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'test-results/results.json' }],
    ['list'],
  ],
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    contextOptions: {
      ignoreHTTPSErrors: true,
    },
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  // 不启动开发服务器，使用已运行的静态服务器
});