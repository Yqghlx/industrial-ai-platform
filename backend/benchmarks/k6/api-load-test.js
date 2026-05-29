import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 5 },   // 起始5用户（降低并发）
    { duration: '1m', target: 10 },    // 逐步增加到10
    { duration: '2m', target: 20 },    // 中等并发20
    { duration: '1m', target: 10 },    // 降压测试
    { duration: '30s', target: 0 },    // 结束
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 目标p95<500ms
    http_req_failed: ['rate<0.05'],   // 目标失败率<5%
  },
};

const BASE_URL = 'http://localhost:8080';

// 每个VU独立获取token
export default function () {
  // 在每次迭代开始时获取新token
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, {
    username: 'admin',
    password: 'Admin@123456',
  });
  
  const body = loginRes.json();
  const token = body?.token || body?.data?.token || '';
  
  const headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };

  // 测试健康检查接口
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health status is 200': (r) => r.status === 200,
  });

  sleep(2);  // 增加sleep时间，减少请求频率

  // 测试设备列表接口（带认证）
  const devicesRes = http.get(`${BASE_URL}/api/v1/devices`, { headers: headers });
  check(devicesRes, {
    'devices status is 200': (r) => r.status === 200,
    'devices has data': (r) => {
      try {
        const data = r.json();
        return data && (Array.isArray(data) || Array.isArray(data?.data));
      } catch (e) {
        return false;
      }
    },
  });

  sleep(2);  // 增加sleep时间，减少请求频率

  // 测试ROI统计接口（带认证）
  const roiRes = http.get(`${BASE_URL}/api/v1/roi/stats`, { headers: headers });
  check(roiRes, {
    'roi status is 200': (r) => r.status === 200,
    'roi has valid data': (r) => {
      try {
        const data = r.json();
        return data && (typeof data.total_devices === 'number' || data.total_devices > 0);
      } catch (e) {
        return false;
      }
    },
  });

  sleep(3);  // 每轮结束增加额外sleep，确保请求频率低于300/min
}