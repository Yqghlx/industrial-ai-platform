import React, { useState, useEffect } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import Skeleton from './Skeleton';
import ExportButton from './ExportButton';
import { BarChart3, TrendingUp, Clock, DollarSign, Activity, RefreshCw } from 'lucide-react';
import { ROIStats } from '../types/api';

export default function ROIDashboard() {
  const { t } = useI18n();
  const [stats, setStats] = useState<ROIStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadStats();
    const interval = setInterval(loadStats, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, []);

  const loadStats = async () => {
    setLoading(true);
    try {
      const res = await api.getROIStats();
      setStats(res as ROIStats);
    } catch (error) {
      console.error('Failed to load ROI stats:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.roi')}</h1>
          <p className="text-slate-400">{t('roi.title')}</p>
        </div>
        <div className="flex items-center gap-2">
          <ExportButton reportType="roi" />
          <button 
            onClick={loadStats}
            className="btn btn-secondary flex items-center gap-2"
          >
            <RefreshCw className="w-5 h-5" />
            <span>{t('common.refresh')}</span>
          </button>
        </div>
      </div>

      {loading ? (
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          {[1, 2, 3, 4, 5, 6].map(i => <Skeleton key={i} variant="card" />)}
        </div>
      ) : stats && (
        <>
          {/* Main metrics */}
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            <div className="metric-card">
              <div className="flex items-center gap-3 mb-2">
                <DollarSign className="w-5 h-5 text-green-500" />
                <span className="metric-label">{t('roi.predictedSavings')}</span>
              </div>
              <div className="metric-value text-green-400">
                ¥{stats.predicted_savings.toLocaleString()}
              </div>
              <div className="text-xs text-slate-400">{t('roi.monthlySavings')}</div>
            </div>

            <div className="metric-card">
              <div className="flex items-center gap-3 mb-2">
                <Activity className="w-5 h-5 text-primary-500" />
                <span className="metric-label">{t('roi.uptime')}</span>
              </div>
              <div className="metric-value text-primary-400">
                {stats.uptime_percentage.toFixed(1)}%
              </div>
            </div>

            <div className="metric-card">
              <div className="flex items-center gap-3 mb-2">
                <Clock className="w-5 h-5 text-orange-500" />
                <span className="metric-label">{t('roi.avgResponseTime')}</span>
              </div>
              <div className="metric-value text-orange-400">
                {stats.avg_response_time_hours.toFixed(1)}h
              </div>
            </div>

            <div className="metric-card">
              <div className="flex items-center gap-3 mb-2">
                <BarChart3 className="w-5 h-5 text-blue-500" />
                <span className="metric-label">{t('roi.totalDevices')}</span>
              </div>
              <div className="metric-value">{stats.total_devices}</div>
            </div>

            <div className="metric-card">
              <div className="flex items-center gap-3 mb-2">
                <TrendingUp className="w-5 h-5 text-yellow-500" />
                <span className="metric-label">{t('roi.activeAlerts')}</span>
              </div>
              <div className="metric-value text-yellow-400">{stats.active_alerts}</div>
            </div>

            <div className="metric-card">
              <div className="flex items-center gap-3 mb-2">
                <TrendingUp className="w-5 h-5 text-green-500" />
                <span className="metric-label">{t('roi.resolvedIssues')}</span>
              </div>
              <div className="metric-value text-green-400">{stats.resolved_issues}</div>
            </div>
          </div>

          {/* ROI summary */}
          <div className="card">
            <div className="card-header">
              <h2 className="text-lg font-semibold text-slate-100">{t('roi.analysisReport')}</h2>
            </div>
            <div className="card-body">
              <div className="space-y-4">
                <div className="flex items-center justify-between p-4 bg-slate-800/50 rounded-lg">
                  <span className="text-slate-300">{t('roi.predictionAccuracy')}</span>
                  <span className="text-2xl font-bold text-green-400">95%</span>
                </div>
                <div className="flex items-center justify-between p-4 bg-slate-800/50 rounded-lg">
                  <span className="text-slate-300">{t('roi.downtimeReduction')}</span>
                  <span className="text-2xl font-bold text-primary-400">40%</span>
                </div>
                <div className="flex items-center justify-between p-4 bg-slate-800/50 rounded-lg">
                  <span className="text-slate-300">{t('roi.costSavings')}</span>
                  <span className="text-2xl font-bold text-green-400">¥{stats.predicted_savings.toLocaleString()}</span>
                </div>
                <div className="flex items-center justify-between p-4 bg-slate-800/50 rounded-lg">
                  <span className="text-slate-300">{t('roi.investmentCycle')}</span>
                  <span className="text-2xl font-bold text-orange-400">3个月</span>
                </div>
              </div>
            </div>
          </div>

          {/* Benefits */}
          <div className="card">
            <div className="card-header">
              <h2 className="text-lg font-semibold text-slate-100">{t('roi.platformValue')}</h2>
            </div>
            <div className="card-body">
              <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                <div className="p-4 bg-green-500/10 border border-green-500/20 rounded-lg text-center">
                  <div className="text-3xl font-bold text-green-400 mb-2">↓40%</div>
                  <div className="text-sm text-slate-300">{t('roi.downtime')}</div>
                </div>
                <div className="p-4 bg-primary-500/10 border border-primary-500/20 rounded-lg text-center">
                  <div className="text-3xl font-bold text-primary-400 mb-2">↑15%</div>
                  <div className="text-sm text-slate-300">{t('roi.efficiency')}</div>
                </div>
                <div className="p-4 bg-green-500/10 border border-green-500/20 rounded-lg text-center">
                  <div className="text-3xl font-bold text-green-400 mb-2">↓25%</div>
                  <div className="text-sm text-slate-300">{t('roi.maintenanceCost')}</div>
                </div>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}