/**
 * Playwright E2E Test Configuration - Simplified
 * 工业AI代理平台端到端测试配置
 */

import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  // 测试目录 - 当前目录
  testDir: '.',
  
  // 测试文件匹配
  testMatch: '*.spec.ts',
  
  // 完全并行运行测试
  fullyParallel: true,
  
  // CI环境下失败时禁止test.only
  forbidOnly: !!process.env.CI,
  
  // CI环境下重试失败测试
  retries: process.env.CI ? 2 : 0,
  
  // 并行workers
  workers: 2,
  
  // 报告配置
  reporter: [
    ['list'],
    ['json', { outputFile: 'results.json' }]
  ],
  
  // 全局配置
  use: {
    // 基础URL - 需要前端服务运行
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:5173',
    
    // 测试超时
    actionTimeout: 10000,
    
    // 收集失败测试的trace
    trace: 'on-first-retries',
    
    // 截图配置
    screenshot: 'only-on-failure',
  },
  
  // 项目配置
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  
  // 超时配置
  timeout: 30000,
  expect: {
    timeout: 5000
  },
});