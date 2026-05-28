import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Mock dependencies
vi.mock('./AuthContext', () => ({
  useAuth: vi.fn().mockReturnValue({
    isAuthenticated: true,
    user: { id: '1', username: 'test' },
  }),
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => vi.fn(),
    useLocation: () => ({ pathname: '/dashboard' }),
  };
});

vi.mock('./Sidebar', () => ({
  default: ({ isOpen }: { isOpen: boolean }) => (
    <div data-testid="sidebar" className={isOpen ? 'open' : 'closed'} />
  ),
}));

vi.mock('./MobileNavBar', () => ({
  default: () => <div data-testid="mobile-nav" />,
}));

vi.mock('./Toast', () => ({
  useToast: () => ({ showToast: vi.fn() }),
}));

vi.mock('../lib/performance', () => ({
  usePerformance: vi.fn(),
}));

vi.mock('../lib/useSwipe', () => ({
  useSwipe: vi.fn(),
  useIsMobile: () => false,
  useViewportHeight: vi.fn(),
}));

vi.mock('../hooks/useWebSocket', () => ({
  useWebSocket: vi.fn().mockReturnValue({
    isConnected: true,
    send: vi.fn(),
    reconnect: vi.fn(),
  }),
}));

vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('./UI/ConfirmDialog', () => ({
  ConfirmDialogProvider: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="confirm-dialog-provider">{children}</div>
  ),
}));

// Import after mocks
import App from './App';

describe('App', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders without crashing when authenticated', () => {
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <App />
      </MemoryRouter>
    );

    // Should render sidebar and outlet container
    expect(screen.getByTestId('sidebar')).toBeInTheDocument();
  });

  it('renders sidebar in closed state by default', () => {
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <App />
      </MemoryRouter>
    );

    const sidebar = screen.getByTestId('sidebar');
    expect(sidebar).toHaveClass('closed');
  });

  it('renders mobile nav on mobile devices', async () => {
    // Re-mock useIsMobile for this test
    vi.mock('../lib/useSwipe', () => ({
      useSwipe: vi.fn(),
      useIsMobile: () => true,
      useViewportHeight: vi.fn(),
    }));

    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <App />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByTestId('mobile-nav')).toBeInTheDocument();
    });
  });
});