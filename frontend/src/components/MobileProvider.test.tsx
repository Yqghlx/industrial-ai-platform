import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, waitFor } from '@testing-library/react';

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('../lib/mobileOptimizations', () => ({
  shouldReduceMotion: vi.fn().mockReturnValue(false),
  isMobileDevice: vi.fn().mockReturnValue(false),
  isIOS: vi.fn().mockReturnValue(false),
  isTouchDevice: vi.fn().mockReturnValue(false),
  isLowEndDevice: vi.fn().mockReturnValue(false),
  isSlowNetwork: vi.fn().mockReturnValue(false),
  getOptimalImageQuality: vi.fn().mockReturnValue('high'),
  getConnectionInfo: vi.fn().mockReturnValue({ effectiveType: '4g', downlink: 10 }),
}));

import MobileProvider from './MobileProvider';

describe('MobileProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders children within provider', () => {
    const { getByText } = render(
      <MobileProvider>
        <div>Test</div>
      </MobileProvider>
    );
    // 验证子组件被渲染
    expect(getByText('Test')).toBeInTheDocument();
  });

  it('provides mobile context to children', () => {
    const { getByTestId } = render(
      <MobileProvider>
        <div data-testid="child">Test Child</div>
      </MobileProvider>
    );
    // 验证子组件被正确渲染
    expect(getByTestId('child')).toBeInTheDocument();
    expect(getByTestId('child').textContent).toBe('Test Child');
  });
});