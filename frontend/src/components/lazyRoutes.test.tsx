import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';

vi.mock('./LoadingSpinner', () => ({
  RouteLoader: () => <div data-testid="route-loader">Loading...</div>,
}));

import { createLazyPage, preloadCriticalRoutes, preloadAllRoutes } from './lazyRoutes';

describe('lazyRoutes', () => {
  it('createLazyPage renders with fallback and loads component', async () => {
    const mockImportFn = vi.fn().mockResolvedValue({
      default: () => <div data-testid="loaded-component">Loaded</div>,
    });

    const result = createLazyPage(mockImportFn);
    render(result);

    // Should show loader initially
    expect(document.querySelector('[data-testid="route-loader"]')).toBeTruthy();

    await waitFor(() => {
      // 验证懒加载组件最终渲染完成
      expect(document.querySelector('[data-testid="loaded-component"]')).toBeTruthy();
    });
  });

  it('preloadCriticalRoutes function exists and is callable', () => {
    expect(preloadCriticalRoutes).toBeDefined();
    expect(typeof preloadCriticalRoutes).toBe('function');
  });

  it('preloadAllRoutes function exists and is callable', () => {
    expect(preloadAllRoutes).toBeDefined();
    expect(typeof preloadAllRoutes).toBe('function');
  });
});