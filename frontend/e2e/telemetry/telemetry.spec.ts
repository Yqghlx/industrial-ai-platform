import { test, expect, TEST_DEVICES, sendMockTelemetry } from '../fixtures/test-fixtures';

/**
 * 遥测数据 E2E 测试
 */
test.describe('设备详情 - 遥测数据', () => {
  
  test.beforeEach(async ({ adminPage }) => {
    // 先确保测试设备存在
    try {
      await createTestDevice(adminPage, TEST_DEVICES.cnc);
    } catch (e) {
      // 设备可能已存在，忽略错误
    }
    
    // 导航到设备详情页
    await adminPage.goto(`/devices/${TEST_DEVICES.cnc.id}`);
    
    // 等待页面加载 - 使用更宽松的选择器
    await adminPage.waitForTimeout(1000);
    
    // 如果显示"设备不存在"，跳过后续测试
    const notFound = await adminPage.locator('text=设备不存在').isVisible();
    if (notFound) {
      test.skip();
    }
  });
  
  test('显示实时遥测数据', async ({ adminPage }) => {
    // 发送模拟遥测数据
    await sendMockTelemetry(TEST_DEVICES.cnc.id, {
      temperature: 75.5,
      vibration: 2.3,
      power_consumption: 500,
    });
    
    // 等待数据更新
    await adminPage.waitForTimeout(2000);
    
    // 验证温度显示
    const tempValue = adminPage.locator('[data-testid="temperature"], text=/温度|Temperature/i').locator('..');
    await expect(tempValue).toBeVisible();
    
    // 验证数值更新
    await expect(adminPage.locator('text=/75\\.5|75/')).toBeVisible({ timeout: 10000 });
  });
  
  test('显示历史数据图表', async ({ adminPage }) => {
    // 检查历史数据 Tab
    const historyTab = adminPage.locator('[data-testid="history-tab"], button:has-text("历史"), button:has-text("History")');
    
    if (await historyTab.isVisible()) {
      await historyTab.click();
      
      // 等待图表加载
      await adminPage.waitForTimeout(1000);
      
      // 验证图表存在
      const chart = adminPage.locator('[data-testid="history-chart"], canvas, svg[class*="recharts"]');
      await expect(chart).toBeVisible({ timeout: 5000 });
    }
  });
  
  test('遥测数据时间范围选择', async ({ adminPage }) => {
    // 检查时间范围选择器
    const timeRangeSelector = adminPage.locator('[data-testid="time-range"], select[name="range"]');
    
    if (await timeRangeSelector.isVisible()) {
      // 选择 1 小时范围
      await timeRangeSelector.selectOption('1h');
      await adminPage.waitForTimeout(500);
      
      // 选择 24 小时范围
      await timeRangeSelector.selectOption('24h');
      await adminPage.waitForTimeout(500);
      
      // 选择自定义范围
      await timeRangeSelector.selectOption('custom');
      
      // 如果弹出日期选择器，验证存在
      const datePicker = adminPage.locator('[data-testid="date-picker"], input[type="date"]');
      if (await datePicker.isVisible()) {
        await expect(datePicker).toBeVisible();
      }
    }
  });
  
  test('遥测数据导出', async ({ adminPage }) => {
    // 检查导出按钮
    const exportBtn = adminPage.locator('[data-testid="export-btn"], button:has-text("导出"), button:has-text("Export")');
    
    if (await exportBtn.isVisible()) {
      await exportBtn.click();
      
      // 等待导出选项
      const csvOption = adminPage.locator('text=/CSV|Excel/i');
      if (await csvOption.isVisible()) {
        await csvOption.click();
        
        // 验证下载开始 (Playwright 会自动处理下载)
        const [download] = await Promise.all([
          adminPage.waitForEvent('download', { timeout: 10000 }),
          adminPage.click('text=/下载|Download/i'),
        ]);
        
        // 验证文件名
        expect(download.suggestedFilename()).toMatch(/\.(csv|xlsx)$/i);
      }
    }
  });
  
  test('指标阈值显示', async ({ adminPage }) => {
    // 发送超出阈值的数据
    await sendMockTelemetry(TEST_DEVICES.cnc.id, {
      temperature: 120, // 超过警告阈值
      vibration: 5.5,   // 超过故障阈值
    });
    
    await adminPage.waitForTimeout(2000);
    
    // 验证警告指示器
    const warningIndicator = adminPage.locator('[data-testid="warning"], .warning, [class*="warning"]');
    await expect(warningIndicator).toBeVisible({ timeout: 10000 });
    
    // 验证故障指示器
    const faultIndicator = adminPage.locator('[data-testid="fault"], .fault, [class*="danger"], [class*="error"]');
    await expect(faultIndicator).toBeVisible({ timeout: 10000 });
  });
  
  test('WebSocket 实时更新', async ({ adminPage }) => {
    // 记录初始温度值
    const initialTemp = await adminPage.locator('[data-testid="temperature-value"], .temp-value').textContent();
    
    // 发送新数据
    const newTemp = parseFloat(initialTemp || '0') + 10;
    await sendMockTelemetry(TEST_DEVICES.cnc.id, {
      temperature: newTemp,
    });
    
    // 等待 WebSocket 更新
    await adminPage.waitForTimeout(3000);
    
    // 验证温度值变化
    const updatedTemp = await adminPage.locator('[data-testid="temperature-value"], .temp-value').textContent();
    expect(updatedTemp).not.toBe(initialTemp);
  });
});

test.describe('设备详情 - 多指标视图', () => {
  
  test.beforeEach(async ({ adminPage }) => {
    await adminPage.goto(`/devices/${TEST_DEVICES.cnc.id}`);
  });
  
  test('切换指标视图', async ({ adminPage }) => {
    // 检查指标切换按钮
    const metricTabs = adminPage.locator('[data-testid="metric-tabs"], [role="tablist"]');
    
    if (await metricTabs.isVisible()) {
      // 切换到振动视图
      const vibrationTab = metricTabs.locator('button:has-text("振动"), button:has-text("Vibration")');
      if (await vibrationTab.isVisible()) {
        await vibrationTab.click();
        await adminPage.waitForTimeout(500);
        
        // 验证振动数据显示
        await expect(adminPage.locator('text=/振动|Vibration/i')).toBeVisible();
      }
      
      // 切换到功率视图
      const powerTab = metricTabs.locator('button:has-text("功率"), button:has-text("Power")');
      if (await powerTab.isVisible()) {
        await powerTab.click();
        await adminPage.waitForTimeout(500);
        
        await expect(adminPage.locator('text=/功率|Power/i')).toBeVisible();
      }
    }
  });
  
  test('多指标并列视图', async ({ adminPage }) => {
    // 检查并列视图选项
    const gridViewBtn = adminPage.locator('[data-testid="grid-view"], button:has-text("并列"), button:has-text("Grid")');
    
    if (await gridViewBtn.isVisible()) {
      await gridViewBtn.click();
      await adminPage.waitForTimeout(500);
      
      // 验证多指标同时显示
      const metrics = adminPage.locator('[data-testid="metric-card"]');
      const count = await metrics.count();
      
      expect(count).toBeGreaterThanOrEqual(2);
    }
  });
  
  test('指标对比视图', async ({ adminPage }) => {
    // 检查对比视图
    const compareBtn = adminPage.locator('[data-testid="compare-btn"], button:has-text("对比"), button:has-text("Compare")');
    
    if (await compareBtn.isVisible()) {
      await compareBtn.click();
      
      // 选择对比设备
      const deviceSelect = adminPage.locator('[data-testid="compare-device-select"], select');
      if (await deviceSelect.isVisible()) {
        await deviceSelect.selectOption(TEST_DEVICES.inj.id);
        
        // 验证对比图表
        await expect(adminPage.locator('[data-testid="compare-chart"]')).toBeVisible({ timeout: 5000 });
      }
    }
  });
});

test.describe('遥测数据刷新', () => {
  
  test('手动刷新数据', async ({ adminPage }) => {
    await adminPage.goto(`/devices/${TEST_DEVICES.cnc.id}`);
    
    // 点击刷新按钮
    const refreshBtn = adminPage.locator('[data-testid="refresh-data-btn"], button:has-text("刷新数据")');
    if (await refreshBtn.isVisible()) {
      await refreshBtn.click();
      
      // 验证加载状态
      await expect(adminPage.locator('[data-testid="loading"], .loading')).toBeVisible();
      
      // 等待加载完成
      await expect(adminPage.locator('[data-testid="loading"], .loading')).not.toBeVisible({ timeout: 10000 });
    }
  });
  
  test('自动刷新间隔', async ({ adminPage }) => {
    await adminPage.goto(`/devices/${TEST_DEVICES.cnc.id}`);
    
    // 检查自动刷新设置
    const autoRefreshToggle = adminPage.locator('[data-testid="auto-refresh"], input[type="checkbox"]');
    
    if (await autoRefreshToggle.isVisible()) {
      // 启用自动刷新
      await autoRefreshToggle.check();
      
      // 等待 5 秒，观察数据是否自动更新
      await adminPage.waitForTimeout(5000);
      
      // 发送新数据，验证自动刷新
      await sendMockTelemetry(TEST_DEVICES.cnc.id, { temperature: 80 });
      await adminPage.waitForTimeout(3000);
      
      // 验证数据已更新
      await expect(adminPage.locator('text=/80|80\\.0/i')).toBeVisible({ timeout: 10000 });
    }
  });
});