import React, { useState, useEffect, useCallback } from 'react';
import { useI18n } from '../i18n';
import { useToast } from './Toast';
import Skeleton from './Skeleton';
import { api } from '../lib/api';
import {
  TrendingUp,
  BarChart3,
  Clock,
  CheckCircle,
  AlertTriangle,
  Download,
  RefreshCw,
} from 'lucide-react';

interface TrendData {
  date: string;
  total: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
  resolved: number;
}

interface DeviceRank {
  device_id: string;
  total: number;
  critical: number;
  active: number;
}

interface Efficiency {
  total_alerts: number;
  resolved_count: number;
  acknowledged_count: number;
  active_count: number;
  resolution_rate: number;
  acknowledgement_rate: number;
  avg_response_min: number;
  avg_resolution_min: number;
  avg_response_str: string;
  avg_resolution_str: string;
}

export default function AlertReportPage() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [trendData, setTrendData] = useState<TrendData[]>([]);
  const [rankingData, setRankingData] = useState<DeviceRank[]>([]);
  const [efficiencyData, setEfficiencyData] = useState<Efficiency | null>(null);
  const [loading, setLoading] = useState(true);
  const [trendDays, setTrendDays] = useState(7);

  // Fetch all report data
  // FE-P3-03: Fixed token storage inconsistency - use 'token' key consistently
  // SEC-LOW-02: Use sessionStorage for better security
  const fetchReports = useCallback(async () => {
    setLoading(true);
    try {
      const token = sessionStorage.getItem('token');
      if (!token) return;

      // Fetch trend data
      try {
        const trendJson = await api.getTrendReport(trendDays);
        setTrendData((trendJson.data || []) as TrendData[]);
      } catch { /* 降级处理 */ }

      // Fetch ranking data
      try {
        const rankJson = await api.getRankingReport(10);
        setRankingData((rankJson.data || []) as DeviceRank[]);
      } catch { /* 降级处理 */ }

      // Fetch efficiency data
      try {
        const effJson = await api.getEfficiencyReport();
        setEfficiencyData(effJson as unknown as Efficiency);
      } catch { /* 降级处理 */ }
    } catch {
      showToast({ type: 'error', message: t('errors.unknown') });
    } finally {
      setLoading(false);
    }
  }, [trendDays, showToast, t]);

  useEffect(() => {
    fetchReports();
  }, [fetchReports]);

  // Export CSV
  const handleExport = () => {
    if (trendData.length === 0) return;

    // Create CSV content
    const headers = [
      t('alertReport.date'),
      t('alertReport.total'),
      t('alertReport.critical'),
      t('alertReport.high'),
      t('alertReport.medium'),
      t('alertReport.low'),
      t('alertReport.resolved'),
    ];
    const rows = trendData.map(d => [
      d.date,
      d.total,
      d.critical,
      d.high,
      d.medium,
      d.low,
      d.resolved,
    ]);

    const csvContent = [
      headers.join(','),
      ...rows.map(r => r.join(',')),
    ].join('\n');

    // Download
    const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = `alert_report_${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
    URL.revokeObjectURL(link.href);

    showToast({ type: 'success', message: t('alertReport.exportSuccess') });
  };

  // Simple bar chart component
  const SimpleBarChart = ({ data, label: _label }: { data: TrendData[]; label: string }) => {
    if (data.length === 0) return null;

    const maxValue = Math.max(...data.map(d => d.total), 1);
    const barWidth = 100 / data.length;

    return (
      <div className="relative h-40 bg-slate-700/50 rounded-lg p-4">
        {/* Bars */}
        <div className="absolute bottom-8 left-4 right-4 flex items-end gap-1 h-24">
          {data.map((d, i) => (
            <div
              key={i}
              className="flex-1 bg-gradient-to-t from-blue-600 to-blue-400 rounded-t"
              style={{ height: `${(d.total / maxValue) * 100}%` }}
              title={`${d.date}: ${d.total}`}
            />
          ))}
        </div>

        {/* X-axis labels */}
        <div className="absolute bottom-0 left-4 right-4 flex justify-between text-xs text-slate-400">
          {data.filter((_, i) => i % 2 === 0).map((d, i) => (
            <span key={i} style={{ width: `${barWidth * 2}%` }}>
              {d.date.slice(5)}
            </span>
          ))}
        </div>

        {/* Y-axis */}
        <div className="absolute top-4 left-0 text-xs text-slate-400">
          {maxValue}
        </div>
      </div>
    );
  };

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <BarChart3 className="h-6 w-6 text-blue-400" />
          <h1 className="text-2xl font-bold text-slate-100">{t('alertReport.title')}</h1>
        </div>
        <div className="flex gap-2">
          <select
            className="px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200"
            value={trendDays}
            onChange={(e) => setTrendDays(Number(e.target.value))}
          >
            <option value="7">{t('alertReport.recent7Days')}</option>
            <option value="14">{t('alertReport.recent14Days')}</option>
            <option value="30">{t('alertReport.recent30Days')}</option>
          </select>
          <button
            className="flex items-center gap-2 px-4 py-2 bg-slate-700 hover:bg-slate-600 rounded-lg text-slate-200 transition-colors"
            onClick={fetchReports}
            disabled={loading}
            aria-label={t('alertReport.refresh')}
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            {t('alertReport.refresh')}
          </button>
          <button
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-white transition-colors"
            onClick={handleExport}
            aria-label={t('alertReport.exportCsv')}
          >
            <Download className="h-4 w-4" />
            {t('alertReport.exportCsv')}
          </button>
        </div>
      </div>

      {/* Efficiency Metrics */}
      {efficiencyData && (
        <div className="grid gap-4 md:grid-cols-4">
          <div className="p-4 bg-slate-800 rounded-lg border border-slate-700">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-slate-400">{t('alertReport.totalAlerts')}</span>
              <AlertTriangle className="h-4 w-4 text-orange-400" />
            </div>
            <div className="text-2xl font-bold text-slate-100">{efficiencyData.total_alerts}</div>
            <div className="text-xs text-slate-400 mt-1">
              {t('alertReport.active')}: {efficiencyData.active_count} | {t('alertReport.acknowledged')}: {efficiencyData.acknowledged_count}
            </div>
          </div>

          <div className="p-4 bg-slate-800 rounded-lg border border-slate-700">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-slate-400">{t('alertReport.resolutionRate')}</span>
              <CheckCircle className="h-4 w-4 text-green-400" />
            </div>
            <div className="text-2xl font-bold text-green-400">
              {efficiencyData.resolution_rate.toFixed(1)}%
            </div>
            <div className="text-xs text-slate-400 mt-1">
              {t('alertReport.resolved')}: {efficiencyData.resolved_count}
            </div>
          </div>

          <div className="p-4 bg-slate-800 rounded-lg border border-slate-700">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-slate-400">{t('alertReport.avgResponseTime')}</span>
              <Clock className="h-4 w-4 text-yellow-400" />
            </div>
            <div className="text-2xl font-bold text-slate-100">
              {efficiencyData.avg_response_str}
            </div>
            <div className="text-xs text-slate-400 mt-1">
              {efficiencyData.avg_response_min.toFixed(1)} {t('alertReport.minutes')}
            </div>
          </div>

          <div className="p-4 bg-slate-800 rounded-lg border border-slate-700">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-slate-400">{t('alertReport.avgResolutionTime')}</span>
              <TrendingUp className="h-4 w-4 text-blue-400" />
            </div>
            <div className="text-2xl font-bold text-slate-100">
              {efficiencyData.avg_resolution_str}
            </div>
            <div className="text-xs text-slate-400 mt-1">
              {efficiencyData.avg_resolution_min.toFixed(1)} {t('alertReport.minutes')}
            </div>
          </div>
        </div>
      )}

      {/* Trend Chart */}
      <div className="p-4 bg-slate-800 rounded-lg border border-slate-700">
        <h2 className="text-lg font-semibold text-slate-100 mb-4 flex items-center gap-2">
          <TrendingUp className="h-5 w-5 text-blue-400" />
          {t('alertReport.alertTrend')}（{t('alertReport.recentDays', { days: trendDays })}）
        </h2>

        {loading ? (
          <Skeleton className="h-40 w-full" />
        ) : trendData.length > 0 ? (
          <>
            <SimpleBarChart data={trendData} label={t('alertReport.alertCount')} />

            {/* Trend table */}
            <div className="mt-4 overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-slate-400 border-b border-slate-600">
                    <th className="px-3 py-2 text-left">{t('alertReport.date')}</th>
                    <th className="px-3 py-2">{t('alertReport.total')}</th>
                    <th className="px-3 py-2 text-red-400">{t('alertReport.critical')}</th>
                    <th className="px-3 py-2 text-orange-400">{t('alertReport.high')}</th>
                    <th className="px-3 py-2 text-yellow-400">{t('alertReport.medium')}</th>
                    <th className="px-3 py-2 text-green-400">{t('alertReport.low')}</th>
                    <th className="px-3 py-2 text-blue-400">{t('alertReport.resolved')}</th>
                  </tr>
                </thead>
                <tbody>
                  {trendData.slice(-5).map((d, i) => (
                    <tr key={i} className="text-slate-200 border-b border-slate-700">
                      <td className="px-3 py-2">{d.date}</td>
                      <td className="px-3 py-2 font-medium">{d.total}</td>
                      <td className="px-3 py-2 text-red-400">{d.critical}</td>
                      <td className="px-3 py-2 text-orange-400">{d.high}</td>
                      <td className="px-3 py-2 text-yellow-400">{d.medium}</td>
                      <td className="px-3 py-2 text-green-400">{d.low}</td>
                      <td className="px-3 py-2 text-blue-400">{d.resolved}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </>
        ) : (
          <div className="text-center py-8 text-slate-400">{t('alertReport.noData')}</div>
        )}
      </div>

      {/* Device Ranking */}
      <div className="p-4 bg-slate-800 rounded-lg border border-slate-700">
        <h2 className="text-lg font-semibold text-slate-100 mb-4 flex items-center gap-2">
          <BarChart3 className="h-5 w-5 text-orange-400" />
          {t('alertReport.deviceRanking')}（{t('alertReport.top10')}）
        </h2>

        {loading ? (
          <Skeleton className="h-32 w-full" />
        ) : rankingData.length > 0 ? (
          <div className="space-y-3">
            {rankingData.map((r, i) => (
              <div
                key={r.device_id}
                className="flex items-center gap-4 p-3 bg-slate-700/50 rounded-lg"
              >
                {/* Rank */}
                <div className={`w-8 h-8 rounded-full flex items-center justify-center font-bold ${
                  i === 0 ? 'bg-yellow-500 text-slate-900' :
                  i === 1 ? 'bg-slate-400 text-slate-900' :
                  i === 2 ? 'bg-orange-600 text-white' :
                  'bg-slate-600 text-slate-200'
                }`}>
                  {i + 1}
                </div>

                {/* Device ID */}
                <div className="flex-1">
                  <span className="text-slate-100 font-medium">{r.device_id}</span>
                </div>

                {/* Stats */}
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-slate-200">
                    <span className="text-slate-400">{t('alertReport.totalLabel')}</span> {r.total}
                  </span>
                  <span className="text-red-400">
                    <span className="text-slate-400">{t('alertReport.criticalLabel')}</span> {r.critical}
                  </span>
                  <span className="text-orange-400">
                    <span className="text-slate-400">{t('alertReport.activeLabel')}</span> {r.active}
                  </span>
                </div>

                {/* Progress bar */}
                <div className="w-20 h-2 bg-slate-600 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-orange-500 to-red-500"
                    style={{ width: `${(r.total / rankingData[0].total) * 100}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-8 text-slate-400">{t('alertReport.noData')}</div>
        )}
      </div>
    </div>
  );
}