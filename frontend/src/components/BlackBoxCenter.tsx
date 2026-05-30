import React, { useState, useEffect } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import { useEscapeKey } from '../lib/hooks';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { Box, Play, Clock } from 'lucide-react';
import { BlackBoxRecord } from '../types/api';

// FIX-005: 类型守卫函数，验证 BlackBoxRecord 必要字段
function isBlackBoxRecord(obj: unknown): obj is BlackBoxRecord {
  if (!obj || typeof obj !== 'object') return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'number' &&
    typeof record.device_id === 'string' &&
    typeof record.trigger_type === 'string' &&
    typeof record.start_time === 'string' &&
    typeof record.end_time === 'string' &&
    typeof record.created_at === 'string'
  );
}

// FE-P0-02: 数组类型守卫，验证 BlackBoxRecord[] 响应
function isBlackBoxRecordArray(data: unknown): data is BlackBoxRecord[] {
  if (!Array.isArray(data)) return false;
  return data.every(isBlackBoxRecord);
}

export default function BlackBoxCenter() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [records, setRecords] = useState<BlackBoxRecord[]>([]);
  const [selectedRecord, setSelectedRecord] = useState<BlackBoxRecord | null>(null);
  const [loading, setLoading] = useState(true);

  // C1: Escape 键关闭模态框
  useEscapeKey(() => setSelectedRecord(null), !!selectedRecord);

  useEffect(() => {
    loadRecords();
  }, []);

  const loadRecords = async () => {
    setLoading(true);
    try {
      const res = await api.getBlackBoxRecords();
      // FE-P0-02: 使用类型守卫替代类型断言
      if (isBlackBoxRecordArray(res.data)) {
        setRecords(res.data);
      } else {
        console.error('Invalid BlackBoxRecord[] response:', res.data);
        setRecords([]);
        showToast({ type: 'error', message: t('errors.loadFailedBlackbox') });
      }
    } catch (error) {
      console.error('Failed to load black box records:', error);
      showToast({ type: 'error', message: t('errors.loadFailedBlackbox') });
    } finally {
      setLoading(false);
    }
  };

  const loadRecordData = async (id: number) => {
    try {
      const res = await api.getBlackBoxData(id);
      // FIX-005: 使用类型守卫替代双重断言
      if (isBlackBoxRecord(res)) {
        setSelectedRecord(res);
      } else {
        console.error('Invalid BlackBoxRecord response:', res);
        showToast({ type: 'error', message: t('errors.loadFailedRecordData') });
      }
    } catch (error) {
      console.error('Failed to load record data:', error);
      showToast({ type: 'error', message: t('errors.loadFailedRecordData') });
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.blackbox')}</h1>
          <p className="text-slate-400">{t('blackbox.title')}</p>
        </div>
      </div>

      {/* Records list */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <Box className="w-5 h-5 text-primary-500" />
            <span className="font-medium text-slate-100">{t('blackbox.records')}</span>
          </div>
        </div>
        <div className="card-body">
          {loading ? (
            <div className="space-y-4">
              {[1, 2, 3].map(i => <Skeleton key={i} variant="card" />)}
            </div>
          ) : records.length === 0 ? (
            <div className="text-center py-8">
              <Box className="w-12 h-12 text-slate-400 mx-auto mb-4" />
              <p className="text-slate-300">{t('blackbox.noRecords')}</p>
            </div>
          ) : (
            <div className="table-container">
              <table className="table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>{t('device.id')}</th>
                    <th>{t('blackbox.triggerType')}</th>
                    <th>{t('blackbox.startTime')}</th>
                    <th>{t('blackbox.endTime')}</th>
                    <th>{t('blackbox.playback')}</th>
                  </tr>
                </thead>
                <tbody>
                  {records.map((r) => (
                    <tr key={r.id}>
                      <td className="font-mono">#${r.id}</td>
                      <td className="font-mono">{r.device_id}</td>
                      <td>
                        <span className="status-badge bg-red-500/20 text-red-400">
                          {r.trigger_type}
                        </span>
                      </td>
                      <td>{new Date(r.start_time).toLocaleString()}</td>
                      <td>{new Date(r.end_time).toLocaleString()}</td>
                      <td>
                        <button
                          onClick={() => loadRecordData(r.id)}
                          className="btn btn-secondary flex items-center gap-2"
                          aria-label={t('blackbox.playback')}
                        >
                          <Play className="w-4 h-4" />
                          <span>{t('blackbox.playback')}</span>
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Detail modal */}
      {selectedRecord && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" role="dialog" aria-modal="true" onClick={(e) => { if (e.target === e.currentTarget) setSelectedRecord(null); }}>
          <div className="card max-w-2xl w-full mx-4">
            <div className="card-header flex items-center justify-between">
              <h2 className="text-lg font-semibold">{t('blackbox.playback')} #{selectedRecord.id}</h2>
              <button
                onClick={() => setSelectedRecord(null)}
                className="text-slate-400 hover:text-slate-200"
                aria-label={t('common.close')}
              >
                ×
              </button>
            </div>
            <div className="card-body space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <span className="text-slate-400">{t('device.id')}:</span>
                  <span className="text-slate-200 ml-2">{selectedRecord.device_id}</span>
                </div>
                <div>
                  <span className="text-slate-400">{t('blackbox.triggerType')}:</span>
                  <span className="text-red-400 ml-2">{selectedRecord.trigger_type}</span>
                </div>
                <div>
                  <span className="text-slate-400">{t('blackbox.startTime')}:</span>
                  <span className="text-slate-200 ml-2">
                    {new Date(selectedRecord.start_time).toLocaleString()}
                  </span>
                </div>
                <div>
                  <span className="text-slate-400">{t('blackbox.endTime')}:</span>
                  <span className="text-slate-200 ml-2">
                    {new Date(selectedRecord.end_time).toLocaleString()}
                  </span>
                </div>
              </div>

              <div className="p-4 bg-slate-800/50 rounded-lg">
                <div className="text-sm text-slate-400 mb-2">{t('blackbox.summary')}</div>
                <div className="text-slate-200">{selectedRecord.summary}</div>
              </div>

              {/* Timeline visualization */}
              <div className="p-4 bg-slate-800/50 rounded-lg">
                <div className="flex items-center gap-4 mb-4">
                  <Clock className="w-5 h-5 text-slate-400" />
                  <span className="text-slate-300">{t('blackbox.timeline')}</span>
                </div>
                <div className="relative h-8 bg-slate-700 rounded-lg">
                  <div 
                    className="absolute left-0 h-full bg-red-500/50 rounded-lg"
                    style={{
                      width: '100%',
                    }}
                  />
                  <div className="absolute left-0 top-1/2 -translate-y-1/2 w-2 h-2 bg-green-500 rounded-full" />
                  <div className="absolute right-0 top-1/2 -translate-y-1/2 w-2 h-2 bg-red-500 rounded-full" />
                </div>
                <div className="flex items-center justify-between mt-2 text-xs text-slate-400">
                  <span>{new Date(selectedRecord.start_time).toLocaleTimeString()}</span>
                  <span>{t('blackbox.faultOccurred')}</span>
                  <span>{new Date(selectedRecord.end_time).toLocaleTimeString()}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}