import { useState, useEffect, useCallback } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { Database, Activity, Server, Clock, CheckCircle, AlertCircle, Settings, Save, Plus, Trash2, Edit3, Zap } from 'lucide-react';
import { SystemStatus as SystemStatusType, LLMConfigItem, LLMConfigCreateRequest, LLMConfigUpdateRequest } from '../types/api';

export default function SystemStatus() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [status, setStatus] = useState<SystemStatusType | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [llmConfigs, setLLMConfigs] = useState<LLMConfigItem[]>([]);
  const [llmLoading, setLLMLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState({ name: '', api_key: '', base_url: '', model: '' });
  const [saving, setSaving] = useState(false);

  const loadStatus = useCallback(async () => {
    setRefreshing(true);
    try {
      const res = await api.getSystemStatus();
      setStatus(res as SystemStatusType);
    } catch (error) {
      console.error('Failed to load system status:', error);
      showToast({ type: 'error', message: t('errors.loadFailedSystem') });
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [showToast, t]);

  const loadLLMConfigs = useCallback(async () => {
    setLLMLoading(true);
    try {
      const res = await api.listLLMConfigs();
      setLLMConfigs(res.data || []);
    } catch {
      // 非管理员可能无法访问
    } finally {
      setLLMLoading(false);
    }
  }, []);

  useEffect(() => {
    loadStatus();
    loadLLMConfigs();
    const interval = setInterval(loadStatus, 60000);
    return () => clearInterval(interval);
  }, [loadStatus, loadLLMConfigs]);

  const openCreateModal = () => {
    setEditingId(null);
    setForm({ name: '', api_key: '', base_url: 'https://open.bigmodel.cn/api/paas/v4', model: '' });
    setShowModal(true);
  };

  const openEditModal = (item: LLMConfigItem) => {
    setEditingId(item.id);
    setForm({ name: item.name, api_key: '', base_url: item.base_url, model: item.model });
    setShowModal(true);
  };

  const closeModal = () => {
    setShowModal(false);
    setEditingId(null);
  };

  const handleSave = async () => {
    if (!form.name || !form.base_url || !form.model) return;
    setSaving(true);
    try {
      if (editingId) {
        const updateData: LLMConfigUpdateRequest = {
          name: form.name,
          base_url: form.base_url,
          model: form.model,
        };
        if (form.api_key) updateData.api_key = form.api_key;
        await api.updateLLMConfigByID(editingId, updateData);
        showToast({ type: 'success', message: t('system.modelUpdated') });
      } else {
        const createData: LLMConfigCreateRequest = {
          name: form.name,
          api_key: form.api_key,
          base_url: form.base_url,
          model: form.model,
        };
        await api.createLLMConfig(createData);
        showToast({ type: 'success', message: t('system.modelCreated') });
      }
      loadLLMConfigs();
      closeModal();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t('system.configSaveFailed');
      showToast({ type: 'error', message: msg });
    } finally {
      setSaving(false);
    }
  };

  const handleActivate = async (id: number) => {
    try {
      await api.activateLLMConfig(id);
      showToast({ type: 'success', message: t('system.modelActivated') });
      loadLLMConfigs();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t('system.configSaveFailed');
      showToast({ type: 'error', message: msg });
    }
  };

  const handleDelete = async (item: LLMConfigItem) => {
    if (item.is_active) {
      showToast({ type: 'error', message: t('system.cannotDeleteActive') });
      return;
    }
    if (!window.confirm(t('system.confirmDeleteModel'))) return;
    try {
      await api.deleteLLMConfig(item.id);
      showToast({ type: 'success', message: t('system.modelDeleted') });
      loadLLMConfigs();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t('system.configSaveFailed');
      showToast({ type: 'error', message: msg });
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.system')}</h1>
          <p className="text-slate-400">{t('system.title')}</p>
        </div>
        <button
          onClick={loadStatus}
          disabled={refreshing}
          className="btn btn-secondary flex items-center gap-2"
          aria-label={t('common.refresh')}
        >
          <Activity className={`w-5 h-5 ${refreshing ? 'animate-pulse' : ''}`} />
          <span>{t('common.refresh')}</span>
        </button>
      </div>

      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[1, 2, 3, 4].map(i => <Skeleton key={i} variant="card" />)}
        </div>
      ) : status && (
        <>
          {/* Status cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="metric-card">
              <div className="flex items-center gap-3 mb-2">
                <Database className="w-5 h-5 text-primary-500" />
                <span className="metric-label">{t('system.database')}</span>
              </div>
              <div className="flex items-center gap-2">
                {status.database === 'healthy' ? (
                  <CheckCircle className="w-5 h-5 text-green-400" />
                ) : (
                  <AlertCircle className="w-5 h-5 text-red-400" />
                )}
                <span className={`text-xl font-bold ${
                  status.database === 'healthy' ? 'text-green-400' : 'text-red-400'
                }`}>
                  {status.database === 'healthy' ? t('system.healthy') : t('system.unhealthy')}
                </span>
              </div>
              <div className="text-sm text-slate-400 mt-1">
                {t('system.latencyLabel')}: {status.db_latency_ms}ms
              </div>
            </div>

            <div className="metric-card">
              <div className="flex items-center gap-3 mb-2">
                <Server className="w-5 h-5 text-purple-500" />
                <span className="metric-label">{t('system.version')}</span>
              </div>
              <div className="metric-value">{status.version}</div>
            </div>

            <div className="metric-card">
              <div className="flex items-center gap-3 mb-2">
                <Clock className="w-5 h-5 text-blue-500" />
                <span className="metric-label">{t('system.uptime')}</span>
              </div>
              <div className="metric-value">{status.uptime}</div>
            </div>

            <div className="metric-card">
              <div className="flex items-center gap-3 mb-2">
                <Activity className="w-5 h-5 text-green-500" />
                <span className="metric-label">{t('device.deviceCount')}</span>
              </div>
              <div className="metric-value">{status.device_count}</div>
            </div>
          </div>

          {/* 多模型配置列表 */}
          <div className="card">
            <div className="card-header">
              <div className="flex items-center justify-between w-full">
                <div className="flex items-center gap-2">
                  <Settings className="w-5 h-5 text-primary-500" />
                  <h2 className="text-lg font-semibold">{t('system.llmConfig')}</h2>
                </div>
                <button
                  onClick={openCreateModal}
                  className="btn btn-primary btn-sm flex items-center gap-1"
                >
                  <Plus className="w-4 h-4" />
                  <span>{t('system.addModel')}</span>
                </button>
              </div>
            </div>
            <div className="card-body">
              {llmLoading ? (
                <div className="space-y-3">
                  {[1, 2].map(i => <Skeleton key={i} variant="card" />)}
                </div>
              ) : llmConfigs.length === 0 ? (
                <div className="text-center text-slate-500 py-8">{t('system.noModels')}</div>
              ) : (
                <div className="space-y-3">
                  {llmConfigs.map(item => (
                    <div
                      key={item.id}
                      className={`flex items-center justify-between p-4 rounded-lg border ${
                        item.is_active
                          ? 'border-primary-500/50 bg-primary-500/5'
                          : 'border-slate-700 bg-slate-800/50'
                      }`}
                    >
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-medium text-slate-100">{item.name}</span>
                          {item.is_active && (
                            <span className="inline-flex items-center gap-1 px-2 py-0.5 text-xs font-medium rounded-full bg-primary-500/20 text-primary-400 border border-primary-500/30">
                              <Zap className="w-3 h-3" />
                              {t('system.activeModel')}
                            </span>
                          )}
                        </div>
                        <div className="flex items-center gap-4 text-sm text-slate-400">
                          <span>{item.model}</span>
                          <span className="truncate max-w-[200px]">{item.base_url}</span>
                          <span>Key: {item.api_key || '—'}</span>
                        </div>
                      </div>
                      <div className="flex items-center gap-2 ml-4">
                        {!item.is_active && (
                          <button
                            onClick={() => handleActivate(item.id)}
                            className="btn btn-sm bg-green-600/20 text-green-400 border border-green-500/30 hover:bg-green-600/30 flex items-center gap-1"
                            title={t('system.activateModel')}
                          >
                            <Zap className="w-3 h-3" />
                            <span className="hidden sm:inline">{t('system.activateModel')}</span>
                          </button>
                        )}
                        <button
                          onClick={() => openEditModal(item)}
                          className="btn btn-sm btn-secondary flex items-center gap-1"
                          title={t('system.editModel')}
                        >
                          <Edit3 className="w-3 h-3" />
                          <span className="hidden sm:inline">{t('common.edit')}</span>
                        </button>
                        {!item.is_active && (
                          <button
                            onClick={() => handleDelete(item)}
                            className="btn btn-sm bg-red-600/20 text-red-400 border border-red-500/30 hover:bg-red-600/30 flex items-center gap-1"
                            title={t('system.deleteModel')}
                          >
                            <Trash2 className="w-3 h-3" />
                            <span className="hidden sm:inline">{t('common.delete')}</span>
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Details */}
          <div className="card">
            <div className="card-header">
              <h2 className="text-lg font-semibold">{t('system.details')}</h2>
            </div>
            <div className="card-body">
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-slate-400">{t('system.latency')}:</span>
                  <span className="text-slate-200 ml-2">{status.db_latency_ms}ms</span>
                </div>
                <div>
                  <span className="text-slate-400">{t('user.title')}:</span>
                  <span className="text-slate-200 ml-2">{status.user_count}</span>
                </div>
                <div>
                  <span className="text-slate-400">{t('device.deviceCount')}:</span>
                  <span className="text-slate-200 ml-2">{status.device_count}</span>
                </div>
                <div>
                  <span className="text-slate-400">{t('telemetry.timestamp')}:</span>
                  <span className="text-slate-200 ml-2">
                    {new Date(status.timestamp).toLocaleString()}
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Health indicator */}
          <div className="card">
            <div className="card-body">
              <div className="flex items-center justify-center gap-8 py-8">
                <div className="text-center">
                  <div className={`w-16 h-16 rounded-full flex items-center justify-center ${
                    status.database === 'healthy'
                      ? 'bg-green-500/20 border-2 border-green-500'
                      : 'bg-red-500/20 border-2 border-red-500'
                  }`}>
                    <Database className={`w-8 h-8 ${
                      status.database === 'healthy' ? 'text-green-400' : 'text-red-400'
                    }`} />
                  </div>
                  <div className="mt-2 font-medium text-slate-100">{t('system.postgresql')}</div>
                  <div className="text-sm text-slate-400">{status.database}</div>
                </div>

                <div className="text-center">
                  <div className="w-16 h-16 rounded-full bg-primary-500/20 border-2 border-primary-500 flex items-center justify-center">
                    <Server className="w-8 h-8 text-primary-400" />
                  </div>
                  <div className="mt-2 font-medium text-slate-100">{t('system.backend')}</div>
                  <div className="text-sm text-slate-400">{t('system.running')}</div>
                </div>

                <div className="text-center">
                  <div className="w-16 h-16 rounded-full bg-green-500/20 border-2 border-green-500 flex items-center justify-center">
                    <Activity className="w-8 h-8 text-green-400 animate-pulse" />
                  </div>
                  <div className="mt-2 font-medium text-slate-100">{t('system.websocket')}</div>
                  <div className="text-sm text-slate-400">{t('system.connected')}</div>
                </div>
              </div>
            </div>
          </div>
        </>
      )}

      {/* 新增/编辑模型弹窗 */}
      {showModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-slate-800 border border-slate-600 rounded-xl p-6 w-full max-w-md mx-4">
            <h3 className="text-lg font-semibold text-slate-100 mb-4">
              {editingId ? t('system.editModel') : t('system.addModel')}
            </h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1">
                  {t('system.modelNameLabel')}
                </label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm(prev => ({ ...prev, name: e.target.value }))}
                  className="input w-full"
                  placeholder="GLM-4"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1">
                  {t('system.baseURL')}
                </label>
                <input
                  type="text"
                  value={form.base_url}
                  onChange={(e) => setForm(prev => ({ ...prev, base_url: e.target.value }))}
                  className="input w-full"
                  placeholder="https://open.bigmodel.cn/api/paas/v4"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1">
                  {t('system.modelName')}
                </label>
                <input
                  type="text"
                  value={form.model}
                  onChange={(e) => setForm(prev => ({ ...prev, model: e.target.value }))}
                  className="input w-full"
                  placeholder="glm-4-flash"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1">
                  API Key
                </label>
                <input
                  type="password"
                  value={form.api_key}
                  onChange={(e) => setForm(prev => ({ ...prev, api_key: e.target.value }))}
                  className="input w-full"
                  placeholder={editingId ? t('system.apiKeyHint') : t('system.apiKeyPlaceholder')}
                />
                {editingId && (
                  <p className="text-xs text-slate-500 mt-1">{t('system.apiKeyHint')}</p>
                )}
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button onClick={closeModal} className="btn btn-secondary">
                {t('common.cancel')}
              </button>
              <button
                onClick={handleSave}
                disabled={saving || !form.name || !form.base_url || !form.model}
                className="btn btn-primary flex items-center gap-2"
              >
                {saving ? (
                  <Activity className="w-4 h-4 animate-spin" />
                ) : (
                  <Save className="w-4 h-4" />
                )}
                <span>{t('common.save')}</span>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
