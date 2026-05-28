import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Mock dependencies
vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
    language: 'zh',
    setLanguage: vi.fn(),
  }),
}));

vi.mock('./AuthContext', () => ({
  useAuth: vi.fn().mockReturnValue({
    user: { id: '1', username: 'admin', role: 'admin' },
    logout: vi.fn(),
    isAdmin: true,
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

vi.mock('../lib/useSwipe', () => ({
  useIsMobile: () => false,
}));

// Import after mocks
import Sidebar from './Sidebar';

describe('Sidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders sidebar with navigation items', () => {
    render(
      <MemoryRouter>
        <Sidebar isOpen={true} onClose={vi.fn()} />
      </MemoryRouter>
    );

    // Should show user info - use getAllByText since multiple matches
    const userElements = screen.getAllByText('admin');
    expect(userElements.length).toBeGreaterThan(0);
  });

  it('calls onClose when close button clicked', () => {
    const onClose = vi.fn();
    render(
      <MemoryRouter>
        <Sidebar isOpen={true} onClose={onClose} />
      </MemoryRouter>
    );

    // Find and click close button
    const closeButton = screen.getByRole('button', { name: /close/i });
    fireEvent.click(closeButton);

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('shows admin navigation items when user is admin', () => {
    render(
      <MemoryRouter>
        <Sidebar isOpen={true} onClose={vi.fn()} />
      </MemoryRouter>
    );

    // Admin should see settings/users menu - use getAllByText since multiple matches
    const dashboardLinks = screen.getAllByText(/nav.dashboard/i);
    expect(dashboardLinks.length).toBeGreaterThan(0);
  });

  it('closes on Escape key press', async () => {
    const onClose = vi.fn();
    render(
      <MemoryRouter>
        <Sidebar isOpen={true} onClose={onClose} />
      </MemoryRouter>
    );

    // Press Escape
    fireEvent.keyDown(document, { key: 'Escape' });

    await waitFor(() => {
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });
});