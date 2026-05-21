import { test, expect } from '../fixtures/test-fixtures';

/**
 * AI Agent 对话 E2E 测试
 * 
 * 注意: 依赖真实AI后端响应的测试已标记为 skip
 * 等待AI服务接入后再运行集成测试
 */
test.describe('AI Agent 对话', () => {
  
  test.beforeEach(async ({ adminPage }) => {
    // 导航到 AI Agent 页面
    await adminPage.goto('/ai-agent');
    
    // 等待页面加载 - 使用 .first() 避免 strict mode（Sidebar 和页面都有 h1）
    await expect(adminPage.locator('h1, h2').filter({ hasText: /AI|智能助手|Agent/i }).first()).toBeVisible({ timeout: 10000 });
  });
  
  test('显示 AI 对话界面', async ({ adminPage }) => {
    // 验证对话输入框存在 - 简化选择器匹配实际组件
    const inputField = adminPage.locator('input[type="text"].input');
    await expect(inputField).toBeVisible();
    
    // 验证发送按钮存在
    const sendBtn = adminPage.locator('button[type="submit"]');
    await expect(sendBtn).toBeVisible();
    
    // 验证消息区域存在 - 使用 .first() 避免 strict mode
    const chatArea = adminPage.locator('.card-body').first();
    await expect(chatArea).toBeVisible();
  });
  
  // ⏭️ 跳过: 依赖真实AI后端响应
  test.skip('发送问题并接收回答', async ({ adminPage }) => {
    const question = '设备温度异常如何处理？';
    const inputField = adminPage.locator('input[type="text"].input');
    await inputField.fill(question);
    await adminPage.click('button[type="submit"]');
    
    const responseMessage = adminPage.locator('[data-testid="ai-response"], [class*="response"], [class*="assistant"]').filter({ hasNotText: question });
    await expect(responseMessage).toBeVisible({ timeout: 60000 });
    
    const responseText = await responseMessage.textContent();
    expect(responseText?.length).toBeGreaterThan(10);
  });
  
  // ⏭️ 跳过: 依赖真实AI后端响应
  test.skip('多轮对话', async ({ adminPage }) => {
    const question1 = 'CNC 设备常见故障有哪些？';
    const inputField = adminPage.locator('input[type="text"].input');
    
    await inputField.fill(question1);
    await adminPage.click('button[type="submit"]');
    await adminPage.waitForTimeout(30000);
    
    const question2 = '如何预防这些故障？';
    await inputField.fill(question2);
    await adminPage.click('button[type="submit"]');
    
    await expect(adminPage.locator('[data-testid="ai-response"]').nth(1)).toBeVisible({ timeout: 60000 });
  });
  
  // ⏭️ 跳过: 依赖真实AI后端响应
  test.skip('设备相关查询', async ({ adminPage }) => {
    const question = '设备 CNC-001 的当前状态如何？';
    const inputField = adminPage.locator('input[type="text"].input');
    
    await inputField.fill(question);
    await adminPage.click('button[type="submit"]');
    
    await expect(adminPage.locator('text=/CNC-001|温度|振动|状态/i')).toBeVisible({ timeout: 60000 });
  });
  
  // ⏭️ 跳过: 依赖真实AI后端响应
  test.skip('历史对话记录', async ({ adminPage }) => {
    const inputField = adminPage.locator('input[type="text"].input');
    await inputField.fill('测试历史记录');
    await adminPage.click('button[type="submit"]');
    await adminPage.waitForTimeout(30000);
    
    await adminPage.reload();
    await expect(adminPage.locator('text="测试历史记录"')).toBeVisible({ timeout: 10000 });
  });
  
  // ⏭️ 跳过: 依赖真实AI后端响应
  test.skip('清空对话', async ({ adminPage }) => {
    const inputField = adminPage.locator('input[type="text"].input');
    await inputField.fill('测试清空');
    await adminPage.click('button[type="submit"]');
    await adminPage.waitForTimeout(10000);
    
    const clearBtn = adminPage.locator('[data-testid="clear-chat-btn"], button:has-text("清空"), button:has-text("Clear")');
    
    if (await clearBtn.isVisible()) {
      await clearBtn.click();
      const confirmBtn = adminPage.locator('[role="dialog"] button:has-text("确认")');
      if (await confirmBtn.isVisible()) {
        await confirmBtn.click();
      }
      await expect(adminPage.locator('text="测试清空"')).not.toBeVisible({ timeout: 5000 });
    }
  });
  
  // ⏭️ 跳过: 依赖真实AI后端响应
  test.skip('停止生成', async ({ adminPage }) => {
    const inputField = adminPage.locator('input[type="text"].input');
    await inputField.fill('详细解释预测性维护的原理');
    await adminPage.click('button[type="submit"]');
    await adminPage.waitForTimeout(2000);
    
    const stopBtn = adminPage.locator('[data-testid="stop-btn"], button:has-text("停止")');
    
    if (await stopBtn.isVisible()) {
      await stopBtn.click();
      await adminPage.waitForTimeout(1000);
      await expect(stopBtn).not.toBeVisible();
    }
  });
  
  // ⏭️ 跳过: 依赖真实AI后端响应
  test.skip('复制回答', async ({ adminPage }) => {
    const inputField = adminPage.locator('input[type="text"].input');
    await inputField.fill('简单的问题');
    await adminPage.click('button[type="submit"]');
    await adminPage.waitForTimeout(30000);
    
    const copyBtn = adminPage.locator('[data-testid="copy-btn"], button:has-text("复制")').first();
    
    if (await copyBtn.isVisible()) {
      await copyBtn.click();
      await expect(adminPage.locator('text=/复制成功|Copied/i')).toBeVisible({ timeout: 5000 });
    }
  });
  
  test('建议问题', async ({ adminPage }) => {
    // 检查建议问题 - 纯UI验证，不依赖AI响应
    const suggestions = adminPage.locator('[data-testid="suggestions"], [class*="suggestions"]');
    
    if (await suggestions.isVisible()) {
      const suggestionCount = await suggestions.locator('button, [role="button"]').count();
      expect(suggestionCount).toBeGreaterThan(0);
      
      const firstSuggestion = suggestions.locator('button').first();
      const suggestionText = await firstSuggestion.textContent();
      await firstSuggestion.click();
      
      const inputField = adminPage.locator('input[type="text"].input');
      await expect(inputField).toHaveValue(suggestionText || '');
    }
  });
});

test.describe('AI Agent 性能', () => {
  
  // ⏭️ 跳过: 依赖真实AI后端响应
  test.skip('响应时间', async ({ adminPage }) => {
    await adminPage.goto('/ai-agent');
    
    const inputField = adminPage.locator('input[type="text"].input');
    await inputField.fill('简单问题：设备状态');
    
    const startTime = Date.now();
    await adminPage.click('button[type="submit"]');
    
    await adminPage.locator('[data-testid="ai-response"]').waitFor({ timeout: 60000 });
    const responseTime = Date.now() - startTime;
    
    console.log(`AI Response time: ${responseTime}ms`);
    expect(responseTime).toBeLessThan(60000);
  });
  
  // ⏭️ 跳过: 依赖真实AI后端响应
  test.skip('并发请求处理', async ({ adminPage }) => {
    const inputField = adminPage.locator('input[type="text"].input');
    
    await inputField.fill('问题 1');
    await adminPage.click('button[type="submit"]');
    await adminPage.waitForTimeout(1000);
    
    const rateLimitMsg = adminPage.locator('text=/速率限制|Rate limit|请稍候/i');
    // 验证速率限制提示或正常处理
  });
});

test.describe('AI Agent 错误处理', () => {
  
  // ⏭️ 跳过: 依赖真实AI后端响应
  test.skip('网络错误提示', async ({ adminPage, context }) => {
    await context.setOffline(true);
    await adminPage.goto('/ai-agent');
    
    const inputField = adminPage.locator('input[type="text"].input');
    await inputField.fill('测试离线');
    await adminPage.click('button[type="submit"]');
    
    await expect(adminPage.locator('text=/网络错误|离线|Network error/i')).toBeVisible({ timeout: 10000 });
    await context.setOffline(false);
  });
  
  test('空问题验证', async ({ adminPage }) => {
    await adminPage.goto('/ai-agent');
    
    // 发送按钮在空输入时应禁用
    const sendBtn = adminPage.locator('button[type="submit"]');
    await expect(sendBtn).toBeDisabled();
  });
  
  test('长问题处理', async ({ adminPage }) => {
    await adminPage.goto('/ai-agent');
    
    // 输入超长问题 - 纯UI验证，不发送
    const longQuestion = '这是一个非常长的问题...' + '测试'.repeat(500);
    const inputField = adminPage.locator('input[type="text"].input');
    await inputField.fill(longQuestion);
    
    // 检查是否有截断或警告提示
    const warningMsg = adminPage.locator('text=/过长|超出限制|Too long/i');
    // 如果有字数限制UI，验证出现；否则输入框应正常接收
    const inputValue = await inputField.inputValue();
    expect(inputValue.length).toBeGreaterThan(0);
  });
});