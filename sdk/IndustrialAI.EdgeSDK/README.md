# Industrial AI Edge SDK

Official C# SDK for Industrial AI Agent Platform edge devices.

## Features

- ✅ **Telemetry Data Submission** - Automatic batch sending with retry
- ✅ **WebSocket Real-time** - Receive alerts and configuration updates
- ✅ **Auto-registration** - Devices automatically register on first connection
- ✅ **Offline Caching** - Queue telemetry when network is unavailable
- ✅ **Heartbeat Management** - Keep connection alive automatically
- ✅ **Multi-tenant Support** - Isolated data per tenant

## Installation

### NuGet

```bash
dotnet add package IndustrialAI.EdgeSDK
```

### From Source

```bash
cd sdk/IndustrialAI.EdgeSDK
dotnet pack
```

## Quick Start

```csharp
using IndustrialAI.EdgeSDK;
using IndustrialAI.EdgeSDK.Models;

// Create SDK instance
var config = new EdgeSDKConfig
{
    BaseUrl = "http://your-platform:8080",
    WebSocketUrl = "ws://your-platform:8080/ws",
    DeviceId = "cnc-001",
    DeviceType = DeviceType.CNC,
    TenantId = "your-tenant-id",  // Optional
    TelemetryIntervalMs = 3000,
    EnableWebSocket = true
};

var sdk = new EdgeSDK(config);

// Initialize
await sdk.InitializeAsync();

// Send telemetry
await sdk.SendTelemetryAsync(new DeviceMetrics
{
    Temperature = 75.5,
    Vibration = 2.3,
    PowerConsumption = 500
});

// Handle alerts
sdk.OnAlertReceived += (sender, alert) =>
{
    Console.WriteLine($"Alert: {alert.Message}");
};

// Shutdown
await sdk.ShutdownAsync();
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `BaseUrl` | Platform API URL | `http://localhost:8080` |
| `WebSocketUrl` | WebSocket URL | `ws://localhost:8080/ws` |
| `DeviceId` | Unique device identifier | Required |
| `DeviceType` | CNC, INJ, ROB, ASM, PKG | CNC |
| `TenantId` | Multi-tenant isolation | null |
| `TelemetryIntervalMs` | Auto-send interval | 3000 |
| `MaxCacheSize` | Offline queue size | 1000 |
| `MaxRetries` | Network retry count | 3 |
| `EnableWebSocket` | Enable real-time | true |
| `EnableHeartbeat` | Enable auto-heartbeat | true |

## Device Types

| Code | Description |
|------|-------------|
| CNC | CNC Machine (数控机床) |
| INJ | Injection Molder (注塑机) |
| ROB | Robot (机器人) |
| ASM | Assembly Line (组装线) |
| PKG | Packaging Machine (包装机) |

## Events

### OnAlertReceived

```csharp
sdk.OnAlertReceived += (sender, alert) =>
{
    Console.WriteLine($"[{alert.Severity}] {alert.RuleName}: {alert.Message}");
};
```

### OnMessageReceived

```csharp
sdk.OnMessageReceived += (sender, message) =>
{
    Console.WriteLine($"Message: {message.Type}");
};
```

### OnConnectionChanged

```csharp
sdk.OnConnectionChanged += (sender, isConnected) =>
{
    Console.WriteLine($"WebSocket: {isConnected}");
};
```

## Advanced Usage

### Manual Telemetry Control

```csharp
// Disable auto-send
config.TelemetryIntervalMs = 0;

// Send manually
await sdk.SendTelemetryAsync(metrics);

// Flush cached data
await sdk.FlushAsync();
```

### Custom Metrics

```csharp
var metrics = new DeviceMetrics
{
    Temperature = 75.5,
    Custom = new Dictionary<string, double>
    {
        ["spindle_speed"] = 12000,
        ["tool_wear"] = 0.15
    }
};
```

### Status Reporting

```csharp
await sdk.SendTelemetryAsync(metrics, DeviceStatus.Warning);
await sdk.SendTelemetryAsync(metrics, DeviceStatus.Fault);
```

## Error Handling

The SDK automatically handles:
- Network failures → Retry with exponential backoff
- Offline scenarios → Cache and send when online
- WebSocket disconnection → Auto-reconnect (5 attempts)

## Example Project

See `sdk/Examples/SimpleDevice` for a complete working example.

## License

MIT License

---

**Version**: 1.0.0  
**Platform**: Industrial AI Agent Platform  
**Author**: Industrial AI Team