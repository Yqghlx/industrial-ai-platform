import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({
  AlertTriangle: () => <span>AlertTriangle</span>,
  Info: () => <span>Info</span>,
  X: () => <span>X</span>,
}));

import { ConfirmDialogProvider, useConfirmDialog } from './ConfirmDialog';

describe('ConfirmDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders ConfirmDialogProvider with children', async () => {
    const { getByText } = render(
      <MemoryRouter>
        <ConfirmDialogProvider>
          <div>Test Content</div>
        </ConfirmDialogProvider>
      </MemoryRouter>
    );
    // 验证子组件被正确渲染
    expect(getByText('Test Content')).toBeInTheDocument();
  });

  it('provides useConfirmDialog hook', async () => {
    const TestComponent = () => {
      const { showConfirm } = useConfirmDialog();
      return (
        <div data-testid="test-component">
          {typeof showConfirm === 'function' ? 'Hook available' : 'Hook missing'}
        </div>
      );
    };

    const { getByTestId } = render(
      <MemoryRouter>
        <ConfirmDialogProvider>
          <TestComponent />
        </ConfirmDialogProvider>
      </MemoryRouter>
    );

    await waitFor(() => {
      // 验证 hook 正确提供了 showConfirm 函数
      expect(getByTestId('test-component').textContent).toBe('Hook available');
    });
  });
});