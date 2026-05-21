import { test, expect, TEST_DEVICES, createTestDevice, deleteTestDevice, cleanupTestData } from '../fixtures/test-fixtures';

/**
 * 设备列表 E2E 测试
 * 基于 DeviceManager.tsx 实际 DOM 结构编写
 *
 * 已跳过的测试（功能尚未实现）：
 * - 批量操作：前端无 checkbox 选择和 bulk-actions
 * - 按类型筛选：前端无类型筛选下拉框（仅有搜索框按名称/ID过滤）
 * - 刷新按钮：前端无独立刷新按钮
 * - 分页：仅有 12 条种子数据，不足 20 条触发分页
 */
test.describe('设备列表页面', () => {

  test.beforeEach(async ({ adminPage }) => {
    await cleanupTestData(adminPage);
  });

  test.afterEach(async ({ adminPage }) => {
    await cleanupTestData(adminPage);
  });

  test('显示设备列表', async ({ adminPage }) => {
    await adminPage.goto('/devices');

    // 验证页面标题 — h1 内容来自 t('nav.devices') = "设备管理"
    await expect(adminPage.locator('h1').filter({ hasText: /设备管理|Devices/i })).toBeVisible();

    // 验证设备表格存在 — DeviceManager 使用 data-testid="device-table"
    const deviceTable = adminPage.locator('[data-testid="device-table"]');
    await expect(deviceTable).toBeVisible({ timeout: 10000 });

    // 验证表格有数据（种子数据至少有 10 条设备）
    const rowCount = await deviceTable.locator('tbody tr').count();
    expect(rowCount).toBeGreaterThan(0);
  });

  test('搜索设备', async ({ adminPage }) => {
    // 先创建一个测试设备
    await createTestDevice(adminPage, TEST_DEVICES.cnc);

    // 回到设备列表页
    await adminPage.goto('/devices');
    await expect(adminPage.locator('[data-testid="device-table"]')).toBeVisible({ timeout: 10000 });

    // 使用搜索框过滤 — DeviceManager 的搜索框 placeholder 是 t('common.search') = "搜索"
    const searchInput = adminPage.locator('input[placeholder="搜索"]');
    await searchInput.fill('Test CNC');

    // 搜索是前端过滤（filter by name/id），无需等 API
    await adminPage.waitForTimeout(300);

    // 验证搜索结果包含目标设备
    await expect(adminPage.locator(`text="${TEST_DEVICES.cnc.name}"`)).toBeVisible();

    // 清空搜索 — 所有设备重新显示
    await searchInput.clear();
    await adminPage.waitForTimeout(300);

    // 验证种子数据设备重新出现
    await expect(adminPage.locator('text="PLC控制器-P1"')).toBeVisible();
  });

  test('按类型筛选设备', async ({ adminPage }) => {
    // 创建一台 CNC 设备
    await createTestDevice(adminPage, TEST_DEVICES.cnc);

    // 回到设备列表页
    await adminPage.goto('/devices');
    await expect(adminPage.locator('[data-testid="device-table"]')).toBeVisible({ timeout: 10000 });

    // 使用类型筛选下拉框选择 CNC
    const typeFilter = adminPage.locator('[data-testid="type-filter"]');
    await typeFilter.selectOption('CNC');

    // 等待前端过滤生效
    await adminPage.waitForTimeout(300);

    // 验证筛选后列表中包含我们创建的 CNC 设备
    await expect(adminPage.locator(`text="${TEST_DEVICES.cnc.name}"`)).toBeVisible();

    // 切换回全部类型
    await typeFilter.selectOption('');
    await adminPage.waitForTimeout(300);

    // 全部类型下，种子数据的 PLC 控制器也应该出现
    await expect(adminPage.locator('text="PLC控制器-P1"')).toBeVisible();
  });

  test('添加设备', async ({ adminPage }) => {
    await adminPage.goto('/devices');
    await expect(adminPage.locator('[data-testid="device-table"]')).toBeVisible({ timeout: 10000 });

    // 点击创建按钮
    await adminPage.click('[data-testid="add-device-btn"]');

    // 等待弹窗出现
    await expect(adminPage.locator('[role="dialog"]')).toBeVisible({ timeout: 5000 });

    // 填写设备信息 — 字段名对应 DeviceManager.tsx 的 name 属性
    await adminPage.fill('[name="device-id"]', TEST_DEVICES.cnc.id);
    await adminPage.fill('[name="device-name"]', TEST_DEVICES.cnc.name);
    await adminPage.selectOption('[name="device-type"]', TEST_DEVICES.cnc.type);
    await adminPage.fill('[name="device-location"]', TEST_DEVICES.cnc.location);

    // 提交
    await adminPage.click('[role="dialog"] button[type="submit"]');

    // 等待弹窗关闭
    await expect(adminPage.locator('[role="dialog"]')).not.toBeVisible({ timeout: 5000 });

    // 等待设备出现在列表
    await expect(adminPage.locator(`text="${TEST_DEVICES.cnc.name}"`)).toBeVisible({ timeout: 10000 });
  });

  test('编辑设备', async ({ adminPage }) => {
    // 创建测试设备
    await createTestDevice(adminPage, TEST_DEVICES.cnc);

    // 找到对应行的编辑按钮并点击
    const row = adminPage.locator(`[data-testid="device-row-${TEST_DEVICES.cnc.id}"]`);
    await row.locator('[data-testid="edit-btn"]').click();

    // 等待编辑弹窗 — 编辑模式下不显示 device-id 字段
    await expect(adminPage.locator('[role="dialog"]')).toBeVisible({ timeout: 5000 });

    // 修改设备名称
    const newName = 'Updated CNC Machine';
    // 编辑模式下 name 字段已有原值，需要先清除再填写
    await adminPage.locator('[name="device-name"]').fill(newName);

    // 提交
    await adminPage.click('[role="dialog"] button[type="submit"]');

    // 等待更新成功 — 验证新名称出现
    await expect(adminPage.locator(`text="${newName}"`)).toBeVisible({ timeout: 10000 });
  });

  test('删除设备', async ({ adminPage }) => {
    // 创建测试设备
    await createTestDevice(adminPage, TEST_DEVICES.cnc);

    // 删除设备 — deleteTestDevice 内部处理 confirm 弹窗
    await deleteTestDevice(adminPage, TEST_DEVICES.cnc.id);

    // 验证设备已从列表消失
    await expect(adminPage.locator(`text="${TEST_DEVICES.cnc.name}"`)).not.toBeVisible({ timeout: 10000 });
  });

  // 批量操作功能尚未实现 — 前端无 checkbox 和 bulk-actions
  test.skip('批量操作', async ({ adminPage }) => {
    // DeviceManager 没有行选择和批量操作功能
  });

  // 刷新按钮不存在 — 前端无独立刷新按钮（数据在页面加载时自动获取）
  test.skip('刷新设备列表', async ({ adminPage }) => {
    // DeviceManager 没有刷新按钮
  });

  test('分页功能', async ({ adminPage }) => {
    await adminPage.goto('/devices');
    await expect(adminPage.locator('[data-testid="device-table"]')).toBeVisible({ timeout: 10000 });

    // 验证分页控件始终可见
    const pagination = adminPage.locator('[data-testid="pagination"]');
    await expect(pagination).toBeVisible();

    // 验证页码信息存在
    const pageInfo = adminPage.locator('[data-testid="page-info"]');
    await expect(pageInfo).toBeVisible();
    const pageText = await pageInfo.textContent();
    expect(pageText).toMatch(/\d+ \/ \d+/);

    // 验证第一页时上一页按钮禁用
    const prevBtn = adminPage.locator('[data-testid="prev-page-btn"]');
    await expect(prevBtn).toBeDisabled();
  });

  test('设备状态指示器', async ({ adminPage }) => {
    await adminPage.goto('/devices');
    await expect(adminPage.locator('[data-testid="device-table"]')).toBeVisible({ timeout: 10000 });

    // 验证种子数据中有状态标签 — DeviceManager 使用 .status-badge 类
    const statusBadges = adminPage.locator('.status-badge');
    const count = await statusBadges.count();
    expect(count).toBeGreaterThan(0);

    // 验证至少有一个在线状态的设备（种子数据 plc-001 是在线的）
    await expect(adminPage.locator('text="在线"').first()).toBeVisible();
  });
});

test.describe('设备列表权限', () => {

  test('管理员可以添加设备', async ({ adminPage }) => {
    await adminPage.goto('/devices');
    await expect(adminPage.locator('[data-testid="add-device-btn"]')).toBeVisible({ timeout: 10000 });
  });

  test('操作员无法删除设备', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/devices');
    await expect(authenticatedPage.locator('[data-testid="device-table"]')).toBeVisible({ timeout: 10000 });

    // DeviceManager 只在 isAdmin 时渲染 delete-btn，operator 应该看不到
    const deleteBtnCount = await authenticatedPage.locator('[data-testid="delete-btn"]').count();
    expect(deleteBtnCount).toBe(0);
  });
});
