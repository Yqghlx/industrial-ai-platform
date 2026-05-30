import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { Link, useLocation } from 'react-router-dom';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { SkeletonGrid } from './Skeleton';
import { useToast } from './Toast';
import { Activity, AlertTriangle, Wrench, TrendingUp, Settings, Bell } from 'lucide-react';
import { getDeviceStatusColor, getDeviceStatusBadgeClass } from '../lib/colorUtils';
import { Device, LatestTelemetry } from '../types/api';
import { useWebSocket } from '../hooks/useWebSocket';
import { isTelemetryData } from '../types/typeGuards';

// FE-P1: 状态数组上限常量
const MAX_TELEMETRY_ENTRIES = 500;

// FE-P2: React.memo 优化设备卡片渲染
interface DeviceCardProps {
  device: Device & { telemetry?: LatestTelemetry };
  t: (key: string) => string;
}

const DeviceCard = React.memo(function DeviceCard({ device, t }: DeviceCardProps) {
  return (
    <Link
      to={`/devices/${device.id}`}
      className="block p-3 lg:p-4 bg-slate-800/50 rounded-lg border border-slate-700 
               hover:border-primary-500 active:border-primary-400 transition-colors touch-manipulation"
    >
      <div className="flex items-start justify-between mb-2 lg:mb-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <div className={`w-2 h-2 rounded-full shrink-0 ${getDeviceStatusColor(device.status)}`} aria-label={t(`device.${device.status}`)} />
            <span className="font-medium text-slate-100 truncate">{device.name}</span>
          </div>
          <span className="text-xs lg:text-sm text-slate-400">{device.id}</span>
        </div>
        <span className={`status-badge ${getDeviceStatusBadgeClass(device.status)} shrink-0`}>
          {t(`device.${device.status}`)}
        </span>
      </div>
      
      {device.telemetry && (
        <div className="grid grid-cols-2 gap-2 text-xs lg:text-sm">
          <div>
            <span className="text-slate-400">{t('telemetry.temperature')}:</span>
            <span className="text-slate-200 ml-1">
              {device.telemetry.temperature?.toFixed(1) || '--'}°C
            </span>
          </div>
          <div>
            <span className="text-slate-400">{t('telemetry.vibration')}:</span>
            <span className="text-slate-200 ml-1">
              {device.telemetry.vibration?.toFixed(2) || '--'} mm/s
            </span>
          </div>
        </div>
      )}
      
      <div className="mt-1.5 lg:mt-2 text-xs text-slate-400 truncate">
        {device.location} · {device.type}
      </div>
    </Link>
  );
});

export default function FleetDashboard() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const location = useLocation();
  const [devices, setDevices] = useState<Device[]>([]);
  const [telemetry, setTelemetry] = useState<LatestTelemetry[]>([]);
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({
    total: 0,
    online: 0,
    warning: 0,
    fault: 0,
  });

  // WebSocket real-time updates - primary data source
  const { isConnected } = useWebSocket({
    onMessage: (message) => {
      if (message.type === 'telemetry') {
        // Update telemetry data in real-time
        if (isTelemetryData(message.payload)) {
          const payload = message.payload as LatestTelemetry;
          setTelemetry(prev => {
            const exists = prev.find(t => t.device_id === payload.device_id);
            let updated: LatestTelemetry[];
            if (exists) {
              updated = prev.map(t => t.device_id === payload.device_id ? payload : t);
            } else {
              updated = [...prev, payload];
            }
            // FE-P1: 限制数组大小，防止内存泄漏
            return updated.slice(-MAX_TELEMETRY_ENTRIES);
          });
        }
      }
    },
  });

  // FIX-006: 使用 useCallback 稳定化 loadData 函数
  const loadData = useCallback(async () => {
    try {
      const [devicesRes, telemetryRes] = await Promise.all([
        api.getDevices(),
        api.getLatestTelemetry(),
      ]);
      
      const devicesData = devicesRes.data ?? [];
      const telemetryData = telemetryRes.data ?? [];
      
      setDevices(devicesData);
      // FE-P1: 限制数组大小，防止内存泄漏
      setTelemetry(telemetryData.slice(-MAX_TELEMETRY_ENTRIES));
      
      // Calculate stats (safe with empty arrays)
      const total = devicesData.length;
      const online = devicesData.filter(d => d.status === 'online').length;
      const warning = devicesData.filter(d => d.status === 'warning').length;
      const fault = devicesData.filter(d => d.status === 'fault').length;
      setStats({ total, online, warning, fault });
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
      showToast({ type: 'error', message: t('errors.loadFailedDashboard') });
    } finally {
      setLoading(false);
    }
  }, [showToast, t]);

  // FE-P1: 使用 ref 跟踪是否已初始化，避免重复加载
  const isInitialMountRef = useRef(true);

  // FIX-006: 合并初始加载和路由变化监听，避免重复 useEffect 触发 loadData
  useEffect(() => {
    // 始终在挂载时加载数据
    if (isInitialMountRef.current) {
      isInitialMountRef.current = false;
    }
    loadData();
  }, [location.pathname, loadData]);

  // Fallback polling when WebSocket is disconnected
  useEffect(() => {
    if (isConnected) {
      // WebSocket connected - no polling needed
      return;
    }
    // WebSocket disconnected - use polling fallback
    const interval = setInterval(loadData, 30000);
    return () => clearInterval(interval);
  }, [isConnected, loadData]);

  // FE-P2-05: 使用 useMemo 优化 devicesWithTelemetry 计算，避免每次渲染重新计算
  // 使用 Map 将 O(n*m) 查找优化为 O(n+m)
  const devicesWithTelemetry = useMemo(() => {
    const telMap = new Map(telemetry.map(t => [t.device_id, t]));
    return devices.map(device => {
      const tel = telMap.get(device.id);
      return { ...device, telemetry: tel };
    });
  }, [devices, telemetry]);

  return (
    <div className="space-y-4 lg:space-y-6 mobile-page">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mobile-page-header">
        <div>
          <h1 className="mobile-page-title">{t('dashboard.overview')}</h1>
          <p className="mobile-page-subtitle">{t('dashboard.fleetStatus')}</p>
        </div>
        <Link to="/digital-twin" className="btn btn-primary flex items-center justify-center gap-2 w-full sm:w-auto">
          <Activity className="w-5 h-5" />
          <span>{t('nav.digitalTwin')}</span>
        </Link>
      </div>

      {/* Stats cards */}
      {loading ? (
        <SkeletonGrid count={4} />
      ) : (
        <div className="responsive-grid grid-cols-2 lg:grid-cols-4">
          <div className="metric-card">
            <div className="flex items-center gap-2 lg:gap-3 mb-1.5 lg:mb-2">
              <Settings className="w-4 h-4 lg:w-5 lg:h-5 text-primary-500" />
              <span className="metric-label">{t('device.deviceCount')}</span>
            </div>
            <div className="metric-value">{stats.total}</div>
          </div>
          
          <div className="metric-card">
            <div className="flex items-center gap-2 lg:gap-3 mb-1.5 lg:mb-2">
              <Activity className="w-4 h-4 lg:w-5 lg:h-5 text-green-500" />
              <span className="metric-label">{t('device.online')}</span>
            </div>
            <div className="metric-value text-green-400">{stats.online}</div>
          </div>
          
          <div className="metric-card">
            <div className="flex items-center gap-2 lg:gap-3 mb-1.5 lg:mb-2">
              <AlertTriangle className="w-4 h-4 lg:w-5 lg:h-5 text-yellow-500" />
              <span className="metric-label">{t('device.warning')}</span>
            </div>
            <div className="metric-value text-yellow-400">{stats.warning}</div>
          </div>
          
          <div className="metric-card">
            <div className="flex items-center gap-2 lg:gap-3 mb-1.5 lg:mb-2">
              <AlertTriangle className="w-4 h-4 lg:w-5 lg:h-5 text-red-500" />
              <span className="metric-label">{t('device.fault')}</span>
            </div>
            <div className="metric-value text-red-400">{stats.fault}</div>
          </div>
        </div>
      )}

      {/* Device grid */}
      <div className="card">
        <div className="card-header flex items-center justify-between">
          <h2 className="text-base lg:text-lg font-semibold text-slate-100">{t('nav.devices')}</h2>
          <Link to="/devices" className="text-sm text-primary-400 hover:text-primary-300 active:text-primary-200">
            {t('device.manageDevices')} →
          </Link>
        </div>
        <div className="card-body">
          {loading ? (
            <SkeletonGrid count={6} />
          ) : (
            <div className="responsive-grid-3">
              {devicesWithTelemetry.map((device) => (
                <DeviceCard
                  key={device.id}
                  device={device}
                  t={t}
                />
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Quick actions */}
      <div className="card">
        <div className="card-header">
          <h2 className="text-base lg:text-lg font-semibold text-slate-100">{t('dashboard.quickActions')}</h2>
        </div>
        <div className="card-body">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 lg:gap-4">
            <Link to="/ai-agent" className="btn btn-secondary flex items-center justify-center gap-2 py-3 text-sm lg:text-base">
              <TrendingUp className="w-4 h-4 lg:w-5 lg:h-5" />
              <span className="truncate">{t('nav.aiAgent')}</span>
            </Link>
            <Link to="/work-orders" className="btn btn-secondary flex items-center justify-center gap-2 py-3 text-sm lg:text-base">
              <Wrench className="w-4 h-4 lg:w-5 lg:h-5" />
              <span className="truncate">{t('nav.workOrders')}</span>
            </Link>
            <Link to="/notifications" className="btn btn-secondary flex items-center justify-center gap-2 py-3 text-sm lg:text-base">
              <Bell className="w-4 h-4 lg:w-5 lg:h-5" />
              <span className="truncate">{t('nav.notifications')}</span>
            </Link>
            <Link to="/reports" className="btn btn-secondary flex items-center justify-center gap-2 py-3 text-sm lg:text-base">
              <Activity className="w-4 h-4 lg:w-5 lg:h-5" />
              <span className="truncate">{t('nav.reports')}</span>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}