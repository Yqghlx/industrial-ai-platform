import React, { useState, useEffect, useMemo, useCallback } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { Bell, Check, Filter } from 'lucide-react';
import { Notification } from '../types/api';

// FE-P1: 状态数组上限常量
const MAX_NOTIFICATIONS_ENTRIES = 500;

export default function NotificationCenter() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState({ type: '', unread: false });

  // FE-P1: 使用 useCallback 包装 loadNotifications，修复依赖问题
  const loadNotifications = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getNotifications(filter);
      // FE-P1: 限制数组大小，防止内存泄漏
      const data = (res.data as Notification[]).slice(0, MAX_NOTIFICATIONS_ENTRIES);
      setNotifications(data);
    } catch (error) {
      console.error('Failed to load notifications:', error);
      showToast({ type: 'error', message: t('errors.loadFailedNotifications') });
    } finally {
      setLoading(false);
    }
  }, [filter, showToast, t]);

  useEffect(() => {
    loadNotifications();
  }, [loadNotifications]);

  const handleMarkRead = async (id: number) => {
    try {
      await api.markNotificationRead(id);
      setNotifications(prev => prev.map(n => 
        n.id === id ? { ...n, read: true } : n
      ));
      showToast({ type: 'success', message: '已标记已读' });
    } catch (error) {
      showToast({ type: 'error', message: '操作失败' });
    }
  };

  const handleMarkAllRead = async () => {
    const unread = notifications.filter(n => !n.read);
    // FE-P2-07: 使用 Promise.all 替代串行 API 调用，提升性能
    try {
      await Promise.all(unread.map(n => api.markNotificationRead(n.id)));
      showToast({ type: 'success', message: '已全部标记已读' });
    } catch (error) {
      showToast({ type: 'error', message: '部分标记失败' });
    }
    loadNotifications();
  };

  // FE-P2-01: 使用 useMemo 优化 unreadCount 计算
  const unreadCount = useMemo(() => notifications.filter(n => !n.read).length, [notifications]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.notifications')}</h1>
          <p className="text-slate-400">{t('notification.title')} ({unreadCount} {t('notification.unread')})</p>
        </div>
        {unreadCount > 0 && (
          <button 
            onClick={handleMarkAllRead}
            className="btn btn-secondary flex items-center gap-2"
            aria-label={t('notification.markAllRead')}
          >
            <Check className="w-5 h-5" />
            <span>{t('notification.markAllRead')}</span>
          </button>
        )}
      </div>

      {/* Filters */}
      <div className="card">
        <div className="card-body">
          <div className="flex items-center gap-4">
            <Filter className="w-5 h-5 text-slate-400" />
            <select
              value={filter.type}
              onChange={(e) => setFilter({ ...filter, type: e.target.value })}
              className="input"
            >
              <option value="">全部类型</option>
              <option value="alert">告警</option>
              <option value="system">系统</option>
              <option value="work_order">工单</option>
            </select>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={filter.unread}
                onChange={(e) => setFilter({ ...filter, unread: e.target.checked })}
                className="w-4 h-4 rounded border-slate-600 text-primary-600 focus:ring-primary-500"
              />
              <span className="text-slate-300">{t('notification.unread')}</span>
            </label>
          </div>
        </div>
      </div>

      {/* Notifications list */}
      <div className="space-y-4">
        {loading ? (
          <div className="space-y-4">
            {[1, 2, 3, 4].map(i => <Skeleton key={i} variant="card" />)}
          </div>
        ) : notifications.length === 0 ? (
          <div className="card">
            <div className="card-body text-center py-8">
              <Bell className="w-12 h-12 text-slate-400 mx-auto mb-4" />
              <p className="text-slate-300">暂无通知</p>
            </div>
          </div>
        ) : (
          notifications.map((n) => (
            <div
              key={n.id}
              className={`card ${n.read ? 'opacity-70' : ''}`}
            >
              <div className="card-body">
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-3">
                    <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                      n.type === 'alert' ? 'bg-red-500/20' :
                      n.type === 'system' ? 'bg-blue-500/20' :
                      'bg-slate-500/20'
                    }`}>
                      <Bell className={`w-5 h-5 ${
                        n.type === 'alert' ? 'text-red-400' :
                        n.type === 'system' ? 'text-blue-400' :
                        'text-slate-400'
                      }`} />
                    </div>
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <span className={`text-sm ${
                          n.read ? 'text-slate-400' : 'text-slate-200 font-medium'
                        }`}>
                          {n.type}
                        </span>
                        {!n.read && (
                          <span className="px-2 py-0.5 bg-primary-500/20 text-primary-400 text-xs rounded-full">
                            {t('notification.unread')}
                          </span>
                        )}
                      </div>
                      <h3 className={`font-medium mb-1 ${
                        n.read ? 'text-slate-400' : 'text-slate-100'
                      }`}>
                        {n.title}
                      </h3>
                      <p className={`text-sm ${
                        n.read ? 'text-slate-500' : 'text-slate-300'
                      }`}>
                        {n.message}
                      </p>
                      <p className="text-xs text-slate-400 mt-2">
                        {new Date(n.created_at).toLocaleString()}
                        {n.device_id && ` · 设备: ${n.device_id}`}
                      </p>
                    </div>
                  </div>
                  {!n.read && (
                    <button
                      onClick={() => handleMarkRead(n.id)}
                      className="btn btn-secondary flex items-center gap-2"
                      aria-label={t('notification.markRead')}
                    >
                      <Check className="w-4 h-4" />
                      <span>{t('notification.markRead')}</span>
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>

    </div>
  );
}