import React, { useState, useEffect } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { useAuth } from './AuthContext';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { Bell, Plus, Edit, ToggleLeft, ToggleRight } from 'lucide-react';
import { AlertRule, AlertSeverity, DeviceType, AlertOperator } from '../types/api';

export default function RuleManager() {
  const { t } = useI18n();
  const { isAdmin } = useAuth();
  const { showToast } = useToast();
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    loadRules();
  }, []);

  const loadRules = async () => {
    setLoading(true);
    try {
      const res = await api.getRules();
      setRules(res.data as AlertRule[]);
    } catch (error) {
      console.error('Failed to load rules:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleToggle = async (id: number, enabled: boolean) => {
    try {
      await api.toggleRule(id, enabled);
      showToast({ type: 'success', message: enabled ? '规则已启用' : '规则已禁用' });
      loadRules();
    } catch (error) {
      showToast({ type: 'error', message: '操作失败' });
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('确认删除此规则？')) return;
    try {
      await api.deleteRule(id);
      showToast({ type: 'success', message: '规则已删除' });
      loadRules();
    } catch (error) {
      showToast({ type: 'error', message: '删除失败' });
    }
  };

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
                        <span className="text-primary-400 ml-1">{rule.threshold}</span>
                      </td>
                      <td>
                        <span className={`status-badge ${getSeverityColor(rule.severity)}`}>
                          {rule.severity}
                        </span>
                      </td>
                      <td>
                        <button
                          onClick={() => handleToggle(rule.id, !rule.enabled)}
                          className="p-1 hover:bg-slate-700 rounded"
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
                            className="p-1 text-slate-400 hover:text-primary-400"
                          >
                            <Edit className="w-4 h-4" />
                          </button>
                          {isAdmin && (
                            <button
                              onClick={() => handleDelete(rule.id)}
                              className="p-1 text-slate-400 hover:text-red-400"
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
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" role="dialog" aria-modal="true">
          <div className="card max-w-md">
            <div className="card-header">
              <h2 className="text-lg font-semibold">
                {editingRule ? t('alert.editRule') : t('alert.createRule')}
              </h2>
            </div>
            <div className="card-body">
              <form onSubmit={async (e) => {
                e.preventDefault();
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

                try {
                  if (editingRule) {
                    await api.updateRule(editingRule.id, data);
                  } else {
                    await api.createRule(data);
                  }
                  showToast({ type: 'success', message: '规则已保存' });
                  setShowCreateModal(false);
                  setEditingRule(null);
                  loadRules();
                } catch (error) {
                  showToast({ type: 'error', message: '保存失败' });
                }
              }}>
                <div className="space-y-4">
                  <div>
                    <label className="label">{t('alert.ruleName')}</label>
                    <input name="name" className="input" required defaultValue={editingRule?.name} />
                  </div>
                  <div>
                    <label className="label">{t('alert.deviceType')}</label>
                    <select name="device_type" className="input" defaultValue={editingRule?.device_type || 'other'}>
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
                    <label className="label">{t('alert.metric')}</label>
                    <select name="metric" className="input" defaultValue={editingRule?.metric}>
                      <option value="temperature">{t('telemetry.temperature')}</option>
                      <option value="vibration">{t('telemetry.vibration')}</option>
                      <option value="pressure">{t('telemetry.pressure')}</option>
                      <option value="power">{t('telemetry.power')}</option>
                    </select>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="label">{t('alert.operator')}</label>
                      <select name="operator" className="input" defaultValue={editingRule?.operator}>
                        <option value=">">&gt;</option>
                        <option value=">=">&gt;=</option>
                        <option value="<">&lt;</option>
                        <option value="<=">&lt;=</option>
                      </select>
                    </div>
                    <div>
                      <label className="label">{t('alert.threshold')}</label>
                      <input name="threshold" type="number" className="input" defaultValue={editingRule?.threshold} />
                    </div>
                  </div>
                  <div>
                    <label className="label">{t('alert.severity')}</label>
                    <select name="severity" className="input" defaultValue={editingRule?.severity}>
                      <option value="critical">{t('alert.critical')}</option>
                      <option value="high">{t('alert.high')}</option>
                      <option value="medium">{t('alert.medium')}</option>
                      <option value="low">{t('alert.low')}</option>
                    </select>
                  </div>
                  <div>
                    <label className="label">{t('alert.cooldown')} (秒)</label>
                    <input name="cooldown_sec" type="number" className="input" defaultValue={editingRule?.cooldown_sec || 300} />
                  </div>
                  <div className="flex gap-2">
                    <button type="submit" className="btn btn-primary flex-1">
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