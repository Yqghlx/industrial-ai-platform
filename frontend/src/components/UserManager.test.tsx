import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, waitFor, fireEvent, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// 使用 vi.hoisted 解决 mock hoisting 问题
const {
  mockShowConfirm, mockShowToast, mockDeleteItem, mockCreate,
  mockCRUDReturn, mockIsAdminReturn,
} = vi.hoisted(() => {
  const _mockShowConfirm = vi.fn().mockResolvedValue(false);
  const _mockShowToast = vi.fn();
  const _mockDeleteItem = vi.fn();
  const _mockCreate = vi.fn();
  const _mockIsAdminReturn = { isAdmin: true, user: { role: 'admin' } };

  const _mockCRUDReturn: [
    { items: unknown[]; loading: boolean; error: string | null; total: number; page: number; pageSize: number },
    { refresh: ReturnType<typeof vi.fn>; create: ReturnType<typeof vi.fn>; delete: ReturnType<typeof vi.fn>; fetchAll: ReturnType<typeof vi.fn>; setPage: vi.fn }
  ] = [
    { items: [], loading: false, error: null, total: 0, page: 1, pageSize: 50 },
    { refresh: vi.fn(), create: _mockCreate, delete: _mockDeleteItem, fetchAll: vi.fn(), setPage: vi.fn() },
  ];

  return {
    mockShowConfirm: _mockShowConfirm,
    mockShowToast: _mockShowToast,
    mockDeleteItem: _mockDeleteItem,
    mockCreate: _mockCreate,
    mockCRUDReturn: _mockCRUDReturn,
    mockIsAdminReturn: _mockIsAdminReturn,
  };
});

vi.mock('./AuthContext', () => ({
  useAuth: () => mockIsAdminReturn,
}));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: mockShowToast }) }));
vi.mock('./UI/ConfirmDialog', () => ({ useConfirmDialog: () => ({ showConfirm: mockShowConfirm }) }));
vi.mock('../hooks/useCRUD', () => ({
  useCRUD: () => mockCRUDReturn,
}));
vi.mock('../lib/api', () => ({
  default: {
    getUsers: vi.fn().mockResolvedValue({ data: [] }),
    createUser: vi.fn().mockResolvedValue({}),
    deleteUser: vi.fn().mockResolvedValue(undefined),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({
  Plus: () => <span>Plus</span>,
  Trash2: () => <span>Trash2</span>,
  Shield: () => <span>Shield</span>,
  User: () => <span>User</span>,
}));

import UserManager from './UserManager';

function renderUserManager() {
  return render(
    <MemoryRouter>
      <UserManager />
    </MemoryRouter>
  );
}

const sampleUsers = [
  {
    id: 1,
    username: 'admin',
    email: 'admin@test.com',
    role: 'admin',
    created_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    username: 'operator',
    email: 'operator@test.com',
    role: 'user',
    created_at: '2024-06-15T00:00:00Z',
  },
];

function setEmptyUsers() {
  mockCRUDReturn[0] = { items: [], loading: false, error: null, total: 0, page: 1, pageSize: 50 };
  mockCRUDReturn[1] = { refresh: vi.fn(), create: mockCreate, delete: mockDeleteItem, fetchAll: vi.fn(), setPage: vi.fn() };
}

function setSampleUsers() {
  mockCRUDReturn[0] = { items: sampleUsers, loading: false, error: null, total: 2, page: 1, pageSize: 50 };
  mockCRUDReturn[1] = { refresh: vi.fn(), create: mockCreate, delete: mockDeleteItem, fetchAll: vi.fn(), setPage: vi.fn() };
}

function setLoading() {
  mockCRUDReturn[0] = { items: [], loading: true, error: null, total: 0, page: 1, pageSize: 50 };
  mockCRUDReturn[1] = { refresh: vi.fn(), create: mockCreate, delete: mockDeleteItem, fetchAll: vi.fn(), setPage: vi.fn() };
}

describe('UserManager - 基础渲染', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockShowConfirm.mockResolvedValue(false);
    mockIsAdminReturn.isAdmin = true;
    mockIsAdminReturn.user = { role: 'admin' };
    setEmptyUsers();
  });

  it('管理员访问时渲染用户管理页面，包含标题和创建按钮', async () => {
    renderUserManager();
    await waitFor(() => {
      expect(screen.getByText('nav.users')).toBeInTheDocument();
      expect(screen.getByLabelText('user.createUser')).toBeInTheDocument();
    });
  });

  it('空用户列表时显示表头', async () => {
    renderUserManager();
    await waitFor(() => {
      expect(screen.getByText('auth.username')).toBeInTheDocument();
      expect(screen.getByText('auth.email')).toBeInTheDocument();
      expect(screen.getByText('user.role')).toBeInTheDocument();
    });
  });

  it('非管理员访问时显示权限提示', async () => {
    mockIsAdminReturn.isAdmin = false;
    mockIsAdminReturn.user = { role: 'user' };
    renderUserManager();

    await waitFor(() => {
      expect(screen.getByText('Shield')).toBeInTheDocument();
      expect(screen.getByText('user.adminRequired')).toBeInTheDocument();
      expect(screen.getByText('user.noPermission')).toBeInTheDocument();
    });
  });
});

describe('UserManager - 用户列表渲染', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockShowConfirm.mockResolvedValue(false);
    mockIsAdminReturn.isAdmin = true;
    mockIsAdminReturn.user = { role: 'admin' };
    setSampleUsers();
  });

  it('渲染用户列表，显示用户名和邮箱', async () => {
    renderUserManager();
    await waitFor(() => {
      expect(screen.getByText('admin')).toBeInTheDocument();
      expect(screen.getByText('operator')).toBeInTheDocument();
      expect(screen.getByText('admin@test.com')).toBeInTheDocument();
      expect(screen.getByText('operator@test.com')).toBeInTheDocument();
    });
  });

  it('管理员用户显示管理员标签', async () => {
    renderUserManager();
    await waitFor(() => {
      expect(screen.getByText('user.admin')).toBeInTheDocument();
      expect(screen.getByText('user.user')).toBeInTheDocument();
    });
  });

  it('每行用户显示删除按钮', async () => {
    renderUserManager();
    await waitFor(() => {
      const deleteButtons = screen.getAllByLabelText('common.delete');
      expect(deleteButtons.length).toBe(2);
    });
  });
});

describe('UserManager - 创建用户', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockShowConfirm.mockResolvedValue(false);
    mockIsAdminReturn.isAdmin = true;
    mockIsAdminReturn.user = { role: 'admin' };
    setEmptyUsers();
  });

  it('点击创建用户按钮打开创建模态框', async () => {
    renderUserManager();

    const createButton = screen.getByLabelText('user.createUser');
    fireEvent.click(createButton);

    await waitFor(() => {
      // 模态框标题使用 h2 标签
      expect(screen.getByRole('heading', { name: 'user.createUser' })).toBeInTheDocument();
    });
  });

  it('创建模态框包含表单字段（用户名、密码、邮箱、角色）', async () => {
    const { container } = renderUserManager();

    const createButton = screen.getByLabelText('user.createUser');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(container.querySelector('input[name="username"]')).toBeInTheDocument();
      expect(container.querySelector('input[name="password"]')).toBeInTheDocument();
      expect(container.querySelector('input[name="email"]')).toBeInTheDocument();
      expect(container.querySelector('select[name="role"]')).toBeInTheDocument();
    });
  });

  it('创建模态框中点击取消按钮关闭模态框', async () => {
    renderUserManager();

    const createButton = screen.getByLabelText('user.createUser');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'user.createUser' })).toBeInTheDocument();
    });

    // 取消按钮有 aria-label
    const cancelButton = screen.getByLabelText('common.cancel');
    fireEvent.click(cancelButton);

    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: 'user.createUser' })).not.toBeInTheDocument();
    });
  });
});

describe('UserManager - 删除用户', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockIsAdminReturn.isAdmin = true;
    mockIsAdminReturn.user = { role: 'admin' };
    setSampleUsers();
  });

  it('确认删除后调用 delete 并显示成功提示', async () => {
    mockShowConfirm.mockResolvedValue(true);
    mockDeleteItem.mockResolvedValue(true);
    renderUserManager();

    await waitFor(() => {
      const deleteButtons = screen.getAllByLabelText('common.delete');
      fireEvent.click(deleteButtons[0]);
    });

    await waitFor(() => {
      expect(mockShowConfirm).toHaveBeenCalledWith(
        expect.objectContaining({
          title: 'user.confirmDeleteTitle',
          variant: 'danger',
        })
      );
      expect(mockDeleteItem).toHaveBeenCalledWith('1');
      expect(mockShowToast).toHaveBeenCalledWith({ type: 'success', message: 'user.deleteSuccess' });
    });
  });

  it('取消删除时不调用 delete', async () => {
    mockShowConfirm.mockResolvedValue(false);
    renderUserManager();

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
    renderUserManager();

    await waitFor(() => {
      const deleteButtons = screen.getAllByLabelText('common.delete');
      fireEvent.click(deleteButtons[0]);
    });

    await waitFor(() => {
      expect(mockShowToast).toHaveBeenCalledWith({ type: 'error', message: 'errors.unknown' });
    });
  });
});

describe('UserManager - 加载状态', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockIsAdminReturn.isAdmin = true;
    mockIsAdminReturn.user = { role: 'admin' };
    setLoading();
  });

  it('加载中显示骨架屏', async () => {
    const { container } = renderUserManager();
    await waitFor(() => {
      expect(container.querySelectorAll('.animate-pulse').length).toBeGreaterThan(0);
    });
  });
});
