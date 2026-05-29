import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({
  useAuth: () => ({ user: { id: 1 }, isAuthenticated: true }),
}));

vi.mock('./Toast', () => ({
  useToast: () => ({ showToast: vi.fn() }),
}));

vi.mock('./Skeleton', () => ({
  default: () => <div data-testid="skeleton" />,
}));

vi.mock('./ExportButton', () => ({
  default: () => <div data-testid="export-button" />,
}));

vi.mock('../lib/api', () => ({
  default: {
    getROIStats: vi.fn().mockResolvedValue({
      total_devices: 50,
      active_alerts: 10,
      open_work_orders: 5,
      resolved_issues: 120,
      predicted_savings: 50000,
      uptime_percentage: 98.5,
      avg_response_time_hours: 2.3,
    }),
  },
}));

vi.mock('../i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}));

vi.mock('../types/typeGuards', () => ({
  asROIStatsSafe: () => ({
    total_devices: 50,
    active_alerts: 10,
    open_work_orders: 5,
    resolved_issues: 120,
    predicted_savings: 50000,
    uptime_percentage: 98.5,
    avg_response_time_hours: 2.3,
  }),
}));

vi.mock('lucide-react', () => ({
  TrendingUp: () => <div />,
  DollarSign: () => <div />,
  BarChart3: () => <div />,
  Clock: () => <div />,
  Activity: () => <div />,
  RefreshCw: () => <div />,
}));

import ROIDashboard from './ROIDashboard';

describe('ROIDashboard', () => {
  beforeEach(() => vi.clearAllMocks());

  it('renders ROI dashboard with header and stats', async () => {
    const { container } = render(<MemoryRouter><ROIDashboard /></MemoryRouter>);
    await waitFor(() => {
      // 验证页面标题已渲染
      expect(container.textContent).toContain('nav.roi');
      // 验证 ROI 数据已渲染（来自 mock，格式化后的值）
      expect(container.textContent).toContain('98.5');
      expect(container.textContent).toContain('50,000');
    });
  });
});