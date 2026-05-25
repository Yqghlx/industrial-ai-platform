import React, { useState, useEffect } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { Activity, Thermometer, Waves, Zap, Settings } from 'lucide-react';
import { useWebSocket } from '../hooks/useWebSocket';
import { getGaugeColor, getGaugeStrokeColor, getGaugePercentage, getTelemetryStatusColor } from '../lib/colorUtils';
import { isTelemetry, asTelemetryArraySafe } from '../types/typeGuards';
import { Telemetry } from '../types/api';

export default function DigitalTwinPanel() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [telemetry, setTelemetry] = useState<Telemetry[]>([]);
  const [selectedDevice, setSelectedDevice] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 5000); // Refresh every 5s
    return () => clearInterval(interval);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

// Use shared WebSocket connection
  useWebSocket({
    onMessage: (message) => {
      if (message.type === 'telemetry') {
        // FE-P1-01: 使用类型守卫替代 as Type 断言
        if (isTelemetry(message.payload)) {
          const payload = message.payload;
          setTelemetry(prev => {
            const exists = prev.find(t => t.device_id === payload.device_id);
            if (exists) {
              return prev.map(t =>
                t.device_id === payload.device_id ? payload : t
              );
            }
            return [...prev, payload];
          });
        }
      }
    },
  });

  const loadData = async () => {
    try {
      const res = await api.getLatestTelemetry();
      // FE-P1-01: 使用类型守卫安全转换数组
      const telemetryData = asTelemetryArraySafe(res.data);
      setTelemetry(telemetryData);
      if (!selectedDevice && telemetryData.length > 0) {
        setSelectedDevice(telemetryData[0].device_id);
      }
    } catch (error) {
      // FIX-023: 使用统一 showError toast 服务
      showToast({ type: 'error', message: t('errors.unknown') });
    } finally {
      setLoading(false);
    }
  };

  const currentDevice = telemetry.find(t => t.device_id === selectedDevice);

  // Note: Gauge color functions are imported from colorUtils.ts

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.digitalTwin')}</h1>
          <p className="text-slate-400">{t('digitalTwin.realTimeMonitoring')}</p>
        </div>
        <div className="flex items-center gap-2">
          <Activity className="w-5 h-5 text-green-400 animate-pulse" />
          <span className="text-sm text-slate-400">{t('digitalTwin.realTimeUpdate')}</span>
        </div>
      </div>

      {/* Device selector */}
      <div className="card">
        <div className="card-body">
          <div className="flex items-center gap-4">
            <Settings className="w-5 h-5 text-slate-400" />
            <select
              value={selectedDevice || ''}
              onChange={(e) => setSelectedDevice(e.target.value)}
              className="input flex-1"
            >
              {telemetry.map((t) => (
                <option key={t.device_id} value={t.device_id}>
                  {t.device_id} - {t.status}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Gauges */}
      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[1, 2, 3].map(i => <Skeleton key={i} variant="card" />)}
        </div>
      ) : currentDevice && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Temperature gauge */}
          <div className="card">
            <div className="card-body text-center">
              <div className="flex items-center justify-center gap-2 mb-4">
                <Thermometer className={`w-6 h-6 ${getGaugeColor(currentDevice.temperature, 'temperature')}`} />
                <span className="text-lg font-medium text-slate-200">{t('telemetry.temperature')}</span>
              </div>
              <div className="relative w-24 h-24 mx-auto mb-4">
                <svg viewBox="0 0 100 100" className="w-full h-full">
                  <circle
                    cx="50"
                    cy="50"
                    r="45"
                    fill="none"
                    stroke="#334155"
                    strokeWidth="8"
                  />
                  <circle
                    cx="50"
                    cy="50"
                    r="45"
                    fill="none"
                    stroke={getGaugeStrokeColor(currentDevice.temperature, 'temperature')}
                    strokeWidth="8"
                    strokeDasharray={`${getGaugePercentage(currentDevice.temperature, 'temperature') * 2.83} 283`}
                    strokeLinecap="round"
                    transform="rotate(-90 50 50)"
                  />
                </svg>
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className={`text-2xl font-bold ${getGaugeColor(currentDevice.temperature, 'temperature')}`}>
                    {currentDevice.temperature?.toFixed(1) || '--'}
                  </span>
                </div>
              </div>
              <div className="text-sm text-slate-400">°C</div>
            </div>
          </div>

          {/* Vibration gauge */}
          <div className="card">
            <div className="card-body text-center">
              <div className="flex items-center justify-center gap-2 mb-4">
                <Waves className={`w-6 h-6 ${getGaugeColor(currentDevice.vibration, 'vibration')}`} />
                <span className="text-lg font-medium text-slate-200">{t('telemetry.vibration')}</span>
              </div>
              <div className="relative w-24 h-24 mx-auto mb-4">
                <svg viewBox="0 0 100 100" className="w-full h-full">
                  <circle
                    cx="50"
                    cy="50"
                    r="45"
                    fill="none"
                    stroke="#334155"
                    strokeWidth="8"
                  />
                  <circle
                    cx="50"
                    cy="50"
                    r="45"
                    fill="none"
                    stroke={getGaugeStrokeColor(currentDevice.vibration, 'vibration')}
                    strokeWidth="8"
                    strokeDasharray={`${getGaugePercentage(currentDevice.vibration, 'vibration') * 2.83} 283`}
                    strokeLinecap="round"
                    transform="rotate(-90 50 50)"
                  />
                </svg>
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className={`text-2xl font-bold ${getGaugeColor(currentDevice.vibration, 'vibration')}`}>
                    {currentDevice.vibration?.toFixed(2) || '--'}
                  </span>
                </div>
              </div>
              <div className="text-sm text-slate-400">mm/s</div>
            </div>
          </div>

          {/* Pressure gauge */}
          <div className="card">
            <div className="card-body text-center">
              <div className="flex items-center justify-center gap-2 mb-4">
                <Zap className={`w-6 h-6 ${getGaugeColor(currentDevice.pressure, 'pressure')}`} />
                <span className="text-lg font-medium text-slate-200">{t('telemetry.pressure')}</span>
              </div>
              <div className="relative w-24 h-24 mx-auto mb-4">
                <svg viewBox="0 0 100 100" className="w-full h-full">
                  <circle
                    cx="50"
                    cy="50"
                    r="45"
                    fill="none"
                    stroke="#334155"
                    strokeWidth="8"
                  />
                  <circle
                    cx="50"
                    cy="50"
                    r="45"
                    fill="none"
                    stroke={getGaugeStrokeColor(currentDevice.pressure, 'pressure')}
                    strokeWidth="8"
                    strokeDasharray={`${getGaugePercentage(currentDevice.pressure, 'pressure') * 2.83} 283`}
                    strokeLinecap="round"
                    transform="rotate(-90 50 50)"
                  />
                </svg>
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className={`text-2xl font-bold ${getGaugeColor(currentDevice.pressure, 'pressure')}`}>
                    {currentDevice.pressure?.toFixed(1) || '--'}
                  </span>
                </div>
              </div>
              <div className="text-sm text-slate-400">bar</div>
            </div>
          </div>
        </div>
      )}

      {/* Status indicator */}
      {currentDevice && (
        <div className="card">
          <div className="card-body">
            <div className="flex items-center justify-center gap-4">
              <div className={`w-4 h-4 rounded-full ${getTelemetryStatusColor(currentDevice.status)}`} />
              <span className="text-xl font-medium text-slate-100">
                {currentDevice.device_id} - {currentDevice.status}
              </span>
              <span className="text-slate-400">
                {new Date(currentDevice.timestamp).toLocaleTimeString()}
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}