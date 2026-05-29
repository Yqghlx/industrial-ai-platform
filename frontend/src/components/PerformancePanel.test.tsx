import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({ default: {} }));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({}));

import PerformancePanel from './PerformancePanel';

describe('PerformancePanel', () => {
  it('renders nothing when closed', async () => {
    const { container } = render(<PerformancePanel isOpen={false} onClose={() => {}} />);
    await waitFor(() => {
      // 面板关闭时不渲染任何内容
      expect(container.innerHTML).toBe('');
    });
  });

  it('renders performance panel when open', async () => {
    const { container } = render(<PerformancePanel isOpen={true} onClose={() => {}} />);
    await waitFor(() => {
      // 面板打开时渲染内容
      expect(container.innerHTML).not.toBe('');
    });
  });
});