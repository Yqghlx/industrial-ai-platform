import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Mock api module
vi.mock('../lib/api', () => ({
  default: {
    getToken: vi.fn().mockReturnValue(null),
    setToken: vi.fn(),
    removeToken: vi.fn(),
    login: vi.fn().mockResolvedValue({
      token: 'mock-jwt-token',
      user: { id: 1, username: 'admin', role: 'admin' },
    }),
    logout: vi.fn(),
  },
}));

// Import after mocks
import { AuthProvider, useAuth } from './AuthContext';

// Test component to access auth context
const TestComponent = () => {
  const { user, isAuthenticated, isAdmin, login, logout } = useAuth();
  return (
    <div>
      <span data-testid="auth-status">{isAuthenticated ? 'authenticated' : 'not-authenticated'}</span>
      <span data-testid="user-name">{user?.username || 'no-user'}</span>
      <span data-testid="is-admin">{isAdmin ? 'admin' : 'not-admin'}</span>
      <button onClick={() => login('test', 'password')} data-testid="login-btn">Login</button>
      <button onClick={logout} data-testid="logout-btn">Logout</button>
    </div>
  );
};

describe('AuthContext', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('provides auth context to children', () => {
    render(
      <MemoryRouter>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </MemoryRouter>
    );

    expect(screen.getByTestId('auth-status')).toBeInTheDocument();
    expect(screen.getByTestId('user-name')).toBeInTheDocument();
  });

  it('starts with no authentication when no token', () => {
    render(
      <MemoryRouter>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </MemoryRouter>
    );

    expect(screen.getByTestId('auth-status').textContent).toBe('not-authenticated');
    expect(screen.getByTestId('user-name').textContent).toBe('no-user');
  });

  it('login updates authentication state', async () => {
    render(
      <MemoryRouter>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </MemoryRouter>
    );

    const loginBtn = screen.getByTestId('login-btn');
    fireEvent.click(loginBtn);

    await waitFor(() => {
      expect(screen.getByTestId('auth-status').textContent).toBe('authenticated');
    });

    await waitFor(() => {
      expect(screen.getByTestId('user-name').textContent).toBe('admin');
    });
  });

  it('logout clears authentication state', async () => {
    // First login
    render(
      <MemoryRouter>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </MemoryRouter>
    );

    const loginBtn = screen.getByTestId('login-btn');
    fireEvent.click(loginBtn);

    await waitFor(() => {
      expect(screen.getByTestId('auth-status').textContent).toBe('authenticated');
    });

    // Then logout
    const logoutBtn = screen.getByTestId('logout-btn');
    fireEvent.click(logoutBtn);

    await waitFor(() => {
      expect(screen.getByTestId('auth-status').textContent).toBe('not-authenticated');
    });
  });

  it('identifies admin role correctly', async () => {
    render(
      <MemoryRouter>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </MemoryRouter>
    );

    const loginBtn = screen.getByTestId('login-btn');
    fireEvent.click(loginBtn);

    await waitFor(() => {
      expect(screen.getByTestId('is-admin').textContent).toBe('admin');
    });
  });
});