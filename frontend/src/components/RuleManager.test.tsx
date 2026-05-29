import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, waitFor, fireEvent, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// 使用 vi.hoisted 将 mock 函数提升，使 vi.mock 工厂可以引用
const {
  mockRefresh, mockCreate, mockUpdate, mockDeleteItem,
  mockShowConfirm, mockShowToast, mockToggleRule,
  mockCRUDReturn,
} = vi.hoisted(() => {
  const _mockRefresh = vi.fn();
  const _mockCreate = vi.fn();
  const _mockUpdate = vi.fn();
  const _mockDeleteItem = vi.fn();
  const _mockShowConfirm = vi.fn().mockResolvedValue(false);
  const _mockShowToast = vi.fn();
  const _mockToggleRule = vi.fn().mockResolvedValue(undefined);

  // 可动态修改的 mock 返回值
  const _mockCRUDReturn: [
    { items: unknown[]; loading: boolean; error: string | null; total: number; page: number; pageSize: number },
    { refresh: ReturnType<typeof vi.fn>; create: ReturnType<typeof vi.fn>; update: ReturnType<typeof vi.fn>; delete: ReturnType<typeof vi.fn>; fetchAll: ReturnType<typeof vi.fn>; setPage: ReturnType<typeof vi.fn> }
  ] = [
    { items: [], loading: false, error: null, total: 0, page: 1, pageSize: 100 },
    { refresh: _mockRefresh, create: _mockCreate, update: _mockUpdate, delete: _mockDeleteItem, fetchAll: vi.fn(), setPage: vi.fn() },
  ];

  return {
    mockRefresh: _mockRefresh,
    mockCreate: _mockCreate,
    mockUpdate: _mockUpdate,
    mockDeleteItem: _mockDeleteItem,
    mockShowConfirm: _mockShowConfirm,
    mockShowToast: _mockShowToast,
    mockToggleRule: _mockToggleRule,
    mockCRUDReturn: _mockCRUDReturn,
  };
});

vi.mock('./AuthContext', () => ({ useAuth: () => ({ isAdmin: true, user: { role: 'admin' } }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: mockShowToast }) }));
vi.mock('../lib/api', () => ({
  default: {
    getRules: vi.fn().mockResolvedValue({ data: [] }),
    getRule: vi.fn().mockResolvedValue({}),
    createRule: vi.fn().mockResolvedValue(undefined),
    updateRule: vi.fn().mockResolvedValue(undefined),
    deleteRule: vi.fn().mockResolvedValue(undefined),
    toggleRule: mockToggleRule,
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('../hooks/useCRUD', () => ({
  useCRUD: () => mockCRUDReturn,
}));
vi.mock('./UI/ConfirmDialog', () => ({ useConfirmDialog: () => ({ showConfirm: mockShowConfirm }) }));
vi.mock('lucide-react', () => ({
  Plus: () => <span>Plus</span>,
  Edit: () => <span>Edit</span>,
  Trash2: () => <span>Trash2</span>,
  Settings: () => <span>Settings</span>,
  Bell: () => <span>Bell</span>,
  ToggleLeft: () => <span>ToggleLeft</span>,
  ToggleRight: () => <span>ToggleRight</span>,
}));

import RuleManager from './RuleManager';

function renderRuleManager() {
  return render(
    <MemoryRouter>
      <RuleManager />
    </MemoryRouter>
  );
}

const sampleRules = [
  {
    id: 1,
    name: '温度过高',
    device_type: 'pump',
    metric: 'temperature',
    operator: '>',
    threshold: 80,
    severity: 'high',
    enabled: true,
    cooldown_sec: 300,
  },
  {
    id: 2,
    name: '振动异常',
    device_type: 'motor',
    metric: 'vibration',
    operator: '>=',
    threshold: 5.0,
    severity: 'medium',
    enabled: false,
    cooldown_sec: 600,
  },
];

function setEmptyRules() {
  mockCRUDReturn[0] = { items: [], loading: false, error: null, total: 0, page: 1, pageSize: 100 };
  mockCRUDReturn[1] = { refresh: mockRefresh, create: mockCreate, update: mockUpdate, delete: mockDeleteItem, fetchAll: vi.fn(), setPage: vi.fn() };
}

function setSampleRules() {
  mockCRUDReturn[0] = { items: sampleRules, loading: false, error: null, total: 2, page: 1, pageSize: 100 };
  mockCRUDReturn[1] = { refresh: mockRefresh, create: mockCreate, update: mockUpdate, delete: mockDeleteItem, fetchAll: vi.fn(), setPage: vi.fn() };
}

function setLoading() {
  mockCRUDReturn[0] = { items: [], loading: true, error: null, total: 0, page: 1, pageSize: 100 };
  mockCRUDReturn[1] = { refresh: mockRefresh, create: mockCreate, update: mockUpdate, delete: mockDeleteItem, fetchAll: vi.fn(), setPage: vi.fn() };
}

describe('RuleManager - 基础渲染', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockShowConfirm.mockResolvedValue(false);
    mockToggleRule.mockResolvedValue(undefined);
    setEmptyRules();
  });

  it('渲染规则管理页面，包含标题和创建按钮', async () => {
    renderRuleManager();
    await waitFor(() => {
      expect(screen.getByText('nav.rules')).toBeInTheDocument();
      expect(screen.getByLabelText('alert.createRule')).toBeInTheDocument();
    });
  });

  it('空规则列表时显示空表格', async () => {
    renderRuleManager();
    await waitFor(() => {
      expect(screen.getByText('alert.ruleName')).toBeInTheDocument();
    });
  });

  it('点击创建规则按钮打开创建模态框', async () => {
    renderRuleManager();

    const createButton = screen.getByLabelText('alert.createRule');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      // 模态框标题使用 h2 标签
      expect(screen.getByRole('heading', { name: 'alert.createRule' })).toBeInTheDocument();
    });
  });

  it('创建模态框中点击取消按钮关闭模态框', async () => {
    renderRuleManager();

    const createButton = screen.getByLabelText('alert.createRule');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    const cancelButtons = screen.getAllByText('common.cancel');
    fireEvent.click(cancelButtons[cancelButtons.length - 1]);

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
  });
});

describe('RuleManager - 规则列表渲染', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockShowConfirm.mockResolvedValue(false);
    mockToggleRule.mockResolvedValue(undefined);
    setSampleRules();
  });

  it('渲染规则列表，显示规则名称和严重级别', async () => {
    renderRuleManager();
    await waitFor(() => {
      expect(screen.getByText('温度过高')).toBeInTheDocument();
      expect(screen.getByText('振动异常')).toBeInTheDocument();
      expect(screen.getByText('high')).toBeInTheDocument();
      expect(screen.getByText('medium')).toBeInTheDocument();
    });
  });

  it('显示规则的 metric 信息', async () => {
    renderRuleManager();
    await waitFor(() => {
      expect(screen.getByText('temperature')).toBeInTheDocument();
      expect(screen.getByText('vibration')).toBeInTheDocument();
    });
  });

  it('显示规则操作符', async () => {
    renderRuleManager();
    await waitFor(() => {
      expect(screen.getByText('>')).toBeInTheDocument();
      expect(screen.getByText('>=')).toBeInTheDocument();
    });
  });

  it('已启用规则显示 ToggleRight，已禁用显示 ToggleLeft', async () => {
    renderRuleManager();
    await waitFor(() => {
      expect(screen.getByText('ToggleRight')).toBeInTheDocument();
      expect(screen.getByText('ToggleLeft')).toBeInTheDocument();
    });
  });

  it('显示编辑和删除按钮', async () => {
    renderRuleManager();
    await waitFor(() => {
      expect(screen.getAllByLabelText('common.edit').length).toBe(2);
      expect(screen.getAllByLabelText('common.delete').length).toBe(2);
    });
  });

  it('点击编辑按钮打开编辑模态框', async () => {
    renderRuleManager();
    await waitFor(() => {
      const editButtons = screen.getAllByLabelText('common.edit');
      fireEvent.click(editButtons[0]);
    });

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      expect(screen.getByText('alert.editRule')).toBeInTheDocument();
    });
  });
});

describe('RuleManager - 启用/禁用切换', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockShowConfirm.mockResolvedValue(false);
    setSampleRules();
  });

  it('点击 toggle 按钮调用 toggleRule API 并显示禁用成功消息', async () => {
    mockToggleRule.mockResolvedValue(undefined);
    renderRuleManager();

    await waitFor(() => {
      const disableButton = screen.getByLabelText('common.disable');
      fireEvent.click(disableButton);
    });

    await waitFor(() => {
      expect(mockToggleRule).toHaveBeenCalledWith(1, false);
      expect(mockRefresh).toHaveBeenCalled();
      expect(mockShowToast).toHaveBeenCalledWith({ type: 'success', message: 'alert.ruleDisabled' });
    });
  });

  it('启用禁用规则时显示启用成功消息', async () => {
    mockToggleRule.mockResolvedValue(undefined);
    renderRuleManager();

    await waitFor(() => {
      const enableButton = screen.getByLabelText('common.enable');
      fireEvent.click(enableButton);
    });

    await waitFor(() => {
      expect(mockToggleRule).toHaveBeenCalledWith(2, true);
      expect(mockShowToast).toHaveBeenCalledWith({ type: 'success', message: 'alert.ruleEnabled' });
    });
  });

  it('toggle 失败时显示错误提示', async () => {
    mockToggleRule.mockRejectedValue(new Error('网络错误'));
    renderRuleManager();

    await waitFor(() => {
      const toggleButton = screen.getByLabelText('common.disable');
      fireEvent.click(toggleButton);
    });

    await waitFor(() => {
      expect(mockShowToast).toHaveBeenCalledWith({ type: 'error', message: 'errors.unknown' });
    });
  });
});

describe('RuleManager - 删除规则', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setSampleRules();
  });

  it('确认删除后调用 delete 并显示成功提示', async () => {
    mockShowConfirm.mockResolvedValue(true);
    mockDeleteItem.mockResolvedValue(true);
    renderRuleManager();

    await waitFor(() => {
      const deleteButtons = screen.getAllByLabelText('common.delete');
      fireEvent.click(deleteButtons[0]);
    });

    await waitFor(() => {
      expect(mockShowConfirm).toHaveBeenCalledWith(
        expect.objectContaining({
          title: 'alert.confirmDeleteTitle',
          variant: 'danger',
        })
      );
      expect(mockDeleteItem).toHaveBeenCalledWith('1');
      expect(mockShowToast).toHaveBeenCalledWith({ type: 'success', message: 'alert.ruleDeleted' });
    });
  });

  it('取消删除时不调用 delete', async () => {
    mockShowConfirm.mockResolvedValue(false);
    renderRuleManager();

    await waitFor(() => {
      const deleteButtons = screen.getAllByLabelText('common.delete');
      fireEvent.click(deleteButtons[0]);
    });

    await waitFor(() => {
      expect(mockShowConfirm).toHaveBeenCalled();
    });

    expect(mockDeleteItem).not.toHaveBeenCalled();
  });

  it('删除失败时显示错误提示', async () => {
    mockShowConfirm.mockResolvedValue(true);
    mockDeleteItem.mockResolvedValue(false);
    renderRuleManager();

    await waitFor(() => {
      const deleteButtons = screen.getAllByLabelText('common.delete');
      fireEvent.click(deleteButtons[0]);
    });

    await waitFor(() => {
      expect(mockShowToast).toHaveBeenCalledWith({ type: 'error', message: 'alert.deleteFailed' });
    });
  });
});

describe('RuleManager - 加载状态', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setLoading();
  });

  it('加载中显示骨架屏', async () => {
    const { container } = renderRuleManager();
    await waitFor(() => {
      expect(container.querySelectorAll('.animate-pulse').length).toBeGreaterThan(0);
    });
  });
});
