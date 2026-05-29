import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    agentQuery: vi.fn().mockResolvedValue({ data: {} }),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('../lib/colorUtils', () => ({ getAgentColor: vi.fn() }));
vi.mock('lucide-react', () => ({
  Bot: () => <div />,
  Send: () => <div />,
  User: () => <div />,
  Loader: () => <div />,
  Sparkles: () => <div />,
}));

import AITeamDashboard from './AITeamDashboard';

describe('AITeamDashboard', () => {
  it('renders AI team dashboard with header and input', async () => {
    const { container } = render(<MemoryRouter><AITeamDashboard /></MemoryRouter>);
    await waitFor(() => {
      // 验证页面标题已渲染
      expect(container.textContent).toContain('nav.aiAgent');
      // 验证输入框已渲染
      expect(container.querySelector('input')).toBeInTheDocument();
      // 验证提交按钮已渲染
      expect(container.querySelector('button[type="submit"]')).toBeInTheDocument();
    });
  });
});