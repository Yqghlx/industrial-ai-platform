import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

// Mock dependencies
vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('./Toast', () => ({
  useToast: () => ({
    showToast: vi.fn(),
  }),
}));

vi.mock('../lib/api', () => ({
  default: {
    exportReport: vi.fn().mockResolvedValue({
      data: new Blob(['test'], { type: 'application/pdf' }),
      mimeType: 'application/pdf',
      filename: 'test-report.pdf',
    }),
  },
}));

vi.mock('lucide-react', () => ({
  Download: () => <div data-testid="download-icon" />,
  FileText: () => <div data-testid="file-text-icon" />,
  FileSpreadsheet: () => <div data-testid="file-spreadsheet-icon" />,
  Loader2: () => <div data-testid="loader-icon" />,
}));

// Import after mocks
import ExportButton from './ExportButton';

describe('ExportButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders export button', () => {
    render(<ExportButton reportType="devices" />);

    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('shows download icon', () => {
    render(<ExportButton reportType="devices" />);

    expect(screen.getByTestId('download-icon')).toBeInTheDocument();
  });

  it('accepts custom className', () => {
    render(<ExportButton reportType="devices" className="custom-class" />);

    const btn = screen.getByRole('button');
    expect(btn.className).toContain('custom-class');
  });

  it('renders with custom label', () => {
    render(<ExportButton reportType="devices" label="Export CSV" />);

    expect(screen.getByText('Export CSV')).toBeInTheDocument();
  });
});