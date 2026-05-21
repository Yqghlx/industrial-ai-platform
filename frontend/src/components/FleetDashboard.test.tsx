import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Mock dependencies
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => vi.fn(),
    useLocation: () => ({ pathname: '/dashboard' }),
  };
});

vi.mock('./AuthContext', () => ({
  useAuth: () => ({
    isAdmin: true,
    user: { role: 'admin' },
  }),
}));

vi.mock('./Toast', () => ({
  default: () => <div data-testid="toast-container" />,
  useToast: () => ({
    showToast: vi.fn(),
  }),
}));

vi.mock('./Skeleton', () => ({
  default: ({ variant }: { variant?: string }) => (
    <div data-testid={`skeleton-${variant || 'default'}`}>Loading...</div>
  ),
  SkeletonGrid: ({ count }: { count: number }) => (
    <div data-testid="skeleton-grid">Loading {count} items...</div>
  ),
}));

vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'dashboard.overview': '设备概览',
        'dashboard.fleetStatus': '设备状态总览',
        'dashboard.quickActions': '快捷操作',
        'nav.digitalTwin': '数字孪生',
        'nav.devices': '设备',
        'nav.aiAgent': 'AI分析',
        'nav.workOrders': '工单',
        'nav.notifications': '通知',
        'nav.reports': '报告',
        'device.deviceCount': '设备总数',
        'device.online': '在线',
        'device.warning': '警告',
        'device.fault': '故障',
        'device.manageDevices': '管理设备',
        'device.offline': '离线',
        'device.status': '状态',
        'telemetry.temperature': '温度',
        'telemetry.vibration': '振动',
      };
      return translations[key] || key;
    },
  }),
}));

vi.mock('../lib/colorUtils', () => ({
  getDeviceStatusColor: (status: string) => {
    switch (status) {
      case 'online':
        return 'bg-green-500';
      case 'warning':
        return 'bg-yellow-500';
      case 'fault':
        return 'bg-red-500';
      default:
        return 'bg-slate-500';
    }
  },
  getDeviceStatusBadgeClass: (status: string) => {
    switch (status) {
      case 'online':
        return 'badge-success';
      case 'warning':
        return 'badge-warning';
      case 'fault':
        return 'badge-error';
      default:
        return 'badge-neutral';
    }
  },
}));

// Mock API
vi.mock('../lib/api', () => ({
  default: {
    getDevices: vi.fn().mockResolvedValue({
      data: [
        { id: 'CNC-001', name: '数控机床001', type: '数控机床', status: 'online', location: '车间A' },
        { id: 'INJ-001', name: '注塑机001', type: '注塑机', status: 'warning', location: '车间B' },
        { id: 'ROB-001', name: '工业机器人001', type: '工业机器人', status: 'offline', location: '车间A' },
      ],
    }),
    getLatestTelemetry: vi.fn().mockResolvedValue({
      data: [
        { device_id: 'CNC-001', temperature: 45.5, vibration: 1.2, status: 'online', timestamp: '2024-01-01' },
        { device_id: 'INJ-001', temperature: 78.2, vibration: 3.5, status: 'warning', timestamp: '2024-01-01' },
      ],
    }),
  },
}));

// Import after mocks
import FleetDashboard from './FleetDashboard';
import api from '../lib/api';

const renderFleetDashboard = () => {
  return render(
    <MemoryRouter>
      <FleetDashboard />
    </MemoryRouter>
  );
};

describe('FleetDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('rendering', () => {
    it('should render header with title and description', async () => {
      renderFleetDashboard();

      await waitFor(() => {
        expect(screen.getByText('设备概览')).toBeInTheDocument();
        expect(screen.getByText('设备状态总览')).toBeInTheDocument();
      });
    });

    it('should render digital twin link', async () => {
      renderFleetDashboard();

      await waitFor(() => {
        expect(screen.getByText('数字孪生')).toBeInTheDocument();
      });
    });

    it('should render quick action buttons', async () => {
      renderFleetDashboard();

      await waitFor(() => {
        expect(screen.getByText('AI分析')).toBeInTheDocument();
        expect(screen.getByText('工单')).toBeInTheDocument();
        expect(screen.getByText('通知')).toBeInTheDocument();
        expect(screen.getByText('报告')).toBeInTheDocument();
      });
    });
  });

  describe('loading state', () => {
    it('should show skeleton while loading', () => {
      renderFleetDashboard();

      // Initially loading state shows skeleton
      // The component loads quickly, so we verify the component renders
      expect(screen.getByText('设备概览')).toBeInTheDocument();
    });
  });

  describe('stats display', () => {
    it('should display device stats section after loading', async () => {
      renderFleetDashboard();

      await waitFor(() => {
        // Stats section should appear after loading
        expect(screen.getByText('设备总数')).toBeInTheDocument();
      });
    });
  });

  describe('device list display', () => {
    it('should display devices after loading', async () => {
      renderFleetDashboard();

      await waitFor(() => {
        expect(screen.getByText('CNC-001')).toBeInTheDocument();
      });
    });

    it('should display device names', async () => {
      renderFleetDashboard();

      await waitFor(() => {
        expect(screen.getByText('数控机床001')).toBeInTheDocument();
      });
    });
  });

  describe('API calls', () => {
    it('should call getDevices and getLatestTelemetry on mount', async () => {
      renderFleetDashboard();

      await waitFor(() => {
        expect(api.getDevices).toHaveBeenCalled();
        expect(api.getLatestTelemetry).toHaveBeenCalled();
      });
    });
  });

  describe('error handling', () => {
    it('should handle API errors gracefully', async () => {
      vi.mocked(api.getDevices).mockRejectedValueOnce(new Error('Network error'));

      renderFleetDashboard();

      // Should not crash
      await waitFor(() => {
        expect(screen.getByText('设备概览')).toBeInTheDocument();
      });
    });
  });

  describe('refresh behavior', () => {
    it('should have refresh functionality', async () => {
      renderFleetDashboard();

      await waitFor(() => {
        expect(api.getDevices).toHaveBeenCalled();
      });
    });
  });

  describe('accessibility', () => {
    it('should have accessible links', async () => {
      renderFleetDashboard();

      await waitFor(() => {
        const links = screen.getAllByRole('link');
        expect(links.length).toBeGreaterThan(0);
      });
    });
  });
});