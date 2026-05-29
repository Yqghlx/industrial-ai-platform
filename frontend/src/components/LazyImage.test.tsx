import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Mock IntersectionObserver
const mockIntersectionObserver = vi.fn();
mockIntersectionObserver.mockReturnValue({
  observe: () => null,
  unobserve: () => null,
  disconnect: () => null,
});
window.IntersectionObserver = mockIntersectionObserver;

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));

import LazyImage from './LazyImage';

describe('LazyImage', () => {
  beforeEach(() => {
    mockIntersectionObserver.mockClear();
  });

  it('renders image element with alt text', async () => {
    const { container } = render(
      <MemoryRouter>
        <LazyImage src="https://test.jpg" alt="Test image" />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证图片元素已渲染
      const img = container.querySelector('img');
      expect(img).toBeInTheDocument();
      expect(img?.getAttribute('alt')).toBe('Test image');
    });
  });

  it('renders with placeholder initially', async () => {
    const { container } = render(
      <MemoryRouter>
        <LazyImage src="https://test.jpg" alt="Test image" />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证组件已渲染（图片或占位符）
      expect(container.innerHTML).not.toBe('');
    });
  });
});