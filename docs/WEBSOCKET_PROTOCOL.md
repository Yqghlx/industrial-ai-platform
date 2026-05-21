# WebSocket 协议文档

本文档详细说明工业AI代理平台的WebSocket连接协议、消息格式、心跳机制和重连策略。

## 连接信息

### WebSocket 端点

```
ws://localhost:8080/ws          # 开发环境
wss://your-domain.com/ws       # 生产环境（推荐使用WSS）
```

### 连接认证

WebSocket连接支持两种认证方式：

#### 1. 查询参数认证（推荐）

```javascript
const ws = new WebSocket(`wss://your-domain.com/ws?token=${accessToken}`);
```

#### 2. 首条消息认证

连接后立即发送认证消息：

```json
{
  "type": "auth",
  "payload": {
    "token": "your_access_token"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 消息格式

### 基础消息结构

所有WebSocket消息都遵循以下JSON结构：

```typescript
interface WSMessage {
  type: string;           // 消息类型
  payload: any;           // 消息负载
  timestamp: string;      // ISO 8601 时间戳
}
```

示例：
```json
{
  "type": "telemetry",
  "payload": {
    "device_id": "DEV001",
    "temperature": 75.5,
    "pressure": 1.2
  },
  "timestamp": "2024-01-15T10:30:00.123Z"
}
```

## 消息类型

### 1. 连接消息

#### 连接成功 (服务端 → 客户端)

```json
{
  "type": "connected",
  "payload": {
    "message": "WebSocket connected",
    "compression": true
  },
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

字段说明：
- `compression`: 是否启用了消息压缩

#### 认证成功 (服务端 → 客户端)

```json
{
  "type": "auth_success",
  "payload": {
    "user_id": 1,
    "tenant_id": "tenant-uuid",
    "session_id": "session-uuid"
  },
  "timestamp": "2024-01-15T10:30:01.000Z"
}
```

#### 认证失败 (服务端 → 客户端)

```json
{
  "type": "auth_failed",
  "payload": {
    "error": "Invalid or expired token",
    "code": "TOKEN_INVALID"
  },
  "timestamp": "2024-01-15T10:30:01.000Z"
}
```

### 2. 心跳消息

#### Ping (服务端 → 客户端)

服务端每30秒发送一次心跳ping：

```json
{
  "type": "ping",
  "payload": null,
  "timestamp": "2024-01-15T10:30:30.000Z"
}
```

#### Pong (客户端 → 服务端)

客户端收到ping后应立即响应pong：

```json
{
  "type": "pong",
  "payload": null,
  "timestamp": "2024-01-15T10:30:30.100Z"
}
```

### 3. 遥测数据消息

#### 实时遥测数据 (服务端 → 客户端)

```json
{
  "type": "telemetry",
  "payload": {
    "device_id": "DEV001",
    "device_name": "CNC加工中心-1号",
    "tenant_id": "tenant-uuid",
    "timestamp": "2024-01-15T10:30:00.123Z",
    "temperature": 75.5,
    "pressure": 1.2,
    "vibration": 0.8,
    "humidity": 45.2,
    "power": 15.6,
    "status": "normal"
  },
  "timestamp": "2024-01-15T10:30:00.123Z"
}
```

字段说明：
- `device_id`: 设备唯一标识
- `device_name`: 设备名称
- `tenant_id`: 租户ID（用于多租户隔离）
- `timestamp`: 数据采集时间
- `temperature`: 温度（℃）
- `pressure`: 压力（MPa）
- `vibration`: 振动（mm/s）
- `humidity`: 湿度（%）
- `power`: 功率（kW）
- `status`: 设备运行状态 (normal/warning/error/maintenance)

#### 批量遥测数据 (服务端 → 客户端)

```json
{
  "type": "telemetry_batch",
  "payload": {
    "device_id": "DEV001",
    "data": [
      {
        "timestamp": "2024-01-15T10:30:00.000Z",
        "temperature": 75.5,
        "pressure": 1.2
      },
      {
        "timestamp": "2024-01-15T10:30:01.000Z",
        "temperature": 76.0,
        "pressure": 1.3
      }
    ]
  },
  "timestamp": "2024-01-15T10:30:01.000Z"
}
```

### 4. 告警消息

#### 新告警 (服务端 → 客户端)

```json
{
  "type": "alert",
  "payload": {
    "id": 123,
    "rule_id": 1,
    "device_id": "DEV001",
    "device_name": "CNC加工中心-1号",
    "tenant_id": "tenant-uuid",
    "message": "温度超过阈值：当前85.2℃，阈值80℃",
    "severity": "warning",
    "status": "active",
    "triggered_at": "2024-01-15T10:30:00.000Z"
  },
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

告警级别：
- `info`: 信息
- `warning`: 警告
- `critical`: 严重
- `emergency`: 紧急

#### 告警状态更新 (服务端 → 客户端)

```json
{
  "type": "alert_update",
  "payload": {
    "id": 123,
    "status": "resolved",
    "resolved_at": "2024-01-15T10:45:00.000Z"
  },
  "timestamp": "2024-01-15T10:45:00.000Z"
}
```

告警状态：
- `active`: 活动中
- `acknowledged`: 已确认
- `resolved`: 已解决

### 5. 通知消息

#### 系统通知 (服务端 → 客户端)

```json
{
  "type": "notification",
  "payload": {
    "id": 456,
    "type": "system",
    "title": "系统维护通知",
    "message": "系统将于今晚23:00-次日02:00进行维护",
    "device_id": null,
    "tenant_id": "tenant-uuid",
    "read": false,
    "created_at": "2024-01-15T09:00:00.000Z"
  },
  "timestamp": "2024-01-15T09:00:00.000Z"
}
```

通知类型：
- `system`: 系统通知
- `device`: 设备通知
- `alert`: 告警通知
- `work_order`: 工单通知

### 6. 设备状态消息

#### 设备状态变更 (服务端 → 客户端)

```json
{
  "type": "device_status",
  "payload": {
    "device_id": "DEV001",
    "device_name": "CNC加工中心-1号",
    "previous_status": "online",
    "current_status": "offline",
    "changed_at": "2024-01-15T10:30:00.000Z",
    "reason": "连接丢失"
  },
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

设备状态：
- `online`: 在线
- `offline`: 离线
- `maintenance`: 维护中
- `error`: 故障

### 7. 工单消息

#### 新工单 (服务端 → 客户端)

```json
{
  "type": "work_order",
  "payload": {
    "id": 789,
    "title": "CNC加工中心-1号 维护工单",
    "description": "温度异常，需要检查冷却系统",
    "device_id": "DEV001",
    "tenant_id": "tenant-uuid",
    "priority": "high",
    "status": "pending",
    "assigned_to": 10,
    "created_at": "2024-01-15T10:35:00.000Z"
  },
  "timestamp": "2024-01-15T10:35:00.000Z"
}
```

工单状态：
- `pending`: 待处理
- `in_progress`: 进行中
- `completed`: 已完成
- `cancelled`: 已取消

工单优先级：
- `low`: 低
- `medium`: 中
- `high`: 高
- `urgent`: 紧急

### 8. 订阅管理消息

#### 订阅请求 (客户端 → 服务端)

```json
{
  "type": "subscribe",
  "payload": {
    "channels": ["telemetry", "alerts", "notifications"],
    "device_ids": ["DEV001", "DEV002"]
  },
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

可用频道：
- `telemetry`: 遥测数据
- `alerts`: 告警消息
- `notifications`: 系统通知
- `device_status`: 设备状态变更
- `work_orders`: 工单消息

#### 订阅确认 (服务端 → 客户端)

```json
{
  "type": "subscribed",
  "payload": {
    "channels": ["telemetry", "alerts", "notifications"],
    "device_ids": ["DEV001", "DEV002"],
    "message": "Successfully subscribed to channels"
  },
  "timestamp": "2024-01-15T10:30:00.100Z"
}
```

#### 取消订阅 (客户端 → 服务端)

```json
{
  "type": "unsubscribe",
  "payload": {
    "channels": ["telemetry"],
    "device_ids": ["DEV001"]
  },
  "timestamp": "2024-01-15T10:35:00.000Z"
}
```

### 9. 错误消息

#### WebSocket错误 (服务端 → 客户端)

```json
{
  "type": "error",
  "payload": {
    "code": "INVALID_MESSAGE_FORMAT",
    "message": "Invalid JSON message format",
    "details": "Expected type field"
  },
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

WebSocket错误码：

| 错误码 | 描述 |
|--------|------|
| `INVALID_MESSAGE_FORMAT` | 消息格式无效 |
| `UNKNOWN_MESSAGE_TYPE` | 未知的消息类型 |
| `UNAUTHORIZED` | 未授权访问 |
| `TOKEN_EXPIRED` | Token已过期 |
| `SUBSCRIPTION_LIMIT_EXCEEDED` | 订阅数量超限 |
| `RATE_LIMITED` | 消息发送过于频繁 |

## 消息压缩

### 压缩机制

当消息大小超过1024字节时，服务端会自动使用gzip压缩消息。客户端应能处理压缩和非压缩两种消息格式。

#### 压缩消息格式

压缩后的消息以二进制帧发送，格式如下：

```
[1 byte: 标志位 0x1F] [压缩数据]
```

#### 解压缩流程

```javascript
// 解压缩示例
ws.onmessage = async (event) => {
  if (event.data instanceof Blob) {
    const buffer = await event.data.arrayBuffer();
    const view = new DataView(buffer);
    const flag = view.getUint8(0);
    
    if (flag === 0x1F) {
      // 压缩消息
      const compressed = new Uint8Array(buffer, 1);
      const decompressed = await decompressGzip(compressed);
      const message = JSON.parse(new TextDecoder().decode(decompressed));
      handleMessage(message);
    } else {
      // 普通消息
      const message = JSON.parse(await event.data.text());
      handleMessage(message);
    }
  }
};
```

## 心跳机制

### 心跳配置

| 参数 | 值 | 说明 |
|------|------|------|
| 心跳间隔 | 30秒 | 服务端发送ping的间隔 |
| 心跳超时 | 60秒 | 未收到pong响应的超时时间 |
| 最大重连次数 | 5次 | 连续重连失败的最大次数 |

### 心跳流程

```
服务端                              客户端
   |                                  |
   |-------- ping (30s间隔) --------->|
   |                                  |
   |<------- pong (立即响应) ---------|
   |                                  |
   |  如果60秒未收到pong:             |
   |  关闭连接                         |
   |                                  |
```

## 重连策略

### 指数退避重连

客户端应实现指数退避重连策略：

```javascript
class WebSocketClient {
  constructor(url) {
    this.url = url;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.baseDelay = 1000; // 1秒
    this.maxDelay = 30000; // 30秒
  }
  
  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onclose = (event) => {
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        const delay = Math.min(
          this.baseDelay * Math.pow(2, this.reconnectAttempts),
          this.maxDelay
        );
        setTimeout(() => this.connect(), delay);
        this.reconnectAttempts++;
      }
    };
    
    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
    };
  }
}
```

### 重连延迟表

| 重连次数 | 延迟时间 |
|----------|----------|
| 1 | 1秒 |
| 2 | 2秒 |
| 3 | 4秒 |
| 4 | 8秒 |
| 5 | 16秒 |

### 重连最佳实践

1. **保存订阅状态**: 断线前保存当前订阅，重连后重新订阅
2. **消息队列**: 断线期间缓存未发送的消息，重连后发送
3. **断线提示**: 向用户显示连接状态
4. **心跳检测**: 主动检测心跳超时并触发重连

## 连接状态

### 状态流转图

```
    CONNECTING
         |
         v
    AUTHENTICATING --[失败]--> CLOSED
         |
         v
    CONNECTED
         |
         v
    [心跳超时/错误] --[重连]--> RECONNECTING
         |                           |
         v                           v
    CLOSED                    CONNECTING
```

### 状态说明

| 状态 | 说明 |
|------|------|
| `CONNECTING` | 正在建立WebSocket连接 |
| `AUTHENTICATING` | 连接已建立，正在认证 |
| `CONNECTED` | 已认证，可正常收发消息 |
| `RECONNECTING` | 正在重新连接 |
| `CLOSED` | 连接已关闭 |

## 限流规则

| 消息类型 | 限制 | 窗口期 |
|----------|------|--------|
| 客户端发送消息 | 100条 | 1分钟 |
| 订阅请求 | 10次 | 1分钟 |
| 单连接订阅频道数 | 10个 | - |
| 单IP连接数 | 10个 | - |

## 安全考虑

### 1. Token安全
- 使用WSS加密传输
- Token有效期为24小时
- Token过期后需重新认证

### 2. 消息验证
- 所有消息必须符合JSON格式
- 消息大小限制为1MB
- 无效消息将被丢弃并记录日志

### 3. 租户隔离
- 每个连接绑定到特定租户
- 只能接收该租户的消息
- 跨租户订阅将被拒绝

## 客户端实现示例

### JavaScript/TypeScript 完整示例

```typescript
interface WSMessage {
  type: string;
  payload: any;
  timestamp: string;
}

class IndustrialAIWebSocket {
  private ws: WebSocket | null = null;
  private url: string;
  private token: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private subscriptions: { channels: string[]; deviceIds: string[] } = {
    channels: [],
    deviceIds: []
  };
  private handlers: Map<string, (payload: any) => void> = new Map();
  private pingTimeout: NodeJS.Timeout | null = null;

  constructor(baseUrl: string, token: string) {
    this.url = `${baseUrl}/ws?token=${token}`;
    this.token = token;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.startHeartbeat();
        resolve();
      };

      this.ws.onmessage = async (event) => {
        const message = await this.parseMessage(event.data);
        this.handleMessage(message);
      };

      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        this.stopHeartbeat();
        this.handleReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        reject(error);
      };
    });
  }

  private async parseMessage(data: Blob | string): Promise<WSMessage> {
    if (typeof data === 'string') {
      return JSON.parse(data);
    }
    
    const buffer = await data.arrayBuffer();
    const view = new DataView(buffer);
    const flag = view.getUint8(0);

    if (flag === 0x1F) {
      const compressed = new Uint8Array(buffer, 1);
      const decompressed = await this.decompress(compressed);
      return JSON.parse(new TextDecoder().decode(decompressed));
    }

    return JSON.parse(new TextDecoder().decode(buffer));
  }

  private async decompress(data: Uint8Array): Promise<Uint8Array> {
    const ds = new DecompressionStream('gzip');
    const writer = ds.writable.getWriter();
    writer.write(data);
    writer.close();
    return new Uint8Array(await new Response(ds.readable).arrayBuffer());
  }

  private handleMessage(message: WSMessage) {
    switch (message.type) {
      case 'ping':
        this.send({ type: 'pong', payload: null, timestamp: new Date().toISOString() });
        break;
      case 'telemetry':
      case 'alert':
      case 'notification':
      case 'device_status':
      case 'work_order':
        const handler = this.handlers.get(message.type);
        if (handler) handler(message.payload);
        break;
    }
  }

  private startHeartbeat() {
    this.pingTimeout = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.send({ type: 'ping', payload: null, timestamp: new Date().toISOString() });
      }
    }, 30000);
  }

  private stopHeartbeat() {
    if (this.pingTimeout) {
      clearInterval(this.pingTimeout);
      this.pingTimeout = null;
    }
  }

  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
      setTimeout(() => {
        this.reconnectAttempts++;
        this.connect().then(() => {
          // 重新订阅
          if (this.subscriptions.channels.length > 0) {
            this.subscribe(this.subscriptions.channels, this.subscriptions.deviceIds);
          }
        });
      }, delay);
    }
  }

  subscribe(channels: string[], deviceIds: string[] = []) {
    this.subscriptions = { channels, deviceIds };
    this.send({
      type: 'subscribe',
      payload: { channels, device_ids: deviceIds },
      timestamp: new Date().toISOString()
    });
  }

  on(type: string, handler: (payload: any) => void) {
    this.handlers.set(type, handler);
  }

  off(type: string) {
    this.handlers.delete(type);
  }

  private send(message: WSMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }

  close() {
    this.stopHeartbeat();
    this.ws?.close();
  }
}

// 使用示例
const ws = new IndustrialAIWebSocket('wss://your-domain.com', 'your-token');

ws.on('telemetry', (data) => {
  console.log('Telemetry update:', data);
});

ws.on('alert', (data) => {
  console.log('New alert:', data);
});

await ws.connect();
ws.subscribe(['telemetry', 'alerts'], ['DEV001']);
```

## 版本历史

| 版本 | 日期 | 变更说明 |
|------|------|----------|
| 1.0.0 | 2024-01-15 | 初始版本 |