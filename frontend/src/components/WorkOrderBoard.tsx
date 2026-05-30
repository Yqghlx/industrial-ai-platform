import React, { useState, useEffect, useCallback } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { useEscapeKey } from '../lib/hooks';
import { SkeletonTable } from './Skeleton';
import { useToast } from './Toast';
import { Plus, Search } from 'lucide-react';
import { WorkOrder, Device } from '../types/api';
import { asWorkOrderArraySafe } from '../types/typeGuards';
import { getWorkOrderStatusColor, getWorkOrderPriorityColor } from '../lib/colorUtils';

// FE-P2: React.memo 优化工单行渲染
interface WorkOrderRowProps {
  order: WorkOrder;
  t: (key: string) => string;
  onUpdateStatus: (id: number, status: string) => void;
}

const WorkOrderRow = React.memo(function WorkOrderRow({ order, t, onUpdateStatus }: WorkOrderRowProps) {
  return (
    <tr>
      <td className="font-mono">#{order.id}</td>
      <td>{order.title}</td>
      <td className="font-mono text-sm">{order.device_id || '--'}</td>
      <td>
        <span className={`status-badge ${getWorkOrderPriorityColor(order.priority)}`}>
          {t(`workOrder.${order.priority}`)}
        </span>
      </td>
      <td>
        <span className={`status-badge ${getWorkOrderStatusColor(order.status)}`}>
          {t(`workOrder.${order.status === 'in_progress' ? 'inProgress' : order.status}`)}
        </span>
      </td>
      <td>{new Date(order.created_at).toLocaleDateString()}</td>
      <td>
        <select
          value={order.status}
          onChange={(e) => onUpdateStatus(order.id, e.target.value)}
          className="input text-sm py-1"
          aria-label={t('workOrder.updateStatus')}
        >
          <option value="pending">{t('workOrder.pending')}</option>
          <option value="in_progress">{t('workOrder.inProgress')}</option>
          <option value="completed">{t('workOrder.completed')}</option>
          <option value="cancelled">{t('workOrder.cancelled')}</option>
        </select>
      </td>
    </tr>
  );
});

export default function WorkOrderBoard() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [orders, setOrders] = useState<WorkOrder[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [creating, setCreating] = useState(false);

  // C1: Escape 键关闭模态框
  const closeCreateModal = useCallback(() => {
    setShowCreateModal(false);
  }, []);

  useEscapeKey(closeCreateModal, showCreateModal);

  const loadOrders = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getWorkOrders({ status: statusFilter });
      setOrders(asWorkOrderArraySafe(res.data));
    } catch (error) {
      showToast({ type: 'error', message: t('errors.loadFailedWorkOrders') });
    } finally {
      setLoading(false);
    }
  }, [statusFilter, showToast, t]);

  // 打开创建模态框时加载设备列表
  useEffect(() => {
    if (!showCreateModal) return;
    api.getDevices(1, 200).then(res => {
      setDevices(res.data || []);
    }).catch(() => {
      setDevices([]);
    });
  }, [showCreateModal]);

  useEffect(() => {
    loadOrders();
  }, [loadOrders]);

  const handleUpdateStatus = useCallback(async (id: number, status: string) => {
    try {
      await api.updateWorkOrderStatus(id, status);
      showToast({ type: 'success', message: t('workOrder.statusUpdated') });
      loadOrders();
    } catch (error) {
      showToast({ type: 'error', message: t('workOrder.updateFailed') });
    }
  }, [loadOrders, showToast, t]);

  const handleSubmit = useCallback(async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.target as HTMLFormElement);
    const data = {
      title: (formData.get('title') as string).trim(),
      description: (formData.get('description') as string).trim(),
      device_id: formData.get('device_id') as string,
      priority: formData.get('priority') as 'urgent' | 'high' | 'medium' | 'low',
    };

    if (!data.title) {
      showToast({ type: 'error', message: t('validation.required') });
      return;
    }
    if (!data.device_id) {
      showToast({ type: 'error', message: t('validation.selectDevice') });
      return;
    }

    setCreating(true);
    try {
      await api.createWorkOrder(data);
      showToast({ type: 'success', message: t('workOrder.created') });
      setShowCreateModal(false);
      loadOrders();
    } catch (error: unknown) {
      const msg = error instanceof Error ? error.message : t('workOrder.createFailed');
      showToast({ type: 'error', message: msg });
    } finally {
      setCreating(false);
    }
  }, [loadOrders, showToast, t]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.workOrders')}</h1>
          <p className="text-slate-400">{t('workOrder.title')}</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="btn btn-primary flex items-center gap-2"
          aria-label={t('workOrder.create')}
        >
          <Plus className="w-5 h-5" />
          <span>{t('workOrder.create')}</span>
        </button>
      </div>

      {/* Filters */}
      <div className="card">
        <div className="card-body">
          <div className="flex items-center gap-4">
            <Search className="w-5 h-5 text-slate-400" />
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="input"
              aria-label={t('workOrder.status')}
            >
              <option value="">{t('workOrder.allStatus')}</option>
              <option value="pending">{t('workOrder.pending')}</option>
              <option value="in_progress">{t('workOrder.inProgress')}</option>
              <option value="completed">{t('workOrder.completed')}</option>
              <option value="cancelled">{t('workOrder.cancelled')}</option>
            </select>
          </div>
        </div>
      </div>

      {/* Orders table */}
      <div className="card">
        <div className="card-body">
          {loading ? (
            <SkeletonTable rows={10} />
          ) : (
            <div className="table-container">
              <table className="table">
                <thead>
                  <tr>
                    <th>{t('workOrder.id')}</th>
                    <th>{t('workOrder.title')}</th>
                    <th>{t('device.id')}</th>
                    <th>{t('workOrder.priority')}</th>
                    <th>{t('workOrder.status')}</th>
                    <th>{t('workOrder.createdAt')}</th>
                    <th>{t('workOrder.updateStatus')}</th>
                  </tr>
                </thead>
<tbody>
                  {orders.map((order) => (
                    <WorkOrderRow
                      key={order.id}
                      order={order}
                      t={t}
                      onUpdateStatus={handleUpdateStatus}
                    />
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" role="dialog" aria-modal="true" onClick={(e) => { if (e.target === e.currentTarget) closeCreateModal(); }}>
          <div className="card max-w-md">
            <div className="card-header">
              <h2 className="text-lg font-semibold">{t('workOrder.createOrder')}</h2>
            </div>
            <div className="card-body">
              <form onSubmit={handleSubmit}>
                <div className="space-y-4">
                  <div>
                    <label className="label" htmlFor="wo-create-title">{t('workOrder.title')}</label>
                    <input id="wo-create-title" name="title" className="input" required minLength={2} maxLength={200} />
                  </div>
                  <div>
                    <label className="label" htmlFor="wo-create-description">{t('device.description')}</label>
                    <textarea id="wo-create-description" name="description" className="input h-20" maxLength={1000} />
                  </div>
                  <div>
                    <label className="label" htmlFor="wo-create-device-id">{t('device.id')}</label>
                    <select id="wo-create-device-id" name="device_id" className="input" required>
                      <option value="">{t('validation.selectDevice')}</option>
                      {devices.map(d => (
                        <option key={d.id} value={d.id}>{d.name} ({d.id})</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="label" htmlFor="wo-create-priority">{t('workOrder.priority')}</label>
                    <select id="wo-create-priority" name="priority" className="input">
                      <option value="urgent">{t('workOrder.urgent')}</option>
                      <option value="high">{t('workOrder.high')}</option>
                      <option value="medium">{t('alert.medium')}</option>
                      <option value="low">{t('workOrder.low')}</option>
                    </select>
                  </div>
                  <div className="flex gap-2">
                    <button type="submit" className="btn btn-primary flex-1" disabled={creating}>
                      {t('common.create')}
                    </button>
                    <button
                      type="button"
                      onClick={() => setShowCreateModal(false)}
                      className="btn btn-secondary flex-1"
                      aria-label={t('common.cancel')}
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
