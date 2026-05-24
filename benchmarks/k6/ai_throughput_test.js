import http from 'k6/http';
import { check, sleep } from 'k6';

// AI Agent throughput test - measures LLM response capacity
export const options = {
  scenarios: {
    // Scenario 1: Sequential queries (realistic usage)
    sequential_queries: {
      executor: 'constant-arrival-rate',
      rate: 5, // 5 queries per second
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 10,
      maxVUs: 20,
    },
    // Scenario 2: Burst queries (stress test)
    burst_queries: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 20,
      maxVUs: 50,
      stages: [
        { target: 20, duration: '30s' },   // Ramp to 20 QPS
        { target: 50, duration: '1m' },    // Peak: 50 QPS
        { target: 10, duration: '30s' },   // Ramp down
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<30000'], // AI queries can take up to 30s
    http_req_failed: ['rate<0.1'],       // Allow higher failure rate (rate limits)
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_PREFIX = '/api/v1';

// Test user credentials (same as api_load_test.js)
const TEST_USER = {
  username: 'k6test',
  password: 'K6Test@12345',
};

export function setup() {
  // Login to get JWT token
  const loginRes = http.post(`${BASE_URL}${API_PREFIX}/auth/login`, JSON.stringify(TEST_USER), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  if (loginRes.status !== 200) {
    console.log('Login failed, test will use mock responses');
    return { token: '' };
  }
  
  return { token: loginRes.json('data.token') || '' };
}

// Sample queries for testing
const SAMPLE_QUERIES = [
  '设备温度异常如何处理？',
  '振动数据超出阈值怎么排查？',
  '如何优化设备运行效率？',
  '预测性维护的最佳实践？',
  '故障诊断流程是什么？',
  '设备能耗优化建议？',
  '生产线瓶颈分析？',
  '质量控制方法？',
  '设备巡检注意事项？',
  '备件管理建议？',
];

export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`,
  };
  
  // Select a random query
  const queryIndex = Math.floor(Math.random() * SAMPLE_QUERIES.length);
  const query = SAMPLE_QUERIES[queryIndex];
  
  const payload = {
    query: query,
    device_ids: [`test-device-${__VU}`],
    context: {
      device_type: 'CNC',
      current_status: 'running',
    },
  };
  
  const res = http.post(`${BASE_URL}${API_PREFIX}/agent/query`, JSON.stringify(payload), { headers });
  
  // Check response (allow 429 rate limit)
  const success = check(res, {
    'agent query completed': (r) => r.status === 200 || r.status === 429,
    'has response or rate limited': (r) => {
      if (r.status === 429) return true;
      const body = r.json();
      return body && (body.data || body.success);
    },
  });
  
  if (res.status === 200) {
    console.log(`Query "${query.substring(0, 20)}..." response time: ${res.timings.duration}ms`);
  } else if (res.status === 429) {
    console.log('Rate limited - this is expected under high load');
  }
  
  sleep(1);
}

export function teardown() {
  console.log('AI Agent throughput test completed');
}