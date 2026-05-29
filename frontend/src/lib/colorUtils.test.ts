import { describe, it, expect } from 'vitest';

import {
  getDeviceStatusColor,
  getDeviceStatusBadgeClass,
  getDeviceStatusTextColor,
  getWorkOrderStatusColor,
  getWorkOrderPriorityColor,
  getAlertSeverityColor,
  getAgentColor,
  getGaugeColor,
  getGaugeStrokeColor,
  getGaugePercentage,
  getTelemetryStatusColor,
} from './colorUtils';

describe('colorUtils', () => {
  describe('getDeviceStatusColor', () => {
    it('returns correct color for each status', () => {
      expect(getDeviceStatusColor('online')).toBe('bg-green-500');
      expect(getDeviceStatusColor('warning')).toBe('bg-yellow-500');
      expect(getDeviceStatusColor('fault')).toBe('bg-red-500');
      expect(getDeviceStatusColor('offline')).toBe('bg-slate-500');
    });

    it('returns default color for unknown status', () => {
      expect(getDeviceStatusColor('unknown')).toBe('bg-slate-500');
      expect(getDeviceStatusColor('')).toBe('bg-slate-500');
    });
  });

  describe('getDeviceStatusBadgeClass', () => {
    it('returns correct badge class for each status', () => {
      expect(getDeviceStatusBadgeClass('online')).toBe('status-online');
      expect(getDeviceStatusBadgeClass('warning')).toBe('status-warning');
      expect(getDeviceStatusBadgeClass('fault')).toBe('status-fault');
      expect(getDeviceStatusBadgeClass('offline')).toBe('status-offline');
    });

    it('returns default badge class for unknown status', () => {
      expect(getDeviceStatusBadgeClass('unknown')).toBe('status-offline');
    });
  });

  describe('getDeviceStatusTextColor', () => {
    it('returns correct text color for each status', () => {
      expect(getDeviceStatusTextColor('online')).toBe('text-green-400');
      expect(getDeviceStatusTextColor('warning')).toBe('text-yellow-400');
      expect(getDeviceStatusTextColor('fault')).toBe('text-red-400');
      expect(getDeviceStatusTextColor('offline')).toBe('text-slate-400');
    });
  });

  describe('getWorkOrderStatusColor', () => {
    it('returns correct color for each status', () => {
      expect(getWorkOrderStatusColor('pending')).toContain('bg-slate-500/20');
      expect(getWorkOrderStatusColor('in_progress')).toContain('bg-primary-500/20');
      expect(getWorkOrderStatusColor('completed')).toContain('bg-green-500/20');
      expect(getWorkOrderStatusColor('cancelled')).toContain('bg-red-500/20');
    });

    it('returns default for unknown status', () => {
      expect(getWorkOrderStatusColor('unknown')).toContain('bg-slate-500/20');
    });
  });

  describe('getWorkOrderPriorityColor', () => {
    it('returns correct color for each priority', () => {
      expect(getWorkOrderPriorityColor('urgent')).toContain('bg-red-500/20');
      expect(getWorkOrderPriorityColor('high')).toContain('bg-orange-500/20');
      expect(getWorkOrderPriorityColor('medium')).toContain('bg-yellow-500/20');
      expect(getWorkOrderPriorityColor('low')).toContain('bg-green-500/20');
    });

    it('returns default for unknown priority', () => {
      expect(getWorkOrderPriorityColor('unknown')).toContain('bg-slate-500/20');
    });
  });

  describe('getAlertSeverityColor', () => {
    it('returns correct color for each severity', () => {
      expect(getAlertSeverityColor('critical')).toContain('bg-red-500/20');
      expect(getAlertSeverityColor('high')).toContain('bg-orange-500/20');
      expect(getAlertSeverityColor('medium')).toContain('bg-yellow-500/20');
      expect(getAlertSeverityColor('low')).toContain('bg-green-500/20');
    });

    it('returns default for unknown severity', () => {
      expect(getAlertSeverityColor('unknown')).toContain('bg-slate-500/20');
    });
  });

  describe('getAgentColor', () => {
    it('returns correct color for device expert', () => {
      expect(getAgentColor('device_expert')).toBe('text-blue-400');
      expect(getAgentColor('设备专家')).toBe('text-blue-400');
    });

    it('returns correct color for maintenance expert', () => {
      expect(getAgentColor('maintenance_expert')).toBe('text-green-400');
      expect(getAgentColor('维护专家')).toBe('text-green-400');
    });

    it('returns correct color for prediction expert', () => {
      expect(getAgentColor('prediction_expert')).toBe('text-purple-400');
      expect(getAgentColor('预测专家')).toBe('text-purple-400');
    });

    it('returns correct color for optimization expert', () => {
      expect(getAgentColor('optimization_expert')).toBe('text-orange-400');
      expect(getAgentColor('优化专家')).toBe('text-orange-400');
    });

    it('returns default for unknown agent', () => {
      expect(getAgentColor('unknown_agent')).toBe('text-slate-400');
    });
  });

  describe('getGaugeColor', () => {
    it('returns green for normal temperature', () => {
      expect(getGaugeColor(50, 'temperature')).toBe('text-green-400');
    });

    it('returns yellow for warning temperature', () => {
      expect(getGaugeColor(100, 'temperature')).toBe('text-yellow-400');
    });

    it('returns red for critical temperature', () => {
      expect(getGaugeColor(120, 'temperature')).toBe('text-red-400');
    });

    it('returns correct colors for vibration', () => {
      expect(getGaugeColor(1, 'vibration')).toBe('text-green-400');
      expect(getGaugeColor(3, 'vibration')).toBe('text-yellow-400');
      expect(getGaugeColor(5, 'vibration')).toBe('text-red-400');
    });

    it('returns correct colors for pressure', () => {
      expect(getGaugeColor(50, 'pressure')).toBe('text-green-400');
      expect(getGaugeColor(150, 'pressure')).toBe('text-yellow-400');
      expect(getGaugeColor(200, 'pressure')).toBe('text-red-400');
    });

    it('returns correct colors for power', () => {
      expect(getGaugeColor(30, 'power')).toBe('text-green-400');
      expect(getGaugeColor(80, 'power')).toBe('text-yellow-400');
      expect(getGaugeColor(100, 'power')).toBe('text-red-400');
    });
  });

  describe('getGaugeStrokeColor', () => {
    it('returns hex colors for temperature', () => {
      expect(getGaugeStrokeColor(50, 'temperature')).toBe('#10b981');
      expect(getGaugeStrokeColor(100, 'temperature')).toBe('#f59e0b');
      expect(getGaugeStrokeColor(120, 'temperature')).toBe('#ef4444');
    });

    it('returns hex colors for vibration', () => {
      expect(getGaugeStrokeColor(1, 'vibration')).toBe('#10b981');
      expect(getGaugeStrokeColor(3, 'vibration')).toBe('#f59e0b');
      expect(getGaugeStrokeColor(5, 'vibration')).toBe('#ef4444');
    });

    it('returns hex colors for pressure', () => {
      expect(getGaugeStrokeColor(50, 'pressure')).toBe('#10b981');
      expect(getGaugeStrokeColor(150, 'pressure')).toBe('#f59e0b');
      expect(getGaugeStrokeColor(200, 'pressure')).toBe('#ef4444');
    });
  });

  describe('getGaugePercentage', () => {
    it('calculates percentage for temperature', () => {
      expect(getGaugePercentage(75, 'temperature')).toBe(50);
      expect(getGaugePercentage(150, 'temperature')).toBe(100);
      expect(getGaugePercentage(200, 'temperature')).toBe(100); // 上限 100
    });

    it('calculates percentage for vibration', () => {
      expect(getGaugePercentage(3, 'vibration')).toBe(50);
      expect(getGaugePercentage(6, 'vibration')).toBe(100);
    });

    it('calculates percentage for pressure', () => {
      expect(getGaugePercentage(125, 'pressure')).toBe(50);
      expect(getGaugePercentage(250, 'pressure')).toBe(100);
    });
  });

  describe('getTelemetryStatusColor', () => {
    it('returns correct color for each status', () => {
      expect(getTelemetryStatusColor('normal')).toBe('bg-green-500 animate-pulse');
      expect(getTelemetryStatusColor('warning')).toBe('bg-yellow-500 animate-pulse');
      expect(getTelemetryStatusColor('fault')).toBe('bg-red-500 animate-pulse');
      expect(getTelemetryStatusColor('unknown')).toBe('bg-slate-500');
    });

    it('returns default for unknown status', () => {
      expect(getTelemetryStatusColor('')).toBe('bg-slate-500');
    });
  });
});
