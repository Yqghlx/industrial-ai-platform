import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    getSystemStatus: vi.fn().mockResolvedValue({
      database: 'healthy',
      version: '1.0.0',
      uptime: '10 days',
      db_latency_ms: 5,
      user_count: 10,
      device_count: 50,
      timestamp: new Date().toISOString(),
    }),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({
  Database: () => <span>Database</span>,
  Activity: () => <span>Activity</span>,
  Server: () => <span>Server</span>,
  Clock: () => <span>Clock</span>,
  CheckCircle: () => <span>CheckCircle</span>,
  AlertCircle: () => <span>AlertCircle</span>,
}));

import SystemStatus from './SystemStatus';

describe('SystemStatus', () => {
  it('renders system status page with header', async () => {
    const { container } = render(
      <MemoryRouter>
        <SystemStatus />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面标题已渲染（i18n key）
      expect(container.textContent).toContain('nav.system');
      // 验证刷新按钮已渲染
      expect(container.textContent).toContain('common.refresh');
    });
  });

  it('displays system status data after loading', async () => {
    const { container } = render(
      <MemoryRouter>
        <SystemStatus />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证 API 返回的数据已渲染
      expect(container.textContent).toContain('1.0.0');
      expect(container.textContent).toContain('10 days');
      expect(container.textContent).toContain('system.healthy');
    });
  });
});