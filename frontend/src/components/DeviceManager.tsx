import React, { useState, useEffect, useCallback } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { useAuth } from './AuthContext';
import { SkeletonTable } from './Skeleton';
import { useToast } from './Toast';
import ExportButton from './ExportButton';
import { Plus, Edit, Trash2, Search, ChevronLeft, ChevronRight } from 'lucide-react';
import { Device, DeviceStatus } from '../types/api';
import { getDeviceStatusBadgeClass } from '../lib/colorUtils';

const PAGE_SIZE = 20;

const DEVICE_TYPES = [
  { value: '', label: '全部类型' },
  { value: 'CNC', label: '数控机床' },
  { value: 'InjectionMolder', label: '注塑机' },
  { value: 'AssemblyRobot', label: '工业机器人' },
  { value: 'Conveyor', label: '装配线' },
  { value: 'Sensor', label: '传感器' },
  { value: 'PLC', label: 'PLC控制器' },
  { value: 'motor', label: '电机' },
  { value: 'pump', label: '泵' },
  { value: 'valve', label: '阀门' },
  { value: 'heater', label: '加热器' },
  { value: 'cooler', label: '冷却器' },
  { value: 'gauge', label: '仪表' },
] as const;

export default function DeviceManager() {
  const { t } = useI18n();
  const { isAdmin } = useAuth();
  const { showToast } = useToast();
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [editingDevice, setEditingDevice] = useState<Device | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);

  const loadDevices = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getDevices(page, 20);
      setDevices(res.data ?? []);
      setTotal(res.total ?? 0);
    } catch (error) {
      console.error('Failed to load devices:', error);
      showToast({ type: 'error', message: '加载设备失败' });
      setDevices([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [page, showToast]);

  useEffect(() => {
    loadDevices();
  }, [loadDevices]);

  const handleDelete = async (id: string) => {
    if (!confirm('确认删除此设备？')) return;
    try {
      await api.deleteDevice(id);
      showToast({ type: 'success', message: '设备已删除' });
      loadDevices();
    } catch (error) {
      showToast({ type: 'error', message: '删除失败' });
    }
  };

  const filteredDevices = devices.filter(d => {
    const matchSearch = d.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      d.id.toLowerCase().includes(searchTerm.toLowerCase());
    const matchType = !typeFilter || d.type === typeFilter;
    return matchSearch && matchType;
  });

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
                    showToast({ type: 'success', message: '设备已更新' });
                  } else {
                    await api.createDevice(data);
                    showToast({ type: 'success', message: '设备已创建' });
                  }
                  setShowCreateModal(false);
                  setEditingDevice(null);
                  loadDevices();
                } catch (error) {
                  showToast({ type: 'error', message: '操作失败' });
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
                      <option value="CNC">数控机床</option>
                      <option value="InjectionMolder">注塑机</option>
                      <option value="AssemblyRobot">工业机器人</option>
                      <option value="Conveyor">装配线</option>
                      <option value="Sensor">传感器</option>
                      <option value="PLC">PLC控制器</option>
                      <option value="motor">电机</option>
                      <option value="pump">泵</option>
                      <option value="valve">阀门</option>
                      <option value="heater">加热器</option>
                      <option value="cooler">冷却器</option>
                      <option value="gauge">仪表</option>
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