import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock fetch globally
const mockFetch = vi.fn();
globalThis.fetch = mockFetch as typeof fetch;

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

// Import after mocking - ApiClient is exported as class and default instance
import { ApiClient } from './api';

describe('ApiClient', () => {
  let apiClient: ApiClient;

  beforeEach(() => {
    apiClient = new ApiClient('/api/v1');
    localStorage.clear();
    mockFetch.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('constructor', () => {
    it('should create instance with default base URL', () => {
      const client = new ApiClient();
      expect(client).toBeInstanceOf(ApiClient);
    });

    it('should create instance with custom base URL', () => {
      const client = new ApiClient('/custom/api');
      expect(client).toBeInstanceOf(ApiClient);
    });
  });

  describe('token management', () => {
    it('should set token', () => {
      apiClient.setToken('test-token');
      expect(apiClient.getToken()).toBe('test-token');
      expect(localStorage.getItem('token')).toBe('test-token');
    });

    it('should remove token when set to null', () => {
      apiClient.setToken('test-token');
      apiClient.setToken(null);
      expect(apiClient.getToken()).toBeNull();
      expect(localStorage.getItem('token')).toBeNull();
    });

    it('should load token from localStorage on construction', () => {
      localStorage.setItem('token', 'stored-token');
      const client = new ApiClient();
      expect(client.getToken()).toBe('stored-token');
    });
  });

  describe('login', () => {
    it('should call login endpoint and set token', async () => {
      const mockResponse = { access_token: 'login-token', refresh_token: 'refresh', expires_in: 3600, token_type: 'Bearer', user: { id: 1, username: 'testuser', role: 'user' } };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await apiClient.login('testuser', 'password123');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:3000/api/v1/auth/login',
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
          body: JSON.stringify({ username: 'testuser', password: 'password123' }),
        })
      );
      expect(result.token).toBe('login-token');
      expect(result.user.username).toBe('testuser');
      expect(apiClient.getToken()).toBe('login-token');
    });

    it('should throw error on login failure', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: () => Promise.resolve({ error: 'Invalid credentials', code: 'AUTH_FAILED' }),
      });

      // Set token before to verify it's cleared
      apiClient.setToken('old-token');

      await expect(apiClient.login('testuser', 'wrongpassword')).rejects.toThrow('Unauthorized');
      expect(apiClient.getToken()).toBeNull();
    });
  });

  describe('register', () => {
    it('should call register endpoint and set token', async () => {
      const mockResponse = { token: 'register-token', user: { id: 2, username: 'newuser' } };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await apiClient.register({
        username: 'newuser',
        password: 'password123',
        email: 'new@example.com',
      });

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:3000/api/v1/auth/register',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({
            username: 'newuser',
            password: 'password123',
            email: 'new@example.com',
          }),
        })
      );
      expect(result).toEqual(mockResponse);
      expect(apiClient.getToken()).toBe('register-token');
    });
  });

  describe('getDevices', () => {
    it('should fetch devices with default pagination', async () => {
      const mockResponse = { devices: [{ id: '1', name: 'Device 1' }], total: 1, page: 1, page_size: 20 };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await apiClient.getDevices();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/devices?page=1&page_size=20'),
        expect.any(Object)
      );
      expect(result).toEqual({
        data: [{ id: '1', name: 'Device 1' }],
        total: 1,
        page: 1,
        page_size: 20,
      });
    });

    it('should fetch devices with custom pagination', async () => {
      const mockResponse = { devices: [], total: 0, page: 2, page_size: 50 };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      await apiClient.getDevices(2, 50);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('page=2&page_size=50'),
        expect.any(Object)
      );
    });
  });

  describe('getDevice', () => {
    it('should fetch single device by id', async () => {
      const mockResponse = { id: 'device-123', name: 'Test Device' };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await apiClient.getDevice('device-123');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/devices/device-123'),
        expect.any(Object)
      );
      expect(result).toEqual(mockResponse);
    });
  });

  describe('createDevice', () => {
    it('should create device with POST request', async () => {
      const deviceData = { id: 'device-1', name: 'New Device', type: 'sensor', status: 'online' as const, location: 'Building A' };
      const mockResponse = { ...deviceData };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await apiClient.createDevice(deviceData);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/devices'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(deviceData),
        })
      );
      expect(result).toEqual(mockResponse);
    });
  });

  describe('deleteDevice', () => {
    it('should delete device with DELETE request', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ message: 'Device deleted' }),
      });

      await apiClient.deleteDevice('device-123');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/devices/device-123'),
        expect.objectContaining({
          method: 'DELETE',
        })
      );
    });
  });

  describe('agentQuery', () => {
    it('should send agent query', async () => {
      apiClient.setToken('test-token');
      const mockResponse = {
        session_id: 'session-1',
        response: 'AI response',
        agent: 'analytics',
      };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await apiClient.agentQuery('分析设备状态', 'device-123');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/agent/query'),
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            Authorization: 'Bearer test-token',
          }),
          body: JSON.stringify({ query: '分析设备状态', device_id: 'device-123' }),
        })
      );
      expect(result).toEqual(mockResponse);
    });
  });

  describe('authorization header', () => {
    it('should include authorization header when token is set', async () => {
      apiClient.setToken('my-auth-token');
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ data: [] }),
      });

      await apiClient.getDevices();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer my-auth-token',
          }),
        })
      );
    });

    it('should not include authorization header when token is not set', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ data: [] }),
      });

      await apiClient.getDevices();

      const call = mockFetch.mock.calls[0];
      expect(call[1].headers).not.toHaveProperty('Authorization');
    });
  });

  describe('error handling', () => {
    it('should throw error with message from API', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: () => Promise.resolve({ error: 'Device not found', code: 'NOT_FOUND' }),
      });

      await expect(apiClient.getDevice('invalid-id')).rejects.toThrow('Device not found');
    });

    it('should throw generic error when no error message', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: () => Promise.resolve({}),
      });

      await expect(apiClient.getDevice('device-id')).rejects.toThrow('Request failed');
    });
  });
});