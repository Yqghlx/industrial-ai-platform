import { describe, it, expect, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    getReports: vi.fn().mockResolvedValue({ data: [] }),
    generateReport: vi.fn().mockResolvedValue(undefined),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({
  FileText: () => <span>FileText</span>,
  Download: () => <span>Download</span>,
  Plus: () => <span>Plus</span>,
}));

import ReportCenter from './ReportCenter';

describe('ReportCenter', () => {
  it('renders report center with header', async () => {
    const { container } = render(
      <MemoryRouter>
        <ReportCenter />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证页面渲染了内容
      expect(container.innerHTML).not.toBe('');
    });
  });
});