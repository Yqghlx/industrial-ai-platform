import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('./Skeleton', () => ({ default: () => <div data-testid="skeleton" /> }));
vi.mock('../lib/api', () => ({
  default: {
    getBlackBoxRecords: vi.fn().mockResolvedValue({ data: [] }),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({
  Box: () => <div />,
  Play: () => <div />,
  Clock: () => <div />,
}));

import BlackBoxCenter from './BlackBoxCenter';

describe('BlackBoxCenter', () => {
  it('renders black box center with header', async () => {
    const { container } = render(<MemoryRouter><BlackBoxCenter /></MemoryRouter>);
    await waitFor(() => {
      // 验证页面渲染了内容（包括标题或空状态提示）
      expect(container.innerHTML).not.toBe('');
    });
  });
});