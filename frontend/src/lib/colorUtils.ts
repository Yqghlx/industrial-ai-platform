/**
 * Color Mapping Utilities for Industrial AI Platform
 * Centralized color mapping functions for consistent styling
 */

import { DeviceStatus, WorkOrderStatus, WorkOrderPriority, AlertSeverity } from '../types/api';

/**
 * Status color for device status indicators (solid background colors)
 */
export function getDeviceStatusColor(status: DeviceStatus | string): string {
  switch (status) {
    case 'online': return 'bg-green-500';
    case 'warning': return 'bg-yellow-500';
    case 'fault': return 'bg-red-500';
    default: return 'bg-slate-500';
  }
}

/**
 * Status badge CSS class for device status badges
 */
export function getDeviceStatusBadgeClass(status: DeviceStatus | string): string {
  switch (status) {
    case 'online': return 'status-online';
    case 'warning': return 'status-warning';
    case 'fault': return 'status-fault';
    default: return 'status-offline';
  }
}

/**
 * Status text color for device status (text colors only)
 */
export function getDeviceStatusTextColor(status: DeviceStatus | string): string {
  switch (status) {
    case 'online': return 'text-green-400';
    case 'warning': return 'text-yellow-400';
    case 'fault': return 'text-red-400';
    default: return 'text-slate-400';
  }
}

/**
 * Work order status color (with background opacity)
 */
export function getWorkOrderStatusColor(status: WorkOrderStatus | string): string {
  switch (status) {
    case 'pending': return 'bg-slate-500/20 text-slate-400';
    case 'in_progress': return 'bg-primary-500/20 text-primary-400';
    case 'completed': return 'bg-green-500/20 text-green-400';
    case 'cancelled': return 'bg-red-500/20 text-red-400';
    default: return 'bg-slate-500/20 text-slate-400';
  }
}

/**
 * Work order priority color (with background opacity)
 */
export function getWorkOrderPriorityColor(priority: WorkOrderPriority | string): string {
  switch (priority) {
    case 'urgent': return 'bg-red-500/20 text-red-400';
    case 'high': return 'bg-orange-500/20 text-orange-400';
    case 'medium': return 'bg-yellow-500/20 text-yellow-400';
    case 'low': return 'bg-green-500/20 text-green-400';
    default: return 'bg-slate-500/20 text-slate-400';
  }
}

/**
 * Alert severity color
 */
export function getAlertSeverityColor(severity: AlertSeverity | string): string {
  switch (severity) {
    case 'critical': return 'bg-red-500/20 text-red-400';
    case 'high': return 'bg-orange-500/20 text-orange-400';
    case 'medium': return 'bg-yellow-500/20 text-yellow-400';
    case 'low': return 'bg-green-500/20 text-green-400';
    default: return 'bg-slate-500/20 text-slate-400';
  }
}

/**
 * Agent name color for AI team dashboard
 * Note: Agent names may be localized, so we use the agent key
 */
export function getAgentColor(agent: string): string {
  // Check for localized names or agent keys
  if (agent.includes('设备') || agent === 'device_expert') return 'text-blue-400';
  if (agent.includes('维护') || agent === 'maintenance_expert') return 'text-green-400';
  if (agent.includes('预测') || agent === 'prediction_expert') return 'text-purple-400';
  if (agent.includes('优化') || agent === 'optimization_expert') return 'text-orange-400';
  return 'text-slate-400';
}

/**
 * Gauge color for telemetry metrics
 * Returns text color based on value thresholds
 */
export function getGaugeColor(
  value: number,
  type: 'temperature' | 'vibration' | 'pressure' | 'power'
): string {
  const thresholds = {
    temperature: { normal: 80, warning: 100, critical: 120 },
    vibration: { normal: 2, warning: 3, critical: 5 },
    pressure: { normal: 100, warning: 150, critical: 200 },
    power: { normal: 50, warning: 80, critical: 100 },
  };
  const t = thresholds[type];
  if (value >= t.critical) return 'text-red-400';
  if (value >= t.warning) return 'text-yellow-400';
  return 'text-green-400';
}

/**
 * Gauge SVG stroke color for telemetry metrics
 * Returns hex color code for SVG elements
 */
export function getGaugeStrokeColor(
  value: number,
  type: 'temperature' | 'vibration' | 'pressure'
): string {
  const thresholds = {
    temperature: { normal: 80, warning: 100, critical: 120 },
    vibration: { normal: 2, warning: 3, critical: 5 },
    pressure: { normal: 100, warning: 150, critical: 200 },
  };
  const t = thresholds[type];
  if (value >= t.critical) return '#ef4444'; // red
  if (value >= t.warning) return '#f59e0b'; // yellow/amber
  return '#10b981'; // green
}

/**
 * Gauge percentage for visualization
 */
export function getGaugePercentage(
  value: number,
  type: 'temperature' | 'vibration' | 'pressure'
): number {
  const max = { temperature: 150, vibration: 6, pressure: 250 };
  return Math.min(100, (value / max[type]) * 100);
}

/**
 * Telemetry status indicator color
 */
export function getTelemetryStatusColor(status: string): string {
  switch (status) {
    case 'normal': return 'bg-green-500 animate-pulse';
    case 'warning': return 'bg-yellow-500 animate-pulse';
    case 'fault': return 'bg-red-500 animate-pulse';
    default: return 'bg-slate-500';
  }
}