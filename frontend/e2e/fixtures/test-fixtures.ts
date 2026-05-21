import { test as base, expect, Page, BrowserContext } from '@playwright/test';

/**
 * 测试用户凭据
 */
export const TEST_USERS = {
  admin: {
    username: 'admin',
    password: process.env.E2E_ADMIN_PASSWORD || 'Admin@123456',
  },
  operator: {
    username: 'operator',
    password: process.env.E2E_OPERATOR_PASSWORD || 'Operator@123',
  },
  viewer: {
    username: 'viewer',
    password: 'Viewer@123',
  },
};

/**
 * 测试设备数据
 * type 值必须匹配后端 ValidDeviceTypes 枚举和 DeviceManager.tsx option value
 */
export const TEST_DEVICES = {
  cnc: {
    id: 'test-cnc-001',
    name: 'Test CNC Machine',
    type: 'CNC',
    location: 'Factory A - Line 1',
  },
  inj: {
    id: 'test-inj-001',
    name: 'Test Injection Molder',
    type: 'InjectionMolder',
    location: 'Factory B - Workshop 2',
  },
};

/**
 * 测试告警规则数据
 * 字段名和值必须匹配 RuleManager.tsx 表单的 name 属性和 option value
 */
export const TEST_RULES = {
  temperature: {
    name: 'Test Temperature Rule',
    device_type: 'sensor',
    metric: 'temperature',
    operator: '>',
    threshold: '100',
    severity: 'high',
  },
  vibration: {
    name: 'Test Vibration Rule',
    device_type: 'motor',
    metric: 'vibration',
    operator: '>',
    threshold: '5',
    severity: 'critical',
  },
};

/**
 * 自定义 fixtures
 */
type MyFixtures = {
  authenticatedPage: Page;
  adminPage: Page;
};

export const test = base.extend<MyFixtures>({
  // 已登录的页面 (operator 用户)
  authenticatedPage: async ({ page }, use) => {
    await login(page, TEST_USERS.operator);
    await use(page);
  },

  // 管理员页面
  adminPage: async ({ page }, use) => {
    await login(page, TEST_USERS.admin);
    await use(page);
  },
});

/**
 * 登录辅助函数
 */
export async function login(page: Page, user: { username: string; password: string }) {
  await page.goto('/login');
  await page.fill('[name="username"]', user.username);
  await page.fill('[name="password"]', user.password);
  await page.click('button[type="submit"]');

  // 等待登录成功并跳转
  await page.waitForURL(url => {
    const path = url.pathname;
    return path === '/' || path === '/dashboard' || path.startsWith('/devices');
  }, { timeout: 15000 });

  // 验证不在登录页
  await expect(page).not.toHaveURL(/\/login/);
}

/**
 * 登出辅助函数
 */
export async function logout(page: Page) {
  await page.locator('[data-testid="user-menu"]').first().click();
  await page.locator('[data-testid="logout-btn"]').first().click();
  await page.waitForURL('/login');
}

/**
 * 等待 API 响应
 */
export async function waitForAPIResponse(page: Page, urlPattern: string | RegExp) {
  return page.waitForResponse(
    (response) => response.url().match(urlPattern) && response.status() === 200
  );
}

/**
 * 模拟设备遥测数据
 */
export async function sendMockTelemetry(deviceId: string, metrics: Record<string, number>) {
  const response = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/devices/telemetry`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      device_id: deviceId,
      device_type: 'CNC',
      timestamp: new Date().toISOString(),
      metrics,
      status: 'running',
    }),
  });
  return response.ok;
}

/**
 * 创建测试设备 — 通过 UI 交互
 * 对应 DeviceManager.tsx 的创建弹窗表单
 */
export async function createTestDevice(page: Page, device: { id: string; name: string; type: string; location?: string }) {
  await page.goto('/devices');
  // 等待设备列表加载完成
  await expect(page.locator('[data-testid="device-table"]')).toBeVisible({ timeout: 10000 });

  // 点击创建按钮
  await page.click('[data-testid="add-device-btn"]');

  // 等待弹窗渲染 — 弹窗有 role="dialog"
  await expect(page.locator('[role="dialog"]')).toBeVisible({ timeout: 5000 });

  // 填写表单（字段名对应 DeviceManager.tsx 中的 name 属性）
  await page.fill('[name="device-id"]', device.id);
  await page.fill('[name="device-name"]', device.name);
  await page.selectOption('[name="device-type"]', device.type);
  if (device.location) {
    await page.fill('[name="device-location"]', device.location);
  }

  // 提交表单 — 等待 API 响应
  const responsePromise = page.waitForResponse(
    resp => resp.url().includes('/api/v1/devices') && (resp.request().method() === 'POST' || resp.request().method() === 'PUT'),
    { timeout: 10000 }
  ).catch(() => null);

  await page.click('[role="dialog"] button[type="submit"]');

  // 等待 API 响应（成功或失败）
  const resp = await responsePromise;

  // 等待弹窗关闭
  await expect(page.locator('[role="dialog"]')).not.toBeVisible({ timeout: 5000 });

  // 等待设备出现在列表中
  await expect(page.locator(`text="${device.name}"`)).toBeVisible({ timeout: 10000 });
}

/**
 * 删除测试设备 — 通过 UI 交互
 * DeviceManager 使用原生 confirm() 弹窗
 */
export async function deleteTestDevice(page: Page, deviceId: string) {
  await page.goto('/devices');

  // 等待列表加载
  await expect(page.locator('[data-testid="device-table"]')).toBeVisible({ timeout: 10000 });

  const row = page.locator(`[data-testid="device-row-${deviceId}"]`);
  if (!(await row.isVisible())) {
    return; // 设备不存在，无需删除
  }

  // 监听原生 confirm() 弹窗并接受
  page.once('dialog', async dialog => {
    await dialog.accept();
  });

  // 点击删除按钮
  await row.locator('[data-testid="delete-btn"]').click();

  // 等待行消失
  await expect(row).not.toBeVisible({ timeout: 5000 });
}

/**
 * 清理所有 test- 前缀的测试设备
 */
export async function cleanupTestData(page: Page) {
  await page.goto('/devices');

  // 等待表格加载
  try {
    await expect(page.locator('[data-testid="device-table"]')).toBeVisible({ timeout: 8000 });
  } catch {
    return; // 表格未加载，无法清理
  }

  const testRows = await page.locator('[data-testid^="device-row-test-"]').all();

  for (const row of testRows) {
    try {
      if (!(await row.isVisible())) continue;

      const deleteBtn = row.locator('[data-testid="delete-btn"]');
      if (!(await deleteBtn.isVisible())) continue;

      // 每次删除都需要处理 confirm 弹窗
      page.once('dialog', async dialog => {
        await dialog.accept();
      });

      await deleteBtn.click();
      await page.waitForTimeout(500);
    } catch {
      // 忽略已消失的行
    }
  }
}

// 导出 expect 供测试使用
export { expect };
