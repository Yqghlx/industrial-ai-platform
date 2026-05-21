# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: login_flow.spec.ts >> Login Responsive Design >> should display correctly on tablet devices
- Location: login_flow.spec.ts:392:7

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5173/login
Call log:
  - navigating to "http://localhost:5173/login", waiting until "load"

```

# Test source

```ts
  295 |     const logoutButton = page.locator('[data-testid="logout"], .logout-button, button:has-text("退出"), button:has-text("Logout")');
  296 |     await logoutButton.click().catch(() => {
  297 |       // 如果找不到登出按钮，尝试从用户菜单登出
  298 |       page.locator('.user-menu, .user-dropdown').click();
  299 |       return page.locator('[data-testid="logout"], button:has-text("退出"), button:has-text("Logout")').click();
  300 |     });
  301 |     
  302 |     // 验证重定向到登录页
  303 |     await expect(page).toHaveURL(/login/, { timeout: TEST_CONFIG.testTimeout });
  304 |     
  305 |     // 验证token已清除
  306 |     const token = await page.evaluate(() => localStorage.getItem('access_token'));
  307 |     expect(token).toBeFalsy();
  308 |   });
  309 | 
  310 |   test('should prevent access to authenticated pages after logout', async ({ page }) => {
  311 |     // 登录然后登出
  312 |     await page.goto('/login');
  313 |     await page.locator('input[name="username"]').fill(TEST_CONFIG.testUser.username);
  314 |     await page.locator('input[name="password"]').fill(TEST_CONFIG.testUser.password);
  315 |     await page.locator('button[type="submit"]').click();
  316 |     
  317 |     await expect(page).toHaveURL(/dashboard|home/, { timeout: TEST_CONFIG.testTimeout });
  318 |     
  319 |     // 清除认证状态
  320 |     await clearAuthState(page);
  321 |     
  322 |     // 尝试访问需要认证的页面
  323 |     await page.goto('/dashboard');
  324 |     
  325 |     // 应被重定向到登录页
  326 |     await page.waitForURL(/login/, { timeout: TEST_CONFIG.testTimeout }).catch(() => {});
  327 |   });
  328 | });
  329 | 
  330 | // 测试套件：登录页面安全性
  331 | test.describe('Login Security', () => {
  332 |   test.beforeEach(async ({ page }) => {
  333 |     await clearAuthState(page);
  334 |   });
  335 | 
  336 |   test('should rate limit login attempts', async ({ page }) => {
  337 |     await page.goto('/login');
  338 |     
  339 |     // 连续尝试失败登录
  340 |     for (let i = 0; i < 6; i++) {
  341 |       await page.locator('input[name="username"]').fill(`invaliduser${i}`);
  342 |       await page.locator('input[name="password"]').fill('invalidpassword');
  343 |       await page.locator('button[type="submit"]').click();
  344 |       
  345 |       await page.waitForTimeout(500);
  346 |     }
  347 |     
  348 |     // 验证账户锁定或速率限制提示
  349 |     await expect(page.locator('.error-message, .alert-error, [role="alert"]')).toContainText(/锁定|locked|限制|limited|rate/i, { timeout: 10000 }).catch(() => {});
  350 |   });
  351 | 
  352 |   test('should use HTTPS in production environment', async ({ page }) => {
  353 |     // 仅在生产环境测试
  354 |     if (process.env.CI && TEST_CONFIG.baseURL.startsWith('https://')) {
  355 |       await page.goto('/login');
  356 |       expect(page.url()).startsWith('https://');
  357 |     }
  358 |   });
  359 | 
  360 |   test('should clear password field on page reload', async ({ page }) => {
  361 |     await page.goto('/login');
  362 |     
  363 |     // 填写密码
  364 |     await page.locator('input[name="password"]').fill('sensitivepassword');
  365 |     
  366 |     // 刷新页面
  367 |     await page.reload();
  368 |     
  369 |     // 验证密码字段已清空
  370 |     await expect(page.locator('input[name="password"]')).toHaveValue('');
  371 |   });
  372 | });
  373 | 
  374 | // 测试套件：登录页面响应式设计
  375 | test.describe('Login Responsive Design', () => {
  376 |   test('should display correctly on mobile devices', async ({ page }) => {
  377 |     // 设置移动设备视口
  378 |     await page.setViewportSize({ width: 375, height: 667 });
  379 |     
  380 |     await page.goto('/login');
  381 |     
  382 |     // 验证登录表单元素仍然可见
  383 |     await expect(page.locator('input[name="username"]')).toBeVisible();
  384 |     await expect(page.locator('input[name="password"]')).toBeVisible();
  385 |     await expect(page.locator('button[type="submit"]')).toBeVisible();
  386 |     
  387 |     // 验证表单在视口中可点击
  388 |     await page.locator('input[name="username"]').click();
  389 |     await expect(page.locator('input[name="username"]')).toBeFocused();
  390 |   });
  391 | 
  392 |   test('should display correctly on tablet devices', async ({ page }) => {
  393 |     await page.setViewportSize({ width: 768, height: 1024 });
  394 |     
> 395 |     await page.goto('/login');
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5173/login
  396 |     
  397 |     await expect(page.locator('input[name="username"]')).toBeVisible();
  398 |     await expect(page.locator('input[name="password"]')).toBeVisible();
  399 |     await expect(page.locator('button[type="submit"]')).toBeVisible();
  400 |   });
  401 | 
  402 |   test('should display correctly on desktop', async ({ page }) => {
  403 |     await page.setViewportSize({ width: 1280, height: 720 });
  404 |     
  405 |     await page.goto('/login');
  406 |     
  407 |     await expect(page.locator('input[name="username"]')).toBeVisible();
  408 |     await expect(page.locator('input[name="password"]')).toBeVisible();
  409 |     await expect(page.locator('button[type="submit"]')).toBeVisible();
  410 |   });
  411 | });
  412 | 
  413 | // 测试套件：登录API响应验证
  414 | test.describe('Login API Response', () => {
  415 |   test('should return correct response structure on successful login', async ({ page }) => {
  416 |     const response = await page.request.post(`${TEST_CONFIG.apiURL}/api/v1/auth/login`, {
  417 |       data: {
  418 |         username: TEST_CONFIG.testUser.username,
  419 |         password: TEST_CONFIG.testUser.password,
  420 |       },
  421 |     });
  422 |     
  423 |     expect(response.status()).toBe(200);
  424 |     
  425 |     const body = await response.json() as LoginResponse;
  426 |     expect(body.access_token).toBeTruthy();
  427 |     expect(body.refresh_token).toBeTruthy();
  428 |     expect(body.expires_in).toBeGreaterThan(0);
  429 |     expect(body.token_type).toBe('bearer');
  430 |     expect(body.user).toBeTruthy();
  431 |     expect(body.user.id).toBeDefined();
  432 |     expect(body.user.username).toBe(TEST_CONFIG.testUser.username);
  433 |   });
  434 | 
  435 |   test('should return correct error code on failed login', async ({ page }) => {
  436 |     const response = await page.request.post(`${TEST_CONFIG.apiURL}/api/v1/auth/login`, {
  437 |       data: {
  438 |         username: 'nonexistent',
  439 |         password: 'wrongpassword',
  440 |       },
  441 |     });
  442 |     
  443 |     expect(response.status()).toBe(401);
  444 |     
  445 |     const body = await response.json() as ErrorResponse;
  446 |     expect(body.code).toBe('AUTH_FAILED');
  447 |     expect(body.error).toBeTruthy();
  448 |   });
  449 | 
  450 |   test('should return validation error for empty fields', async ({ page }) => {
  451 |     const response = await page.request.post(`${TEST_CONFIG.apiURL}/api/v1/auth/login`, {
  452 |       data: {
  453 |         username: '',
  454 |         password: '',
  455 |       },
  456 |     });
  457 |     
  458 |     expect(response.status()).toBe(400);
  459 |     
  460 |     const body = await response.json() as ErrorResponse;
  461 |     expect(body.code).toBe('INVALID_INPUT');
  462 |     expect(body.error).toBeTruthy();
  463 |   });
  464 | });
  465 | 
  466 | /**
  467 |  * 测试运行说明：
  468 |  * 
  469 |  * 1. 运行所有测试:
  470 |  *    npx playwright test tests/e2e/login_flow.spec.ts
  471 |  * 
  472 |  * 2. 运行特定浏览器测试:
  473 |  *    npx playwright test tests/e2e/login_flow.spec.ts --project=chromium
  474 |  * 
  475 |  * 3. 运行特定测试:
  476 |  *    npx playwright test tests/e2e/login_flow.spec.ts -g "should login successfully"
  477 |  * 
  478 |  * 4. 调试测试:
  479 |  *    npx playwright test tests/e2e/login_flow.spec.ts --debug
  480 |  * 
  481 |  * 5. 生成测试报告:
  482 |  *    npx playwright show-report
  483 |  * 
  484 |  * 注意：运行测试前需要确保：
  485 |  * - 后端API服务已启动 (localhost:8080)
  486 |  * - 前端服务已启动 (localhost:3000)
  487 |  * - 测试用户已创建 (或使用环境变量配置)
  488 |  */
```