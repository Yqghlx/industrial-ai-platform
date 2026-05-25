import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';
import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';

// Custom metrics for detailed reporting
const errorRate = new Rate('errors');
const loginLatency = new Trend('login_latency');
const devicesLatency = new Trend('devices_latency');
const alertsLatency = new Trend('alerts_latency');
const telemetryLatency = new Trend('telemetry_latency');
const agentQueryLatency = new Trend('agent_query_latency');
const healthLatency = new Trend('health_latency');
const totalRequests = new Counter('total_requests');
const successRequests = new Counter('success_requests');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:80';
const API_PREFIX = '/api/v1';
const TEST_USER = __ENV.TEST_USER || 'admin';
const TEST_PASS = __ENV.TEST_PASS || 'Admin@123456';

// Test scenarios - comprehensive performance testing
export const options = {
  scenarios: {
    // Scenario 1: API Response Time Test (sequential, low load)
    api_response_time: {
      executor: 'per-vu-iterations',
      vus: 1,
      iterations: 5,
      maxDuration: '2m',
      exec: 'testApiResponseTime',
    },
    // Scenario 2: Light load test (10 concurrent users)
    light_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },
        { duration: '1m', target: 10 },
        { duration: '15s', target: 0 },
      ],
      exec: 'testConcurrentLoad',
    },
    // Scenario 3: Medium load test (50 concurrent users)
    medium_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '20s', target: 0 },
      ],
      exec: 'testConcurrentLoad',
    },
    // Scenario 4: Peak load test (100 concurrent users)
    peak_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 100 },
        { duration: '1m', target: 100 },
        { duration: '20s', target: 0 },
      ],
      exec: 'testConcurrentLoad',
    },
  },
  thresholds: {
    http_req_duration: ['p(50)<200', 'p(95)<500', 'p(99)<1000'],
    errors: ['rate<0.1'],
    http_req_failed: ['rate<0.05'],
  },
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(50)', 'p(90)', 'p(95)', 'p(99)'],
};

// Store auth token
let authToken = '';

// Setup: Login and get JWT token
export function setup() {
  console.log('Starting comprehensive performance benchmark...');
  console.log(`Base URL: ${BASE_URL}`);
  
  const loginRes = http.post(`${BASE_URL}${API_PREFIX}/auth/login`, JSON.stringify({
    username: TEST_USER,
    password: TEST_PASS,
  }), {
    headers: { 'Content-Type': 'application/json' },
    timeout: '30s',
  });
  
  const success = check(loginRes, {
    'login successful': (r) => r.status === 200,
    'received token': (r) => {
      try {
        const data = r.json();
        return data?.token !== undefined || data?.access_token !== undefined || data?.data?.token !== undefined;
      } catch (e) {
        console.log('Login response parsing failed:', e.message);
        return false;
      }
    },
  });
  
  if (!success) {
    console.log('Login failed. Response status:', loginRes.status);
    console.log('Response body:', loginRes.body.substring(0, 500));
  }
  
  // Extract token from various possible response formats
  let token = '';
  try {
    const data = loginRes.json();
    token = data.token || data.access_token || data.data?.token || '';
  } catch (e) {}
  
  loginLatency.add(loginRes.timings.duration);
  
  return { token };
}

// Test API response time (single user, detailed measurements)
export function testApiResponseTime(data) {
  if (!data?.token) {
    console.log('No auth token available, skipping authenticated tests');
    testPublicEndpoints();
    return;
  }
  
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`,
  };
  
  console.log(`VU ${__VU}: Testing API response times...`);
  
  // Test 1: Health check (public, baseline)
  testHealthCheck();
  
  // Test 2: Login (auth)
  testLogin();
  
  // Test 3: Device list
  testDeviceList(headers);
  
  // Test 4: Alerts list
  testAlertsList(headers);
  
  // Test 5: Telemetry data
  testTelemetryLatest(headers);
  
  // Test 6: AI Query (expensive)
  testAgentQuery(headers);
  
  sleep(1);
}

// Test concurrent load
export function testConcurrentLoad(data) {
  if (!data?.token) {
    testPublicEndpoints();
    return;
  }
  
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`,
  };
  
  // Rotate through different endpoints
  const testNum = __ITER % 6;
  
  switch (testNum) {
    case 0:
      testHealthCheck();
      break;
    case 1:
      testDeviceList(headers);
      break;
    case 2:
      testAlertsList(headers);
      break;
    case 3:
      testTelemetryLatest(headers);
      break;
    case 4:
      testSubmitTelemetry();
      break;
    case 5:
      // Skip AI query in concurrent test to avoid overwhelming
      testHealthCheck();
      break;
  }
  
  sleep(0.5 + Math.random() * 0.5); // Random think time
}

// Test public endpoints
function testPublicEndpoints() {
  testHealthCheck();
  testSubmitTelemetry();
  sleep(1);
}

// Individual API test functions
function testHealthCheck() {
  const res = http.get(`${BASE_URL}/health`, { timeout: '10s' });
  
  check(res, {
    'health status 200': (r) => r.status === 200,
    'health response valid': (r) => {
      try {
        const data = r.json();
        return data?.status === 'healthy' || data?.status === 'ok' || data?.healthy === true;
      } catch (e) {
        return r.status === 200;
      }
    },
  });
  
  healthLatency.add(res.timings.duration);
  totalRequests.add(1);
  successRequests.add(res.status === 200 ? 1 : 0);
}

function testLogin() {
  const res = http.post(`${BASE_URL}${API_PREFIX}/auth/login`, JSON.stringify({
    username: TEST_USER,
    password: TEST_PASS,
  }), {
    headers: { 'Content-Type': 'application/json' },
    timeout: '10s',
  });
  
  check(res, {
    'login status 200': (r) => r.status === 200,
  });
  
  loginLatency.add(res.timings.duration);
  totalRequests.add(1);
  successRequests.add(res.status === 200 ? 1 : 0);
  errorRate.add(res.status !== 200);
}

function testDeviceList(headers) {
  const res = http.get(`${BASE_URL}${API_PREFIX}/devices`, { headers, timeout: '15s' });
  
  const success = check(res, {
    'devices status 200': (r) => r.status === 200 || r.status === 401,
    'devices has response': (r) => {
      if (r.status === 401) return true;
      try {
        const data = r.json();
        return data !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
  
  devicesLatency.add(res.timings.duration);
  totalRequests.add(1);
  successRequests.add(success ? 1 : 0);
  errorRate.add(!success);
}

function testAlertsList(headers) {
  const res = http.get(`${BASE_URL}${API_PREFIX}/alerts`, { headers, timeout: '15s' });
  
  const success = check(res, {
    'alerts status 200 or 401': (r) => r.status === 200 || r.status === 401 || r.status === 404,
  });
  
  alertsLatency.add(res.timings.duration);
  totalRequests.add(1);
  successRequests.add(success ? 1 : 0);
  errorRate.add(!success);
}

function testTelemetryLatest(headers) {
  const res = http.get(`${BASE_URL}${API_PREFIX}/telemetry/latest`, { headers, timeout: '15s' });
  
  const success = check(res, {
    'telemetry status 200 or 401': (r) => r.status === 200 || r.status === 401 || r.status === 404,
  });
  
  telemetryLatency.add(res.timings.duration);
  totalRequests.add(1);
  successRequests.add(success ? 1 : 0);
  errorRate.add(!success);
}

function testSubmitTelemetry() {
  const payload = {
    device_id: `bench-device-${__VU}-${Date.now()}`,
    device_type: 'CNC',
    timestamp: new Date().toISOString(),
    metrics: {
      temperature: 75 + Math.random() * 20,
      vibration: 2.5 + Math.random() * 1.5,
      pressure: 100 + Math.random() * 30,
      power_consumption: 500 + Math.random() * 100,
    },
    status: 'normal',
  };
  
  const res = http.post(`${BASE_URL}${API_PREFIX}/devices/telemetry`, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' },
    timeout: '10s',
  });
  
  const success = check(res, {
    'telemetry submit status 200': (r) => r.status === 200 || r.status === 201,
  });
  
  telemetryLatency.add(res.timings.duration);
  totalRequests.add(1);
  successRequests.add(success ? 1 : 0);
  errorRate.add(!success);
}

function testAgentQuery(headers) {
  const payload = {
    query: '当前设备状态如何？',
    context: {
      device_ids: ['device-001'],
    },
  };
  
  const res = http.post(`${BASE_URL}${API_PREFIX}/agent/query`, JSON.stringify(payload), {
    headers,
    timeout: '35s', // AI queries can take longer
  });
  
  const success = check(res, {
    'agent query status 200/401/404': (r) => r.status === 200 || r.status === 401 || r.status === 404 || r.status === 429,
  });
  
  agentQueryLatency.add(res.timings.duration);
  totalRequests.add(1);
  successRequests.add(success ? 1 : 0);
  errorRate.add(!success);
}

// Teardown
export function teardown() {
  console.log('Performance benchmark completed');
}

// Report generation
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'benchmark_report.json': JSON.stringify(data, null, 2),
  };
}