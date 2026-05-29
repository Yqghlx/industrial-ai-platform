import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({ default: {} }));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({}));

import LazyWrapper from './LazyWrapper';

describe('LazyWrapper', () => {
  it('renders children within LazyWrapper', async () => {
    const { container } = render(
      <MemoryRouter>
        <LazyWrapper>
          <div data-testid="child">Loaded Content</div>
        </LazyWrapper>
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证子组件被渲染
      expect(container.textContent).toContain('Loaded Content');
    });
  });
});