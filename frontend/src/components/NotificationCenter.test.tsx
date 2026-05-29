import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    getNotifications: vi.fn().mockResolvedValue({ data: [] }),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({
  Bell: () => <span>Bell</span>,
  Check: () => <span>Check</span>,
  Trash2: () => <span>Trash2</span>,
  Filter: () => <span>Filter</span>,
}));

import NotificationCenter from './NotificationCenter';

describe('NotificationCenter', () => {
  it('renders notification center with header and empty state', async () => {
    const { container } = render(
      <MemoryRouter>
        <NotificationCenter />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面标题已渲染
      expect(container.textContent).toContain('nav.notifications');
      // 验证通知标题和筛选区已渲染
      expect(container.textContent).toContain('notification.title');
      // API 返回空数据，应显示"无通知"提示
      expect(container.textContent).toContain('notification.noNotifications');
    });
  });
});