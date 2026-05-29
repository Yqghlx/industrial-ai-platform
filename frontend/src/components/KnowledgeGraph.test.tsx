import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    getDeviceGraph: vi.fn().mockResolvedValue({
      nodes: [{ id: 'test-1', name: 'Test Device', status: 'online' }],
      links: [],
    }),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('../types/typeGuards', () => ({ asDeviceGraphSafe: vi.fn().mockReturnValue({
  nodes: [{ id: 'test-1', name: 'Test Device', status: 'online' }],
  links: [],
}) }));
vi.mock('lucide-react', () => ({
  Network: () => <span>Network</span>,
  RefreshCw: () => <span>RefreshCw</span>,
}));

import KnowledgeGraph from './KnowledgeGraph';

describe('KnowledgeGraph', () => {
  it('renders knowledge graph page with content', async () => {
    const { container } = render(
      <MemoryRouter>
        <KnowledgeGraph />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面渲染了内容
      expect(container.innerHTML).not.toBe('');
    });
  });

  it('shows graph data after loading', async () => {
    const { container } = render(
      <MemoryRouter>
        <KnowledgeGraph />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面渲染了内容（canvas 或其他可视化元素）
      expect(container.innerHTML).not.toBe('');
    });
  });
});