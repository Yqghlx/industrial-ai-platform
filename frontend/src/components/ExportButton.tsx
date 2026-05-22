import React, { useState } from 'react';
import { Download, FileText, FileSpreadsheet, Loader2 } from 'lucide-react';
import { useToast } from './Toast';
import api from '../lib/api';
import { useI18n } from '../i18n';

interface ExportButtonProps {
  reportType: 'devices' | 'alerts' | 'roi';
  className?: string;
  startDate?: string;
  endDate?: string;
  label?: string;
}

export default function ExportButton({
  reportType,
  className = '',
  startDate,
  endDate,
  label,
}: ExportButtonProps) {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [showModal, setShowModal] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [selectedFormat, setSelectedFormat] = useState<'pdf' | 'xlsx'>('pdf');

  const handleExport = async (format: 'pdf' | 'xlsx') => {
    setExporting(true);
    setSelectedFormat(format);

    try {
      const result = await api.exportReport(reportType, format, startDate, endDate);
      
      // Create download link
      const blob = new Blob([result.data], { type: result.mimeType });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = result.filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      showToast({ type: 'success', message: t('export.exportSuccess') });
      setShowModal(false);
    } catch (error) {
      console.error('Export failed:', error);
      showToast({ type: 'error', message: t('export.exportFailed') });
    } finally {
      setExporting(false);
    }
  };

  const getReportLabel = () => {
    switch (reportType) {
      case 'devices':
        return t('export.deviceReport');
      case 'alerts':
        return t('export.alertReport');
      case 'roi':
        return t('export.roiReport');
      default:
        return t('export.exportReport');
    }
  };

  return (
    <>
      <button
        onClick={() => setShowModal(true)}
        className={`btn btn-secondary flex items-center gap-2 ${className}`}
        title={getReportLabel()}
        aria-label={t('common.export')}
      >
        <Download className="w-5 h-5" />
        <span>{label || t('common.export')}</span>
      </button>

      {/* Export Format Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" role="dialog" aria-modal="true">
          <div className="card max-w-md">
            <div className="card-header">
              <h2 className="text-lg font-semibold text-slate-100">
                {t('export.selectExportFormat')}
              </h2>
            </div>
            <div className="card-body">
              <p className="text-slate-400 mb-4">
                {getReportLabel()}
              </p>
              
              <div className="space-y-3">
                {/* PDF Option */}
                <button
                  onClick={() => handleExport('pdf')}
                  disabled={exporting}
                  className={`w-full p-4 rounded-lg border transition-colors flex items-center gap-4 ${
                    selectedFormat === 'pdf' && exporting
                      ? 'bg-blue-500/20 border-blue-500'
                      : 'bg-slate-800/50 border-slate-700 hover:border-blue-500'
                  }`}
                >
                  {exporting && selectedFormat === 'pdf' ? (
                    <Loader2 className="w-6 h-6 text-blue-400 animate-spin" />
                  ) : (
                    <FileText className="w-6 h-6 text-blue-400" />
                  )}
                  <div className="flex-1">
                    <div className="font-medium text-slate-100">{t('export.pdfFormat')}</div>
                    <div className="text-sm text-slate-400">
                      {t('export.pdfDescription')}
                    </div>
                  </div>
                </button>

                {/* Excel Option */}
                <button
                  onClick={() => handleExport('xlsx')}
                  disabled={exporting}
                  className={`w-full p-4 rounded-lg border transition-colors flex items-center gap-4 ${
                    selectedFormat === 'xlsx' && exporting
                      ? 'bg-green-500/20 border-green-500'
                      : 'bg-slate-800/50 border-slate-700 hover:border-green-500'
                  }`}
                >
                  {exporting && selectedFormat === 'xlsx' ? (
                    <Loader2 className="w-6 h-6 text-green-400 animate-spin" />
                  ) : (
                    <FileSpreadsheet className="w-6 h-6 text-green-400" />
                  )}
                  <div className="flex-1">
                    <div className="font-medium text-slate-100">{t('export.xlsxFormat')}</div>
                    <div className="text-sm text-slate-400">
                      {t('export.xlsxDescription')}
                    </div>
                  </div>
                </button>
              </div>

              {/* Cancel Button */}
              <button
                onClick={() => setShowModal(false)}
                disabled={exporting}
                className="btn btn-secondary w-full mt-4"
                aria-label={t('common.cancel')}
              >
                {t('common.cancel')}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}