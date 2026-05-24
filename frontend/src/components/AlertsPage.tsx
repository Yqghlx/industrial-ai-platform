import React, { useState, useEffect, useCallback } from 'react';
import { useI18n } from '../i18n';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { useWebSocket } from '../hooks/useWebSocket';
import {
  AlertTriangle,
  CheckCircle,
  Clock,
  Filter,
  RefreshCw,
  Bell,
  AlertCircle,
  Info,
  BarChart3,
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { Alert as AlertType } from '../types/api';
import { isAlert, asAlertStatusSafe, isAlertStatusPayload } from '../types/typeGuards';

interface AlertStats {
  active_count: number;
  total_count: number;
  by_severity: Record<string, number>;
  by_status: Record<string, number>;
}

const severityConfig = {
  critical: { label: '紧急', bgColor: 'bg-red-500', textColor: 'text-red-500', icon: AlertTriangle },
  high: { label: '高', bgColor: 'bg-orange-500', textColor: 'text-orange-500', icon: AlertCircle },
  medium: { label: '中', bgColor: 'bg-yellow-500', textColor: 'text-yellow-500', icon: Info },
  low: { label: '低', bgColor: 'bg-green-500', textColor: 'text-green-500', icon: Bell },
};

const statusConfig = {
  active: { label: '活跃', bgColor: 'bg-red-100', textColor: 'text-red-700', icon: AlertTriangle },
  acknowledged: { label: '已确认', bgColor: 'bg-yellow-100', textColor: 'text-yellow-700', icon: Clock },
  resolved: { label: '已解决', bgColor: 'bg-green-100', textColor: 'text-green-700', icon: CheckCircle },
};

export default function AlertsPage() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const navigate = useNavigate();
  const [alerts, setAlerts] = useState<AlertType[]>([]);
  const [stats, setStats] = useState<AlertStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [severityFilter, setSeverityFilter] = useState<string>('all');

  // WebSocket for real-time alerts
  useWebSocket({
    onMessage: (message) => {
      if (message.type === 'alert') {
        // FE-P1-01: 使用类型守卫替代 as Type 断言
        if (isAlert(message.payload)) {
          const newAlert = message.payload;
          setAlerts(prev => [newAlert, ...prev]);
          setStats(prev => prev ? {
            ...prev,
            active_count: prev.active_count + 1,
            total_count: prev.total_count + 1,
            by_severity: {
              ...prev.by_severity,
              [newAlert.severity]: (prev.by_severity[newAlert.severity] || 0) + 1,
            },
            by_status: {
              ...prev.by_status,
              active: (prev.by_status.active || 0) + 1,
            },
          } : null);

          showToast({
            type: newAlert.severity === 'critical' ? 'error' : 'info',
            message: `${severityConfig[newAlert.severity].label}: ${newAlert.message || ''}`,
          });
        }
      } else if (message.type === 'alert_resolved' || message.type === 'alert_acknowledged') {
        // FE-P1-01: 使用类型守卫替代 as Type 断言
        if (isAlertStatusPayload(message.payload)) {
          const alertId = message.payload.id;
          const newStatus = asAlertStatusSafe(message.payload.status);
          if (newStatus) {
            setAlerts(prev => prev.map(a => 
              a.id === alertId ? { ...a, status: newStatus } : a
            ));
          }
        }
        fetchStats();
      }
    },
  });

  // Fetch alerts
  const fetchAlerts = useCallback(async () => {
    try {
      const params = new URLSearchParams();
      if (statusFilter !== 'all') params.append('status', statusFilter);
      if (severityFilter !== 'all') params.append('severity', severityFilter);

      const response = await fetch(`/api/v1/alerts?${params.toString()}`, {
        headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
      });

      if (!response.ok) throw new Error('Failed to fetch alerts');

      const data = await response.json();
      setAlerts(data.data || []);
    } catch {
      showToast({ type: 'error', message: t('errors.unknown') });
    }
  }, [statusFilter, severityFilter, showToast, t]);

  // Fetch stats
  const fetchStats = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/alerts/stats', {
        headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
      });

      if (!response.ok) throw new Error('Failed to fetch stats');

      const data = await response.json();
      setStats(data);
    } catch {
      console.error('Failed to fetch alert stats');
      showToast({ type: 'error', message: t('errors.loadFailedAlertStats') });
    }
  }, [showToast, t]);

  // Initial load
  useEffect(() => {
    const load = async () => {
      setLoading(true);
      await Promise.all([fetchAlerts(), fetchStats()]);
      setLoading(false);
    };
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Reload on filter change
  useEffect(() => {
    if (!loading) {
      fetchAlerts();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [statusFilter, severityFilter]);

  // Refresh
  const handleRefresh = async () => {
    setRefreshing(true);
    await Promise.all([fetchAlerts(), fetchStats()]);
    setRefreshing(false);
  };

  // Resolve alert
  const handleResolve = async (alertId: number) => {
    try {
      const response = await fetch(`/api/v1/alerts/${alertId}/resolve`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
      });

      if (!response.ok) throw new Error('Failed to resolve alert');

      setAlerts(prev => prev.map(a => 
        a.id === alertId ? { ...a, status: 'resolved', resolved_at: new Date().toISOString() } : a
      ));
      fetchStats();

      showToast({ type: 'success', message: t('alert.alertResolved', { id: alertId }) });
    } catch {
      showToast({ type: 'error', message: t('errors.unknown') });
    }
  };

  // Acknowledge alert
  const handleAcknowledge = async (alertId: number) => {
    try {
      const response = await fetch(`/api/v1/alerts/${alertId}/acknowledge`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
      });

      if (!response.ok) throw new Error('Failed to acknowledge alert');

      setAlerts(prev => prev.map(a => 
        a.id === alertId ? { ...a, status: 'acknowledged' } : a
      ));
      fetchStats();

      showToast({ type: 'success', message: t('alert.alertAcknowledged', { id: alertId }) });
    } catch {
      showToast({ type: 'error', message: t('errors.unknown') });
    }
  };

  // Format timestamp
  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <AlertTriangle className="h-6 w-6 text-orange-500" />
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.alerts')}</h1>
        </div>
        <button
          className="flex items-center gap-2 px-4 py-2 bg-slate-700 hover:bg-slate-600 rounded-lg text-slate-200 transition-colors"
          onClick={handleRefresh}
          disabled={refreshing}
          aria-label={t('common.refresh')}
        >
          <RefreshCw className={`h-4 w-4 ${refreshing ? 'animate-spin' : ''}`} />
          {t('common.refresh')}
        </button>
        <button
          className=\"flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-white transition-colors"
          onClick={() => navigate('/alerts/report')}
          aria-label={t('report.generate')}
        >
          <BarChart3 className="h-4 w-4" />
          {t('report.analysis')}
        </button>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid gap-4 md:grid-cols-4">
          <div className="p-4 bg-slate-800 rounded-lg border border-slate-700">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-slate-400">{t('alert.activeAlerts')}</span>
              <AlertTriangle className="h-4 w-4 text-red-500" />
            </div>
            <div className="text-2xl font-bold text-red-500">{stats.active_count}</div>
          </div>

          <div className="p-4 bg-slate-800 rounded-lg border border-slate-700">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-slate-400">{t('alert.criticalAlerts')}</span>
              <AlertCircle className="h-4 w-4 text-red-500" />
            </div>
            <div className="text-2xl font-bold text-slate-100">{stats.by_severity.critical || 0}</div>
          </div>

          <div className="p-4 bg-slate-800 rounded-lg border border-slate-700">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-slate-400">{t('alert.acknowledged')}</span>
              <Clock className="h-4 w-4 text-yellow-500" />
            </div>
            <div className="text-2xl font-bold text-slate-100">{stats.by_status.acknowledged || 0}</div>
          </div>

          <div className="p-4 bg-slate-800 rounded-lg border border-slate-700">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-slate-400">{t('alert.resolved')}</span>
              <CheckCircle className="h-4 w-4 text-green-500" />
            </div>
            <div className="text-2xl font-bold text-slate-100">{stats.by_status.resolved || 0}</div>
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex items-center gap-2">
          <Filter className="h-4 w-4 text-slate-400" />
          <select
            className="px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 focus:outline-none focus:border-slate-500"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
          >
            <option value="all">{t('alert.allStatuses')}</option>
            <option value="active">{t('alert.activeLabel')}</option>
            <option value="acknowledged">{t('alert.acknowledgedLabel')}</option>
            <option value="resolved">{t('alert.resolvedLabel')}</option>
          </select>
        </div>

        <select
          className="px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 focus:outline-none focus:border-slate-500"
          value={severityFilter}
          onChange={(e) => setSeverityFilter(e.target.value)}
        >
          <option value="all">{t('alert.allSeverities')}</option>
          <option value="critical">{t('alert.criticalLabel')}</option>
          <option value="high">{t('alert.highLabel')}</option>
          <option value="medium">{t('alert.mediumLabel')}</option>
          <option value="low">{t('alert.lowLabel')}</option>
        </select>
      </div>

      {/* Alerts List */}
      <div className="p-4 bg-slate-800 rounded-lg border border-slate-700">
        <h2 className="text-lg font-semibold text-slate-100 mb-4">{t('alert.alertList')}</h2>

        {loading ? (
          <div className="space-y-4">
            {[1, 2, 3].map(i => (
              <div key={i} className="flex items-center gap-4">
                <Skeleton className="h-12 w-12 rounded-full" />
                <div className="space-y-2 flex-1">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-4 w-1/2" />
                </div>
              </div>
            ))}
          </div>
        ) : alerts.length === 0 ? (
          <div className="text-center py-8 text-slate-400">
            <Bell className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>{t('alert.noAlerts')}</p>
          </div>
        ) : (
          <div className="space-y-4">
            {alerts.map(alert => {
              const severity = severityConfig[alert.severity];
              const status = statusConfig[alert.status];
              const SeverityIcon = severity.icon;

              return (
                <div
                  key={alert.id}
                  className="flex items-start gap-4 p-4 rounded-lg border border-slate-600 bg-slate-700/50 hover:bg-slate-700 transition-colors"
                >
                  {/* Severity Icon */}
                  <div className={`p-2 rounded-full ${severity.bgColor}`}>
                    <SeverityIcon className="h-5 w-5 text-white" />
                  </div>

                  {/* Alert Content */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className={`px-2 py-0.5 rounded text-xs font-medium ${severity.bgColor} text-white`}>
                        {severity.label}
                      </span>
                      <span className={`px-2 py-0.5 rounded text-xs font-medium ${status.bgColor} ${status.textColor}`}>
                        {status.label}
                      </span>
                      <span className="text-xs text-slate-400">
                        #{alert.id}
                      </span>
                    </div>
                    <p className="font-medium text-slate-100 mb-1">{alert.message}</p>
                    <div className="flex items-center gap-4 text-sm text-slate-400">
                      <span>{t('alert.device')}: {alert.device_id}</span>
                      <span>{t('alert.triggered')}: {formatTime(alert.triggered_at)}</span>
                      {alert.resolved_at && (
                        <span>{t('alert.resolvedTime')}: {formatTime(alert.resolved_at)}</span>
                      )}
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex gap-2">
                    {alert.status === 'active' && (
                      <>
<button
                        className="flex items-center gap-1 px-3 py-1.5 bg-yellow-600 hover:bg-yellow-500 rounded text-sm text-white transition-colors"
                        onClick={() => handleAcknowledge(alert.id)}
                        aria-label={t('alert.acknowledgedLabel')}
                      >
                        <Clock className="h-4 w-4" />
                        {t('alert.acknowledge')}
                      </button>
                      <button
                        className="flex items-center gap-1 px-3 py-1.5 bg-green-600 hover:bg-green-500 rounded text-sm text-white transition-colors"
                        onClick={() => handleResolve(alert.id)}
                        aria-label={t('alert.resolvedLabel')}
                      >
                        <CheckCircle className="h-4 w-4" />
                        {t('alert.resolve')}
                      </button>
                      </>
                    )}
                    {alert.status === 'acknowledged' && (
                      <button
                        className="flex items-center gap-1 px-3 py-1.5 bg-green-600 hover:bg-green-500 rounded text-sm text-white transition-colors"
                        onClick={() => handleResolve(alert.id)}
                        aria-label={t('alert.resolvedLabel')}
                      >
                        <CheckCircle className="h-4 w-4" />
                        {t('alert.resolve')}
                      </button>
                    )}
                    {alert.status === 'resolved' && (
                      <span className="flex items-center gap-1 px-3 py-1.5 bg-green-100 text-green-700 rounded text-sm">
                        <CheckCircle className="h-4 w-4" />
                        {t('alert.processed')}
                      </span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}