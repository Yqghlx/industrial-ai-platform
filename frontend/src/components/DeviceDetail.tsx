import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import api from '../lib/api';
import { useI18n } from '../i18n';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { Thermometer, Waves, BarChart3 } from 'lucide-react';
import { Device, Telemetry, DeviceStats } from '../types/api';
import { asDeviceSafe, asTelemetryArraySafe, asDeviceStatsSafe } from '../types/typeGuards';

export default function DeviceDetail() {
  const { id } = useParams<{ id: string }>();
  const { t } = useI18n();
  const { showToast } = useToast();
  const [device, setDevice] = useState<Device | null>(null);
  const [telemetry, setTelemetry] = useState<Telemetry[]>([]);
  const [stats, setStats] = useState<DeviceStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [timeRange, setTimeRange] = useState('24h');

  // Stable loadData function for useEffect dependencies
  const loadData = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const [deviceRes, telemetryRes, statsRes] = await Promise.all([
        api.getDevice(id),
        api.getDeviceTelemetry(id, timeRange, 500),
        api.getDeviceStats(id, timeRange),
      ]);
      
      // FE-P1-01: 使用类型守卫安全转换
      const device = asDeviceSafe(deviceRes);
      const telemetry = asTelemetryArraySafe(telemetryRes.data);
      const stats = asDeviceStatsSafe(statsRes);
      
      setDevice(device);
      setTelemetry(telemetry);
      setStats(stats);
    } catch (error) {
      // FIX-023: 使用统一 showError toast 服务
      showToast({ type: 'error', message: t('errors.unknown') });
    } finally {
      setLoading(false);
    }
  }, [id, timeRange, showToast, t]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // FE-P2-03: 使用 useMemo 缓存 timeRangeOptions，避免每次渲染重新创建
  const timeRangeOptions = useMemo(() => [
    { value: '1h', label: t('telemetry.range1h') },
    { value: '6h', label: t('telemetry.range6h') },
    { value: '24h', label: t('telemetry.range24h') },
    { value: '7d', label: t('telemetry.range7d') },
  ], [t]);

  // FE-P2-04: 使用 useMemo 优化 chartData 计算，避免每次渲染重新映射
  const chartData = useMemo(() => telemetry.map(t => ({
    time: new Date(t.timestamp || t.time || Date.now()).toLocaleTimeString(),
    temperature: t.temperature,
    vibration: t.vibration,
    pressure: t.pressure,
    power: t.power,
  })), [telemetry]);

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'online': return 'status-online';
      case 'warning': return 'status-warning';
      case 'fault': return 'status-fault';
      default: return 'status-offline';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      {loading ? (
        <Skeleton variant="text" height={40} />
      ) : device ? (
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-slate-100">{device.name}</h1>
              <span className={`status-badge ${getStatusBadgeClass(device.status)}`}>
                {device.status}
              </span>
            </div>
            <p className="text-slate-400 mt-1">
              {device.id} · {device.type} · {device.location}
            </p>
          </div>
          
          {/* Time range selector */}
          <div className="flex gap-2">
            {timeRangeOptions.map((opt) => (
              <button
                key={opt.value}
                onClick={() => setTimeRange(opt.value)}
                className={`px-3 py-1 rounded-lg text-sm transition-colors ${
                  timeRange === opt.value
                    ? 'bg-primary-600 text-white'
                    : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                }`}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>
      ) : (
        <div className="text-red-400">{t('device.notFound')}</div>
      )}

      {/* Stats cards */}
      {loading ? (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map(i => <Skeleton key={i} variant="card" />)}
        </div>
      ) : stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="metric-card">
            <div className="flex items-center gap-2 mb-2">
              <Thermometer className="w-5 h-5 text-red-400" />
              <span className="metric-label">{t('telemetry.avg')} {t('telemetry.temperature')}</span>
            </div>
            <div className="metric-value">{stats.avg_temperature?.toFixed(1) || '--'}°C</div>
          </div>
          
          <div className="metric-card">
            <div className="flex items-center gap-2 mb-2">
              <Thermometer className="w-5 h-5 text-orange-400" />
              <span className="metric-label">{t('telemetry.max')} {t('telemetry.temperature')}</span>
            </div>
            <div className="metric-value text-orange-400">{stats.max_temperature?.toFixed(1) || '--'}°C</div>
          </div>
          
          <div className="metric-card">
            <div className="flex items-center gap-2 mb-2">
              <Waves className="w-5 h-5 text-blue-400" />
              <span className="metric-label">{t('telemetry.avg')} {t('telemetry.vibration')}</span>
            </div>
            <div className="metric-value">{stats.avg_vibration?.toFixed(2) || '--'} mm/s</div>
          </div>
          
          <div className="metric-card">
            <div className="flex items-center gap-2 mb-2">
              <BarChart3 className="w-5 h-5 text-purple-400" />
              <span className="metric-label">{t('telemetry.trend')}</span>
            </div>
            <div className="metric-value">{stats.data_points || 0}</div>
            <div className="text-xs text-slate-400">{t('telemetry.value')}</div>
          </div>
        </div>
      )}

      {/* Charts */}
      <div className="card">
        <div className="card-header">
          <h2 className="text-lg font-semibold text-slate-100">{t('telemetry.trend')}</h2>
        </div>
        <div className="card-body">
          {loading ? (
            <Skeleton variant="chart" />
          ) : chartData.length === 0 ? (
            <div className="h-64 flex flex-col items-center justify-center text-slate-400">
              <BarChart3 className="w-12 h-12 text-slate-500 mb-3" />
              <p className="text-lg text-slate-300">{t('telemetry.noHistoryData')}</p>
            </div>
          ) : (
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData}>
                  <XAxis dataKey="time" stroke="#64748b" />
                  <YAxis stroke="#64748b" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#1e293b',
                      border: '1px solid #334155',
                      borderRadius: '8px'
                    }}
                  />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="temperature"
                    stroke="#ef4444"
                    name={t('telemetry.temperature')}
                    dot={false}
                  />
                  <Line
                    type="monotone"
                    dataKey="vibration"
                    stroke="#3b82f6"
                    name={t('telemetry.vibration')}
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>
      </div>

      {/* Recent telemetry table */}
      <div className="card">
        <div className="card-header">
          <h2 className="text-lg font-semibold text-slate-100">{t('telemetry.history')}</h2>
        </div>
        <div className="card-body">
          {loading ? (
            <Skeleton variant="card" />
          ) : telemetry.length === 0 ? (
            <div className="py-12 text-center text-slate-400">
              <BarChart3 className="w-12 h-12 text-slate-500 mx-auto mb-3" />
              <p className="text-lg text-slate-300">{t('telemetry.noHistoryData')}</p>
            </div>
          ) : (
            <div className="table-container">
              <table className="table">
                <thead>
                  <tr>
                    <th>{t('telemetry.timestamp')}</th>
                    <th>{t('telemetry.temperature')}</th>
                    <th>{t('telemetry.vibration')}</th>
                    <th>{t('telemetry.pressure')}</th>
                    <th>{t('device.status')}</th>
                  </tr>
                </thead>
                <tbody>
                  {telemetry.slice(0, 10).map((t) => (
                    <tr key={t.timestamp || t.time || `${t.temperature}-${t.vibration}`}>
                      <td>{new Date(t.timestamp || t.time || Date.now()).toLocaleString()}</td>
                      <td>{t.temperature?.toFixed(1) || '--'}°C</td>
                      <td>{t.vibration?.toFixed(2) || '--'} mm/s</td>
                      <td>{t.pressure?.toFixed(1) || '--'} bar</td>
                      <td>
                        <span className={`status-badge ${getStatusBadgeClass(t.status)}`}>
                          {t.status}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}