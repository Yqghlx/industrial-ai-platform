import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Mock dependencies
vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    getDevices: vi.fn().mockResolvedValue({ data: [] }),
    getLatestTelemetry: vi.fn().mockResolvedValue({ data: [] }),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('../hooks/useWebSocket', () => ({ useWebSocket: vi.fn().mockReturnValue({ isConnected: false }) }));
vi.mock('../lib/colorUtils', () => ({
  getDeviceStatusColor: vi.fn().mockReturnValue('bg-green-500'),
  getDeviceStatusBadgeClass: vi.fn().mockReturnValue('bg-green-500/20 text-green-400'),
}));
vi.mock('lucide-react', () => ({
  Activity: () => <span>Activity</span>,
  AlertTriangle: () => <span>AlertTriangle</span>,
  Wrench: () => <span>Wrench</span>,
  TrendingUp: () => <span>TrendingUp</span>,
  Settings: () => <span>Settings</span>,
  Bell: () => <span>Bell</span>,
}));

import FleetDashboard from './FleetDashboard';

describe('FleetDashboard', () => {
  it('renders without crashing', async () => {
    const { container } = render(
      <MemoryRouter>
        <FleetDashboard />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面渲染了仪表盘内容（包含 dashboard.overview 标题）
      expect(container.textContent).toContain('dashboard.overview');
    });
  });

  it('shows stats cards after loading', async () => {
    const { container } = render(
      <MemoryRouter>
        <FleetDashboard />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证统计卡片渲染（设备数量、在线数等 i18n key）
      expect(container.textContent).toContain('device.deviceCount');
      expect(container.textContent).toContain('device.online');
      expect(container.textContent).toContain('device.warning');
      expect(container.textContent).toContain('device.fault');
    });
  });
});