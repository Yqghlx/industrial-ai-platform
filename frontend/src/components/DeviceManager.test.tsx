import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('./AuthContext', () => ({ useAuth: () => ({ user: { id: 1, role: 'admin' } }) }));
vi.mock('./Toast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('./Skeleton', () => ({ SkeletonTable: () => <div data-testid="skeleton-table" /> }));
vi.mock('./ExportButton', () => ({ __esModule: true, default: () => <div data-testid="export-button" /> }));
vi.mock('./UI/ConfirmDialog', () => ({ useConfirmDialog: () => ({ showConfirm: vi.fn() }) }));
vi.mock('../lib/api', () => ({
  default: {
    getDevices: vi.fn().mockResolvedValue({ data: [], total: 0 }),
    getDevice: vi.fn().mockResolvedValue({}),
    createDevice: vi.fn().mockResolvedValue({}),
    updateDevice: vi.fn().mockResolvedValue({}),
    deleteDevice: vi.fn().mockResolvedValue({}),
  },
}));
vi.mock('../i18n', () => ({ useI18n: () => ({ t: (k: string) => k }) }));
vi.mock('../lib/colorUtils', () => ({ getDeviceStatusBadgeClass: vi.fn() }));
vi.mock('../hooks/useCRUD', () => ({
  useCRUD: () => [
    {
      items: [],
      loading: false,
      error: null,
      total: 0,
      page: 1,
      pageSize: 20,
    },
    {
      fetchAll: vi.fn(),
      fetchOne: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      refresh: vi.fn(),
      setPage: vi.fn(),
      setPageSize: vi.fn(),
      resetError: vi.fn(),
    },
  ],
}));
vi.mock('lucide-react', () => ({
  Plus: () => <div />,
  Edit: () => <div />,
  Trash2: () => <div />,
  Search: () => <div />,
  ChevronLeft: () => <div />,
  ChevronRight: () => <div />,
}));

import DeviceManager from './DeviceManager';

describe('DeviceManager', () => {
  beforeEach(() => vi.clearAllMocks());

  it('renders device management page with header and table', async () => {
    const { container } = render(<MemoryRouter><DeviceManager /></MemoryRouter>);
    await waitFor(() => {
      // 验证页面标题已渲染（i18n key：nav.devices）
      expect(container.textContent).toContain('nav.devices');
      // 验证搜索框和类型筛选器已渲染
      expect(container.querySelector('input[type="text"]')).toBeInTheDocument();
      // 验证设备表格容器已渲染（loading=false 时）
      expect(container.querySelector('[data-testid="device-table"]')).toBeInTheDocument();
    });
  });
});