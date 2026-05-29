import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Mock all dependencies
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useParams: () => ({ id: 'device-1' }),
  };
});

vi.mock('./AuthContext', () => ({
  useAuth: () => ({ user: { id: 1 }, isAuthenticated: true }),
}));

vi.mock('./Toast', () => ({
  useToast: () => ({ showToast: vi.fn() }),
}));

vi.mock('./Skeleton', () => ({
  default: () => <div data-testid="skeleton" />,
}));

vi.mock('../lib/api', () => ({
  default: {
    getDevice: vi.fn().mockResolvedValue({ name: 'Test Device' }),
    getDeviceTelemetry: vi.fn().mockResolvedValue({ data: [] }),
    getDeviceStats: vi.fn().mockResolvedValue({}),
  },
}));

vi.mock('../i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}));

vi.mock('lucide-react', () => ({
  ArrowLeft: () => <div />,
  Settings: () => <div />,
  Thermometer: () => <div />,
  Waves: () => <div />,
  BarChart3: () => <div />,
}));

vi.mock('recharts', () => ({
  LineChart: () => <div />,
  Line: () => <div />,
  XAxis: () => <div />,
  YAxis: () => <div />,
  Tooltip: () => <div />,
  ResponsiveContainer: () => <div />,
  Legend: () => <div />,
}));

import DeviceDetail from './DeviceDetail';

describe('DeviceDetail', () => {
  beforeEach(() => vi.clearAllMocks());

  it('renders device detail page with telemetry sections', async () => {
    const { container } = render(<MemoryRouter><DeviceDetail /></MemoryRouter>);
    await waitFor(() => {
      // 验证遥测趋势图表标题已渲染
      expect(container.textContent).toContain('telemetry.trend');
      // 验证遥测历史表格已渲染
      expect(container.textContent).toContain('telemetry.history');
      // 验证页面渲染了内容（设备信息或 notFound 提示）
      expect(container.innerHTML).not.toBe('');
    });
  });
});