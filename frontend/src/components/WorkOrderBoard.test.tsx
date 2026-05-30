import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import React from 'react';

// ============== 使用 vi.hoisted() 定义 mock 函数，避免提升问题 ==============
const { mockShowToast, mockGetWorkOrders, mockCreateWorkOrder, mockUpdateWorkOrderStatus } =
  vi.hoisted(() => ({
    mockShowToast: vi.fn(),
    mockGetWorkOrders: vi.fn(),
    mockCreateWorkOrder: vi.fn(),
    mockUpdateWorkOrderStatus: vi.fn(),
  }));

// ============== Mock 依赖模块 ==============

vi.mock('./Toast', () => ({
  default: () => <div data-testid="toast-container" />,
  useToast: () => ({
    showToast: mockShowToast,
  }),
}));

vi.mock('../lib/api', () => ({
  default: {
    getWorkOrders: mockGetWorkOrders,
    createWorkOrder: mockCreateWorkOrder,
    updateWorkOrderStatus: mockUpdateWorkOrderStatus,
  },
}));

vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      // 工单相关翻译键值映射
      const translations: Record<string, string> = {
        'nav.workOrders': '工单管理',
        'workOrder.title': '工单',
        'workOrder.id': '工单ID',
        'workOrder.priority': '优先级',
        'workOrder.status': '状态',
        'workOrder.pending': '待处理',
        'workOrder.inProgress': '进行中',
        'workOrder.completed': '已完成',
        'workOrder.cancelled': '已取消',
        'workOrder.urgent': '紧急',
        'workOrder.high': '高',
        'workOrder.medium': '中',
        'workOrder.low': '低',
        'workOrder.createdAt': '创建时间',
        'workOrder.updateStatus': '更新状态',
        'workOrder.create': '创建工单',
        'workOrder.createOrder': '创建工单',
        'workOrder.allStatus': '全部状态',
        'workOrder.statusUpdated': '状态已更新',
        'workOrder.updateFailed': '更新失败',
        'workOrder.created': '工单已创建',
        'workOrder.createFailed': '创建失败',
        'device.id': '设备ID',
        'device.description': '描述',
        'common.create': '创建',
        'common.cancel': '取消',
        'alert.medium': '中',
        'errors.loadFailedWorkOrders': '工单数据加载失败',
      };
      return translations[key] || key;
    },
  }),
}));

vi.mock('lucide-react', () => ({
  Plus: () => <span data-testid="icon-plus">+</span>,
  Search: () => <span data-testid="icon-search">搜索</span>,
}));

vi.mock('../types/typeGuards', () => ({
  asWorkOrderArraySafe: (data: unknown) => {
    // 简单模拟：如果是数组则原样返回，否则返回空数组
    if (Array.isArray(data)) return data;
    return [];
  },
}));

vi.mock('../lib/colorUtils', () => ({
  getWorkOrderStatusColor: (status: string) => `status-color-${status}`,
  getWorkOrderPriorityColor: (priority: string) => `priority-color-${priority}`,
}));

// ============== 模拟数据 ==============

const mockWorkOrders = [
  {
    id: 1,
    title: 'CNC-001 温度异常维修',
    description: 'CNC-001 设备温度持续过高，需要紧急维修',
    device_id: 'CNC-001',
    priority: 'urgent' as const,
    status: 'pending' as const,
    created_at: '2024-01-15T10:30:00Z',
    updated_at: '2024-01-15T10:30:00Z',
  },
  {
    id: 2,
    title: 'INJ-001 振动异常检测',
    description: '注塑机 INJ-001 振动超标，需安排检测',
    device_id: 'INJ-001',
    priority: 'high' as const,
    status: 'in_progress' as const,
    created_at: '2024-01-14T09:00:00Z',
    updated_at: '2024-01-14T12:00:00Z',
  },
  {
    id: 3,
    title: 'ROB-001 定期维护',
    description: '工业机器人 ROB-001 定期维护保养',
    device_id: 'ROB-001',
    priority: 'medium' as const,
    status: 'completed' as const,
    created_at: '2024-01-13T08:00:00Z',
    updated_at: '2024-01-13T16:00:00Z',
  },
  {
    id: 4,
    title: 'PLC-002 程序升级',
    description: 'PLC控制器固件升级任务，已取消',
    device_id: 'PLC-002',
    priority: 'low' as const,
    status: 'cancelled' as const,
    created_at: '2024-01-12T14:00:00Z',
    updated_at: '2024-01-12T18:00:00Z',
  },
];

// ============== 辅助渲染函数 ==============

function renderWorkOrderBoard() {
  return render(
    <MemoryRouter>
      <WorkOrderBoard />
    </MemoryRouter>
  );
}

// ============== 在 mock 之后导入被测组件 ==============

import WorkOrderBoard from './WorkOrderBoard';

// ============== 测试用例 ==============

describe('WorkOrderBoard 组件交互测试', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // 默认 mock：返回工单列表
    mockGetWorkOrders.mockResolvedValue({
      data: mockWorkOrders,
      total: mockWorkOrders.length,
      page: 1,
      page_size: 20,
    });
  });

  // -------------------------------------------------------
  // 1. 渲染组件基础测试
  // -------------------------------------------------------
  it('应正确渲染工单看板组件', async () => {
    renderWorkOrderBoard();

    // 验证标题存在
    const heading = screen.getByText('工单管理');
    expect(heading).toBeTruthy();

    // 验证副标题存在
    const subtitle = screen.getByText('工单');
    expect(subtitle).toBeTruthy();

    // 验证创建按钮存在
    const createBtn = screen.getByText('创建工单');
    expect(createBtn).toBeTruthy();

    // 等待数据加载完成
    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledTimes(1);
    });
  });

  // -------------------------------------------------------
  // 2. 加载骨架屏测试
  // -------------------------------------------------------
  it('初始加载时显示骨架屏', () => {
    // 让 API 永远不 resolve，保持 loading 状态
    mockGetWorkOrders.mockReturnValue(new Promise(() => {}));

    renderWorkOrderBoard();

    // 验证没有真实数据行出现
    const noData = screen.queryByText('CNC-001 温度异常维修');
    expect(noData).toBeNull();

    // 验证 API 已被调用
    expect(mockGetWorkOrders).toHaveBeenCalledWith({ status: '' });
  });

  // -------------------------------------------------------
  // 3. 获取并展示工单列表
  // -------------------------------------------------------
  it('成功获取并展示工单列表', async () => {
    renderWorkOrderBoard();

    // 等待数据加载完成并渲染工单行
    await waitFor(() => {
      // 验证所有工单标题都显示出来
      expect(screen.getByText('CNC-001 温度异常维修')).toBeTruthy();
      expect(screen.getByText('INJ-001 振动异常检测')).toBeTruthy();
      expect(screen.getByText('ROB-001 定期维护')).toBeTruthy();
      expect(screen.getByText('PLC-002 程序升级')).toBeTruthy();
    });

    // 验证设备ID显示
    expect(screen.getByText('CNC-001')).toBeTruthy();
    expect(screen.getByText('INJ-001')).toBeTruthy();
    expect(screen.getByText('ROB-001')).toBeTruthy();
    expect(screen.getByText('PLC-002')).toBeTruthy();

    // 验证优先级显示（getAllByText 因为文本在多行下拉框中重复出现）
    expect(screen.getAllByText('紧急').length).toBeGreaterThan(0);
    expect(screen.getAllByText('高').length).toBeGreaterThan(0);
    expect(screen.getAllByText('中').length).toBeGreaterThan(0);
    expect(screen.getAllByText('低').length).toBeGreaterThan(0);

    // 验证状态显示
    expect(screen.getAllByText('待处理').length).toBeGreaterThan(0);
    expect(screen.getAllByText('进行中').length).toBeGreaterThan(0);
    expect(screen.getAllByText('已完成').length).toBeGreaterThan(0);
    expect(screen.getAllByText('已取消').length).toBeGreaterThan(0);
  });

  // -------------------------------------------------------
  // 4. 状态筛选功能
  // -------------------------------------------------------
  it('状态筛选器切换后重新获取数据', async () => {
    renderWorkOrderBoard();

    // 等待初始加载完成
    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledTimes(1);
    });

    // 查找筛选下拉框（包含"全部状态"选项的那个）
    const filterSelect = screen.getByDisplayValue('全部状态');
    expect(filterSelect).toBeTruthy();

    // 切换到"进行中"状态
    fireEvent.change(filterSelect, { target: { value: 'in_progress' } });

    // 验证 API 使用新状态参数重新调用
    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledWith({ status: 'in_progress' });
    });

    // 切换到"已完成"状态
    fireEvent.change(filterSelect, { target: { value: 'completed' } });

    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledWith({ status: 'completed' });
    });

    // 切换回"全部状态"
    fireEvent.change(filterSelect, { target: { value: '' } });

    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledWith({ status: '' });
    });
  });

  // -------------------------------------------------------
  // 5. 创建新工单（打开模态框、填写表单、提交）
  // -------------------------------------------------------
  it('打开创建工单模态框、填写表单并提交', async () => {
    mockCreateWorkOrder.mockResolvedValue({
      id: 5,
      title: '新建设备维修工单',
      description: '测试描述内容',
      device_id: 'TEST-001',
      priority: 'high',
      status: 'pending',
      created_at: '2024-01-16T10:00:00Z',
    });

    renderWorkOrderBoard();

    // 等待初始加载完成
    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledTimes(1);
    });

    // 点击"创建工单"按钮打开模态框
    // 页面上有多个"创建工单"文本（按钮和 aria-label）
    const createButtons = screen.getAllByText('创建工单');
    // 使用第一个（在 header 区域的按钮）
    fireEvent.click(createButtons[0]);

    // 验证模态框已显示
    const modal = screen.getByRole('dialog');
    expect(modal).toBeTruthy();

    // 验证模态框标题（"创建工单"可能多匹配，使用 within modal 查找）
    const modalTitles = within(modal).getAllByText('创建工单');
    expect(modalTitles.length).toBeGreaterThan(0);

    // 填写表单 - title 字段
    const titleInput = modal.querySelector('input[name="title"]') as HTMLInputElement;
    expect(titleInput).toBeTruthy();
    fireEvent.change(titleInput, { target: { value: '新建设备维修工单' } });

    // 填写描述（textarea）
    const descriptionTextarea = modal.querySelector('textarea[name="description"]');
    expect(descriptionTextarea).toBeTruthy();
    fireEvent.change(descriptionTextarea!, {
      target: { value: '测试描述内容' },
    });

    // 填写设备ID
    const deviceInput = modal.querySelector('input[name="device_id"]') as HTMLInputElement;
    expect(deviceInput).toBeTruthy();
    fireEvent.change(deviceInput, { target: { value: 'TEST-001' } });

    // 选择优先级
    const prioritySelect = modal.querySelector('select[name="priority"]');
    expect(prioritySelect).toBeTruthy();
    fireEvent.change(prioritySelect!, { target: { value: 'high' } });

    // 提交表单
    const submitButton = screen.getByText('创建');
    fireEvent.click(submitButton);

    // 验证创建 API 调用
    await waitFor(() => {
      expect(mockCreateWorkOrder).toHaveBeenCalledWith({
        title: '新建设备维修工单',
        description: '测试描述内容',
        device_id: 'TEST-001',
        priority: 'high',
      });
    });

    // 验证成功 toast 提示
    await waitFor(() => {
      expect(mockShowToast).toHaveBeenCalledWith({
        type: 'success',
        message: '工单已创建',
      });
    });

    // 验证重新加载工单列表（至少 2 次：初始加载 + 创建后刷新）
    await waitFor(() => {
      expect(mockGetWorkOrders.mock.calls.length).toBeGreaterThanOrEqual(2);
    });
  });

  // -------------------------------------------------------
  // 5.1 关闭创建工单模态框
  // -------------------------------------------------------
  it('点击取消按钮关闭创建工单模态框', async () => {
    renderWorkOrderBoard();

    // 等待初始加载完成
    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledTimes(1);
    });

    // 打开模态框
    const createButtons = screen.getAllByText('创建工单');
    fireEvent.click(createButtons[0]);

    // 验证模态框已显示
    let modal = screen.getByRole('dialog');
    expect(modal).toBeTruthy();

    // 点击取消按钮
    const cancelButton = screen.getByLabelText('取消');
    fireEvent.click(cancelButton);

    // 验证模态框已关闭
    modal = screen.queryByRole('dialog') as HTMLElement;
    expect(modal).toBeNull();

    // 验证没有调用创建 API
    expect(mockCreateWorkOrder).not.toHaveBeenCalled();
  });

  // -------------------------------------------------------
  // 5.2 创建工单失败处理
  // -------------------------------------------------------
  it('创建工单失败时显示错误提示且模态框保持打开', async () => {
    mockCreateWorkOrder.mockRejectedValue(new Error('创建失败'));

    renderWorkOrderBoard();

    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledTimes(1);
    });

    // 打开模态框
    const createButtons = screen.getAllByText('创建工单');
    fireEvent.click(createButtons[0]);

    // 填写必填字段（title）
    const modal = screen.getByRole('dialog');
    const titleInput = modal.querySelector('input[name="title"]') as HTMLInputElement;
    fireEvent.change(titleInput, { target: { value: '测试工单' } });

    // 提交表单
    const submitButton = screen.getByText('创建');
    fireEvent.click(submitButton);

    // 验证错误 toast 提示
    await waitFor(() => {
      expect(mockShowToast).toHaveBeenCalledWith({
        type: 'error',
        message: '创建失败',
      });
    });

    // 模态框应仍然打开（未关闭）
    const modalAfterError = screen.queryByRole('dialog');
    expect(modalAfterError).toBeTruthy();
  });

  // -------------------------------------------------------
  // 6. 错误处理（API 失败显示错误 toast）
  // -------------------------------------------------------
  it('获取工单列表失败时显示错误 toast', async () => {
    mockGetWorkOrders.mockRejectedValue(new Error('网络错误'));

    renderWorkOrderBoard();

    await waitFor(() => {
      // 验证错误 toast 被调用
      expect(mockShowToast).toHaveBeenCalledWith({
        type: 'error',
        message: '工单数据加载失败',
      });
    });
  });

  // -------------------------------------------------------
  // 7. 空状态（无工单数据）
  // -------------------------------------------------------
  it('无工单数据时显示空表格', async () => {
    mockGetWorkOrders.mockResolvedValue({
      data: [],
      total: 0,
      page: 1,
      page_size: 20,
    });

    renderWorkOrderBoard();

    // 等待加载完成
    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledTimes(1);
    });

    // 验证表格头部仍然存在
    expect(screen.getByText('工单ID')).toBeTruthy();

    // 验证没有工单数据行（不应出现具体工单内容）
    const noOrders = screen.queryByText('CNC-001 温度异常维修');
    expect(noOrders).toBeNull();

    // 验证筛选器仍然存在
    const filterSelect = screen.getByDisplayValue('全部状态');
    expect(filterSelect).toBeTruthy();
  });

  // -------------------------------------------------------
  // 8. 工单卡片展示正确信息
  // -------------------------------------------------------
  it('每条工单行显示正确的信息（ID、标题、设备ID、优先级、状态）', async () => {
    renderWorkOrderBoard();

    await waitFor(() => {
      expect(screen.getByText('CNC-001 温度异常维修')).toBeTruthy();
    });

    // 验证工单 ID 显示（#1 格式）
    expect(screen.getByText('#1')).toBeTruthy();
    expect(screen.getByText('#2')).toBeTruthy();
    expect(screen.getByText('#3')).toBeTruthy();
    expect(screen.getByText('#4')).toBeTruthy();

    // 验证设备 ID 显示
    expect(screen.getByText('CNC-001')).toBeTruthy();
    expect(screen.getByText('INJ-001')).toBeTruthy();
    expect(screen.getByText('ROB-001')).toBeTruthy();
    expect(screen.getByText('PLC-002')).toBeTruthy();

    // 验证状态颜色类已应用（getAllByText 因为下拉框中也有重复文本）
    const pendingBadges = screen.getAllByText('待处理');
    // 找到带 status-color 类的那个（badge 元素，非 option）
    const pendingBadge = pendingBadges.find(el => el.className.includes('status-color'));
    expect(pendingBadge).toBeTruthy();
    expect(pendingBadge!.className).toContain('status-color-pending');

    const inProgressBadges = screen.getAllByText('进行中');
    const inProgressBadge = inProgressBadges.find(el => el.className.includes('status-color'));
    expect(inProgressBadge).toBeTruthy();
    expect(inProgressBadge!.className).toContain('status-color-in_progress');

    const completedBadges = screen.getAllByText('已完成');
    const completedBadge = completedBadges.find(el => el.className.includes('status-color'));
    expect(completedBadge).toBeTruthy();
    expect(completedBadge!.className).toContain('status-color-completed');

    const cancelledBadges = screen.getAllByText('已取消');
    const cancelledBadge = cancelledBadges.find(el => el.className.includes('status-color'));
    expect(cancelledBadge).toBeTruthy();
    expect(cancelledBadge!.className).toContain('status-color-cancelled');

    // 验证优先级颜色类已应用
    const urgentBadges = screen.getAllByText('紧急');
    const urgentBadge = urgentBadges.find(el => el.className.includes('priority-color'));
    expect(urgentBadge).toBeTruthy();
    expect(urgentBadge!.className).toContain('priority-color-urgent');

    const highBadges = screen.getAllByText('高');
    const highBadge = highBadges.find(el => el.className.includes('priority-color'));
    expect(highBadge).toBeTruthy();
    expect(highBadge!.className).toContain('priority-color-high');
  });

  // -------------------------------------------------------
  // 9. 更新工单状态成功
  // -------------------------------------------------------
  it('通过行内下拉框更新工单状态成功', async () => {
    mockUpdateWorkOrderStatus.mockResolvedValue({ message: '状态已更新' });

    renderWorkOrderBoard();

    // 等待工单列表加载
    await waitFor(() => {
      expect(screen.getByText('CNC-001 温度异常维修')).toBeTruthy();
    });

    // 找到所有 combobox 元素
    const allSelects = screen.getAllByRole('combobox');
    // 第一个是筛选器，后面的是每行的状态更新下拉框
    expect(allSelects.length).toBeGreaterThan(1);

    // 修改第一行的状态（从 pending 改为 in_progress）
    const rowStatusSelect = allSelects[1];
    fireEvent.change(rowStatusSelect, { target: { value: 'in_progress' } });

    // 验证 API 调用
    await waitFor(() => {
      expect(mockUpdateWorkOrderStatus).toHaveBeenCalledWith(1, 'in_progress');
    });

    // 验证成功 toast
    await waitFor(() => {
      expect(mockShowToast).toHaveBeenCalledWith({
        type: 'success',
        message: '状态已更新',
      });
    });

    // 验证重新加载工单列表（至少 2 次：初始加载 + 更新后刷新）
    await waitFor(() => {
      expect(mockGetWorkOrders.mock.calls.length).toBeGreaterThanOrEqual(2);
    });
  });

  // -------------------------------------------------------
  // 9.1 更新工单状态失败
  // -------------------------------------------------------
  it('更新工单状态失败时显示错误 toast', async () => {
    mockUpdateWorkOrderStatus.mockRejectedValue(new Error('更新失败'));

    renderWorkOrderBoard();

    await waitFor(() => {
      expect(screen.getByText('CNC-001 温度异常维修')).toBeTruthy();
    });

    const allSelects = screen.getAllByRole('combobox');
    const rowStatusSelect = allSelects[1];
    fireEvent.change(rowStatusSelect, { target: { value: 'completed' } });

    await waitFor(() => {
      expect(mockShowToast).toHaveBeenCalledWith({
        type: 'error',
        message: '更新失败',
      });
    });
  });

  // -------------------------------------------------------
  // 10. 筛选后仅展示筛选结果
  // -------------------------------------------------------
  it('筛选后仅展示匹配的工单数据', async () => {
    // 模拟"进行中"筛选后的结果
    mockGetWorkOrders.mockResolvedValue({
      data: [mockWorkOrders[1]], // 仅 in_progress 的工单
      total: 1,
      page: 1,
      page_size: 20,
    });

    renderWorkOrderBoard();

    await waitFor(() => {
      expect(screen.getByText('INJ-001 振动异常检测')).toBeTruthy();
    });

    // 筛选结果中不应有其他工单
    expect(screen.queryByText('CNC-001 温度异常维修')).toBeNull();
    expect(screen.queryByText('ROB-001 定期维护')).toBeNull();
    expect(screen.queryByText('PLC-002 程序升级')).toBeNull();
  });

  // -------------------------------------------------------
  // 11. 搜索图标和创建按钮图标渲染
  // -------------------------------------------------------
  it('搜索图标和创建按钮图标正确渲染', async () => {
    renderWorkOrderBoard();

    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledTimes(1);
    });

    // 验证 Plus 图标渲染
    const plusIcon = screen.getByTestId('icon-plus');
    expect(plusIcon).toBeTruthy();

    // 验证 Search 图标渲染
    const searchIcon = screen.getByTestId('icon-search');
    expect(searchIcon).toBeTruthy();
  });

  // -------------------------------------------------------
  // 12. 初始加载时 API 调用参数正确
  // -------------------------------------------------------
  it('初始加载以空状态参数调用 API', async () => {
    renderWorkOrderBoard();

    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledWith({ status: '' });
    });
  });

  // -------------------------------------------------------
  // 13. device_id 为空时显示 '--' 回退值
  // -------------------------------------------------------
  it('device_id 为空字符串时显示回退值', async () => {
    mockGetWorkOrders.mockResolvedValue({
      data: [
        {
          id: 99,
          title: '无关联设备的工单',
          description: '测试描述',
          device_id: '',
          priority: 'medium' as const,
          status: 'pending' as const,
          created_at: '2024-01-16T10:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    });

    renderWorkOrderBoard();

    await waitFor(() => {
      expect(screen.getByText('无关联设备的工单')).toBeTruthy();
    });

    // device_id 为空时应显示 '--'
    expect(screen.getByText('--')).toBeTruthy();
  });

  // -------------------------------------------------------
  // 14. 创建按钮具有 aria-label 无障碍属性
  // -------------------------------------------------------
  it('创建工单按钮具有正确的 aria-label', async () => {
    renderWorkOrderBoard();

    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledTimes(1);
    });

    // 找到具有 aria-label="创建工单" 的按钮
    const createButton = screen.getByLabelText('创建工单');
    expect(createButton).toBeTruthy();
  });

  // -------------------------------------------------------
  // 15. 状态筛选器包含所有状态选项
  // -------------------------------------------------------
  it('状态筛选下拉框包含全部五个状态选项', async () => {
    renderWorkOrderBoard();

    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledTimes(1);
    });

    // 筛选器的 select 应包含所有选项
    const filterSelect = screen.getByDisplayValue('全部状态');
    expect(filterSelect).toBeTruthy();

    const options = within(filterSelect as HTMLElement).getAllByRole('option');
    // 全部状态、待处理、进行中、已完成、已取消 = 5 个选项
    expect(options.length).toBe(5);

    // 验证选项值
    expect(options[0].getAttribute('value')).toBe('');
    expect(options[1].getAttribute('value')).toBe('pending');
    expect(options[2].getAttribute('value')).toBe('in_progress');
    expect(options[3].getAttribute('value')).toBe('completed');
    expect(options[4].getAttribute('value')).toBe('cancelled');
  });

  // -------------------------------------------------------
  // 16. 工单行内状态更新下拉框包含所有状态选项
  // -------------------------------------------------------
  it('工单行内状态下拉框包含四个可选状态', async () => {
    renderWorkOrderBoard();

    await waitFor(() => {
      expect(screen.getByText('CNC-001 温度异常维修')).toBeTruthy();
    });

    // 获取所有 combobox，第二个开始是行内状态下拉框
    const allSelects = screen.getAllByRole('combobox');
    const rowSelect = allSelects[1]; // 第一行的状态更新 select

    const options = within(rowSelect as HTMLElement).getAllByRole('option');
    expect(options.length).toBe(4);
    expect(options[0].getAttribute('value')).toBe('pending');
    expect(options[1].getAttribute('value')).toBe('in_progress');
    expect(options[2].getAttribute('value')).toBe('completed');
    expect(options[3].getAttribute('value')).toBe('cancelled');
  });

  // -------------------------------------------------------
  // 17. 创建工单表单字段完整性验证
  // -------------------------------------------------------
  it('创建工单表单包含所有必要字段', async () => {
    renderWorkOrderBoard();

    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledTimes(1);
    });

    // 打开模态框
    const createButtons = screen.getAllByText('创建工单');
    fireEvent.click(createButtons[0]);

    const modal = screen.getByRole('dialog');

    // 验证 title 字段存在且有 required 属性
    const titleInput = modal.querySelector('input[name="title"]') as HTMLInputElement;
    expect(titleInput).toBeTruthy();
    expect(titleInput.hasAttribute('required')).toBe(true);

    // 验证 description textarea 存在
    const descriptionTextarea = modal.querySelector('textarea[name="description"]');
    expect(descriptionTextarea).toBeTruthy();

    // 验证 device_id input 存在
    const deviceIdInput = modal.querySelector('input[name="device_id"]');
    expect(deviceIdInput).toBeTruthy();

    // 验证 priority select 存在
    const prioritySelect = modal.querySelector('select[name="priority"]');
    expect(prioritySelect).toBeTruthy();
  });

  // -------------------------------------------------------
  // 18. 创建模态框 ARIA 无障碍属性
  // -------------------------------------------------------
  it('创建模态框具有正确的 ARIA 无障碍属性', async () => {
    renderWorkOrderBoard();

    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledTimes(1);
    });

    // 打开模态框
    const createButtons = screen.getAllByText('创建工单');
    fireEvent.click(createButtons[0]);

    // 验证 dialog 角色和 aria-modal 属性
    const modal = screen.getByRole('dialog');
    expect(modal).toBeTruthy();
    expect(modal.getAttribute('aria-modal')).toBe('true');
  });

  // -------------------------------------------------------
  // 19. 多条工单渲染为对应行数
  // -------------------------------------------------------
  it('工单数据正确渲染为对应行数', async () => {
    renderWorkOrderBoard();

    await waitFor(() => {
      expect(screen.getByText('CNC-001 温度异常维修')).toBeTruthy();
    });

    // 验证表格行数等于工单数据条数
    const tableBody = document.querySelector('tbody');
    expect(tableBody).toBeTruthy();

    const rows = tableBody!.querySelectorAll('tr');
    expect(rows.length).toBe(mockWorkOrders.length);
  });

  // -------------------------------------------------------
  // 20. 加载失败后仍可切换筛选器重新加载
  // -------------------------------------------------------
  it('加载失败后切换筛选器可重新触发加载', async () => {
    // 第一次调用失败
    mockGetWorkOrders.mockRejectedValueOnce(new Error('网络错误'));

    renderWorkOrderBoard();

    // 等待错误 toast 显示
    await waitFor(() => {
      expect(mockShowToast).toHaveBeenCalledWith({
        type: 'error',
        message: '工单数据加载失败',
      });
    });

    // 恢复正常响应
    mockGetWorkOrders.mockResolvedValue({
      data: mockWorkOrders,
      total: mockWorkOrders.length,
      page: 1,
      page_size: 20,
    });

    // 切换筛选器
    const filterSelect = screen.getByDisplayValue('全部状态');
    fireEvent.change(filterSelect, { target: { value: 'pending' } });

    // 验证数据加载成功
    await waitFor(() => {
      expect(mockGetWorkOrders).toHaveBeenCalledWith({ status: 'pending' });
    });

    await waitFor(() => {
      expect(screen.getByText('CNC-001 温度异常维修')).toBeTruthy();
    });
  });
});
