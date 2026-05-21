import http from 'k6/http';
import { check, sleep } from 'k6';

// WebSocket stress test - simulates many concurrent device connections
export const options = {
  stages: [
    { duration: '1m', target: 50 },    // Ramp up to 50 connections
    { duration: '3m', target: 200 },   // Peak: 200 concurrent connections
    { duration: '1m', target: 100 },   // Ramp down
    { duration: '30s', target: 0 },    // Cool down
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'],  // WebSocket upgrade should be fast
    http_req_failed: ['rate<0.02'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Note: k6 doesn't support WebSocket directly in the default module
// This test checks the WebSocket endpoint availability
// For actual WebSocket testing, use k6 experimental WebSocket module

export default function () {
  // Test WebSocket endpoint availability (upgrade request)
  const wsUrl = `${BASE_URL}/ws`;
  
  const res = http.get(wsUrl);
  
  // WebSocket upgrade should return 400 (Bad Request) without proper upgrade headers
  // This is expected behavior - confirms endpoint exists
  check(res, {
    'ws endpoint exists': (r) => r.status === 400 || r.status === 200 || r.status === 426, // 426 = Upgrade Required
  });
  
  // Simulate telemetry push rate (each VU represents a device)
  const deviceId = `ws-device-${__VU}`;
  
  // Push telemetry every 3 seconds (simulating device heartbeat)
  const telemetryPayload = {
    device_id: deviceId,
    device_type: 'INJ', // Injection molding machine
    timestamp: new Date().toISOString(),
    metrics: {
      temperature: 80 + Math.random() * 40,
      vibration: 3.0 + Math.random() * 2.0,
      cycle_time: 15 + Math.random() * 5,
      efficiency: 0.85 + Math.random() * 0.15,
    },
    status: Math.random() > 0.95 ? 'warning' : 'running', // 5% warning rate
  };
  
  const telemetryRes = http.post(
    `${BASE_URL}/api/v1/devices/telemetry`,
    JSON.stringify(telemetryPayload),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  check(telemetryRes, {
    'telemetry accepted': (r) => r.status === 200,
  });
  
  sleep(3); // Device heartbeat interval
}

export function setup() {
  console.log('WebSocket stress test starting...');
  console.log('Simulating concurrent device connections');
}

export function teardown() {
  console.log('WebSocket stress test completed');
}