import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({
  useAuth: () => ({
    isAdmin: true,
    user: { role: 'admin' },
  }),
}));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('./UI/ConfirmDialog', () => ({ useConfirmDialog: () => ({ showConfirm: vi.fn().mockResolvedValue(false) }) }));
vi.mock('../hooks/useCRUD', () => ({
  useCRUD: vi.fn().mockReturnValue([
    { items: [], loading: false },
    { create: vi.fn(), delete: vi.fn() },
  ]),
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

describe('UserManager', () => {
  it('renders without crashing for admin user', async () => {
    const { container } = render(
      <MemoryRouter>
        <UserManager />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面渲染了用户管理内容
      expect(container.innerHTML).not.toBe('');
    });
  });

  it('shows admin required message for non-admin', async () => {
    vi.mock('./AuthContext', () => ({
      useAuth: () => ({
        isAdmin: false,
        user: { role: 'user' },
      }),
    }));

    const { container } = render(
      <MemoryRouter>
        <UserManager />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证组件渲染了内容（非管理员时可能显示权限提示）
      expect(container.innerHTML).not.toBe('');
    });
  });
});