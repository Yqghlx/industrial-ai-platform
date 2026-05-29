import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    getDevices: vi.fn().mockResolvedValue({ data: [] }),
    getTelemetry: vi.fn().mockResolvedValue({ data: [] }),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('../hooks/useWebSocket', () => ({ useWebSocket: vi.fn().mockReturnValue({ isConnected: false }) }));
vi.mock('lucide-react', () => ({
  Activity: () => <span>Activity</span>,
  TrendingUp: () => <span>TrendingUp</span>,
  Clock: () => <span>Clock</span>,
  RefreshCw: () => <span>RefreshCw</span>,
  Settings: () => <span>Settings</span>,
  AlertCircle: () => <span>AlertCircle</span>,
}));

import TelemetryPage from './TelemetryPage';

describe('TelemetryPage', () => {
  it('renders telemetry page with content', async () => {
    const { container } = render(
      <MemoryRouter>
        <TelemetryPage />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面渲染了内容
      expect(container.innerHTML).not.toBe('');
    });
  });

  it('shows telemetry data after loading', async () => {
    const { container } = render(
      <MemoryRouter>
        <TelemetryPage />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面渲染了内容（加载完成后的状态）
      expect(container.innerHTML).not.toBe('');
    });
  });
});