import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ isAdmin: false, user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    getRules: vi.fn().mockResolvedValue({ data: [] }),
    getRule: vi.fn().mockResolvedValue({}),
    createRule: vi.fn().mockResolvedValue(undefined),
    updateRule: vi.fn().mockResolvedValue(undefined),
    deleteRule: vi.fn().mockResolvedValue(undefined),
    toggleRule: vi.fn().mockResolvedValue(undefined),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('../hooks/useCRUD', () => ({
  useCRUD: () => [
    { items: [], loading: false, error: null, total: 0, page: 1, pageSize: 100 },
    { refresh: vi.fn(), create: vi.fn(), update: vi.fn(), delete: vi.fn(), fetchAll: vi.fn(), setPage: vi.fn() },
  ],
}));
vi.mock('./UI/ConfirmDialog', () => ({ useConfirmDialog: () => ({ showConfirm: vi.fn() }) }));
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

describe('RuleManager', () => {
  it('renders rule management page with header and table', async () => {
    const { container } = render(
      <MemoryRouter>
        <RuleManager />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面标题已渲染（i18n key）
      expect(container.textContent).toContain('nav.rules');
      // 验证创建规则按钮已渲染
      expect(container.textContent).toContain('alert.createRule');
      // 验证规则表格已渲染（表头包含 alert.ruleName）
      expect(container.textContent).toContain('alert.ruleName');
    });
  });
});