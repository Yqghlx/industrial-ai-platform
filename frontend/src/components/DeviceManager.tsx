import React, { useState, useEffect, useCallback, useMemo } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { useAuth } from './AuthContext';
import { SkeletonTable } from './Skeleton';
import { useToast } from './Toast';
import ExportButton from './ExportButton';
import { Plus, Edit, Trash2, Search, ChevronLeft, ChevronRight } from 'lucide-react';
import { Device, DeviceStatus } from '../types/api';
import { getDeviceStatusBadgeClass } from '../lib/colorUtils';
import { useConfirmDialog } from './UI/ConfirmDialog';

const PAGE_SIZE = 20;

export default function DeviceManager() {
  const { t } = useI18n();
  const { isAdmin } = useAuth();
  const { showToast } = useToast();
  const { showConfirm } = useConfirmDialog();
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [editingDevice, setEditingDevice] = useState<Device | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);

  // FE-P1-01: 使用 useMemo 包裹设备类型数组，避免每次渲染创建新对象
  const DEVICE_TYPES = useMemo(() => [
    { value: '', label: t('device.allTypes') },
    { value: 'CNC', label: t('device.cnc') },
    { value: 'InjectionMolder', label: t('device.injectionMolder') },
    { value: 'AssemblyRobot', label: t('device.assemblyRobot') },
    { value: 'Conveyor', label: t('device.assemblyLine') },
    { value: 'Sensor', label: t('device.sensor') },
    { value: 'PLC', label: t('device.plc') },
    { value: 'motor', label: t('device.motor') },
    { value: 'pump', label: t('device.pump') },
    { value: 'valve', label: t('device.valve') },
    { value: 'heater', label: t('device.heater') },
    { value: 'cooler', label: t('device.cooler') },
    { value: 'gauge', label: t('device.gauge') },
  ] as const, [t]);

  // FE-P1-01: 使用 useMemo 包裹设备类型选项数组，避免每次渲染创建新对象
  const DEVICE_TYPE_OPTIONS = useMemo(() => [
    { value: 'CNC', label: t('device.cnc') },
    { value: 'InjectionMolder', label: t('device.injectionMolder') },
    { value: 'AssemblyRobot', label: t('device.assemblyRobot') },
    { value: 'Conveyor', label: t('device.assemblyLine') },
    { value: 'Sensor', label: t('device.sensor') },
    { value: 'PLC', label: t('device.plc') },
    { value: 'motor', label: t('device.motor') },
    { value: 'pump', label: t('device.pump') },
    { value: 'valve', label: t('device.valve') },
    { value: 'heater', label: t('device.heater') },
    { value: 'cooler', label: t('device.cooler') },
    { value: 'gauge', label: t('device.gauge') },
  ] as const, [t]);

  const loadDevices = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getDevices(page, 20);
      setDevices(res.data ?? []);
      setTotal(res.total ?? 0);
    } catch (error) {
      console.error('Failed to load devices:', error);
      showToast({ type: 'error', message: t('device.loadFailed') });
      setDevices([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [page, showToast, t]);

  useEffect(() => {
    loadDevices();
  }, [loadDevices]);

  const handleDelete = async (id: string) => {
    // FE-P2-11: 使用自定义确认框替代原生 confirm()
    const confirmed = await showConfirm({
      title: t('device.deleteConfirmTitle'),
      message: t('device.deleteConfirm'),
      variant: 'danger',
      confirmText: t('common.delete'),
      cancelText: t('common.cancel'),
    });
    if (!confirmed) return;
    try {
      await api.deleteDevice(id);
      showToast({ type: 'success', message: t('device.deleteSuccess') });
      loadDevices();
    } catch (error) {
      showToast({ type: 'error', message: t('device.deleteFailed') });
    }
  };

  // FE-P2-02: 使用 useMemo 优化 filteredDevices 过滤计算
  const filteredDevices = useMemo(() => devices.filter(d => {
    const matchSearch = d.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      d.id.toLowerCase().includes(searchTerm.toLowerCase());
    const matchType = !typeFilter || d.type === typeFilter;
    return matchSearch && matchType;
  }), [devices, searchTerm, typeFilter]);

  const totalPages = Math.ceil(total / PAGE_SIZE) || 1;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.devices')}</h1>
          <p className="text-slate-400">{t('device.manageDevices')}</p>
        </div>
        <div className="flex items-center gap-2">
          <ExportButton reportType="devices" />
          {isAdmin && (
            <button 
              data-testid="add-device-btn"
              onClick={() => setShowCreateModal(true)}
              className="btn btn-primary flex items-center gap-2"
            >
              <Plus className="w-5 h-5" />
              <span>{t('common.create')}</span>
            </button>
          )}
        </div>
      </div>

      {/* Search & Filter */}
      <div className="card">
        <div className="card-body">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="input pl-10"
                placeholder={t('common.search')}
              />
            </div>
            <select
              data-testid="type-filter"
              value={typeFilter}
              onChange={(e) => { setTypeFilter(e.target.value); setPage(1); }}
              className="input w-auto min-w-[140px]"
              aria-label={t('device.type')}
            >
              {DEVICE_TYPES.map(dt => (
                <option key={dt.value} value={dt.value}>{dt.value ? dt.label : `${t('device.type')}: ${dt.label}`}</option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Device table */}
      <div className="card">
        <div className="card-body">
          {loading ? (
            <SkeletonTable rows={10} />
          ) : (
            <div className="table-container" data-testid="device-table">
              <table className="table">
                <thead>
                  <tr>
                    <th>{t('device.id')}</th>
                    <th>{t('device.name')}</th>
                    <th>{t('device.type')}</th>
                    <th>{t('device.location')}</th>
                    <th>{t('device.status')}</th>
                    <th>{t('common.edit')}</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredDevices.map((device) => (
                    <tr key={device.id} data-testid={`device-row-${device.id}`}>
                      <td className="font-mono text-sm">{device.id}</td>
                      <td>{device.name}</td>
                      <td>{device.type}</td>
                      <td>{device.location}</td>
                      <td>
                        <span className={`status-badge ${getDeviceStatusBadgeClass(device.status)}`}>
                          {t(`device.${device.status}`)}
                        </span>
                      </td>
                      <td>
                        <div className="flex items-center gap-2">
                          <button
                            data-testid="edit-btn"
                            onClick={() => setEditingDevice(device)}
                            className="p-1 text-slate-400 hover:text-primary-400"
                          >
                            <Edit className="w-4 h-4" />
                          </button>
                          {isAdmin && (
                            <button
                              data-testid="delete-btn"
                              onClick={() => handleDelete(device.id)}
                              className="p-1 text-slate-400 hover:text-red-400"
                            >
                              <Trash2 className="w-4 h-4" />
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

          {/* Pagination */}
          <div className="flex items-center justify-between mt-4" data-testid="pagination">
            <span className="text-sm text-slate-400">
              {t('common.all')} {total} {t('device.deviceCount')} · {t('common.filter')}: {filteredDevices.length}
            </span>
            <div className="flex items-center gap-2">
              <button
                data-testid="prev-page-btn"
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="btn btn-secondary disabled:opacity-50 flex items-center gap-1"
              >
                <ChevronLeft className="w-4 h-4" />
                {t('common.prev')}
              </button>
              <span className="text-sm text-slate-300 px-2" data-testid="page-info">
                {page} / {totalPages}
              </span>
              <button
                data-testid="next-page-btn"
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
                className="btn btn-secondary disabled:opacity-50 flex items-center gap-1"
              >
                {t('common.next')}
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Create/Edit Modal */}
      {(showCreateModal || editingDevice) && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" role="dialog" aria-modal="true">
          <div className="card max-w-md">
            <div className="card-header">
              <h2 className="text-lg font-semibold">
                {editingDevice ? t('device.edit') : t('device.create')}
              </h2>
            </div>
            <div className="card-body">
              <form onSubmit={async (e) => {
                e.preventDefault();
                const formData = new FormData(e.target as HTMLFormElement);
                const data = {
                  id: formData.get('device-id') as string,
                  name: formData.get('device-name') as string,
                  type: formData.get('device-type') as string,
                  location: formData.get('device-location') as string,
                  status: formData.get('status') as DeviceStatus,
                  description: formData.get('description') as string,
                };

                try {
                  if (editingDevice) {
                    await api.updateDevice(editingDevice.id, data);
                    showToast({ type: 'success', message: t('device.updateSuccess') });
                  } else {
                    await api.createDevice(data);
                    showToast({ type: 'success', message: t('device.createSuccess') });
                  }
                  setShowCreateModal(false);
                  setEditingDevice(null);
                  loadDevices();
                } catch (error) {
                  showToast({ type: 'error', message: t('device.operationFailed') });
                }
              }}>
                <div className="space-y-4">
                  {!editingDevice && (
                    <div>
                      <label className="label">{t('device.id')}</label>
                      <input name="device-id" className="input" required defaultValue="" />
                    </div>
                  )}
                  <div>
                    <label className="label">{t('device.name')}</label>
                    <input name="device-name" className="input" required defaultValue={editingDevice?.name} />
                  </div>
                  <div>
                    <label className="label">{t('device.type')}</label>
                    <select name="device-type" className="input" defaultValue={editingDevice?.type}>
                      {DEVICE_TYPE_OPTIONS.map(dt => (
                        <option key={dt.value} value={dt.value}>{dt.label}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="label">{t('device.location')}</label>
                    <input name="device-location" className="input" defaultValue={editingDevice?.location} />
                  </div>
                  <div>
                    <label className="label">{t('device.status')}</label>
                    <select name="status" className="input" defaultValue={editingDevice?.status}>
                      <option value="online">{t('device.online')}</option>
                      <option value="warning">{t('device.warning')}</option>
                      <option value="fault">{t('device.fault')}</option>
                      <option value="offline">{t('device.offline')}</option>
                    </select>
                  </div>
                  <div className="flex gap-2">
                    <button type="submit" className="btn btn-primary flex-1">
                      {t('common.save')}
                    </button>
                    <button 
                      type="button"
                      onClick={() => { setShowCreateModal(false); setEditingDevice(null); }}
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