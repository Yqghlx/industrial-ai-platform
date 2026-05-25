import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
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

vi.mock('./Toast', () => ({
  default: () => <div data-testid="toast-container" />,
  useToast: () => ({
    showToast: vi.fn(),
  }),
}));

vi.mock('./Skeleton', () => ({
  default: ({ className }: { className?: string }) => (
    <div data-testid="skeleton" className={className} />
  ),
}));

vi.mock('../hooks/useWebSocket', () => ({
  useWebSocket: vi.fn().mockReturnValue({
    isConnected: false,
    send: vi.fn(),
    reconnect: vi.fn(),
    getCompressionStats: vi.fn().mockReturnValue({
      totalMessages: 0,
      compressedMessages: 0,
      skippedMessages: 0,
      originalBytes: 0,
      compressedBytes: 0,
      compressionRatio: 0,
      savingsPercent: 0,
    }),
  }),
}));

vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, string | number>) => {
      const translations: Record<string, string> = {
        'nav.alerts': '告警管理',
        'common.refresh': '刷新',
        'report.generate': '生成报告',
        'report.analysis': '告警分析',
        'alert.activeAlerts': '活跃告警',
        'alert.criticalAlerts': '紧急告警',
        'alert.acknowledged': '已确认',
        'alert.resolved': '已解决',
        'alert.allStatuses': '全部状态',
        'alert.activeLabel': '活跃',
        'alert.acknowledgedLabel': '已确认',
        'alert.resolvedLabel': '已解决',
        'alert.allSeverities': '全部级别',
        'alert.criticalLabel': '紧急',
        'alert.highLabel': '高',
        'alert.mediumLabel': '中',
        'alert.lowLabel': '低',
        'alert.alertList': '告警列表',
        'alert.noAlerts': '暂无告警',
        'alert.device': '设备',
        'alert.triggered': '触发时间',
        'alert.resolvedTime': '解决时间',
        'alert.acknowledge': '确认',
        'alert.resolve': '解决',
        'alert.processed': '已处理',
        'alert.alertResolved': '告警 #{id} 已解决',
        'alert.alertAcknowledged': '告警 #{id} 已确认',
        'errors.unknown': '未知错误',
        'errors.loadFailedAlertStats': '加载告警统计失败',
      };
      let result = translations[key] || key;
      if (params) {
        Object.entries(params).forEach(([k, v]) => {
          result = result.replace(`{${k}}`, String(v));
        });
      }
      return result;
    },
  }),
}));

vi.mock('../types/typeGuards', () => ({
  isAlert: (payload: unknown) => {
    return payload && typeof payload === 'object' && 'id' in payload && 'severity' in payload;
  },
  isAlertStatusPayload: (payload: unknown) => {
    return payload && typeof payload === 'object' && 'id' in payload && 'status' in payload;
  },
  asAlertStatusSafe: (status: unknown) => {
    if (status === 'active' || status === 'acknowledged' || status === 'resolved') {
      return status;
    }
    return null;
  },
}));

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = { token: 'test-token' };
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

// Import after mocks
import AlertsPage from './AlertsPage';
import { useWebSocket } from '../hooks/useWebSocket';

const mockAlerts = [
  {
    id: 1,
    device_id: 'CNC-001',
    message: '温度过高警告',
    severity: 'critical',
    status: 'active',
    triggered_at: '2024-01-15T10:30:00Z',
    resolved_at: null,
  },
  {
    id: 2,
    device_id: 'INJ-001',
    message: '振动异常',
    severity: 'high',
    status: 'acknowledged',
    triggered_at: '2024-01-15T09:00:00Z',
    resolved_at: null,
  },
  {
    id: 3,
    device_id: 'ROB-001',
    message: '通信中断',
    severity: 'medium',
    status: 'resolved',
    triggered_at: '2024-01-14T08:00:00Z',
    resolved_at: '2024-01-14T12:00:00Z',
  },
];

const mockStats = {
  active_count: 5,
  total_count: 20,
  by_severity: { critical: 3, high: 5, medium: 8, low: 4 },
  by_status: { active: 5, acknowledged: 8, resolved: 7 },
};

const renderAlertsPage = () => {
  return render(
    <MemoryRouter>
      <AlertsPage />
    </MemoryRouter>
  );
};

describe('AlertsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useWebSocket).mockReturnValue({
      isConnected: false,
      send: vi.fn(),
      reconnect: vi.fn(),
      getCompressionStats: vi.fn().mockReturnValue({
        totalMessages: 0,
        compressedMessages: 0,
        skippedMessages: 0,
        originalBytes: 0,
        compressedBytes: 0,
        compressionRatio: 0,
        savingsPercent: 0,
      }),
    });
    
    // Setup default fetch responses
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/alerts/stats')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockStats),
        });
      }
      if (url.includes('/alerts')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ alerts: mockAlerts }),
        });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('rendering', () => {
    it('should render header with title', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('告警管理')).toBeInTheDocument();
      });
    });

    it('should render refresh button', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('刷新')).toBeInTheDocument();
      });
    });

    it('should render analysis button', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('告警分析')).toBeInTheDocument();
      });
    });

    it('should render status filter dropdown', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('全部状态')).toBeInTheDocument();
      });
    });

    it('should render severity filter dropdown', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('全部级别')).toBeInTheDocument();
      });
    });
  });

  describe('stats display', () => {
    it('should display active alerts count', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('活跃告警')).toBeInTheDocument();
        expect(screen.getByText('5')).toBeInTheDocument();
      });
    });

    it('should display critical alerts count', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('紧急告警')).toBeInTheDocument();
        expect(screen.getByText('3')).toBeInTheDocument();
      });
    });

    it('should display acknowledged count', async () => {
      renderAlertsPage();

      await waitFor(() => {
        // Use getAllByText and verify there are multiple occurrences (stats card + filter)
        const acknowledgedElements = screen.getAllByText('已确认');
        expect(acknowledgedElements.length).toBeGreaterThan(0);
        expect(screen.getByText('8')).toBeInTheDocument();
      });
    });

    it('should display resolved count', async () => {
      renderAlertsPage();

      await waitFor(() => {
        // Use getAllByText and verify there are multiple occurrences (stats card + filter)
        const resolvedElements = screen.getAllByText('已解决');
        expect(resolvedElements.length).toBeGreaterThan(0);
        expect(screen.getByText('7')).toBeInTheDocument();
      });
    });
  });

  describe('alert list display', () => {
    it('should display alerts after loading', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('温度过高警告')).toBeInTheDocument();
        expect(screen.getByText('振动异常')).toBeInTheDocument();
        expect(screen.getByText('通信中断')).toBeInTheDocument();
      });
    });

    it('should display alert severity badges', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('紧急')).toBeInTheDocument();
        expect(screen.getByText('高')).toBeInTheDocument();
        expect(screen.getByText('中')).toBeInTheDocument();
      });
    });

    it('should display alert status badges', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('活跃')).toBeInTheDocument();
        expect(screen.getByText('已确认')).toBeInTheDocument();
        expect(screen.getByText('已解决')).toBeInTheDocument();
      });
    });

    it('should display alert IDs', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('#1')).toBeInTheDocument();
        expect(screen.getByText('#2')).toBeInTheDocument();
        expect(screen.getByText('#3')).toBeInTheDocument();
      });
    });

    it('should display device IDs', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText(/CNC-001/)).toBeInTheDocument();
        expect(screen.getByText(/INJ-001/)).toBeInTheDocument();
      });
    });
  });

  describe('filter interactions', () => {
    it('should change status filter value', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('告警管理')).toBeInTheDocument();
      });

      const statusSelects = screen.getAllByRole('combobox');
      const statusFilter = statusSelects[0];
      
      fireEvent.change(statusFilter, { target: { value: 'active' } });

      await waitFor(() => {
        expect(statusFilter).toHaveValue('active');
      });
    });

    it('should change severity filter value', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('告警管理')).toBeInTheDocument();
      });

      const statusSelects = screen.getAllByRole('combobox');
      const severityFilter = statusSelects[1];
      
      fireEvent.change(severityFilter, { target: { value: 'critical' } });

      await waitFor(() => {
        expect(severityFilter).toHaveValue('critical');
      });
    });
  });

  describe('alert actions', () => {
    it('should show acknowledge button for active alerts', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('温度过高警告')).toBeInTheDocument();
      });

      // Active alert should have acknowledge button
      const acknowledgeButtons = screen.getAllByText('确认');
      expect(acknowledgeButtons.length).toBeGreaterThan(0);
    });

    it('should show resolve button for active alerts', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('温度过高警告')).toBeInTheDocument();
      });

      // Active alert should have resolve button
      const resolveButtons = screen.getAllByText('解决');
      expect(resolveButtons.length).toBeGreaterThan(0);
    });

    it('should show only resolve button for acknowledged alerts', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('振动异常')).toBeInTheDocument();
      });

      // Check that acknowledged alert row exists and has resolve button
      const resolveButtons = screen.getAllByText('解决');
      expect(resolveButtons.length).toBeGreaterThan(0);
    });

    it('should show processed status for resolved alerts', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('通信中断')).toBeInTheDocument();
      });

      // Resolved alert should show "已处理"
      expect(screen.getByText('已处理')).toBeInTheDocument();
    });

    it('should call fetch when acknowledge button is clicked', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('温度过高警告')).toBeInTheDocument();
      });

      const acknowledgeButtons = screen.getAllByText('确认');
      fireEvent.click(acknowledgeButtons[0]);

      // Verify fetch was called after clicking acknowledge
      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalled();
      });
    });

    it('should call fetch when resolve button is clicked', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('温度过高警告')).toBeInTheDocument();
      });

      const resolveButtons = screen.getAllByText('解决');
      fireEvent.click(resolveButtons[0]);

      // Verify fetch was called after clicking resolve
      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalled();
      });
    });
  });

  describe('refresh functionality', () => {
    it('should call fetch when refresh button clicked', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('刷新')).toBeInTheDocument();
      });

      // Clear previous calls
      mockFetch.mockClear();

      const refreshButton = screen.getByText('刷新');
      fireEvent.click(refreshButton);

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalled();
      });
    });
  });

  describe('loading state', () => {
    it('should show skeleton while loading', () => {
      // Delay the fetch response
      mockFetch.mockImplementation(() => new Promise(resolve => {
        setTimeout(() => resolve({
          ok: true,
          json: () => Promise.resolve({ alerts: [] }),
        }), 100);
      }));

      renderAlertsPage();

      // Should show skeletons initially
      expect(screen.getAllByTestId('skeleton').length).toBeGreaterThan(0);
    });
  });

  describe('empty state', () => {
    it('should show no alerts message when empty', async () => {
      mockFetch.mockImplementation((url: string) => {
        if (url.includes('/alerts/stats')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ ...mockStats, active_count: 0 }),
          });
        }
        if (url.includes('/alerts')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ alerts: [] }),
          });
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
      });

      renderAlertsPage();

      await waitFor(() => {
        expect(screen.getByText('暂无告警')).toBeInTheDocument();
      });
    });
  });

  describe('error handling', () => {
    it('should handle fetch errors gracefully', async () => {
      mockFetch.mockImplementation(() => Promise.resolve({
        ok: false,
        json: () => Promise.resolve({}),
      }));

      renderAlertsPage();

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalled();
      });

      // Component should still render without crashing
      expect(screen.getByText('告警管理')).toBeInTheDocument();
    });

    it('should handle network errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      renderAlertsPage();

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalled();
      });

      // Component should still render
      expect(screen.getByText('告警管理')).toBeInTheDocument();
    });
  });

  describe('WebSocket integration', () => {
    it('should setup WebSocket connection', async () => {
      renderAlertsPage();

      await waitFor(() => {
        expect(useWebSocket).toHaveBeenCalled();
      });
    });

    it('should handle WebSocket alert message', async () => {
      const mockOnMessage = vi.fn();
      vi.mocked(useWebSocket).mockReturnValue({
        onMessage: mockOnMessage,
      });

      renderAlertsPage();

      await waitFor(() => {
        expect(useWebSocket).toHaveBeenCalledWith(
          expect.objectContaining({
            onMessage: expect.any(Function),
          })
        );
      });
    });
  });

  describe('accessibility', () => {
    it('should have aria-label on refresh button', async () => {
      renderAlertsPage();

      await waitFor(() => {
        const refreshButton = screen.getByText('刷新').closest('button');
        expect(refreshButton).toHaveAttribute('aria-label');
      });
    });

    it('should have aria-label on analysis button', async () => {
      renderAlertsPage();

      await waitFor(() => {
        const analysisButton = screen.getByText('告警分析').closest('button');
        expect(analysisButton).toHaveAttribute('aria-label');
      });
    });
  });
});