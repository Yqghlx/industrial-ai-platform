import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('./Skeleton', () => ({ default: () => <div data-testid="skeleton" /> }));
vi.mock('../lib/api', () => ({
  default: {
    getLatestTelemetry: vi.fn().mockResolvedValue({ data: [] }),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('../hooks/useWebSocket', () => ({
  useWebSocket: vi.fn().mockReturnValue({
    isConnected: false,
    send: vi.fn(),
    reconnect: vi.fn(),
  }),
}));
vi.mock('../lib/colorUtils', () => ({
  getGaugeColor: vi.fn(),
  getGaugeStrokeColor: vi.fn(),
  getGaugePercentage: vi.fn(),
  getTelemetryStatusColor: vi.fn(),
}));
vi.mock('../types/typeGuards', () => ({
  isTelemetry: vi.fn(),
  asTelemetryArraySafe: vi.fn().mockReturnValue([]),
}));
vi.mock('lucide-react', () => ({
  Activity: () => <div />,
  Thermometer: () => <div />,
  Waves: () => <div />,
  Zap: () => <div />,
  Settings: () => <div />,
}));

import DigitalTwinPanel from './DigitalTwinPanel';

describe('DigitalTwinPanel', () => {
  it('renders digital twin panel with content', async () => {
    const { container } = render(<MemoryRouter><DigitalTwinPanel /></MemoryRouter>);
    await waitFor(() => {
      // 验证组件渲染了内容（数字孪生面板）
      expect(container.innerHTML).not.toBe('');
    });
  });
});