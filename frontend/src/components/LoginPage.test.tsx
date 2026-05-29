import { describe, it, expect, vi } from 'vitest';
import { render, waitFor, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  };
});

vi.mock('./AuthContext', () => ({
  useAuth: () => ({
    login: vi.fn().mockResolvedValue(undefined),
    user: null,
  }),
}));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    register: vi.fn().mockResolvedValue(undefined),
  },
}));
vi.mock('../lib/errorHelper', () => ({
  parseApiError: vi.fn().mockReturnValue('Error'),
  ErrorType: { NETWORK: 'network', TIMEOUT: 'timeout', RATE_LIMIT: 'rate_limit', UNAUTHORIZED: 'unauthorized', VALIDATION: 'validation' },
  getErrorType: vi.fn().mockReturnValue('unknown'),
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({
  Activity: () => <span>Activity</span>,
  Lock: () => <span>Lock</span>,
  User: () => <span>User</span>,
  Mail: () => <span>Mail</span>,
  AlertCircle: () => <span>AlertCircle</span>,
}));

import LoginPage from './LoginPage';

describe('LoginPage', () => {
  it('renders without crashing', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );
    // 验证登录页面渲染了平台标题
    await waitFor(() => {
      expect(screen.getByText('Industrial AI Platform')).toBeInTheDocument();
    });
  });

  it('renders login tab', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );
    expect(screen.getByTestId('login-tab')).toBeInTheDocument();
  });

  it('renders register tab', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );
    expect(screen.getByTestId('register-tab')).toBeInTheDocument();
  });

  it('switches to register mode when clicking register tab', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );
    
    fireEvent.click(screen.getByTestId('register-tab'));
    
    await waitFor(() => {
      expect(screen.getByTestId('register-tab')).toHaveClass('bg-primary-600');
    });
  });

  it('renders submit button', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );
    expect(screen.getByTestId('submit-button')).toBeInTheDocument();
  });
});