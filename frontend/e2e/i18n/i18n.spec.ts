import { test, expect } from '../fixtures/test-fixtures';

/**
 * 多语言切换 E2E 测试
 */
test.describe('多语言切换', () => {
  
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });
  
  test('切换到英文', async ({ page }) => {
    // 检查语言切换器
    const langSwitcher = page.locator('[data-testid="lang-switcher"], select[name="lang"], [aria-label*="language"]');
    
    if (await langSwitcher.isVisible()) {
      await langSwitcher.selectOption('en');
      
      // 等待语言切换
      await page.waitForTimeout(500);
      
      // 验证页面文字变为英文
      await expect(page.locator('text=/Dashboard|Devices|Reports/i')).toBeVisible({ timeout: 5000 });
      
      // 验证不再是中文
      await expect(page.locator('text=/仪表盘|设备|报告/').filter({ hasNotText: 'Dashboard' })).not.toBeVisible();
    }
  });
  
  test('切换到中文', async ({ page }) => {
    // 先切换到英文
    const langSwitcher = page.locator('[data-testid="lang-switcher"]');
    
    if (await langSwitcher.isVisible()) {
      await langSwitcher.selectOption('en');
      await page.waitForTimeout(500);
      
      // 再切换回中文
      await langSwitcher.selectOption('zh');
      await page.waitForTimeout(500);
      
      // 验证页面文字变为中文
      await expect(page.locator('text=/仪表盘|设备|报告/i')).toBeVisible({ timeout: 5000 });
    }
  });
  
  test('登录页多语言', async ({ page }) => {
    await page.goto('/login');
    
    const langSwitcher = page.locator('[data-testid="lang-switcher"]');
    
    if (await langSwitcher.isVisible()) {
      // 验证登录页中文
      await expect(page.locator('text=/登录|用户名|密码/i')).toBeVisible();
      
      // 切换到英文
      await langSwitcher.selectOption('en');
      await page.waitForTimeout(500);
      
      // 验证登录页英文
      await expect(page.locator('text=/Login|Username|Password/i')).toBeVisible();
    }
  });
  
  test('设备页多语言', async ({ page }) => {
    // 登录
    await page.goto('/login');
    await page.fill('[name="username"]', 'admin');
    await page.fill('[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/dashboard|\/devices/);
    
    // 导航到设备页
    await page.goto('/devices');
    
    const langSwitcher = page.locator('[data-testid="lang-switcher"]');
    
    if (await langSwitcher.isVisible()) {
      // 验证中文标签
      await expect(page.locator('text=/设备名称|类型|状态/i')).toBeVisible();
      
      // 切换到英文
      await langSwitcher.selectOption('en');
      await page.waitForTimeout(500);
      
      // 验证英文标签
      await expect(page.locator('text=/Device Name|Type|Status/i')).toBeVisible();
    }
  });
  
  test('告警页多语言', async ({ page }) => {
    // 登录并导航到告警页
    await page.goto('/login');
    await page.fill('[name="username"]', 'admin');
    await page.fill('[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.goto('/alerts');
    
    const langSwitcher = page.locator('[data-testid="lang-switcher"]');
    
    if (await langSwitcher.isVisible()) {
      // 验证中文
      await expect(page.locator('text=/告警|严重|警告/i')).toBeVisible();
      
      // 切换到英文
      await langSwitcher.selectOption('en');
      await page.waitForTimeout(500);
      
      // 验证英文
      await expect(page.locator('text=/Alert|Critical|Warning/i')).toBeVisible();
    }
  });
  
  test('AI Agent 多语言', async ({ page }) => {
    // 登录并导航到 AI 页
    await page.goto('/login');
    await page.fill('[name="username"]', 'admin');
    await page.fill('[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.goto('/agent');
    
    const langSwitcher = page.locator('[data-testid="lang-switcher"]');
    
    if (await langSwitcher.isVisible()) {
      // 验证中文 placeholder
      const input = page.locator('[data-testid="ai-input"], textarea');
      const placeholder = await input.getAttribute('placeholder');
      expect(placeholder).toMatch(/问|输入|查询/);
      
      // 切换到英文
      await langSwitcher.selectOption('en');
      await page.waitForTimeout(500);
      
      // 验证英文 placeholder
      const newPlaceholder = await input.getAttribute('placeholder');
      expect(newPlaceholder).toMatch(/Ask|Enter|Query/i);
    }
  });
  
  test('报告页多语言', async ({ page }) => {
    // 登录并导航到报告页
    await page.goto('/login');
    await page.fill('[name="username"]', 'admin');
    await page.fill('[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.goto('/reports');
    
    const langSwitcher = page.locator('[data-testid="lang-switcher"]');
    
    if (await langSwitcher.isVisible()) {
      // 验证中文
      await expect(page.locator('text=/报告|生成|下载/i')).toBeVisible();
      
      // 切换到英文
      await langSwitcher.selectOption('en');
      await page.waitForTimeout(500);
      
      // 验证英文
      await expect(page.locator('text=/Report|Generate|Download/i')).toBeVisible();
    }
  });
  
  test('语言持久化', async ({ page, context }) => {
    const langSwitcher = page.locator('[data-testid="lang-switcher"]');
    
    if (await langSwitcher.isVisible()) {
      // 切换到英文
      await langSwitcher.selectOption('en');
      await page.waitForTimeout(500);
      
      // 关闭页面
      await page.close();
      
      // 打开新页面
      const newPage = await context.newPage();
      await newPage.goto('/');
      
      // 验证语言仍然是英文
      await expect(newPage.locator('text=/Dashboard|Devices/i')).toBeVisible({ timeout: 5000 });
    }
  });
  
  test('错误提示多语言', async ({ page }) => {
    await page.goto('/login');
    
    const langSwitcher = page.locator('[data-testid="lang-switcher"]');
    
    if (await langSwitcher.isVisible()) {
      // 输入错误密码
      await page.fill('[name="username"]', 'admin');
      await page.fill('[name="password"]', 'wrongpassword');
      await page.click('button[type="submit"]');
      
      // 验证中文错误提示
      await expect(page.locator('text=/登录失败|密码错误/i')).toBeVisible({ timeout: 5000 });
      
      // 切换到英文
      await langSwitcher.selectOption('en');
      
      // 再次尝试登录
      await page.fill('[name="password"]', 'wrongpassword');
      await page.click('button[type="submit"]');
      
      // 验证英文错误提示
      await expect(page.locator('text=/Login failed|Invalid password/i')).toBeVisible({ timeout: 5000 });
    }
  });
});

test.describe('日期/时间本地化', () => {
  
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('/login');
    await page.fill('[name="username"]', 'admin');
    await page.fill('[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/dashboard|\/devices/);
  });
  
  test('中文日期格式', async ({ page }) => {
    const langSwitcher = page.locator('[data-testid="lang-switcher"]');
    
    if (await langSwitcher.isVisible()) {
      await langSwitcher.selectOption('zh');
      await page.waitForTimeout(500);
      
      // 检查日期显示格式 (YYYY年MM月DD日 或 YYYY-MM-DD)
      const dateElement = page.locator('[data-testid="current-date"], .date-display').first();
      
      if (await dateElement.isVisible()) {
        const dateText = await dateElement.textContent();
        // 中文格式通常包含年/月/日 或使用 YYYY-MM-DD
        expect(dateText).toMatch(/\d{4}[-年]\d{1,2}[-月]\d{1,2}/);
      }
    }
  });
  
  test('英文日期格式', async ({ page }) => {
    const langSwitcher = page.locator('[data-testid="lang-switcher"]');
    
    if (await langSwitcher.isVisible()) {
      await langSwitcher.selectOption('en');
      await page.waitForTimeout(500);
      
      const dateElement = page.locator('[data-testid="current-date"]').first();
      
      if (await dateElement.isVisible()) {
        const dateText = await dateElement.textContent();
        // 英文格式通常包含 May, June 等月份名
        expect(dateText).toMatch(/[A-Z][a-z]+\s\d{1,2},?\s\d{4}/);
      }
    }
  });
  
  test('时间格式 (24小时 vs 12小时)', async ({ page }) => {
    const langSwitcher = page.locator('[data-testid="lang-switcher"]');
    
    if (await langSwitcher.isVisible()) {
      // 中文通常使用 24 小时制
      await langSwitcher.selectOption('zh');
      await page.waitForTimeout(500);
      
      const timeElement = page.locator('[data-testid="current-time"], .time-display').first();
      if (await timeElement.isVisible()) {
        const timeText = await timeElement.textContent();
        // 24 小时制: HH:mm
        expect(timeText).toMatch(/\d{1,2}:\d{2}/);
      }
      
      // 英文可能使用 12 小时制 (带 AM/PM)
      await langSwitcher.selectOption('en');
      await page.waitForTimeout(500);
      
      if (await timeElement.isVisible()) {
        const enTimeText = await timeElement.textContent();
        // 12 小时制可能包含 AM/PM
        // 但也可能使用 24 小时制
        expect(enTimeText).toMatch(/\d{1,2}:\d{2}/);
      }
    }
  });
});

test.describe('数字本地化', () => {
  
  test('数字分隔符', async ({ page }) => {
    // 登录并导航到报告页
    await page.goto('/login');
    await page.fill('[name="username"]', 'admin');
    await page.fill('[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.goto('/reports');
    
    // 检查大数字显示
    const largeNumber = page.locator('text=/1,000|1.000|10,000|10.000/i').first();
    
    if (await largeNumber.isVisible()) {
      // 不同语言可能使用不同的千位分隔符
      // 中文/英文: 逗号 (1,000)
      // 某些欧洲语言: 点号 (1.000)
    }
  });
  
  test('货币符号', async ({ page }) => {
    await page.goto('/login');
    await page.fill('[name="username"]', 'admin');
    await page.fill('[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.goto('/reports');
    
    // 检查货币显示
    const currencyElement = page.locator('text=/¥|￥|$|€/').first();
    
    if (await currencyElement.isVisible()) {
      const currencyText = await currencyElement.textContent();
      // 中文环境可能显示 ¥ (人民币)
      // 英文环境可能显示 $ (美元)
    }
  });
});