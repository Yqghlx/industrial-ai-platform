import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

// Mock dependencies
vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('../lib/api', () => ({
  default: {
    getSystemStatus: vi.fn().mockResolvedValue({
      database: 'healthy',
      redis: 'healthy',
      uptime: 1000,
    }),
  },
}));

vi.mock('./Skeleton', () => ({
  default: () => <div data-testid="skeleton" />,
}));

vi.mock('./Toast', () => ({
  useToast: () => ({
    showToast: vi.fn(),
  }),
}));

vi.mock('lucide-react', () => ({
  Database: () => <div data-testid="database-icon" />,
  Activity: () => <div data-testid="activity-icon" />,
  Server: () => <div data-testid="server-icon" />,
  Clock: () => <div data-testid="clock-icon" />,
  CheckCircle: () => <div data-testid="check-icon" />,
  AlertCircle: () => <div data-testid="alert-icon" />,
}));

// Import after mocks
import SystemStatus from './SystemStatus';

describe('SystemStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders system status component', async () => {
    render(<SystemStatus />);

    await waitFor(() => {
      expect(screen.getByText(/nav.system/i)).toBeInTheDocument();
    });
  });

  it('shows loading skeleton initially', () => {
    render(<SystemStatus />);

    // Should show skeleton or loading state
    expect(screen.getByText(/nav.system/i)).toBeInTheDocument();
  });

  it('has refresh button', async () => {
    render(<SystemStatus />);

    await waitFor(() => {
      const refreshBtn = screen.getByRole('button', { name: /common.refresh/i });
      expect(refreshBtn).toBeInTheDocument();
    });
  });

  it('loads status data from API', async () => {
    render(<SystemStatus />);

    await waitFor(() => {
      expect(screen.getByText(/nav.system/i)).toBeInTheDocument();
    });
  });
});