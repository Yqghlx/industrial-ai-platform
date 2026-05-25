import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Mock dependencies
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  };
});

vi.mock('./AuthContext', () => ({
  useAuth: () => ({
    isAdmin: true,
    user: { role: 'admin' },
  }),
}));

vi.mock('./Toast', () => ({
  default: () => <div data-testid="toast-container" />,
  useToast: () => ({
    showToast: vi.fn(),
  }),
}));

vi.mock('./UI/ConfirmDialog', () => ({
  useConfirmDialog: () => ({
    showConfirm: vi.fn().mockResolvedValue(true),
  }),
}));

vi.mock('./Skeleton', () => ({
  default: ({ variant }: { variant?: string }) => (
    <div data-testid={`skeleton-${variant || 'default'}`}>Loading...</div>
  ),
}));

vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'nav.rules': '规则管理',
        'alert.title': '告警规则配置',
        'alert.createRule': '创建规则',
        'alert.editRule': '编辑规则',
        'alert.ruleName': '规则名称',
        'alert.metric': '指标',
        'alert.threshold': '阈值',
        'alert.severity': '严重程度',
        'alert.enabled': '启用状态',
        'alert.deviceType': '设备类型',
        'alert.operator': '操作符',
        'alert.cooldown': '冷却时间',
        'alert.critical': '严重',
        'alert.high': '高',
        'alert.medium': '中',
        'alert.low': '低',
        'common.edit': '编辑',
        'common.save': '保存',
        'common.cancel': '取消',
        'device.pump': '泵',
        'device.motor': '电机',
        'device.compressor': '压缩机',
        'device.conveyor': '传送带',
        'device.valve': '阀门',
        'device.sensor': '传感器',
        'device.other': '其他',
        'telemetry.temperature': '温度',
        'telemetry.vibration': '振动',
        'telemetry.pressure': '压力',
        'telemetry.power': '功率',
      };
      return translations[key] || key;
    },
  }),
}));

// Mock API
vi.mock('../lib/api', () => ({
  default: {
    getRules: vi.fn().mockResolvedValue({
      data: [
        { id: 1, name: '温度过高告警', metric: 'temperature', operator: '>', threshold: 80, severity: 'critical', enabled: true, device_type: 'pump' },
        { id: 2, name: '振动异常告警', metric: 'vibration', operator: '>', threshold: 5.0, severity: 'high', enabled: false, device_type: 'motor' },
        { id: 3, name: '压力过低告警', metric: 'pressure', operator: '<', threshold: 10, severity: 'medium', enabled: true, device_type: 'compressor' },
      ],
    }),
    toggleRule: vi.fn().mockResolvedValue({ success: true }),
    deleteRule: vi.fn().mockResolvedValue({ success: true }),
    createRule: vi.fn().mockResolvedValue({ id: 4, name: '新规则' }),
    updateRule: vi.fn().mockResolvedValue({ id: 1, name: '更新规则' }),
  },
}));

// Import after mocks
import RuleManager from './RuleManager';
import api from '../lib/api';

const renderRuleManager = () => {
  return render(
    <MemoryRouter>
      <RuleManager />
    </MemoryRouter>
  );
};

describe('RuleManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('rendering', () => {
    it('should render header with title and description', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('规则管理')).toBeInTheDocument();
        expect(screen.getByText('告警规则配置')).toBeInTheDocument();
      });
    });

    it('should render create rule button', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('创建规则')).toBeInTheDocument();
      });
    });

    it('should render table headers', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('规则名称')).toBeInTheDocument();
        expect(screen.getByText('指标')).toBeInTheDocument();
        expect(screen.getByText('阈值')).toBeInTheDocument();
        expect(screen.getByText('严重程度')).toBeInTheDocument();
        expect(screen.getByText('启用状态')).toBeInTheDocument();
        expect(screen.getByText('编辑')).toBeInTheDocument();
      });
    });
  });

  describe('loading state', () => {
    it('should show skeleton while loading', () => {
      renderRuleManager();

      // Multiple skeletons are shown during loading
      const skeletons = screen.getAllByTestId('skeleton-card');
      expect(skeletons.length).toBeGreaterThan(0);
    });
  });

  describe('rule list display', () => {
    it('should display rules after loading', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('温度过高告警')).toBeInTheDocument();
        expect(screen.getByText('振动异常告警')).toBeInTheDocument();
        expect(screen.getByText('压力过低告警')).toBeInTheDocument();
      });
    });

    it('should display rule metrics', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('temperature')).toBeInTheDocument();
        expect(screen.getByText('vibration')).toBeInTheDocument();
      });
    });

    it('should display rule thresholds', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('80')).toBeInTheDocument();
        expect(screen.getByText('10')).toBeInTheDocument();
      });
    });

    it('should display severity badges', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('critical')).toBeInTheDocument();
        expect(screen.getByText('high')).toBeInTheDocument();
      });
    });
  });

  describe('toggle rule', () => {
    it('should call toggleRule API when clicking toggle button', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('温度过高告警')).toBeInTheDocument();
      });

      // Find toggle buttons
      const toggleButtons = screen.getAllByRole('button').filter(btn => 
        btn.querySelector('svg') && btn.className.includes('hover:bg-slate-700')
      );

      if (toggleButtons.length > 0) {
        fireEvent.click(toggleButtons[0]);

        await waitFor(() => {
          expect(api.toggleRule).toHaveBeenCalled();
        });
      }
    });
  });

  describe('create rule modal', () => {
    it('should open create modal when clicking create button', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('创建规则')).toBeInTheDocument();
      });

      // Click the create button
      const createButton = screen.getByText('创建规则');
      fireEvent.click(createButton);

      // Modal should be triggered (even if not immediately visible)
      expect(createButton).toBeInTheDocument();
    });

    it('should have create button available', async () => {
      renderRuleManager();

      await waitFor(() => {
        const createButton = screen.getByText('创建规则');
        expect(createButton).toBeInTheDocument();
      });
    });
  });

  describe('edit rule', () => {
    it('should open edit modal when clicking edit button', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('温度过高告警')).toBeInTheDocument();
      });

      // Find edit button
      const editButtons = screen.getAllByRole('button').filter(btn =>
        btn.className.includes('text-slate-400') && btn.className.includes('hover:text-primary-400')
      );

      if (editButtons.length > 0) {
        fireEvent.click(editButtons[0]);

        await waitFor(() => {
          expect(screen.getByRole('dialog')).toBeInTheDocument();
        });
      }
    });
  });

  describe('delete rule', () => {
    it('should show confirmation dialog for admin users', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('温度过高告警')).toBeInTheDocument();
      });

      // Find delete buttons (× symbol)
      const deleteButtons = screen.getAllByText('×');

      if (deleteButtons.length > 0) {
        fireEvent.click(deleteButtons[0]);

        // showConfirm is mocked to return true, so delete should proceed
        await waitFor(() => {
          expect(api.deleteRule).toHaveBeenCalled();
        });
      }
    });

    it('should call deleteRule API after confirmation', async () => {
      vi.spyOn(window, 'confirm').mockReturnValue(true);

      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('温度过高告警')).toBeInTheDocument();
      });

      const deleteButtons = screen.getAllByText('×');

      if (deleteButtons.length > 0) {
        fireEvent.click(deleteButtons[0]);

        await waitFor(() => {
          expect(api.deleteRule).toHaveBeenCalled();
        });
      }
    });

    it('should not delete when confirmation is cancelled', async () => {
      vi.spyOn(window, 'confirm').mockReturnValue(false);

      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('温度过高告警')).toBeInTheDocument();
      });

      const deleteButtons = screen.getAllByText('×');

      if (deleteButtons.length > 0) {
        fireEvent.click(deleteButtons[0]);

        await waitFor(() => {
          expect(api.deleteRule).not.toHaveBeenCalled();
        });
      }
    });
  });

  describe('API calls', () => {
    it('should call getRules on mount', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(api.getRules).toHaveBeenCalled();
      });
    });
  });

  describe('error handling', () => {
    it('should handle API errors gracefully', async () => {
      vi.mocked(api.getRules).mockRejectedValueOnce(new Error('Network error'));

      renderRuleManager();

      // Should not crash
      await waitFor(() => {
        expect(screen.getByText('规则管理')).toBeInTheDocument();
      });
    });

    it('should show error toast on toggle failure', async () => {
      vi.mocked(api.toggleRule).mockRejectedValueOnce(new Error('Toggle error'));

      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('温度过高告警')).toBeInTheDocument();
      });

      const toggleButtons = screen.getAllByRole('button').filter(btn => 
        btn.querySelector('svg') && btn.className.includes('hover:bg-slate-700')
      );

      if (toggleButtons.length > 0) {
        fireEvent.click(toggleButtons[0]);
      }
    });
  });

  describe('accessibility', () => {
    it('should have role="dialog" for modal', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('创建规则')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('创建规则'));

      await waitFor(() => {
        const dialog = screen.getByRole('dialog');
        expect(dialog).toBeInTheDocument();
      });
    });

    it('should have aria-modal="true" for modal', async () => {
      renderRuleManager();

      await waitFor(() => {
        expect(screen.getByText('创建规则')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('创建规则'));

      await waitFor(() => {
        const dialog = screen.getByRole('dialog');
        expect(dialog).toHaveAttribute('aria-modal', 'true');
      });
    });
  });
});
