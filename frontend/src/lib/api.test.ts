import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ApiClient, TimeoutError, DEFAULT_TIMEOUT, AGENT_TIMEOUT } from './api';

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

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
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

// Mock sessionStorage (SEC-LOW-02: API uses sessionStorage for tokens)
const sessionStorageMock = (() => {
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
Object.defineProperty(window, 'sessionStorage', { value: sessionStorageMock });

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    origin: 'http://localhost:3000',
  },
  writable: true,
});

// Mock AbortController
class MockAbortController {
  signal = { aborted: false };
  abort() {
    this.signal.aborted = true;
  }
}
vi.stubGlobal('AbortController', MockAbortController);

describe('ApiClient', () => {
  let api: ApiClient;

  beforeEach(() => {
    vi.clearAllMocks();
    localStorageMock.clear();
    sessionStorageMock.clear();
    api = new ApiClient();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('constructor', () => {
    it('should create instance with default base URL', () => {
      const client = new ApiClient();
      expect(client).toBeDefined();
    });

    it('should create instance with custom base URL', () => {
      const client = new ApiClient('/custom-api');
      expect(client).toBeDefined();
    });
  });

  describe('token management', () => {
    it('should set token and store in sessionStorage', () => {
      api.setToken('test-token-123');
      expect(api.getToken()).toBe('test-token-123');
      expect(sessionStorageMock.getItem('token')).toBe('test-token-123');
    });

    it('should remove token from sessionStorage when set to null', () => {
      api.setToken('test-token');
      api.setToken(null);
      expect(api.getToken()).toBe(null);
      expect(sessionStorageMock.getItem('token')).toBe(null);
    });

    it('should load token from sessionStorage on construction', () => {
      sessionStorageMock.setItem('token', 'stored-token');
      const newApi = new ApiClient();
      expect(newApi.getToken()).toBe('stored-token');
    });
  });

  describe('request cancellation', () => {
    it('should cancel specific request', () => {
      const requestId = 'test-request';
      // The cancelRequest method should work without throwing
      api.cancelRequest(requestId);
      expect(api).toBeDefined();
    });

    it('should cancel all requests', () => {
      api.cancelAllRequests();
      expect(api).toBeDefined();
    });
  });

  describe('login', () => {
    it('should login successfully and store token', async () => {
      const mockResponse = {
        token: 'login-token',
        user: {
          id: 1,
          username: 'admin',
          role: 'admin',
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.login('admin', 'password');

      expect(result.token).toBe('login-token');
      expect(result.user.username).toBe('admin');
      expect(api.getToken()).toBe('login-token');
    });

    it('should handle login error', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: () => Promise.resolve({ error: 'Invalid credentials' }),
      });

      await expect(api.login('wrong', 'wrong')).rejects.toThrow();
    });

    it('should throw error when no token in response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ user: { id: 1 } }),
      });

      await expect(api.login('admin', 'password')).rejects.toThrow('Login failed: no token in response');
    });
  });

  describe('register', () => {
    it('should register successfully', async () => {
      const mockResponse = {
        token: 'register-token',
        user: { id: 2, username: 'newuser' },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.register({
        username: 'newuser',
        password: 'password',
        email: 'new@example.com',
      });

      expect(result.token).toBe('register-token');
      expect(api.getToken()).toBe('register-token');
    });
  });

  describe('getDevices', () => {
    it('should fetch devices successfully', async () => {
      const mockResponse = {
        data: [
          { id: 'CNC-001', name: 'CNC Machine', status: 'online' },
          { id: 'INJ-001', name: 'Injection Molder', status: 'warning' },
        ],
        total: 2,
        page: 1,
        page_size: 20,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.getDevices(1, 20);

      expect(result.data).toHaveLength(2);
      expect(result.total).toBe(2);
      expect(result.page).toBe(1);
    });

    it('should fetch devices with pagination params', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ devices: [], total: 0, page: 2, page_size: 10 }),
      });

      await api.getDevices(2, 10);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('page=2'),
        expect.any(Object)
      );
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('page_size=10'),
        expect.any(Object)
      );
    });
  });

  describe('getDevice', () => {
    it('should fetch single device', async () => {
      const mockDevice = { id: 'CNC-001', name: 'CNC Machine', status: 'online' };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDevice),
      });

      const result = await api.getDevice('CNC-001');

      expect(result.id).toBe('CNC-001');
    });
  });

  describe('createDevice', () => {
    it('should create device successfully', async () => {
      const mockDevice = { id: 'NEW-001', name: 'New Device', status: 'offline' };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDevice),
      });

      const result = await api.createDevice({
        id: 'NEW-001',
        name: 'New Device',
        type: 'CNC',
        status: 'offline',
      });

      expect(result.id).toBe('NEW-001');
    });
  });

  describe('updateDevice', () => {
    it('should update device successfully', async () => {
      const mockDevice = { id: 'CNC-001', name: 'Updated Name', status: 'online' };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDevice),
      });

      const result = await api.updateDevice('CNC-001', { name: 'Updated Name' });

      expect(result.name).toBe('Updated Name');
    });
  });

  describe('deleteDevice', () => {
    it('should delete device successfully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ message: 'Device deleted' }),
      });

      const result = await api.deleteDevice('CNC-001');

      expect(result.message).toBe('Device deleted');
    });
  });

  describe('getRules', () => {
    it('should fetch alert rules', async () => {
      const mockResponse = {
        data: [
          { id: 1, name: 'Temperature Rule', enabled: true },
          { id: 2, name: 'Vibration Rule', enabled: false },
        ],
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.getRules();

      expect(result.data).toHaveLength(2);
    });
  });

  describe('agentQuery', () => {
    it('should send agent query with extended timeout', async () => {
      const mockResponse = {
        response: 'AI analysis result',
        confidence: 0.95,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.agentQuery('分析设备状态', 'CNC-001');

      expect(result.response).toBe('AI analysis result');
    });

    it('should send agent query without device ID', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ response: 'General analysis' }),
      });

      await api.agentQuery('分析整体状态');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          body: expect.stringContaining('分析整体状态'),
        })
      );
    });
  });

  describe('getWorkOrders', () => {
    it('should fetch work orders', async () => {
      const mockResponse = {
        data: [
          { id: 1, title: 'Repair CNC', status: 'pending' },
        ],
        total: 1,
        page: 1,
        page_size: 20,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.getWorkOrders();

      expect(result.data).toHaveLength(1);
    });

    it('should fetch work orders with status filter', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ data: [], total: 0, page: 1, page_size: 20 }),
      });

      await api.getWorkOrders({ status: 'pending' });

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('status=pending'),
        expect.any(Object)
      );
    });
  });

  describe('getNotifications', () => {
    it('should fetch notifications', async () => {
      const mockResponse = {
        data: [{ id: 1, type: 'alert', message: 'New alert' }],
        total: 1,
        page: 1,
        page_size: 20,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.getNotifications();

      expect(result.data).toHaveLength(1);
    });

    it('should fetch unread notifications', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ data: [], total: 0, page: 1, page_size: 20 }),
      });

      await api.getNotifications({ unread: true });

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('unread=true'),
        expect.any(Object)
      );
    });
  });

  describe('getROIStats', () => {
    it('should fetch ROI stats', async () => {
      const mockResponse = {
        total_savings: 50000,
        maintenance_reduction: 30,
        uptime_improvement: 15,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.getROIStats();

      expect(result.total_savings).toBe(50000);
    });
  });

  describe('getUsers', () => {
    it('should fetch users', async () => {
      const mockResponse = {
        data: [
          { id: 1, username: 'admin', role: 'admin' },
          { id: 2, username: 'operator', role: 'operator' },
        ],
        total: 2,
        page: 1,
        page_size: 20,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.getUsers();

      expect(result.data).toHaveLength(2);
    });
  });

  describe('healthCheck', () => {
    it('should check health status', async () => {
      const mockResponse = {
        status: 'healthy',
        version: '1.0.0',
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.healthCheck();

      expect(result.status).toBe('healthy');
    });
  });

  describe('error handling', () => {
    it('should throw error on 401 unauthorized', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: () => Promise.resolve({ error: 'Unauthorized' }),
      });

      await expect(api.getDevices()).rejects.toThrow('Unauthorized');
    });

    it('should throw error on network failure', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(api.getDevices()).rejects.toThrow('Network error');
    });

    it('should clear token on 401', async () => {
      api.setToken('test-token');
      
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: () => Promise.resolve({ error: 'Unauthorized' }),
      });

      try {
        await api.getDevices();
      } catch {
        // Expected to throw
      }

      expect(api.getToken()).toBe(null);
    });
  });

  describe('timeout handling', () => {
    it('should use default timeout for regular requests', () => {
      expect(DEFAULT_TIMEOUT).toBe(30000);
    });

    it('should use extended timeout for agent queries', () => {
      expect(AGENT_TIMEOUT).toBe(60000);
    });
  });

  describe('TimeoutError', () => {
    it('should create TimeoutError with default message', () => {
      const error = new TimeoutError();
      expect(error.message).toBe('请求超时，请稍后重试');
      expect(error.name).toBe('TimeoutError');
    });

    it('should create TimeoutError with custom message', () => {
      const error = new TimeoutError('Custom timeout message');
      expect(error.message).toBe('Custom timeout message');
    });
  });

  describe('authorization header', () => {
    it('should include authorization header when token is set', async () => {
      api.setToken('auth-token');
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ devices: [], total: 0, page: 1, page_size: 20 }),
      });

      await api.getDevices();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer auth-token',
          }),
        })
      );
    });

    it('should not include authorization header when token is null', async () => {
      api.setToken(null);
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ devices: [], total: 0, page: 1, page_size: 20 }),
      });

      await api.getDevices();

      const callArgs = mockFetch.mock.calls[0][1];
      expect(callArgs.headers.Authorization).toBeUndefined();
    });
  });

  describe('exportReport', () => {
    it('should export report and return blob', async () => {
      const mockBlob = new Blob(['report content'], { type: 'application/pdf' });
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        blob: () => Promise.resolve(mockBlob),
        headers: {
          get: (name: string) => {
            if (name === 'Content-Disposition') return 'attachment; filename="report.pdf"';
            if (name === 'Content-Type') return 'application/pdf';
            return null;
          },
        },
      });

      api.setToken('export-token');
      const result = await api.exportReport('devices', 'pdf');

      expect(result.filename).toBe('report.pdf');
      expect(result.mimeType).toBe('application/pdf');
    });

    it('should export report with date range', async () => {
      const mockBlob = new Blob(['report'], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        blob: () => Promise.resolve(mockBlob),
        headers: {
          get: () => null,
        },
      });

      await api.exportReport('alerts', 'xlsx', '2024-01-01', '2024-01-31');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('start_date=2024-01-01'),
        expect.any(Object)
      );
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('end_date=2024-01-31'),
        expect.any(Object)
      );
    });
  });

  describe('getBlackBoxRecords', () => {
    it('should fetch black box records', async () => {
      const mockResponse = {
        data: [{ id: 1, device_id: 'CNC-001', timestamp: '2024-01-15T10:00:00Z' }],
        total: 1,
        page: 1,
        page_size: 20,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.getBlackBoxRecords();

      expect(result.data).toHaveLength(1);
    });
  });

  describe('getReports', () => {
    it('should fetch reports', async () => {
      const mockResponse = {
        data: [{ id: 1, type: 'daily', created_at: '2024-01-15' }],
        total: 1,
        page: 1,
        page_size: 20,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await api.getReports();

      expect(result.data).toHaveLength(1);
    });
  });
});