import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// 使用 vi.hoisted 确保 mock 函数在 vi.mock 工厂中可用
const { mockShowToast, mockExportReport, mockDownloadBlob } = vi.hoisted(() => ({
  mockShowToast: vi.fn(),
  mockExportReport: vi.fn(),
  mockDownloadBlob: vi.fn(),
}));

vi.mock('./Toast', () => ({ useToast: () => ({ showToast: mockShowToast }) }));
vi.mock('../lib/api', () => ({
  default: { exportReport: mockExportReport },
}));
vi.mock('../lib/fileDownload', () => ({
  downloadBlob: mockDownloadBlob,
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('lucide-react', () => ({
  Download: () => <span>Download</span>,
  FileText: () => <span>FileText</span>,
  FileSpreadsheet: () => <span>FileSpreadsheet</span>,
  Loader2: () => <span>Loader2</span>,
}));

import ExportButton from './ExportButton';

function renderExportButton(reportType: 'devices' | 'alerts' | 'roi' = 'devices') {
  return render(
    <MemoryRouter>
      <ExportButton reportType={reportType} />
    </MemoryRouter>
  );
}

describe('ExportButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExportReport.mockResolvedValue({
      data: new Uint8Array([1, 2, 3]),
      filename: 'report.pdf',
      mimeType: 'application/pdf',
    });
  });

  it('渲染导出按钮', () => {
    renderExportButton();
    expect(screen.getByLabelText('common.export')).toBeTruthy();
  });

  it('点击导出按钮打开格式选择模态框', async () => {
    renderExportButton();
    fireEvent.click(screen.getByLabelText('common.export'));

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeTruthy();
      expect(screen.getByText('export.selectExportFormat')).toBeTruthy();
    });
  });

  it('模态框包含 PDF 和 Excel 选项', async () => {
    renderExportButton();
    fireEvent.click(screen.getByLabelText('common.export'));

    await waitFor(() => {
      expect(screen.getByText('export.pdfFormat')).toBeTruthy();
      expect(screen.getByText('export.xlsxFormat')).toBeTruthy();
    });
  });

  it('点击 PDF 按钮触发导出', async () => {
    renderExportButton();
    fireEvent.click(screen.getByLabelText('common.export'));

    await waitFor(() => screen.getByRole('dialog'));

    const pdfButtons = screen.getAllByText('export.pdfFormat');
    fireEvent.click(pdfButtons[pdfButtons.length - 1]);

    await waitFor(() => {
      expect(mockExportReport).toHaveBeenCalledWith('devices', 'pdf', undefined, undefined);
    });
  });

  it('导出成功时下载文件并显示成功提示', async () => {
    renderExportButton();
    fireEvent.click(screen.getByLabelText('common.export'));

    await waitFor(() => screen.getByRole('dialog'));

    const pdfButtons = screen.getAllByText('export.pdfFormat');
    fireEvent.click(pdfButtons[pdfButtons.length - 1]);

    await waitFor(() => {
      expect(mockDownloadBlob).toHaveBeenCalledWith(
        expect.any(Uint8Array),
        'report.pdf',
        'application/pdf'
      );
      expect(mockShowToast).toHaveBeenCalledWith({ type: 'success', message: 'export.exportSuccess' });
    });
  });

  it('导出失败时显示错误提示', async () => {
    mockExportReport.mockRejectedValue(new Error('Network error'));
    renderExportButton();
    fireEvent.click(screen.getByLabelText('common.export'));

    await waitFor(() => screen.getByRole('dialog'));

    const pdfButtons = screen.getAllByText('export.pdfFormat');
    fireEvent.click(pdfButtons[pdfButtons.length - 1]);

    await waitFor(() => {
      expect(mockShowToast).toHaveBeenCalledWith({ type: 'error', message: 'export.exportFailed' });
    });
  });

  it('点击取消按钮关闭模态框', async () => {
    renderExportButton();
    fireEvent.click(screen.getByLabelText('common.export'));

    await waitFor(() => screen.getByRole('dialog'));

    const cancelButton = screen.getByLabelText('common.cancel');
    fireEvent.click(cancelButton);

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull();
    });
  });

  it('根据 reportType 显示正确的报告标签', async () => {
    renderExportButton('alerts');
    fireEvent.click(screen.getByLabelText('common.export'));

    await waitFor(() => {
      expect(screen.getByText('export.alertReport')).toBeTruthy();
    });
  });
});
