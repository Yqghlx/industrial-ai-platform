import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({ default: {} }));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({}));

import LoadingSpinner from './LoadingSpinner';

describe('LoadingSpinner', () => {
  it('renders loading spinner with text', async () => {
    const { container } = render(<MemoryRouter><LoadingSpinner /></MemoryRouter>);
    await waitFor(() => {
      // 验证加载动画已渲染（包含旋转圆圈）
      expect(container.querySelector('.animate-spin')).toBeInTheDocument();
      // 验证默认加载文本
      expect(container.textContent).toContain('common.loading');
    });
  });
});