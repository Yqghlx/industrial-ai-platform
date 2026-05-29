import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ isAuthenticated: false, user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({ default: {} }));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({}));
vi.mock('./Sidebar', () => ({ default: () => <div data-testid="sidebar" /> }));
vi.mock('./MobileNavBar', () => ({ default: () => <div data-testid="mobile-nav" /> }));
vi.mock('./PerformancePanel', () => ({ PerformanceButton: () => <div data-testid="perf-btn" /> }));
vi.mock('../lib/performance', () => ({ usePerformance: () => {} }));
vi.mock('../lib/useSwipe', () => ({ useSwipe: () => ({}), useIsMobile: () => false, useViewportHeight: () => {} }));
vi.mock('../hooks/useWebSocket', () => ({ useWebSocket: () => ({ isConnected: false }) }));
vi.mock('./UI/ConfirmDialog', () => ({ ConfirmDialogProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div> }));

import App from './App';

describe('App', () => {
  it('renders nothing when not authenticated', async () => {
    const { container } = render(<MemoryRouter><App /></MemoryRouter>);
    // 未认证时组件返回 null
    await waitFor(() => {
      expect(container.innerHTML).toBe('');
    });
  });
});