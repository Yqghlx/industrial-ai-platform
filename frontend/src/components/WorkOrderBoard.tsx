import React, { useState, useEffect, useCallback } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { SkeletonTable } from './Skeleton';
import { useToast } from './Toast';
import { Plus, Search } from 'lucide-react';
import { WorkOrder } from '../types/api';
import { asWorkOrderArraySafe } from '../types/typeGuards';
import { getWorkOrderStatusColor, getWorkOrderPriorityColor } from '../lib/colorUtils';

export default function WorkOrderBoard() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [orders, setOrders] = useState<WorkOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);

  const loadOrders = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getWorkOrders({ status: statusFilter });
      // FE-P1-01: 使用类型守卫安全转换数组
      setOrders(asWorkOrderArraySafe(res.data));
    } catch (error) {
      console.error('Failed to load work orders:', error);
      showToast({ type: 'error', message: t('errors.loadFailedWorkOrders') });
    } finally {
      setLoading(false);
    }
  }, [statusFilter, showToast, t]);

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

  // Note: Color functions are imported from colorUtils.ts

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
          aria-label={t('workOrder.createOrder')}
        >
          <Plus className="w-5 h-5" />
          <span>{t('workOrder.createOrder')}</span>
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
                    <tr key={order.id}>
                      <td className="font-mono">#${order.id}</td>
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
                          onChange={(e) => handleUpdateStatus(order.id, e.target.value)}
                          className="input text-sm py-1"
                        >
                          <option value="pending">{t('workOrder.pending')}</option>
                          <option value="in_progress">{t('workOrder.inProgress')}</option>
                          <option value="completed">{t('workOrder.completed')}</option>
                          <option value="cancelled">{t('workOrder.cancelled')}</option>
                        </select>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" role="dialog" aria-modal="true">
          <div className="card max-w-md">
            <div className="card-header">
              <h2 className="text-lg font-semibold">{t('workOrder.createOrder')}</h2>
            </div>
            <div className="card-body">
              <form onSubmit={async (e) => {
                e.preventDefault();
                const formData = new FormData(e.target as HTMLFormElement);
                const data = {
                  title: formData.get('title') as string,
                  description: formData.get('description') as string,
                  device_id: formData.get('device_id') as string,
                  priority: formData.get('priority') as 'urgent' | 'high' | 'medium' | 'low',
                };

                try {
                  await api.createWorkOrder(data);
                  showToast({ type: 'success', message: t('workOrder.created') });
                  setShowCreateModal(false);
                  loadOrders();
                } catch (error) {
                  showToast({ type: 'error', message: t('workOrder.createFailed') });
                }
              }}>
                <div className="space-y-4">
                  <div>
                    <label className="label">{t('workOrder.title')}</label>
                    <input name="title" className="input" required />
                  </div>
                  <div>
                    <label className="label">{t('device.description')}</label>
                    <textarea name="description" className="input h-20" />
                  </div>
                  <div>
                    <label className="label">{t('device.id')}</label>
                    <input name="device_id" className="input" />
                  </div>
                  <div>
                    <label className="label">{t('workOrder.priority')}</label>
                    <select name="priority" className="input">
                      <option value="urgent">{t('workOrder.urgent')}</option>
                      <option value="high">{t('workOrder.high')}</option>
                      <option value="medium">{t('alert.medium')}</option>
                      <option value="low">{t('workOrder.low')}</option>
                    </select>
                  </div>
                  <div className="flex gap-2">
                    <button type="submit" className="btn btn-primary flex-1">
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