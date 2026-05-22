import React, { useState, useEffect } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { Database, Activity, Server, Clock, CheckCircle, AlertCircle } from 'lucide-react';
import { SystemStatus as SystemStatusType } from '../types/api';

export default function SystemStatus() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [status, setStatus] = useState<SystemStatusType | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    loadStatus();
    const interval = setInterval(loadStatus, 60000); // Refresh every minute
    return () => clearInterval(interval);
  }, []);

  const loadStatus = async () => {
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
    </div>
  );
}