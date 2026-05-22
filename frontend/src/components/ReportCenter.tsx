import React, { useState, useEffect, useMemo } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import ExportButton from './ExportButton';
import { FileText, Download, Plus } from 'lucide-react';
import { Report } from '../types/api';

export default function ReportCenter() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [reports, setReports] = useState<Report[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedReport, setSelectedReport] = useState<Report | null>(null);
  const [showGenerateModal, setShowGenerateModal] = useState(false);

  useEffect(() => {
    loadReports();
  }, []);

  const loadReports = async () => {
    setLoading(true);
    try {
      const res = await api.getReports();
      setReports((res.data as Report[]) || []);
    } catch (error) {
      console.error('Failed to load reports:', error);
      showToast({ type: 'error', message: t('errors.loadFailedReports') });
    } finally {
      setLoading(false);
    }
  };

  const handleGenerate = async (type: string, deviceId?: string) => {
    try {
      await api.generateReport(type, deviceId);
      showToast({ type: 'success', message: '报告已生成' });
      setShowGenerateModal(false);
      loadReports();
    } catch (error) {
      showToast({ type: 'error', message: '生成失败' });
    }
  };

  // FE-P2-08: 使用完整类名映射替代动态类名，避免 Tailwind purge 问题
  const getReportTypeButtonClass = (color: string): string => {
    const classMap: Record<string, string> = {
      blue: 'w-full p-4 bg-blue-500/20 border border-blue-500/30 rounded-lg text-left hover:bg-blue-500/30 transition-colors',
      green: 'w-full p-4 bg-green-500/20 border border-green-500/30 rounded-lg text-left hover:bg-green-500/30 transition-colors',
      orange: 'w-full p-4 bg-orange-500/20 border border-orange-500/30 rounded-lg text-left hover:bg-orange-500/30 transition-colors',
      red: 'w-full p-4 bg-red-500/20 border border-red-500/30 rounded-lg text-left hover:bg-red-500/30 transition-colors',
      purple: 'w-full p-4 bg-purple-500/20 border border-purple-500/30 rounded-lg text-left hover:bg-purple-500/30 transition-colors',
    };
    return classMap[color] || 'w-full p-4 bg-slate-500/20 border border-slate-500/30 rounded-lg text-left hover:bg-slate-500/30 transition-colors';
  };

  // FE-P2-06: 使用 useMemo 缓存报告类型配置
  const reportTypes = useMemo(() => [
    { type: 'daily', label: t('report.daily'), icon: '📊', color: 'blue' },
    { type: 'device', label: t('report.device'), icon: '⚙️', color: 'green' },
    { type: 'maintenance', label: t('report.maintenance'), icon: '🔧', color: 'orange' },
    { type: 'anomaly', label: t('report.anomaly'), icon: '⚠️', color: 'red' },
    { type: 'comprehensive', label: t('report.comprehensive'), icon: '📈', color: 'purple' },
  ], [t]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.reports')}</h1>
          <p className="text-slate-400">{t('report.title')}</p>
        </div>
        <div className="flex items-center gap-2">
          <ExportButton reportType="alerts" label="导出告警报告" />
          <button 
            onClick={() => setShowGenerateModal(true)}
            className="btn btn-primary flex items-center gap-2"
            aria-label={t('report.generate')}
          >
            <Plus className="w-5 h-5" />
            <span>{t('report.generate')}</span>
          </button>
        </div>
      </div>

      {/* Reports list */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <FileText className="w-5 h-5 text-primary-500" />
            <span className="font-medium text-slate-100">{t('report.list')}</span>
          </div>
        </div>
        <div className="card-body">
          {loading ? (
            <div className="space-y-4">
              {[1, 2, 3].map(i => <Skeleton key={i} variant="card" />)}
            </div>
          ) : (reports?.length ?? 0) === 0 ? (
            <div className="text-center py-8">
              <FileText className="w-12 h-12 text-slate-400 mx-auto mb-4" />
              <p className="text-slate-300">{t('report.noReports')}</p>
              <button
                onClick={() => setShowGenerateModal(true)}
                className="btn btn-primary mt-4"
                aria-label={t('report.generate')}
              >
                {t('report.generate')}
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {reports.map((r) => (
                <div
                  key={r.id}
                  className="p-4 bg-slate-800/50 rounded-lg border border-slate-700 hover:border-primary-500 transition-colors cursor-pointer"
                  onClick={() => setSelectedReport(r)}
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-3">
                      <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                        r.type === 'daily' ? 'bg-blue-500/20' :
                        r.type === 'device' ? 'bg-green-500/20' :
                        r.type === 'maintenance' ? 'bg-orange-500/20' :
                        r.type === 'anomaly' ? 'bg-red-500/20' :
                        'bg-slate-500/20'
                      }`}>
                        <FileText className={`w-5 h-5 ${
                          r.type === 'daily' ? 'text-blue-400' :
                          r.type === 'device' ? 'text-green-400' :
                          r.type === 'maintenance' ? 'text-orange-400' :
                          r.type === 'anomaly' ? 'text-red-400' :
                          'text-slate-400'
                        }`} />
                      </div>
                      <div>
                        <div className="font-medium text-slate-100">{r.title}</div>
                        <div className="text-sm text-slate-400">
                          {t(`report.${r.type}`)} · {new Date(r.generated_at).toLocaleDateString()}
                        </div>
                      </div>
                    </div>
                    <button className="p-2 text-slate-400 hover:text-slate-200"
                      aria-label={t('common.download')}
                    >
                      <Download className="w-5 h-5" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Generate Modal */}
      {showGenerateModal && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" role="dialog" aria-modal="true">
          <div className="card max-w-md">
            <div className="card-header">
              <h2 className="text-lg font-semibold">{t('report.generate')}</h2>
            </div>
            <div className="card-body">
              <div className="space-y-4">
                {reportTypes.map((item) => (
                  <button
                    key={item.type}
                    onClick={() => handleGenerate(item.type)}
                    className={getReportTypeButtonClass(item.color)}
                  >
                    <div className="flex items-center gap-3">
                      <span className="text-2xl">{item.icon}</span>
                      <span className="font-medium text-slate-100">{item.label}</span>
                    </div>
                  </button>
                ))}
                <button
                  onClick={() => setShowGenerateModal(false)}
                  className="btn btn-secondary w-full"
                  aria-label={t('common.cancel')}
                >
                  {t('common.cancel')}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Report Detail Modal */}
      {selectedReport && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" role="dialog" aria-modal="true">
          <div className="card max-w-2xl w-full mx-4 max-h-[80vh] overflow-hidden">
            <div className="card-header flex items-center justify-between">
              <h2 className="text-lg font-semibold">{selectedReport.title}</h2>
              <button
                onClick={() => setSelectedReport(null)}
                className="text-slate-400 hover:text-slate-200"
                aria-label={t('common.close')}
              >
                ×
              </button>
            </div>
            <div className="card-body overflow-y-auto scrollbar-thin">
              <div className="prose prose-invert max-w-none">
                <div className="whitespace-pre-wrap text-slate-200">
                  {selectedReport.content}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}