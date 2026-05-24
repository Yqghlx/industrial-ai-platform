import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');
const requestsPerSecond = new Counter('requests_per_second');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Warm up: 10 users
    { duration: '1m', target: 50 },    // Ramp up: 50 users
    { duration: '2m', target: 100 },   // Peak: 100 users
    { duration: '1m', target: 50 },    // Ramp down: 50 users
    { duration: '30s', target: 0 },    // Cool down: 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95% of requests < 500ms
    errors: ['rate<0.05'],              // Error rate < 5%
    http_req_failed: ['rate<0.01'],     // Failed requests < 1%
  },
};

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_PREFIX = '/api/v1';

// Test data
const TEST_USER = {
  username: 'admin',
  password: 'admin123',  // In production, use environment variable
};

let authToken = '';

// Login and get JWT token
export function setup() {
  const loginRes = http.post(`${BASE_URL}${API_PREFIX}/auth/login`, JSON.stringify(TEST_USER), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(loginRes, {
    'login successful': (r) => r.status === 200,
    'received token': (r) => r.json('data.token') !== undefined,
  });
  
  return { token: loginRes.json('data.token') };
}

// Main test function
export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`,
  };

  // Test 1: List devices (high frequency read)
  testListDevices(headers);
  
  // Test 2: Get device details
  testGetDevice(headers);
  
  // Test 3: Submit telemetry data (write operation)
  testSubmitTelemetry(headers);
  
  // Test 4: Get latest telemetry
  testGetLatestTelemetry(headers);
  
  // Test 5: List alert rules
  testListRules(headers);
  
  // Test 6: AI Agent query (expensive operation)
  testAgentQuery(headers);
  
  // Test 7: ROI stats (cached read)
  testROIStats(headers);
  
  // Test 8: Health check (public endpoint)
  testHealthCheck();
  
  sleep(1); // Think time
}

function testListDevices(headers) {
  const res = http.get(`${BASE_URL}${API_PREFIX}/devices`, { headers });
  
  const success = check(res, {
    'list devices status 200': (r) => r.status === 200,
    'list devices has data': (r) => r.json('data') !== undefined,
  });
  
  errorRate.add(!success);
  responseTime.add(res.timings.duration);
  requestsPerSecond.add(1);
}

function testGetDevice(headers) {
  // Get first device from list
  const deviceId = 'device-001'; // Mock device ID
  
  const res = http.get(`${BASE_URL}${API_PREFIX}/devices/${deviceId}`, { headers });
  
  check(res, {
    'get device status 200 or 404': (r) => r.status === 200 || r.status === 404,
  });
  
  errorRate.add(res.status >= 500);
  responseTime.add(res.timings.duration);
}

function testSubmitTelemetry(headers) {
  const payload = {
    device_id: `sim-device-${__VU}`,
    device_type: 'CNC',
    timestamp: new Date().toISOString(),
    metrics: {
      temperature: 75 + Math.random() * 20,
      vibration: 2.5 + Math.random() * 1.5,
      pressure: 100 + Math.random() * 30,
      power_consumption: 500 + Math.random() * 100,
    },
    status: 'running',
  };
  
  const res = http.post(`${BASE_URL}${API_PREFIX}/devices/telemetry`, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' }, // Public endpoint, no auth needed
  });
  
  const success = check(res, {
    'telemetry submit status 200': (r) => r.status === 200,
  });
  
  errorRate.add(!success);
  responseTime.add(res.timings.duration);
  requestsPerSecond.add(1);
}

function testGetLatestTelemetry(headers) {
  const res = http.get(`${BASE_URL}${API_PREFIX}/devices/latest`, { headers });
  
  check(res, {
    'latest telemetry status 200': (r) => r.status === 200,
  });
  
  errorRate.add(res.status >= 500);
  responseTime.add(res.timings.duration);
}

function testListRules(headers) {
  const res = http.get(`${BASE_URL}${API_PREFIX}/rules`, { headers });
  
  check(res, {
    'list rules status 200': (r) => r.status === 200,
  });
  
  errorRate.add(res.status >= 500);
  responseTime.add(res.timings.duration);
}

function testAgentQuery(headers) {
  const payload = {
    query: '设备温度异常如何处理？',
    device_ids: ['device-001'],
  };
  
  const res = http.post(`${BASE_URL}${API_PREFIX}/agent/query`, JSON.stringify(payload), { headers, timeout: '35s' });
  
  check(res, {
    'agent query status 200': (r) => r.status === 200 || r.status === 429,
    'agent has response': (r) => {
      try {
        const data = r.json();
        const resp = data?.response?.response || data?.response?.data || data?.response;
        return (resp !== undefined && resp !== '') || r.status === 429;
      } catch (e) {
        return r.status === 429;  // Rate limited or timeout is OK
      }
    },
  });
  
  errorRate.add(res.status >= 500);
  responseTime.add(res.timings.duration);
}

function testROIStats(headers) {
  const res = http.get(`${BASE_URL}${API_PREFIX}/roi/stats`, { headers, timeout: '10s' });
  
  check(res, {
    'ROI stats status 200': (r) => r.status === 200,
    'ROI stats has data': (r) => {
      try {
        const data = r.json();
        return data?.total_devices !== undefined || r.status === 429;
      } catch (e) {
        return r.status === 429;
      }
    },
  });
  
  errorRate.add(res.status >= 500);
  responseTime.add(res.timings.duration);
}

function testHealthCheck() {
  const res = http.get(`${BASE_URL}/health`);
  
  check(res, {
    'health check status 200': (r) => r.status === 200,
    'health check has status': (r) => r.json('status') === 'healthy',
  });
  
  responseTime.add(res.timings.duration);
}

// Teardown: cleanup test data
export function teardown() {
  console.log('Performance test completed');
}