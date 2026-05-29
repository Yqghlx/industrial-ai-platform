import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({
  Home: () => <span>Home</span>,
  Settings: () => <span>Settings</span>,
  Menu: () => <span>Menu</span>,
  X: () => <span>X</span>,
  LayoutDashboard: () => <span>LayoutDashboard</span>,
  Activity: () => <span>Activity</span>,
  Bell: () => <span>Bell</span>,
  FileText: () => <span>FileText</span>,
  Users: () => <span>Users</span>,
  Bot: () => <span>Bot</span>,
}));

import MobileNavBar from './MobileNavBar';

describe('MobileNavBar', () => {
  it('renders mobile navigation bar with nav items', async () => {
    const { container } = render(
      <MemoryRouter>
        <MobileNavBar />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证导航栏已渲染（包含 nav 元素）
      expect(container.querySelector('nav')).toBeInTheDocument();
      // 验证导航链接已渲染（i18n key）
      expect(container.textContent).toContain('nav.dashboard');
      expect(container.textContent).toContain('nav.devices');
      expect(container.textContent).toContain('nav.digitalTwin');
      expect(container.textContent).toContain('nav.notifications');
      expect(container.textContent).toContain('nav.aiAgent');
    });
  });
});