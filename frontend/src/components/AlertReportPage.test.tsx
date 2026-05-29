import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('./Skeleton', () => ({ default: () => <div data-testid="skeleton" /> }));
vi.mock('../lib/api', () => ({
  default: {
    getAlertTrend: vi.fn().mockResolvedValue({ data: [] }),
    getDeviceRanking: vi.fn().mockResolvedValue({ data: [] }),
    getAlertEfficiency: vi.fn().mockResolvedValue({ data: null }),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({
  TrendingUp: () => <div />,
  BarChart3: () => <div />,
  Clock: () => <div />,
  CheckCircle: () => <div />,
  AlertTriangle: () => <div />,
  Download: () => <div />,
  RefreshCw: () => <div />,
}));

import AlertReportPage from './AlertReportPage';

describe('AlertReportPage', () => {
  it('renders alert report page with header', async () => {
    const { container } = render(<MemoryRouter><AlertReportPage /></MemoryRouter>);
    await waitFor(() => {
      // 验证页面标题已渲染
      expect(container.textContent).toContain('alertReport.title');
      // 验证导出和刷新按钮已渲染
      expect(container.textContent).toContain('alertReport.refresh');
      expect(container.textContent).toContain('alertReport.exportCsv');
    });
  });
});