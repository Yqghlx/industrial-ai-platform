import React, { useState, useEffect, useCallback } from 'react';
import { useI18n } from '../i18n';
import { performanceMonitor, PerformanceMetrics, ComponentPerformance } from '../lib/performance';

interface PerformancePanelProps {
  isOpen: boolean;
  onClose: () => void;
}

/**
 * 性能监控面板
 * 用于开发环境查看性能指标
 */
export function PerformancePanel({ isOpen, onClose }: PerformancePanelProps) {
  const { t } = useI18n();
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null);
  const [componentMetrics, setComponentMetrics] = useState<ComponentPerformance[]>([]);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    if (!isOpen) return;

    const unsubscribe = performanceMonitor.onMetrics(setMetrics);
    setComponentMetrics(performanceMonitor.getComponentMetrics());

    return unsubscribe;
  }, [isOpen, refreshKey]);

  const refresh = useCallback(() => {
    setRefreshKey(k => k + 1);
  }, []);

  if (!isOpen) return null;

  // 性能评分
  const getScore = (value: number, thresholds: { good: number; needsWork: number }) => {
    if (value <= thresholds.good) return { label: t('performance.scoreGood'), color: 'text-green-400' };
    if (value <= thresholds.needsWork) return { label: t('performance.scoreNeedsWork'), color: 'text-yellow-400' };
    return { label: t('performance.scorePoor'), color: 'text-red-400' };
  };

  const lcpScore = metrics?.lcp ? getScore(metrics.lcp, { good: 2500, needsWork: 4000 }) : null;
  const fidScore = metrics?.fid ? getScore(metrics.fid, { good: 100, needsWork: 300 }) : null;
  const clsScore = metrics?.cls ? getScore(metrics.cls, { good: 0.1, needsWork: 0.25 }) : null;
  const fcpScore = metrics?.fcp ? getScore(metrics.fcp, { good: 1800, needsWork: 3000 }) : null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
      <div 
        className="bg-slate-800 rounded-lg shadow-xl w-full max-w-4xl max-h-[80vh] overflow-hidden"
        onClick={e => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-slate-700">
          <h2 className="text-lg font-semibold text-slate-100">{t('performance.panelTitle')}</h2>
          <div className="flex items-center gap-2">
            <button
              onClick={refresh}
              className="px-3 py-1 text-sm bg-slate-700 rounded hover:bg-slate-600 text-slate-300"
              aria-label={t('performance.refresh')}
            >
              {t('performance.refresh')}
            </button>
            <button
              onClick={onClose}
              className="p-1 text-slate-400 hover:text-slate-200"
              aria-label={t('common.close')}
            >
              ✕
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="p-4 overflow-y-auto max-h-[calc(80vh-60px)]">
          {/* Core Web Vitals */}
          <div className="mb-6">
            <h3 className="text-md font-medium text-slate-200 mb-3">{t('performance.coreWebVitals')}</h3>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              {/* LCP */}
              <div className="bg-slate-700/50 rounded-lg p-4">
                <div className="text-sm text-slate-400 mb-1">{t('performance.lcp')}</div>
                <div className={`text-2xl font-bold ${lcpScore?.color || 'text-slate-300'}`}>
                  {metrics?.lcp ? `${(metrics.lcp / 1000).toFixed(2)}s` : 'N/A'}
                </div>
                <div className="text-xs text-slate-500 mt-1">
                  {lcpScore?.label || '--'}
                </div>
              </div>

              {/* FID */}
              <div className="bg-slate-700/50 rounded-lg p-4">
                <div className="text-sm text-slate-400 mb-1">{t('performance.fid')}</div>
                <div className={`text-2xl font-bold ${fidScore?.color || 'text-slate-300'}`}>
                  {metrics?.fid ? `${metrics.fid.toFixed(0)}ms` : 'N/A'}
                </div>
                <div className="text-xs text-slate-500 mt-1">
                  {fidScore?.label || '--'}
                </div>
              </div>

              {/* CLS */}
              <div className="bg-slate-700/50 rounded-lg p-4">
                <div className="text-sm text-slate-400 mb-1">{t('performance.cls')}</div>
                <div className={`text-2xl font-bold ${clsScore?.color || 'text-slate-300'}`}>
                  {metrics?.cls ? metrics.cls.toFixed(3) : 'N/A'}
                </div>
                <div className="text-xs text-slate-500 mt-1">
                  {clsScore?.label || '--'}
                </div>
              </div>

              {/* FCP */}
              <div className="bg-slate-700/50 rounded-lg p-4">
                <div className="text-sm text-slate-400 mb-1">{t('performance.fcp')}</div>
                <div className={`text-2xl font-bold ${fcpScore?.color || 'text-slate-300'}`}>
                  {metrics?.fcp ? `${(metrics.fcp / 1000).toFixed(2)}s` : 'N/A'}
                </div>
                <div className="text-xs text-slate-500 mt-1">
                  {fcpScore?.label || '--'}
                </div>
              </div>
            </div>
          </div>

          {/* Navigation Timing */}
          <div className="mb-6">
            <h3 className="text-md font-medium text-slate-200 mb-3">{t('performance.navigationTiming')}</h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="bg-slate-700/50 rounded-lg p-3">
                <div className="text-sm text-slate-400">{t('performance.ttfb')}</div>
                <div className="text-lg font-medium text-slate-200">
                  {metrics?.ttfb ? `${metrics.ttfb.toFixed(0)}ms` : 'N/A'}
                </div>
              </div>
              <div className="bg-slate-700/50 rounded-lg p-3">
                <div className="text-sm text-slate-400">{t('performance.domContentLoaded')}</div>
                <div className="text-lg font-medium text-slate-200">
                  {metrics?.domContentLoaded ? `${metrics.domContentLoaded.toFixed(0)}ms` : 'N/A'}
                </div>
              </div>
              <div className="bg-slate-700/50 rounded-lg p-3">
                <div className="text-sm text-slate-400">{t('performance.loadComplete')}</div>
                <div className="text-lg font-medium text-slate-200">
                  {metrics?.loadComplete ? `${metrics.loadComplete.toFixed(0)}ms` : 'N/A'}
                </div>
              </div>
              <div className="bg-slate-700/50 rounded-lg p-3">
                <div className="text-sm text-slate-400">{t('performance.tti')}</div>
                <div className="text-lg font-medium text-slate-200">
                  {metrics?.tti ? `${metrics.tti.toFixed(0)}ms` : 'N/A'}
                </div>
              </div>
            </div>
          </div>

          {/* Memory Usage */}
          {metrics?.jsHeapSize && (
            <div className="mb-6">
              <h3 className="text-md font-medium text-slate-200 mb-3">{t('performance.memoryUsage')}</h3>
              <div className="bg-slate-700/50 rounded-lg p-3">
                <div className="text-sm text-slate-400">{t('performance.jsHeapSize')}</div>
                <div className="text-lg font-medium text-slate-200">
                  {(metrics.jsHeapSize / 1024 / 1024).toFixed(2)} MB
                </div>
              </div>
            </div>
          )}

          {/* Component Performance */}
          {componentMetrics.length > 0 && (
            <div>
              <h3 className="text-md font-medium text-slate-200 mb-3">{t('performance.componentPerformance')}</h3>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-left text-slate-400 border-b border-slate-700">
                      <th className="pb-2 pr-4">{t('performance.component')}</th>
                      <th className="pb-2 pr-4">{t('performance.mountTime')}</th>
                      <th className="pb-2 pr-4">{t('performance.updateTime')}</th>
                      <th className="pb-2 pr-4">{t('performance.renderCount')}</th>
                    </tr>
                  </thead>
                  <tbody className="text-slate-300">
                    {componentMetrics.map((comp) => (
                      <tr key={comp.name} className="border-b border-slate-700/50">
                        <td className="py-2 pr-4">{comp.name}</td>
                        <td className="py-2 pr-4">
                          <span className={comp.mountTime > 16 ? 'text-yellow-400' : ''}>
                            {comp.mountTime.toFixed(2)}ms
                          </span>
                        </td>
                        <td className="py-2 pr-4">
                          {comp.updateTime ? (
                            <span className={comp.updateTime > 16 ? 'text-yellow-400' : ''}>
                              {comp.updateTime.toFixed(2)}ms
                            </span>
                          ) : '-'}
                        </td>
                        <td className="py-2 pr-4">
                          <span className={comp.renderCount > 10 ? 'text-yellow-400' : ''}>
                            {comp.renderCount}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

/**
 * 性能监控按钮
 * 用于触发性能面板
 */
export function PerformanceButton() {
  const { t } = useI18n();
  const [isOpen, setIsOpen] = useState(false);

  // 仅在开发环境显示
  if (!import.meta.env.DEV) return null;

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className="fixed bottom-4 right-4 p-3 bg-slate-700 rounded-full shadow-lg hover:bg-slate-600 transition-colors z-40"
        title={t('performance.monitorButton')}
      >
        <svg 
          className="w-5 h-5 text-slate-300" 
          fill="none" 
          viewBox="0 0 24 24" 
          stroke="currentColor"
        >
          <path 
            strokeLinecap="round" 
            strokeLinejoin="round" 
            strokeWidth={2} 
            d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" 
          />
        </svg>
      </button>
      <PerformancePanel isOpen={isOpen} onClose={() => setIsOpen(false)} />
    </>
  );
}

export default PerformancePanel;