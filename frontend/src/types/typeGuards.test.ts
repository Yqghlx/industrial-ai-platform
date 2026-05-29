/**
 * typeGuards.ts 测试
 * 测试所有类型守卫函数的正确性
 */

import { describe, it, expect } from 'vitest';
import {
  hasProperty,
  isDeviceStatus,
  isDevice,
  isDeviceArray,
  isTelemetry,
  isTelemetryArray,
  isDeviceStats,
  isAlertSeverity,
  isAlert,
  isAlertArray,
  isAlertStatus,
  isAlertOperator,
  isAlertRule,
  isAlertRuleArray,
  isWorkOrder,
  isWorkOrderArray,
  isNotification,
  isNotificationArray,
  isReport,
  isReportArray,
  isUserRole,
  isUser,
  isUserArray,
  isROIStats,
  isSystemStatus,
  isGraphNode,
  isGraphLink,
  isDeviceGraph,
  isAgentResponse,
  isAlertStatusPayload,
  isTelemetryData,
  isTelemetryDataArray,
  isTelemetryHistory,
  isTelemetryHistoryArray,
  // 安全转换函数
  asDeviceSafe,
  asTelemetryArraySafe,
  asDeviceStatsSafe,
  asAlertRuleArraySafe,
  asWorkOrderArraySafe,
  asNotificationArraySafe,
  asReportArraySafe,
  asUserArraySafe,
  asROIStatsSafe,
  asSystemStatusSafe,
  asDeviceGraphSafe,
  asAgentResponseSafe,
  asAlertStatusSafe,
  asTelemetryDataSafe,
  asTelemetryDataArraySafe,
  asTelemetryHistoryArraySafe,
  asAlertStatusPayloadSafe,
} from './typeGuards';

// ============== 测试数据 ==============

const validDevice = {
  id: 'device-1',
  name: '测试设备',
  type: 'pump',
  status: 'online' as const,
  location: '车间A',
};

const validTelemetry = {
  device_id: 'device-1',
  timestamp: '2024-01-01T00:00:00Z',
  temperature: 25.5,
  pressure: 101.3,
  vibration: 0.5,
  power: 1500,
  status: 'online' as const,
};

const validDeviceStats = {
  device_id: 'device-1',
  avg_temperature: 25.0,
  avg_vibration: 0.3,
  max_temperature: 30.0,
  max_vibration: 0.8,
  data_points: 100,
};

const validAlert = {
  id: 1,
  rule_id: 1,
  device_id: 'device-1',
  metric: 'temperature',
  value: 100,
  threshold: 80,
  severity: 'high' as const,
  status: 'active',
  triggered_at: '2024-01-01T00:00:00Z',
};

const validAlertRule = {
  id: 1,
  name: '温度过高规则',
  device_type: 'pump',
  metric: 'temperature',
  operator: '>' as const,
  threshold: 80,
  severity: 'high' as const,
  enabled: true,
  cooldown_sec: 300,
};

const validWorkOrder = {
  id: 1,
  title: '维修工单',
  description: '设备异常维修',
  device_id: 'device-1',
  priority: 'high',
  status: 'pending',
  created_at: '2024-01-01T00:00:00Z',
};

const validNotification = {
  id: 1,
  type: 'alert',
  title: '告警通知',
  message: '设备温度过高',
  read: false,
  created_at: '2024-01-01T00:00:00Z',
};

const validReport = {
  id: 1,
  title: '日报',
  type: 'daily',
  content: '报告内容',
  generated_at: '2024-01-01T00:00:00Z',
};

const validUser = {
  id: 1,
  username: 'admin',
  email: 'admin@test.com',
  role: 'admin' as const,
  created_at: '2024-01-01T00:00:00Z',
};

const validROIStats = {
  total_devices: 10,
  active_alerts: 3,
  open_work_orders: 2,
  resolved_issues: 50,
  predicted_savings: 100000,
  uptime_percentage: 99.5,
  avg_response_time_hours: 2.5,
};

const validSystemStatus = {
  database: 'healthy',
  db_latency_ms: 5,
  uptime: '10d',
  version: '1.0.0',
  timestamp: '2024-01-01T00:00:00Z',
};

const validGraphNode = {
  id: 'node-1',
  name: '节点1',
  type: 'pump',
  status: 'online' as const,
};

const validGraphLink = {
  source: 'node-1',
  target: 'node-2',
  type: 'dependency',
};

const validDeviceGraph = {
  nodes: [validGraphNode],
  links: [validGraphLink],
};

const validAgentResponse = {
  session_id: 'session-1',
  response: '分析结果',
  agent: 'device-expert',
};

const validAlertStatusPayload = {
  id: 1,
  status: 'resolved',
};

const validTelemetryData = {
  device_id: 'device-1',
  timestamp: '2024-01-01T00:00:00Z',
  status: 'online',
  temperature: 25.5,
};

const validTelemetryHistory = {
  id: 1,
  device_id: 'device-1',
  timestamp: '2024-01-01T00:00:00Z',
  temperature: 25.5,
};

// ============== hasProperty ==============

describe('hasProperty', () => {
  it('有指定属性时返回 true', () => {
    expect(hasProperty({ name: 'test' }, 'name')).toBe(true);
  });

  it('无指定属性时返回 false', () => {
    expect(hasProperty({ name: 'test' }, 'age')).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(hasProperty(null, 'name')).toBe(false);
  });

  it('传入 undefined 返回 false', () => {
    expect(hasProperty(undefined, 'name')).toBe(false);
  });

  it('传入字符串返回 false', () => {
    expect(hasProperty('hello', 'length')).toBe(false);
  });

  it('传入数字返回 false', () => {
    expect(hasProperty(42, 'toString')).toBe(false);
  });
});

// ============== isDeviceStatus ==============

describe('isDeviceStatus', () => {
  it('有效状态值返回 true', () => {
    expect(isDeviceStatus('online')).toBe(true);
    expect(isDeviceStatus('warning')).toBe(true);
    expect(isDeviceStatus('fault')).toBe(true);
    expect(isDeviceStatus('offline')).toBe(true);
  });

  it('无效状态值返回 false', () => {
    expect(isDeviceStatus('unknown')).toBe(false);
    expect(isDeviceStatus('')).toBe(false);
  });

  it('非字符串返回 false', () => {
    expect(isDeviceStatus(123)).toBe(false);
    expect(isDeviceStatus(null)).toBe(false);
    expect(isDeviceStatus(undefined)).toBe(false);
  });
});

// ============== isDevice ==============

describe('isDevice', () => {
  it('有效 Device 对象返回 true', () => {
    expect(isDevice(validDevice)).toBe(true);
  });

  it('缺少必要字段返回 false', () => {
    expect(isDevice({ id: '1', name: 'test' })).toBe(false);
  });

  it('status 类型错误返回 false', () => {
    expect(isDevice({ ...validDevice, status: 'invalid' })).toBe(false);
  });

  it('字段类型错误返回 false', () => {
    expect(isDevice({ ...validDevice, id: 123 })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isDevice(null)).toBe(false);
  });

  it('传入 undefined 返回 false', () => {
    expect(isDevice(undefined)).toBe(false);
  });

  it('传入数组返回 false', () => {
    expect(isDevice([])).toBe(false);
  });
});

// ============== isDeviceArray ==============

describe('isDeviceArray', () => {
  it('有效 Device 数组返回 true', () => {
    expect(isDeviceArray([validDevice])).toBe(true);
  });

  it('空数组返回 true', () => {
    expect(isDeviceArray([])).toBe(true);
  });

  it('包含无效元素的数组返回 false', () => {
    expect(isDeviceArray([validDevice, { bad: true }])).toBe(false);
  });

  it('非数组返回 false', () => {
    expect(isDeviceArray(null)).toBe(false);
    expect(isDeviceArray('not array')).toBe(false);
  });
});

// ============== isTelemetry ==============

describe('isTelemetry', () => {
  it('有效 Telemetry 对象返回 true', () => {
    expect(isTelemetry(validTelemetry)).toBe(true);
  });

  it('仅包含必填字段时返回 true', () => {
    expect(isTelemetry({
      device_id: 'device-1',
      timestamp: '2024-01-01T00:00:00Z',
    })).toBe(true);
  });

  it('缺少 device_id 返回 false', () => {
    expect(isTelemetry({ timestamp: '2024-01-01T00:00:00Z' })).toBe(false);
  });

  it('字段类型错误返回 false', () => {
    expect(isTelemetry({ ...validTelemetry, device_id: 123 })).toBe(false);
  });

  it('temperature 类型错误返回 false', () => {
    expect(isTelemetry({ ...validTelemetry, temperature: 'hot' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isTelemetry(null)).toBe(false);
  });
});

// ============== isTelemetryArray ==============

describe('isTelemetryArray', () => {
  it('有效数组返回 true', () => {
    expect(isTelemetryArray([validTelemetry])).toBe(true);
  });

  it('非数组返回 false', () => {
    expect(isTelemetryArray(validTelemetry)).toBe(false);
  });
});

// ============== isDeviceStats ==============

describe('isDeviceStats', () => {
  it('有效 DeviceStats 返回 true', () => {
    expect(isDeviceStats(validDeviceStats)).toBe(true);
  });

  it('仅包含必填字段时返回 true', () => {
    expect(isDeviceStats({
      device_id: 'device-1',
      avg_temperature: 25.0,
      avg_vibration: 0.3,
    })).toBe(true);
  });

  it('缺少必要字段返回 false', () => {
    expect(isDeviceStats({ device_id: '1' })).toBe(false);
  });

  it('字段类型错误返回 false', () => {
    expect(isDeviceStats({ ...validDeviceStats, avg_temperature: 'hot' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isDeviceStats(null)).toBe(false);
  });
});

// ============== isAlertSeverity ==============

describe('isAlertSeverity', () => {
  it('有效严重级别返回 true', () => {
    expect(isAlertSeverity('low')).toBe(true);
    expect(isAlertSeverity('medium')).toBe(true);
    expect(isAlertSeverity('high')).toBe(true);
    expect(isAlertSeverity('critical')).toBe(true);
  });

  it('无效值返回 false', () => {
    expect(isAlertSeverity('unknown')).toBe(false);
    expect(isAlertSeverity(123)).toBe(false);
    expect(isAlertSeverity(null)).toBe(false);
  });
});

// ============== isAlert ==============

describe('isAlert', () => {
  it('有效 Alert 返回 true', () => {
    expect(isAlert(validAlert)).toBe(true);
  });

  it('metric 为可选字段', () => {
    const { metric, ...withoutMetric } = validAlert;
    expect(isAlert(withoutMetric)).toBe(true);
  });

  it('缺少必要字段返回 false', () => {
    expect(isAlert({ id: 1 })).toBe(false);
  });

  it('severity 无效返回 false', () => {
    expect(isAlert({ ...validAlert, severity: 'invalid' })).toBe(false);
  });

  it('status 无效返回 false', () => {
    expect(isAlert({ ...validAlert, status: 'invalid' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isAlert(null)).toBe(false);
  });
});

// ============== isAlertArray ==============

describe('isAlertArray', () => {
  it('有效数组返回 true', () => {
    expect(isAlertArray([validAlert])).toBe(true);
  });

  it('包含无效元素返回 false', () => {
    expect(isAlertArray([validAlert, {}])).toBe(false);
  });
});

// ============== isAlertStatus ==============

describe('isAlertStatus', () => {
  it('有效状态返回 true', () => {
    expect(isAlertStatus('active')).toBe(true);
    expect(isAlertStatus('acknowledged')).toBe(true);
    expect(isAlertStatus('resolved')).toBe(true);
  });

  it('无效状态返回 false', () => {
    expect(isAlertStatus('invalid')).toBe(false);
    expect(isAlertStatus(123)).toBe(false);
  });
});

// ============== isAlertOperator ==============

describe('isAlertOperator', () => {
  it('有效操作符返回 true', () => {
    expect(isAlertOperator('>')).toBe(true);
    expect(isAlertOperator('>=')).toBe(true);
    expect(isAlertOperator('<')).toBe(true);
    expect(isAlertOperator('<=')).toBe(true);
    expect(isAlertOperator('==')).toBe(true);
    expect(isAlertOperator('!=')).toBe(true);
  });

  it('无效操作符返回 false', () => {
    expect(isAlertOperator('===')).toBe(false);
    expect(isAlertOperator('><')).toBe(false);
  });
});

// ============== isAlertRule ==============

describe('isAlertRule', () => {
  it('有效 AlertRule 返回 true', () => {
    expect(isAlertRule(validAlertRule)).toBe(true);
  });

  it('缺少必要字段返回 false', () => {
    expect(isAlertRule({ id: 1, name: 'test' })).toBe(false);
  });

  it('operator 无效返回 false', () => {
    expect(isAlertRule({ ...validAlertRule, operator: '===' })).toBe(false);
  });

  it('enabled 非布尔返回 false', () => {
    expect(isAlertRule({ ...validAlertRule, enabled: 'yes' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isAlertRule(null)).toBe(false);
  });
});

// ============== isAlertRuleArray ==============

describe('isAlertRuleArray', () => {
  it('有效数组返回 true', () => {
    expect(isAlertRuleArray([validAlertRule])).toBe(true);
  });

  it('非数组返回 false', () => {
    expect(isAlertRuleArray(null)).toBe(false);
  });
});

// ============== isWorkOrder ==============

describe('isWorkOrder', () => {
  it('有效 WorkOrder 返回 true', () => {
    expect(isWorkOrder(validWorkOrder)).toBe(true);
  });

  it('缺少必要字段返回 false', () => {
    expect(isWorkOrder({ id: 1 })).toBe(false);
  });

  it('priority 无效返回 false', () => {
    expect(isWorkOrder({ ...validWorkOrder, priority: 'super-urgent' })).toBe(false);
  });

  it('status 无效返回 false', () => {
    expect(isWorkOrder({ ...validWorkOrder, status: 'unknown' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isWorkOrder(null)).toBe(false);
  });
});

// ============== isWorkOrderArray ==============

describe('isWorkOrderArray', () => {
  it('有效数组返回 true', () => {
    expect(isWorkOrderArray([validWorkOrder])).toBe(true);
  });

  it('非数组返回 false', () => {
    expect(isWorkOrderArray(null)).toBe(false);
  });
});

// ============== isNotification ==============

describe('isNotification', () => {
  it('有效 Notification 返回 true', () => {
    expect(isNotification(validNotification)).toBe(true);
  });

  it('type 无效返回 false', () => {
    expect(isNotification({ ...validNotification, type: 'invalid' })).toBe(false);
  });

  it('read 非布尔返回 false', () => {
    expect(isNotification({ ...validNotification, read: 'yes' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isNotification(null)).toBe(false);
  });
});

// ============== isNotificationArray ==============

describe('isNotificationArray', () => {
  it('有效数组返回 true', () => {
    expect(isNotificationArray([validNotification])).toBe(true);
  });

  it('非数组返回 false', () => {
    expect(isNotificationArray(null)).toBe(false);
  });
});

// ============== isReport ==============

describe('isReport', () => {
  it('有效 Report 返回 true', () => {
    expect(isReport(validReport)).toBe(true);
  });

  it('type 无效返回 false', () => {
    expect(isReport({ ...validReport, type: 'invalid' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isReport(null)).toBe(false);
  });
});

// ============== isReportArray ==============

describe('isReportArray', () => {
  it('有效数组返回 true', () => {
    expect(isReportArray([validReport])).toBe(true);
  });

  it('非数组返回 false', () => {
    expect(isReportArray(null)).toBe(false);
  });
});

// ============== isUserRole ==============

describe('isUserRole', () => {
  it('有效角色返回 true', () => {
    expect(isUserRole('admin')).toBe(true);
    expect(isUserRole('user')).toBe(true);
    expect(isUserRole('viewer')).toBe(true);
  });

  it('无效角色返回 false', () => {
    expect(isUserRole('superadmin')).toBe(false);
    expect(isUserRole(123)).toBe(false);
    expect(isUserRole(null)).toBe(false);
  });
});

// ============== isUser ==============

describe('isUser', () => {
  it('有效 User 返回 true', () => {
    expect(isUser(validUser)).toBe(true);
  });

  it('role 无效返回 false', () => {
    expect(isUser({ ...validUser, role: 'superadmin' })).toBe(false);
  });

  it('字段类型错误返回 false', () => {
    expect(isUser({ ...validUser, id: 'not-number' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isUser(null)).toBe(false);
  });
});

// ============== isUserArray ==============

describe('isUserArray', () => {
  it('有效数组返回 true', () => {
    expect(isUserArray([validUser])).toBe(true);
  });

  it('非数组返回 false', () => {
    expect(isUserArray(null)).toBe(false);
  });
});

// ============== isROIStats ==============

describe('isROIStats', () => {
  it('有效 ROIStats 返回 true', () => {
    expect(isROIStats(validROIStats)).toBe(true);
  });

  it('缺少字段返回 false', () => {
    expect(isROIStats({ total_devices: 10 })).toBe(false);
  });

  it('字段类型错误返回 false', () => {
    expect(isROIStats({ ...validROIStats, total_devices: '10' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isROIStats(null)).toBe(false);
  });
});

// ============== isSystemStatus ==============

describe('isSystemStatus', () => {
  it('有效 SystemStatus 返回 true', () => {
    expect(isSystemStatus(validSystemStatus)).toBe(true);
  });

  it('database 值无效返回 false', () => {
    expect(isSystemStatus({ ...validSystemStatus, database: 'ok' })).toBe(false);
  });

  it('字段类型错误返回 false', () => {
    expect(isSystemStatus({ ...validSystemStatus, db_latency_ms: 'fast' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isSystemStatus(null)).toBe(false);
  });
});

// ============== isGraphNode / isGraphLink / isDeviceGraph ==============

describe('isGraphNode', () => {
  it('有效 GraphNode 返回 true', () => {
    expect(isGraphNode(validGraphNode)).toBe(true);
  });

  it('status 无效返回 false', () => {
    expect(isGraphNode({ ...validGraphNode, status: 'invalid' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isGraphNode(null)).toBe(false);
  });
});

describe('isGraphLink', () => {
  it('有效 GraphLink 返回 true', () => {
    expect(isGraphLink(validGraphLink)).toBe(true);
  });

  it('缺少字段返回 false', () => {
    expect(isGraphLink({ source: 'a' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isGraphLink(null)).toBe(false);
  });
});

describe('isDeviceGraph', () => {
  it('有效 DeviceGraph 返回 true', () => {
    expect(isDeviceGraph(validDeviceGraph)).toBe(true);
  });

  it('nodes 包含无效元素返回 false', () => {
    expect(isDeviceGraph({ nodes: [{ bad: true }], links: [] })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isDeviceGraph(null)).toBe(false);
  });
});

// ============== isAgentResponse ==============

describe('isAgentResponse', () => {
  it('有效 AgentResponse 返回 true', () => {
    expect(isAgentResponse(validAgentResponse)).toBe(true);
  });

  it('字段类型错误返回 false', () => {
    expect(isAgentResponse({ ...validAgentResponse, session_id: 123 })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isAgentResponse(null)).toBe(false);
  });
});

// ============== isAlertStatusPayload ==============

describe('isAlertStatusPayload', () => {
  it('有效 payload 返回 true', () => {
    expect(isAlertStatusPayload(validAlertStatusPayload)).toBe(true);
  });

  it('字段类型错误返回 false', () => {
    expect(isAlertStatusPayload({ id: 'not-number', status: 'ok' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isAlertStatusPayload(null)).toBe(false);
  });
});

// ============== isTelemetryData ==============

describe('isTelemetryData', () => {
  it('有效 TelemetryData 返回 true', () => {
    expect(isTelemetryData(validTelemetryData)).toBe(true);
  });

  it('仅包含必填字段时返回 true', () => {
    expect(isTelemetryData({
      device_id: 'device-1',
      timestamp: '2024-01-01T00:00:00Z',
      status: 'online',
    })).toBe(true);
  });

  it('temperature 类型错误返回 false', () => {
    expect(isTelemetryData({ ...validTelemetryData, temperature: 'hot' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isTelemetryData(null)).toBe(false);
  });
});

// ============== isTelemetryDataArray ==============

describe('isTelemetryDataArray', () => {
  it('有效数组返回 true', () => {
    expect(isTelemetryDataArray([validTelemetryData])).toBe(true);
  });

  it('非数组返回 false', () => {
    expect(isTelemetryDataArray(null)).toBe(false);
  });
});

// ============== isTelemetryHistory ==============

describe('isTelemetryHistory', () => {
  it('有效 TelemetryHistory 返回 true', () => {
    expect(isTelemetryHistory(validTelemetryHistory)).toBe(true);
  });

  it('id 类型错误返回 false', () => {
    expect(isTelemetryHistory({ ...validTelemetryHistory, id: 'not-number' })).toBe(false);
  });

  it('传入 null 返回 false', () => {
    expect(isTelemetryHistory(null)).toBe(false);
  });
});

// ============== isTelemetryHistoryArray ==============

describe('isTelemetryHistoryArray', () => {
  it('有效数组返回 true', () => {
    expect(isTelemetryHistoryArray([validTelemetryHistory])).toBe(true);
  });

  it('非数组返回 false', () => {
    expect(isTelemetryHistoryArray(null)).toBe(false);
  });
});

// ============== 安全转换函数 ==============

describe('安全转换函数', () => {
  it('asDeviceSafe - 有效对象返回对象', () => {
    expect(asDeviceSafe(validDevice)).toEqual(validDevice);
  });

  it('asDeviceSafe - 无效对象返回 null', () => {
    expect(asDeviceSafe(null)).toBeNull();
    expect(asDeviceSafe({ bad: true })).toBeNull();
  });

  it('asTelemetryArraySafe - 有效数组返回数组', () => {
    expect(asTelemetryArraySafe([validTelemetry])).toEqual([validTelemetry]);
  });

  it('asTelemetryArraySafe - 无效数据返回空数组', () => {
    expect(asTelemetryArraySafe(null)).toEqual([]);
    expect(asTelemetryArraySafe([{}])).toEqual([]);
  });

  it('asDeviceStatsSafe - 有效返回对象', () => {
    expect(asDeviceStatsSafe(validDeviceStats)).toEqual(validDeviceStats);
  });

  it('asDeviceStatsSafe - 无效返回 null', () => {
    expect(asDeviceStatsSafe(null)).toBeNull();
  });

  it('asAlertRuleArraySafe - 有效返回数组', () => {
    expect(asAlertRuleArraySafe([validAlertRule])).toEqual([validAlertRule]);
  });

  it('asAlertRuleArraySafe - 无效返回空数组', () => {
    expect(asAlertRuleArraySafe(null)).toEqual([]);
  });

  it('asWorkOrderArraySafe - 有效返回数组', () => {
    expect(asWorkOrderArraySafe([validWorkOrder])).toEqual([validWorkOrder]);
  });

  it('asWorkOrderArraySafe - 无效返回空数组', () => {
    expect(asWorkOrderArraySafe(null)).toEqual([]);
  });

  it('asNotificationArraySafe - 有效返回数组', () => {
    expect(asNotificationArraySafe([validNotification])).toEqual([validNotification]);
  });

  it('asNotificationArraySafe - 无效返回空数组', () => {
    expect(asNotificationArraySafe(null)).toEqual([]);
  });

  it('asReportArraySafe - 有效返回数组', () => {
    expect(asReportArraySafe([validReport])).toEqual([validReport]);
  });

  it('asReportArraySafe - 无效返回空数组', () => {
    expect(asReportArraySafe(null)).toEqual([]);
  });

  it('asUserArraySafe - 有效返回数组', () => {
    expect(asUserArraySafe([validUser])).toEqual([validUser]);
  });

  it('asUserArraySafe - 无效返回空数组', () => {
    expect(asUserArraySafe(null)).toEqual([]);
  });

  it('asROIStatsSafe - 有效返回对象', () => {
    expect(asROIStatsSafe(validROIStats)).toEqual(validROIStats);
  });

  it('asROIStatsSafe - 无效返回 null', () => {
    expect(asROIStatsSafe(null)).toBeNull();
  });

  it('asSystemStatusSafe - 有效返回对象', () => {
    expect(asSystemStatusSafe(validSystemStatus)).toEqual(validSystemStatus);
  });

  it('asSystemStatusSafe - 无效返回 null', () => {
    expect(asSystemStatusSafe(null)).toBeNull();
  });

  it('asDeviceGraphSafe - 有效返回对象', () => {
    expect(asDeviceGraphSafe(validDeviceGraph)).toEqual(validDeviceGraph);
  });

  it('asDeviceGraphSafe - 无效返回 null', () => {
    expect(asDeviceGraphSafe(null)).toBeNull();
  });

  it('asAgentResponseSafe - 有效返回对象', () => {
    expect(asAgentResponseSafe(validAgentResponse)).toEqual(validAgentResponse);
  });

  it('asAgentResponseSafe - 无效返回 null', () => {
    expect(asAgentResponseSafe(null)).toBeNull();
  });

  it('asAlertStatusSafe - 有效返回值', () => {
    expect(asAlertStatusSafe('active')).toBe('active');
  });

  it('asAlertStatusSafe - 无效返回 null', () => {
    expect(asAlertStatusSafe('invalid')).toBeNull();
    expect(asAlertStatusSafe(null)).toBeNull();
  });

  it('asTelemetryDataSafe - 有效返回对象', () => {
    expect(asTelemetryDataSafe(validTelemetryData)).toEqual(validTelemetryData);
  });

  it('asTelemetryDataSafe - 无效返回 null', () => {
    expect(asTelemetryDataSafe(null)).toBeNull();
  });

  it('asTelemetryDataArraySafe - 有效返回数组', () => {
    expect(asTelemetryDataArraySafe([validTelemetryData])).toEqual([validTelemetryData]);
  });

  it('asTelemetryDataArraySafe - 无效返回空数组', () => {
    expect(asTelemetryDataArraySafe(null)).toEqual([]);
  });

  it('asTelemetryHistoryArraySafe - 有效返回数组', () => {
    expect(asTelemetryHistoryArraySafe([validTelemetryHistory])).toEqual([validTelemetryHistory]);
  });

  it('asTelemetryHistoryArraySafe - 无效返回空数组', () => {
    expect(asTelemetryHistoryArraySafe(null)).toEqual([]);
  });

  it('asAlertStatusPayloadSafe - 有效返回对象', () => {
    expect(asAlertStatusPayloadSafe(validAlertStatusPayload)).toEqual(validAlertStatusPayload);
  });

  it('asAlertStatusPayloadSafe - 无效返回 null', () => {
    expect(asAlertStatusPayloadSafe(null)).toBeNull();
  });
});
