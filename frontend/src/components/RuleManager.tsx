import React, { useState, useCallback } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { useEscapeKey } from '../lib/hooks';
import { useAuth } from './AuthContext';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { Bell, Plus, Edit, ToggleLeft, ToggleRight } from 'lucide-react';
import { AlertRule, AlertSeverity, DeviceType, AlertOperator, AlertRuleCreateInput, AlertRuleUpdateInput } from '../types/api';
import { useConfirmDialog } from './UI/ConfirmDialog';
import { useCRUD } from '../hooks/useCRUD';

export default function RuleManager() {
  const { t } = useI18n();
  const { isAdmin } = useAuth();
  const { showToast } = useToast();
  const { showConfirm } = useConfirmDialog();
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [saving, setSaving] = useState(false);

  // C1: Escape 键关闭模态框
  const closeRuleModal = useCallback(() => {
    setShowCreateModal(false);
    setEditingRule(null);
  }, []);

  useEscapeKey(closeRuleModal, !!(showCreateModal || editingRule));

  // FE-P2-09: 使用通用 useCRUD hook 替代重复的 CRUD 逻辑
  // 由于 Rule API 返回格式不同（无分页），需要适配
  const [state, actions] = useCRUD<AlertRule>({
    apiGetAll: async () => {
      const res = await api.getRules();
      return { data: res.data ?? [], total: res.data?.length ?? 0 };
    },
    apiGetOne: (id) => api.getRule(Number(id)),
    apiCreate: (data) => api.createRule(data as AlertRuleCreateInput),
    apiUpdate: (id, data) => api.updateRule(Number(id), data as AlertRuleUpdateInput),
    apiDelete: (id) => api.deleteRule(Number(id)),
    entityName: 'AlertRule',
    initialPageSize: 100,
    onError: (error) => showToast({ type: 'error', message: error }),
    onSuccess: (_action) => {},
  });

  const { items: rules, loading } = state;
  const { refresh, create, update, delete: deleteItem } = actions;

  const handleToggle = useCallback(async (id: number, enabled: boolean) => {
    try {
      await api.toggleRule(id, enabled);
      showToast({ type: 'success', message: enabled ? t('alert.ruleEnabled') : t('alert.ruleDisabled') });
      refresh();
    } catch (error) {
      showToast({ type: 'error', message: t('errors.unknown') });
    }
  }, [refresh, showToast, t]);

  const handleDelete = useCallback(async (id: number) => {
    // FE-P2-11: 使用自定义确认框替代原生 confirm()
    const confirmed = await showConfirm({
      title: t('alert.confirmDeleteTitle'),
      message: t('alert.confirmDelete'),
      variant: 'danger',
      confirmText: t('common.delete'),
      cancelText: t('common.cancel'),
    });
    if (!confirmed) return;
    const success = await deleteItem(String(id));
    if (success) {
      showToast({ type: 'success', message: t('alert.ruleDeleted') });
    } else {
      showToast({ type: 'error', message: t('alert.deleteFailed') });
    }
  }, [showConfirm, deleteItem, showToast, t]);

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'bg-red-500/20 text-red-400';
      case 'high': return 'bg-orange-500/20 text-orange-400';
      case 'medium': return 'bg-yellow-500/20 text-yellow-400';
      case 'low': return 'bg-green-500/20 text-green-400';
      default: return 'bg-slate-500/20 text-slate-400';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.rules')}</h1>
          <p className="text-slate-400">{t('alert.title')}</p>
        </div>
        <button 
          onClick={() => setShowCreateModal(true)}
          className="btn btn-primary flex items-center gap-2"
          aria-label={t('alert.createRule')}
        >
          <Plus className="w-5 h-5" />
          <span>{t('alert.createRule')}</span>
        </button>
      </div>

      {/* Rules table */}
      <div className="card">
        <div className="card-body">
          {loading ? (
            <div className="space-y-4">
              {[1, 2, 3, 4].map(i => <Skeleton key={i} variant="card" />)}
            </div>
          ) : (
            <div className="table-container">
              <table className="table">
                <thead>
                  <tr>
                    <th>{t('alert.ruleName')}</th>
                    <th>{t('alert.metric')}</th>
                    <th>{t('alert.threshold')}</th>
                    <th>{t('alert.severity')}</th>
                    <th>{t('alert.enabled')}</th>
                    <th>{t('common.edit')}</th>
                  </tr>
                </thead>
                <tbody>
                  {rules.map((rule) => (
                    <tr key={rule.id}>
                      <td>
                        <div className="flex items-center gap-2">
                          <Bell className="w-4 h-4 text-slate-400" />
                          <span className="font-medium">{rule.name}</span>
                        </div>
                      </td>
                      <td>
                        <span className="text-slate-300">{rule.metric}</span>
                        <span className="text-slate-400 ml-1">{rule.operator}</span>
                      </td>
                      <td>
                        <span className="text-primary-400 font-medium">{rule.threshold}</span>
                      </td>
                      <td>
                        <span className={`status-badge ${getSeverityColor(rule.severity)}`}>
                          {rule.severity}
                        </span>
                      </td>
                      <td>
                        <button
                          onClick={() => handleToggle(rule.id, !rule.enabled)}
                          className="p-1 hover:bg-slate-700 rounded min-w-[44px] min-h-[44px] flex items-center justify-center"
                          aria-label={rule.enabled ? t('common.disable') : t('common.enable')}
                        >
                          {rule.enabled ? (
                            <ToggleRight className="w-6 h-6 text-green-400" />
                          ) : (
                            <ToggleLeft className="w-6 h-6 text-slate-400" />
                          )}
                        </button>
                      </td>
                      <td>
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => setEditingRule(rule)}
                            className="p-1 text-slate-400 hover:text-primary-400 min-w-[44px] min-h-[44px] flex items-center justify-center"
                            aria-label={t('common.edit')}
                          >
                            <Edit className="w-4 h-4" />
                          </button>
                          {isAdmin && (
                            <button
                              onClick={() => handleDelete(rule.id)}
                              className="p-1 text-slate-400 hover:text-red-400 min-w-[44px] min-h-[44px] flex items-center justify-center"
                              aria-label={t('common.delete')}
                            >
                              ×
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Create/Edit Modal */}
      {(showCreateModal || editingRule) && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" role="dialog" aria-modal="true" onClick={(e) => { if (e.target === e.currentTarget) closeRuleModal(); }}>
          <div className="card max-w-md">
            <div className="card-header">
              <h2 className="text-lg font-semibold">
                {editingRule ? t('alert.editRule') : t('alert.createRule')}
              </h2>
            </div>
            <div className="card-body">
              <form onSubmit={async (e) => {
                e.preventDefault();
                setSaving(true);
                const formData = new FormData(e.target as HTMLFormElement);
                const data = {
                  name: formData.get('name') as string,
                  device_type: formData.get('device_type') as DeviceType,
                  metric: formData.get('metric') as string,
                  operator: formData.get('operator') as AlertOperator,
                  threshold: parseFloat(formData.get('threshold') as string),
                  severity: formData.get('severity') as AlertSeverity,
                  enabled: formData.get('enabled') === 'true',
                  cooldown_sec: parseInt(formData.get('cooldown_sec') as string) || 300,
                  actions: '[{"type": "notification"}]',
                };

                // FE-P2-09: 使用 useCRUD hook 的 create/update 方法
                let success = false;
                if (editingRule) {
                  success = await update(String(editingRule.id), data) !== null;
                } else {
                  success = await create(data) !== null;
                }
                if (success) {
                  showToast({ type: 'success', message: t('alert.ruleSaved') });
                } else {
                  showToast({ type: 'error', message: t('alert.saveFailed') });
                }
                setSaving(false);
                setShowCreateModal(false);
                setEditingRule(null);
              }}>
                <div className="space-y-4">
                  <div>
                    <label className="label" htmlFor="rule-name">{t('alert.ruleName')}</label>
                    <input id="rule-name" name="name" className="input" required defaultValue={editingRule?.name} />
                  </div>
                  <div>
                    <label className="label" htmlFor="rule-device-type">{t('alert.deviceType')}</label>
                    <select id="rule-device-type" name="device_type" className="input" defaultValue={editingRule?.device_type || 'other'}>
                      <option value="pump">{t('device.pump')}</option>
                      <option value="motor">{t('device.motor')}</option>
                      <option value="compressor">{t('device.compressor')}</option>
                      <option value="conveyor">{t('device.conveyor')}</option>
                      <option value="valve">{t('device.valve')}</option>
                      <option value="sensor">{t('device.sensor')}</option>
                      <option value="other">{t('device.other')}</option>
                    </select>
                  </div>
                  <div>
                    <label className="label" htmlFor="rule-metric">{t('alert.metric')}</label>
                    <select id="rule-metric" name="metric" className="input" defaultValue={editingRule?.metric}>
                      <option value="temperature">{t('telemetry.temperature')}</option>
                      <option value="vibration">{t('telemetry.vibration')}</option>
                      <option value="pressure">{t('telemetry.pressure')}</option>
                      <option value="power">{t('telemetry.power')}</option>
                    </select>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="label" htmlFor="rule-operator">{t('alert.operator')}</label>
                      <select id="rule-operator" name="operator" className="input" defaultValue={editingRule?.operator}>
                        <option value=">">&gt;</option>
                        <option value=">=">&gt;=</option>
                        <option value="<">&lt;</option>
                        <option value="<=">&lt;=</option>
                      </select>
                    </div>
                    <div>
                      <label className="label" htmlFor="rule-threshold">{t('alert.threshold')}</label>
                      <input id="rule-threshold" name="threshold" type="number" className="input" required min="0" step="0.1" defaultValue={editingRule?.threshold} />
                    </div>
                  </div>
                  <div>
                    <label className="label" htmlFor="rule-severity">{t('alert.severity')}</label>
                    <select id="rule-severity" name="severity" className="input" defaultValue={editingRule?.severity}>
                      <option value="critical">{t('alert.critical')}</option>
                      <option value="high">{t('alert.high')}</option>
                      <option value="medium">{t('alert.medium')}</option>
                      <option value="low">{t('alert.low')}</option>
                    </select>
                  </div>
                  <div>
                    <label className="label" htmlFor="rule-cooldown">{t('alert.cooldown')} ({t('alert.cooldownUnit')})</label>
                    <input id="rule-cooldown" name="cooldown_sec" type="number" className="input" required min="0" step="1" defaultValue={editingRule?.cooldown_sec || 300} />
                  </div>
                  <div className="flex gap-2">
                    <button type="submit" className="btn btn-primary flex-1" disabled={saving}>
                      {t('common.save')}
                    </button>
                    <button
                      type="button"
                      onClick={() => { setShowCreateModal(false); setEditingRule(null); }}
                      className="btn btn-secondary flex-1"
                    >
                      {t('common.cancel')}
                    </button>
                  </div>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
      
    </div>
  );
}