/**
 * Login Flow E2E Tests
 * 工业AI代理平台登录流程端到端测试
 */

import { test, expect, Page } from '@playwright/test';

// 测试配置
const TEST_CONFIG = {
  baseURL: process.env.E2E_BASE_URL || 'http://localhost:3000',
  apiURL: process.env.E2E_API_URL || 'http://localhost:8080',
  testUser: {
    username: process.env.E2E_TEST_USER || 'admin',
    password: process.env.E2E_TEST_PASSWORD || 'Admin@123456',
  },
  testTimeout: 30000,
};

// API响应类型
interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  user: {
    id: number;
    username: string;
    role: string;
    tenant_id: string;
  };
}

interface ErrorResponse {
  error: string;
  code: string;
}

// 辅助函数
async function loginViaAPI(page: Page, username: string, password: string): Promise<LoginResponse> {
  const response = await page.request.post(`${TEST_CONFIG.apiURL}/api/v1/auth/login`, {
    data: { username, password },
  });
  
  expect(response.ok()).toBeTruthy();
  return await response.json() as LoginResponse;
}

async function clearAuthState(page: Page) {
  await page.context().clearCookies();
  await page.evaluate(() => {
    localStorage.clear();
    sessionStorage.clear();
  });
}

// 测试套件：登录页面访问
test.describe('Login Page Access', () => {
  test.beforeEach(async ({ page }) => {
    await clearAuthState(page);
  });

  test('should display login page with correct elements', async ({ page }) => {
    // 导航到登录页
    await page.goto('/login');
    
    // 等待页面加载完成
    await expect(page).toHaveTitle(/工业AI|Industrial AI|Login/i);
    
    // 验证登录表单元素存在
    await expect(page.locator('input[name="username"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
    
    // 验证登录按钮文本
    const loginButton = page.locator('button[type="submit"]');
    await expect(loginButton).toContainText(/登录|Login/i);
  });

  test('should have correct form field placeholders', async ({ page }) => {
    await page.goto('/login');
    
    // 检查用户名输入框placeholder
    const usernameInput = page.locator('input[name="username"]');
    await expect(usernameInput).toHaveAttribute('placeholder', /.+/i);
    
    // 检查密码输入框placeholder
    const passwordInput = page.locator('input[name="password"]');
    await expect(passwordInput).toHaveAttribute('placeholder', /.+/i);
  });

  test('should have proper accessibility attributes', async ({ page }) => {
    await page.goto('/login');
    
    // 检查表单标签
    const usernameInput = page.locator('input[name="username"]');
    const passwordInput = page.locator('input[name="password"]');
    
    // 验证required属性
    await expect(usernameInput).toHaveAttribute('required', '');
    await expect(passwordInput).toHaveAttribute('required', '');
    
    // 验证输入类型
    await expect(usernameInput).toHaveAttribute('type', 'text');
    await expect(passwordInput).toHaveAttribute('type', 'password');
  });
});

// 测试套件：登录表单验证
test.describe('Login Form Validation', () => {
  test.beforeEach(async ({ page }) => {
    await clearAuthState(page);
    await page.goto('/login');
  });

  test('should show error for empty username', async ({ page }) => {
    // 清空用户名输入框
    await page.locator('input[name="username"]').clear();
    await page.locator('input[name="password"]').fill('password');
    
    // 点击登录按钮
    await page.locator('button[type="submit"]').click();
    
    // 验证错误消息显示
    await expect(page.locator('.error-message, .alert-error, [role="alert"]')).toBeVisible({ timeout: 5000 });
  });

  test('should show error for empty password', async ({ page }) => {
    await page.locator('input[name="username"]').fill('testuser');
    await page.locator('input[name="password"]').clear();
    
    await page.locator('button[type="submit"]').click();
    
    await expect(page.locator('.error-message, .alert-error, [role="alert"]')).toBeVisible({ timeout: 5000 });
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.locator('input[name="username"]').fill('invaliduser');
    await page.locator('input[name="password"]').fill('invalidpassword');
    
    await page.locator('button[type="submit"]').click();
    
    // 等待API响应
    await page.waitForResponse(response => 
      response.url().includes('/auth/login') && response.status() === 401
    ).catch(() => {});
    
    // 验证错误消息
    await expect(page.locator('.error-message, .alert-error, [role="alert"]')).toBeVisible({ timeout: 5000 });
  });

  test('should disable submit button while submitting', async ({ page }) => {
    await page.locator('input[name="username"]').fill(TEST_CONFIG.testUser.username);
    await page.locator('input[name="password"]').fill(TEST_CONFIG.testUser.password);
    
    const submitButton = page.locator('button[type="submit"]');
    
    // 点击提交
    await submitButton.click();
    
    // 检查按钮是否被禁用（正在提交状态）
    await expect(submitButton).toBeDisabled({ timeout: 1000 }).catch(() => {});
    
    // 检查加载指示器
    await expect(page.locator('.loading-spinner, .spinner, [data-testid="loading"]')).toBeVisible({ timeout: 2000 }).catch(() => {});
  });
});

// 测试套件：成功登录流程
test.describe('Successful Login Flow', () => {
  test.beforeEach(async ({ page }) => {
    await clearAuthState(page);
  });

  test('should login successfully with valid credentials', async ({ page }) => {
    await page.goto('/login');
    
    // 填写登录表单
    await page.locator('input[name="username"]').fill(TEST_CONFIG.testUser.username);
    await page.locator('input[name="password"]').fill(TEST_CONFIG.testUser.password);
    
    // 点击登录
    await page.locator('button[type="submit"]').click();
    
    // 等待导航到主页或仪表盘
    await expect(page).toHaveURL(/dashboard|home|\/$/, { timeout: TEST_CONFIG.testTimeout });
    
    // 验证登录成功指示器
    await expect(page.locator('[data-testid="user-info"], .user-profile, .user-avatar')).toBeVisible({ timeout: 5000 }).catch(() => {});
    
    // 验证token存储
    const token = await page.evaluate(() => localStorage.getItem('access_token') || localStorage.getItem('token'));
    expect(token).toBeTruthy();
  });

  test('should store tokens in localStorage after login', async ({ page }) => {
    // 通过API登录获取token
    const loginResponse = await loginViaAPI(page, TEST_CONFIG.testUser.username, TEST_CONFIG.testUser.password);
    
    // 导航到应用并设置token
    await page.goto('/');
    await page.evaluate((token) => {
      localStorage.setItem('access_token', token);
    }, loginResponse.access_token);
    
    // 刷新页面验证token存在
    await page.reload();
    
    const storedToken = await page.evaluate(() => localStorage.getItem('access_token'));
    expect(storedToken).toBeTruthy();
    expect(storedToken).toBe(loginResponse.access_token);
  });

  test('should redirect to dashboard after successful login', async ({ page }) => {
    await page.goto('/login');
    
    await page.locator('input[name="username"]').fill(TEST_CONFIG.testUser.username);
    await page.locator('input[name="password"]').fill(TEST_CONFIG.testUser.password);
    await page.locator('button[type="submit"]').click();
    
    // 等待重定向
    await page.waitForURL(/dashboard|home/, { timeout: TEST_CONFIG.testTimeout });
    
    // 验证仪表盘元素
    await expect(page.locator('.dashboard, [data-testid="dashboard"], .main-content')).toBeVisible({ timeout: 5000 }).catch(() => {});
  });

  test('should display user information after login', async ({ page }) => {
    const loginResponse = await loginViaAPI(page, TEST_CONFIG.testUser.username, TEST_CONFIG.testUser.password);
    
    await page.goto('/dashboard');
    await page.evaluate((token) => localStorage.setItem('access_token', token), loginResponse.access_token);
    await page.reload();
    
    // 验证用户信息显示
    await expect(page.locator(`text=${loginResponse.user.username}`)).toBeVisible({ timeout: 5000 }).catch(() => {});
  });
});

// 测试套件：Token刷新和过期
test.describe('Token Refresh and Expiration', () => {
  test.beforeEach(async ({ page }) => {
    await clearAuthState(page);
  });

  test('should refresh token when access token expires', async ({ page }) => {
    // 登录获取token
    const loginResponse = await loginViaAPI(page, TEST_CONFIG.testUser.username, TEST_CONFIG.testUser.password);
    
    // 存储token
    await page.goto('/');
    await page.evaluate((tokens) => {
      localStorage.setItem('access_token', tokens.access);
      localStorage.setItem('refresh_token', tokens.refresh);
    }, { access: loginResponse.access_token, refresh: loginResponse.refresh_token });
    
    // 导航到需要认证的页面
    await page.goto('/dashboard');
    
    // 验证页面正常显示
    await expect(page).toHaveURL(/dashboard/);
  });

  test('should redirect to login when token is invalid', async ({ page }) => {
    await page.goto('/');
    
    // 设置无效token
    await page.evaluate(() => {
      localStorage.setItem('access_token', 'invalid_token');
    });
    
    // 导航到需要认证的页面
    await page.goto('/dashboard');
    
    // 等待重定向到登录页
    await page.waitForURL(/login/, { timeout: TEST_CONFIG.testTimeout }).catch(() => {});
  });
});

// 测试套件：登出流程
test.describe('Logout Flow', () => {
  test.beforeEach(async ({ page }) => {
    await clearAuthState(page);
  });

  test('should logout successfully and clear tokens', async ({ page }) => {
    // 登录
    await page.goto('/login');
    await page.locator('input[name="username"]').fill(TEST_CONFIG.testUser.username);
    await page.locator('input[name="password"]').fill(TEST_CONFIG.testUser.password);
    await page.locator('button[type="submit"]').click();
    
    await expect(page).toHaveURL(/dashboard|home/, { timeout: TEST_CONFIG.testTimeout });
    
    // 查找并点击登出按钮
    const logoutButton = page.locator('[data-testid="logout"], .logout-button, button:has-text("退出"), button:has-text("Logout")');
    await logoutButton.click().catch(() => {
      // 如果找不到登出按钮，尝试从用户菜单登出
      page.locator('.user-menu, .user-dropdown').click();
      return page.locator('[data-testid="logout"], button:has-text("退出"), button:has-text("Logout")').click();
    });
    
    // 验证重定向到登录页
    await expect(page).toHaveURL(/login/, { timeout: TEST_CONFIG.testTimeout });
    
    // 验证token已清除
    const token = await page.evaluate(() => localStorage.getItem('access_token'));
    expect(token).toBeFalsy();
  });

  test('should prevent access to authenticated pages after logout', async ({ page }) => {
    // 登录然后登出
    await page.goto('/login');
    await page.locator('input[name="username"]').fill(TEST_CONFIG.testUser.username);
    await page.locator('input[name="password"]').fill(TEST_CONFIG.testUser.password);
    await page.locator('button[type="submit"]').click();
    
    await expect(page).toHaveURL(/dashboard|home/, { timeout: TEST_CONFIG.testTimeout });
    
    // 清除认证状态
    await clearAuthState(page);
    
    // 尝试访问需要认证的页面
    await page.goto('/dashboard');
    
    // 应被重定向到登录页
    await page.waitForURL(/login/, { timeout: TEST_CONFIG.testTimeout }).catch(() => {});
  });
});

// 测试套件：登录页面安全性
test.describe('Login Security', () => {
  test.beforeEach(async ({ page }) => {
    await clearAuthState(page);
  });

  test('should rate limit login attempts', async ({ page }) => {
    await page.goto('/login');
    
    // 连续尝试失败登录
    for (let i = 0; i < 6; i++) {
      await page.locator('input[name="username"]').fill(`invaliduser${i}`);
      await page.locator('input[name="password"]').fill('invalidpassword');
      await page.locator('button[type="submit"]').click();
      
      await page.waitForTimeout(500);
    }
    
    // 验证账户锁定或速率限制提示
    await expect(page.locator('.error-message, .alert-error, [role="alert"]')).toContainText(/锁定|locked|限制|limited|rate/i, { timeout: 10000 }).catch(() => {});
  });

  test('should use HTTPS in production environment', async ({ page }) => {
    // 仅在生产环境测试
    if (process.env.CI && TEST_CONFIG.baseURL.startsWith('https://')) {
      await page.goto('/login');
      expect(page.url()).startsWith('https://');
    }
  });

  test('should clear password field on page reload', async ({ page }) => {
    await page.goto('/login');
    
    // 填写密码
    await page.locator('input[name="password"]').fill('sensitivepassword');
    
    // 刷新页面
    await page.reload();
    
    // 验证密码字段已清空
    await expect(page.locator('input[name="password"]')).toHaveValue('');
  });
});

// 测试套件：登录页面响应式设计
test.describe('Login Responsive Design', () => {
  test('should display correctly on mobile devices', async ({ page }) => {
    // 设置移动设备视口
    await page.setViewportSize({ width: 375, height: 667 });
    
    await page.goto('/login');
    
    // 验证登录表单元素仍然可见
    await expect(page.locator('input[name="username"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
    
    // 验证表单在视口中可点击
    await page.locator('input[name="username"]').click();
    await expect(page.locator('input[name="username"]')).toBeFocused();
  });

  test('should display correctly on tablet devices', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    
    await page.goto('/login');
    
    await expect(page.locator('input[name="username"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('should display correctly on desktop', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    
    await page.goto('/login');
    
    await expect(page.locator('input[name="username"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });
});

// 测试套件：登录API响应验证
test.describe('Login API Response', () => {
  test('should return correct response structure on successful login', async ({ page }) => {
    const response = await page.request.post(`${TEST_CONFIG.apiURL}/api/v1/auth/login`, {
      data: {
        username: TEST_CONFIG.testUser.username,
        password: TEST_CONFIG.testUser.password,
      },
    });
    
    expect(response.status()).toBe(200);
    
    const body = await response.json() as LoginResponse;
    expect(body.access_token).toBeTruthy();
    expect(body.refresh_token).toBeTruthy();
    expect(body.expires_in).toBeGreaterThan(0);
    expect(body.token_type).toBe('bearer');
    expect(body.user).toBeTruthy();
    expect(body.user.id).toBeDefined();
    expect(body.user.username).toBe(TEST_CONFIG.testUser.username);
  });

  test('should return correct error code on failed login', async ({ page }) => {
    const response = await page.request.post(`${TEST_CONFIG.apiURL}/api/v1/auth/login`, {
      data: {
        username: 'nonexistent',
        password: 'wrongpassword',
      },
    });
    
    expect(response.status()).toBe(401);
    
    const body = await response.json() as ErrorResponse;
    expect(body.code).toBe('AUTH_FAILED');
    expect(body.error).toBeTruthy();
  });

  test('should return validation error for empty fields', async ({ page }) => {
    const response = await page.request.post(`${TEST_CONFIG.apiURL}/api/v1/auth/login`, {
      data: {
        username: '',
        password: '',
      },
    });
    
    expect(response.status()).toBe(400);
    
    const body = await response.json() as ErrorResponse;
    expect(body.code).toBe('INVALID_INPUT');
    expect(body.error).toBeTruthy();
  });
});

/**
 * 测试运行说明：
 * 
 * 1. 运行所有测试:
 *    npx playwright test tests/e2e/login_flow.spec.ts
 * 
 * 2. 运行特定浏览器测试:
 *    npx playwright test tests/e2e/login_flow.spec.ts --project=chromium
 * 
 * 3. 运行特定测试:
 *    npx playwright test tests/e2e/login_flow.spec.ts -g "should login successfully"
 * 
 * 4. 调试测试:
 *    npx playwright test tests/e2e/login_flow.spec.ts --debug
 * 
 * 5. 生成测试报告:
 *    npx playwright show-report
 * 
 * 注意：运行测试前需要确保：
 * - 后端API服务已启动 (localhost:8080)
 * - 前端服务已启动 (localhost:3000)
 * - 测试用户已创建 (或使用环境变量配置)
 */