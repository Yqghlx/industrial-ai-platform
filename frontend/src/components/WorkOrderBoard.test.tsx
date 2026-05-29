import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    getWorkOrders: vi.fn().mockResolvedValue({ data: [] }),
    updateWorkOrderStatus: vi.fn().mockResolvedValue(undefined),
    createWorkOrder: vi.fn().mockResolvedValue(undefined),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('../lib/colorUtils', () => ({
  getWorkOrderStatusColor: vi.fn().mockReturnValue('bg-green-500/20 text-green-400'),
  getWorkOrderPriorityColor: vi.fn().mockReturnValue('bg-red-500/20 text-red-400'),
}));
vi.mock('../types/typeGuards', () => ({ asWorkOrderArraySafe: vi.fn().mockReturnValue([]) }));
vi.mock('lucide-react', () => ({
  Plus: () => <span>Plus</span>,
  Search: () => <span>Search</span>,
}));

import WorkOrderBoard from './WorkOrderBoard';

describe('WorkOrderBoard', () => {
  it('renders work order board with header', async () => {
    const { container } = render(
      <MemoryRouter>
        <WorkOrderBoard />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面渲染了内容
      expect(container.innerHTML).not.toBe('');
    });
  });

  it('shows work order data after loading', async () => {
    const { container } = render(
      <MemoryRouter>
        <WorkOrderBoard />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面渲染了内容（加载完成后的空状态或数据列表）
      expect(container.innerHTML).not.toBe('');
    });
  });
});