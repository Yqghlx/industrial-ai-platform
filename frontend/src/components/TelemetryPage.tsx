import React, { useState, useEffect } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { Activity, Thermometer, Waves, Gauge, Zap, Clock, TrendingUp, AlertCircle } from 'lucide-react';
import { useWebSocket } from '../hooks/useWebSocket';
import { getGaugeColor, getTelemetryStatusColor } from '../lib/colorUtils';
import { isTelemetryData, isTelemetryDataArray, isTelemetryHistoryArray } from '../types/typeGuards';

interface TelemetryData {
  device_id: string;
  timestamp: string;
  temperature?: number;
  pressure?: number;
  vibration?: number;
  power?: number;
  humidity?: number;
  status: string;
}

interface TelemetryHistory {
  id: number;
  device_id: string;
  timestamp: string;
  temperature?: number;
  pressure?: number;
  vibration?: number;
  power?: number;
  humidity?: number;
}

export default function TelemetryPage() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [latestTelemetry, setLatestTelemetry] = useState<TelemetryData[]>([]);
  const [selectedDevice, setSelectedDevice] = useState<string | null>(null);
  const [history, setHistory] = useState<TelemetryHistory[]>([]);
  const [timeRange, setTimeRange] = useState<'1h' | '6h' | '24h'>('1h');
  const [loading, setLoading] = useState(true);
  const [historyLoading, setHistoryLoading] = useState(false);

  // WebSocket real-time updates - primary data source
  const { isConnected } = useWebSocket({
    onMessage: (message) => {
      if (message.type === 'telemetry') {
        // FE-P1-02: 使用类型守卫替代 as Type 断言
        if (isTelemetryData(message.payload)) {
          const payload = message.payload;
          setLatestTelemetry(prev => {
            const exists = prev.find(t => t.device_id === payload.device_id);
            if (exists) {
              return prev.map(d => d.device_id === payload.device_id ? payload : d);
            }
            return [...prev, payload];
          });
        }
      }
    },
  });

  // Load history when device or time range changes
  useEffect(() => {
    if (selectedDevice) {
      loadHistory(selectedDevice, timeRange);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedDevice, timeRange]);

  // Initial load
  useEffect(() => {
    loadLatestTelemetry();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Fallback polling when WebSocket is disconnected
  useEffect(() => {
    if (isConnected) {
      // WebSocket connected - no polling needed
      return;
    }
    // WebSocket disconnected - use polling fallback
    const interval = setInterval(loadLatestTelemetry, 10000);
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isConnected]);

  const loadLatestTelemetry = async () => {
    try {
      const res = await api.getLatestTelemetry();
      // FE-P1-02: 使用类型守卫替代 as Type 断言
      const data = isTelemetryDataArray(res.data) ? res.data : [];
      setLatestTelemetry(data);
      if (!selectedDevice && data.length > 0) {
        setSelectedDevice(data[0].device_id);
      }
    } catch {
      showToast({ type: 'error', message: t('errors.unknown') });
    } finally {
      setLoading(false);
    }
  };

  const loadHistory = async (deviceId: string, range: string) => {
    setHistoryLoading(true);
    try {
      const res = await api.getDeviceTelemetry(deviceId, range, 100);
      // FE-P1-02: 使用类型守卫替代 as Type 断言
      const data = isTelemetryHistoryArray(res.data) ? res.data : [];
      setHistory(data);
    } catch {
      showToast({ type: 'error', message: t('errors.unknown') });
    } finally {
      setHistoryLoading(false);
    }
  };

  const formatTimestamp = (ts: string) => {
    const date = new Date(ts);
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  const formatValue = (value?: number, unit?: string) => {
    if (value === undefined || value === null) return '--';
    return `${value.toFixed(2)}${unit || ''}`;
  };

  // Stats calculation
  const calculateStats = (field: keyof TelemetryHistory) => {
    const values = history.map(h => h[field]).filter((v): v is number => v !== undefined);
    if (values.length === 0) return { avg: '--', min: '--', max: '--' };
    const avg = values.reduce((a, b) => a + b, 0) / values.length;
    const min = Math.min(...values);
    const max = Math.max(...values);
    return {
      avg: avg.toFixed(2),
      min: min.toFixed(2),
      max: max.toFixed(2),
    };
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  const currentDevice = latestTelemetry.find(d => d.device_id === selectedDevice);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.telemetry')}</h1>
          <p className="text-slate-400">{t('telemetry.history')}</p>
        </div>
        <div className="flex items-center gap-2">
          <Activity className="w-5 h-5 text-green-400 animate-pulse" />
          <span className="text-sm text-slate-400">{t('common.connected')}</span>
        </div>
      </div>

      {/* Device Selector */}
      <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
        <label className="text-sm text-slate-400 mb-2 block">{t('device.id')}</label>
        <div className="flex flex-wrap gap-2">
          {latestTelemetry.map((d) => (
            <button
              key={d.device_id}
              onClick={() => setSelectedDevice(d.device_id)}
              className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                selectedDevice === d.device_id
                  ? 'bg-primary-600 text-white'
                  : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
              }`}
              aria-label={`${t('device.id')}: ${d.device_id}`}
            >
              {d.device_id}
              <span className={`ml-2 w-2 h-2 rounded-full inline-block ${getTelemetryStatusColor(d.status)}`} />
            </button>
          ))}
        </div>
      </div>

      {/* Current Metrics */}
      {currentDevice && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {/* Temperature */}
          <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
            <div className="flex items-center gap-2 mb-2">
              <Thermometer className={`w-5 h-5 ${getGaugeColor(currentDevice.temperature || 0, 'temperature')}`} />
              <span className="text-slate-400 text-sm">{t('telemetry.temperature')}</span>
            </div>
            <div className="text-2xl font-bold text-slate-100">
              {formatValue(currentDevice.temperature, '°C')}
            </div>
          </div>

          {/* Pressure */}
          <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
            <div className="flex items-center gap-2 mb-2">
              <Gauge className={`w-5 h-5 ${getGaugeColor(currentDevice.pressure || 0, 'pressure')}`} />
              <span className="text-slate-400 text-sm">{t('telemetry.pressure')}</span>
            </div>
            <div className="text-2xl font-bold text-slate-100">
              {formatValue(currentDevice.pressure, 'kPa')}
            </div>
          </div>

          {/* Vibration */}
          <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
            <div className="flex items-center gap-2 mb-2">
              <Waves className={`w-5 h-5 ${getGaugeColor(currentDevice.vibration || 0, 'vibration')}`} />
              <span className="text-slate-400 text-sm">{t('telemetry.vibration')}</span>
            </div>
            <div className="text-2xl font-bold text-slate-100">
              {formatValue(currentDevice.vibration, 'Hz')}
            </div>
          </div>

          {/* Power */}
          <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
            <div className="flex items-center gap-2 mb-2">
              <Zap className={`w-5 h-5 ${getGaugeColor(currentDevice.power || 0, 'power')}`} />
              <span className="text-slate-400 text-sm">{t('telemetry.power')}</span>
            </div>
            <div className="text-2xl font-bold text-slate-100">
              {formatValue(currentDevice.power, 'kW')}
            </div>
          </div>
        </div>
      )}

      {/* Time Range Selector */}
      <div className="flex items-center gap-2">
        <Clock className="w-4 h-4 text-slate-400" />
        <div className="flex gap-2">
          {(['1h', '6h', '24h'] as const).map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range)}
              className={`px-3 py-1 rounded text-sm ${
                timeRange === range
                  ? 'bg-primary-600 text-white'
                  : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
              }`}
            >
              {range === '1h' ? t('telemetry.range1h') : range === '6h' ? t('telemetry.range6h') : t('telemetry.range24h')}
            </button>
          ))}
        </div>
      </div>

      {/* Stats Summary */}
      {history.length > 0 && (
        <div className="bg-slate-800 rounded-lg p-4 border border-slate-700">
          <div className="flex items-center gap-2 mb-3">
            <TrendingUp className="w-5 h-5 text-primary-400" />
            <span className="text-slate-100 font-medium">{t('telemetry.stats')}</span>
          </div>
          <div className="grid grid-cols-3 md:grid-cols-6 gap-4 text-sm">
            {/* Temperature Stats */}
            <div>
              <span className="text-slate-400">{t('telemetry.temperature')} {t('telemetry.avg')}</span>
              <div className="text-slate-100 font-medium">{calculateStats('temperature').avg}°C</div>
            </div>
            <div>
              <span className="text-slate-400">{t('telemetry.temperature')} Min</span>
              <div className="text-slate-100 font-medium">{calculateStats('temperature').min}°C</div>
            </div>
            {/* Vibration Stats */}
            <div>
              <span className="text-slate-400">{t('telemetry.vibration')} {t('telemetry.avg')}</span>
              <div className="text-slate-100 font-medium">{calculateStats('vibration').avg}Hz</div>
            </div>
            <div>
              <span className="text-slate-400">{t('telemetry.vibration')} Max</span>
              <div className="text-slate-100 font-medium">{calculateStats('vibration').max}Hz</div>
            </div>
            {/* Pressure Stats */}
            <div>
              <span className="text-slate-400">{t('telemetry.pressure')} {t('telemetry.avg')}</span>
              <div className="text-slate-100 font-medium">{calculateStats('pressure').avg}kPa</div>
            </div>
            {/* Power Stats */}
            <div>
              <span className="text-slate-400">{t('telemetry.power')} {t('telemetry.avg')}</span>
              <div className="text-slate-100 font-medium">{calculateStats('power').avg}kW</div>
            </div>
          </div>
        </div>
      )}

      {/* History Table */}
      <div className="bg-slate-800 rounded-lg border border-slate-700 overflow-hidden">
        <div className="p-4 border-b border-slate-700">
          <span className="text-slate-100 font-medium">{t('telemetry.history')}</span>
          <span className="text-slate-400 text-sm ml-2">({t('telemetry.records', { count: history.length })})</span>
        </div>
        {historyLoading ? (
          <div className="p-8 text-center text-slate-400">
            <Activity className="w-6 h-6 animate-spin mx-auto mb-2" />
            {t('common.loading')}
          </div>
        ) : history.length === 0 ? (
          <div className="p-8 text-center">
            <AlertCircle className="w-8 h-8 text-slate-500 mx-auto mb-2" />
            <p className="text-slate-400">{selectedDevice ? t('telemetry.noHistoryData') : t('telemetry.selectDevice')}</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-slate-700/50">
                <tr>
                  <th className="px-4 py-2 text-left text-slate-400">{t('telemetry.timestamp')}</th>
                  <th className="px-4 py-2 text-left text-slate-400">{t('telemetry.temperature')}</th>
                  <th className="px-4 py-2 text-left text-slate-400">{t('telemetry.pressure')}</th>
                  <th className="px-4 py-2 text-left text-slate-400">{t('telemetry.vibration')}</th>
                  <th className="px-4 py-2 text-left text-slate-400">{t('telemetry.power')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                {history.slice(0, 20).map((h, i) => (
                  <tr key={h.id || i} className="hover:bg-slate-700/30">
                    <td className="px-4 py-2 text-slate-300">{formatTimestamp(h.timestamp)}</td>
                    <td className="px-4 py-2 text-slate-100">{formatValue(h.temperature, '°C')}</td>
                    <td className="px-4 py-2 text-slate-100">{formatValue(h.pressure, 'kPa')}</td>
                    <td className="px-4 py-2 text-slate-100">{formatValue(h.vibration, 'Hz')}</td>
                    <td className="px-4 py-2 text-slate-100">{formatValue(h.power, 'kW')}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}