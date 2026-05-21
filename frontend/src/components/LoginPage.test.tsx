import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Mock dependencies
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
  }),
}));

vi.mock('./Toast', () => ({
  default: () => <div data-testid="toast-container" />,
  useToast: () => ({
    showToast: vi.fn(),
  }),
}));

vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'auth.login': '登录',
        'auth.register': '注册',
        'auth.username': '用户名',
        'auth.password': '密码',
        'auth.email': '邮箱',
        'auth.loginSuccess': '登录成功',
        'auth.registerSuccess': '注册成功',
        'auth.loginFailed': '登录失败',
        'common.loading': '加载中...',
      };
      return translations[key] || key;
    },
  }),
}));

vi.mock('../lib/api', () => ({
  default: {
    register: vi.fn().mockResolvedValue({ token: 'test-token', user: {} }),
  },
}));

// Import after mocks
import LoginPage from './LoginPage';
import api from '../lib/api';

const renderLoginPage = () => {
  return render(
    <MemoryRouter>
      <LoginPage />
    </MemoryRouter>
  );
};

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('rendering', () => {
    it('should render login form by default', () => {
      renderLoginPage();
      
      expect(screen.getByText('Industrial AI Platform')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('用户名')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('密码')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '登录' })).toBeInTheDocument();
    });

    it('should render login and register tabs', () => {
      renderLoginPage();
      
      expect(screen.getByRole('button', { name: '登录' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '注册' })).toBeInTheDocument();
    });

    it('should not show email field in login mode', () => {
      renderLoginPage();
      
      expect(screen.queryByPlaceholderText('邮箱')).not.toBeInTheDocument();
    });

    it('should show email field when switching to register mode', () => {
      renderLoginPage();
      
      const registerTab = screen.getByRole('button', { name: '注册' });
      fireEvent.click(registerTab);
      
      expect(screen.getByPlaceholderText('邮箱')).toBeInTheDocument();
    });
  });

  describe('form interactions', () => {
    it('should update username input value', () => {
      renderLoginPage();
      
      const usernameInput = screen.getByPlaceholderText('用户名');
      fireEvent.change(usernameInput, { target: { value: 'testuser' } });
      
      expect(usernameInput).toHaveValue('testuser');
    });

    it('should update password input value', () => {
      renderLoginPage();
      
      const passwordInput = screen.getByPlaceholderText('密码');
      fireEvent.change(passwordInput, { target: { value: 'password123' } });
      
      expect(passwordInput).toHaveValue('password123');
    });

    it('should update email input value in register mode', () => {
      renderLoginPage();
      
      // Switch to register mode
      fireEvent.click(screen.getByRole('button', { name: '注册' }));
      
      const emailInput = screen.getByPlaceholderText('邮箱');
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
      
      expect(emailInput).toHaveValue('test@example.com');
    });
  });

  describe('tab switching', () => {
    it('should highlight login tab by default', () => {
      renderLoginPage();
      
      const loginTab = screen.getByRole('button', { name: '登录' });
      const registerTab = screen.getByRole('button', { name: '注册' });
      
      expect(loginTab).toHaveClass('bg-primary-600');
      expect(registerTab).not.toHaveClass('bg-primary-600');
    });

    it('should switch to register mode', () => {
      renderLoginPage();
      
      const registerTab = screen.getByRole('button', { name: '注册' });
      fireEvent.click(registerTab);
      
      expect(registerTab).toHaveClass('bg-primary-600');
      expect(screen.getByRole('button', { type: 'submit' })).toHaveTextContent('注册');
    });
  });

  describe('form submission', () => {
    it('should have submit button with type submit', () => {
      renderLoginPage();
      
      const submitButton = screen.getByRole('button', { type: 'submit' });
      expect(submitButton).toBeInTheDocument();
    });

    it('should show loading state during submission', async () => {
      renderLoginPage();
      
      const usernameInput = screen.getByPlaceholderText('用户名');
      const passwordInput = screen.getByPlaceholderText('密码');
      
      fireEvent.change(usernameInput, { target: { value: 'admin' } });
      fireEvent.change(passwordInput, { target: { value: 'admin123' } });
      
      const submitButton = screen.getByRole('button', { type: 'submit' });
      
      // Submit form
      fireEvent.click(submitButton);
      
      // Check loading state appears immediately
      await waitFor(() => {
        expect(screen.getByText('加载中...')).toBeInTheDocument();
      });
    });
  });

  describe('demo hint', () => {
    it('should show demo credentials', () => {
      renderLoginPage();
      
      expect(screen.getByText(/演示账户/)).toBeInTheDocument();
      expect(screen.getByText(/admin \/ admin123/)).toBeInTheDocument();
    });
  });

  describe('register flow', () => {
    it('should call api.register with correct data', async () => {
      const mockRegister = vi.mocked(api.register);
      renderLoginPage();
      
      // Switch to register mode
      fireEvent.click(screen.getByRole('button', { name: '注册' }));
      
      // Fill form
      fireEvent.change(screen.getByPlaceholderText('用户名'), { target: { value: 'newuser' } });
      fireEvent.change(screen.getByPlaceholderText('密码'), { target: { value: 'password123' } });
      fireEvent.change(screen.getByPlaceholderText('邮箱'), { target: { value: 'new@example.com' } });
      
      // Submit
      fireEvent.click(screen.getByRole('button', { type: 'submit' }));
      
      await waitFor(() => {
        expect(mockRegister).toHaveBeenCalledWith({
          username: 'newuser',
          password: 'password123',
          email: 'new@example.com',
        });
      });
    });
  });
});