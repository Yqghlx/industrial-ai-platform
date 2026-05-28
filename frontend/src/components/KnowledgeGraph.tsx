import React, { useState, useEffect, useRef, useCallback } from 'react';
import api from '../lib/api';
import { useI18n } from '../i18n';
import Skeleton from './Skeleton';
import { useToast } from './Toast';
import { Network, RefreshCw } from 'lucide-react';
import { DeviceGraph } from '../types/api';
import { asDeviceGraphSafe } from '../types/typeGuards';

export default function KnowledgeGraph() {
  const { t } = useI18n();
  const { showToast } = useToast();
  const [graphData, setGraphData] = useState<DeviceGraph | null>(null);
  const [loading, setLoading] = useState(true);
  const containerRef = useRef<HTMLDivElement>(null);

  // Stable loadGraph function for useEffect dependencies
  const loadGraph = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getDeviceGraph();
      // FE-P1-01: 使用类型守卫安全转换
      setGraphData(asDeviceGraphSafe(res));
    } catch (error) {
      // FIX-023: 使用统一 showError toast 服务
      showToast({ type: 'error', message: t('errors.unknown') });
    } finally {
      setLoading(false);
    }
  }, [showToast, t]);

  // Initial load - use ref to ensure only runs once on mount
  const isMountedRef = useRef(false);
  useEffect(() => {
    if (!isMountedRef.current) {
      isMountedRef.current = true;
      loadGraph();
    }
  }, [loadGraph]);

  // Simple graph visualization using canvas
  useEffect(() => {
    if (!graphData || !containerRef.current) return;

    const canvas = document.createElement('canvas');
    canvas.width = containerRef.current.clientWidth;
    canvas.height = 500;
    // MINOR-01: 使用 textContent 替代 innerHTML 清空容器，更安全
    containerRef.current.textContent = '';
    containerRef.current.appendChild(canvas);

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Calculate positions (simple grid layout)
    const nodePositions: Map<string, { x: number; y: number }> = new Map();
    const cols = Math.ceil(Math.sqrt(graphData.nodes.length));
    const cellWidth = canvas.width / cols;
    const cellHeight = canvas.height / cols;

    graphData.nodes.forEach((node, i) => {
      const row = Math.floor(i / cols);
      const col = i % cols;
      nodePositions.set(node.id, {
        x: col * cellWidth + cellWidth / 2,
        y: row * cellHeight + cellHeight / 2,
      });
    });

    // Draw links
    ctx.strokeStyle = '#475569';
    ctx.lineWidth = 2;
    graphData.links.forEach((link) => {
      const source = nodePositions.get(link.source);
      const target = nodePositions.get(link.target);
      if (source && target) {
        ctx.beginPath();
        ctx.moveTo(source.x, source.y);
        ctx.lineTo(target.x, target.y);
        ctx.stroke();
      }
    });

    // Draw nodes
    graphData.nodes.forEach((node) => {
      const pos = nodePositions.get(node.id);
      if (!pos) return;

      // Node circle
      ctx.beginPath();
      ctx.arc(pos.x, pos.y, 30, 0, 2 * Math.PI);
      ctx.fillStyle = node.status === 'online' ? '#10b981' :
                     node.status === 'warning' ? '#f59e0b' :
                     node.status === 'fault' ? '#ef4444' :
                     '#64748b';
      ctx.fill();

      // Node label
      ctx.fillStyle = '#f1f5f9';
      ctx.font = '12px Inter';
      ctx.textAlign = 'center';
      ctx.fillText(node.name.substring(0, 10), pos.x, pos.y + 50);
      ctx.fillStyle = '#94a3b8';
      ctx.font = '10px Inter';
      ctx.fillText(node.id, pos.x, pos.y + 65);
    });
  }, [graphData]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">{t('nav.knowledgeGraph')}</h1>
          <p className="text-slate-400">设备拓扑关系图</p>
        </div>
        <button 
          onClick={loadGraph}
          className="btn btn-secondary flex items-center gap-2"
          aria-label={t('common.refresh')}
        >
          <RefreshCw className="w-5 h-5" />
          <span>{t('common.refresh')}</span>
        </button>
      </div>

      {/* Graph container */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <Network className="w-5 h-5 text-primary-500" />
            <span className="font-medium text-slate-100">设备关系网络</span>
          </div>
        </div>
        <div className="card-body">
          {loading ? (
            <Skeleton variant="chart" height={500} />
          ) : (
            <div 
              ref={containerRef}
              className="w-full h-[500px] bg-slate-800/50 rounded-lg"
            />
          )}
        </div>
      </div>

      {/* Legend */}
      <div className="card">
        <div className="card-body">
          <div className="flex items-center justify-center gap-8">
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded-full bg-green-500" />
              <span className="text-sm text-slate-300">{t('device.online')}</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded-full bg-yellow-500" />
              <span className="text-sm text-slate-300">{t('device.warning')}</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded-full bg-red-500" />
              <span className="text-sm text-slate-300">{t('device.fault')}</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded-full bg-slate-500" />
              <span className="text-sm text-slate-300">{t('device.offline')}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}