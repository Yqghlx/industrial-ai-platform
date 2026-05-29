import { describe, it, expect, vi } from 'vitest';
import { render, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    exportReport: vi.fn().mockResolvedValue({ data: { url: 'test.pdf' } }),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({
  Download: () => <div />,
  FileText: () => <div />,
  FileSpreadsheet: () => <div />,
  Loader2: () => <div />,
}));

import ExportButton from './ExportButton';

describe('ExportButton', () => {
  it('renders export button with default label', async () => {
    const { container } = render(
      <MemoryRouter>
        <ExportButton reportType="alerts" />
      </MemoryRouter>
    );
    await waitFor(() => {
      // 验证按钮已渲染且包含导出文本
      const button = container.querySelector('button');
      expect(button).toBeInTheDocument();
      expect(button?.textContent).toContain('common.export');
    });
  });

  it('shows button with default label', async () => {
    const { container } = render(
      <MemoryRouter>
        <ExportButton reportType="devices" />
      </MemoryRouter>
    );
    await waitFor(() => {
      expect(container.querySelector('button')).toBeInTheDocument();
    });
  });
});