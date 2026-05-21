using Microsoft.Extensions.Logging;
using IndustrialAI.EdgeSDK;
using IndustrialAI.EdgeSDK.Models;

// 简单设备示例 - 展示 SDK 基本用法

// 创建日志
using var loggerFactory = LoggerFactory.Create(builder => builder.AddConsole());
var logger = loggerFactory.CreateLogger<EdgeSDK>();

// 配置 SDK
var config = new EdgeSDKConfig
{
    BaseUrl = Environment.GetEnvironmentVariable("PLATFORM_URL") ?? "http://localhost:8080",
    WebSocketUrl = Environment.GetEnvironmentVariable("WS_URL") ?? "ws://localhost:8080/ws",
    DeviceId = "cnc-demo-001",
    DeviceType = DeviceType.CNC,
    TenantId = Environment.GetEnvironmentVariable("TENANT_ID"),
    TelemetryIntervalMs = 3000,
    MaxCacheSize = 500,
    EnableWebSocket = true,
    EnableHeartbeat = true
};

Console.WriteLine("=== Industrial AI Edge SDK Demo ===");
Console.WriteLine($"Device: {config.DeviceId}");
Console.WriteLine($"Type: {config.DeviceType}");
Console.WriteLine($"Platform: {config.BaseUrl}");
Console.WriteLine();

// 创建 SDK 实例
var sdk = new EdgeSDK(config, logger);

// 注册事件
sdk.OnAlertReceived += (sender, alert) =>
{
    Console.WriteLine($"[ALERT] {alert.Severity}: {alert.Message}");
    Console.WriteLine($"  Device: {alert.DeviceId}");
    Console.WriteLine($"  Rule: {alert.RuleName}");
};

sdk.OnConnectionChanged += (sender, isConnected) =>
{
    Console.WriteLine($"[WS] Connection: {isConnected}");
};

// 初始化
Console.WriteLine("Initializing...");
var success = await sdk.InitializeAsync();
Console.WriteLine($"Initialized: {success}");
Console.WriteLine();

if (!success)
{
    Console.WriteLine("Failed to initialize. Check platform connection.");
    return;
}

// 模拟设备运行
Console.WriteLine("Starting telemetry simulation...");
Console.WriteLine("Press Ctrl+C to stop.");
Console.WriteLine();

var random = new Random();
var running = true;

// Ctrl+C 处理
Console.CancelKeyPress += (s, e) =>
{
    e.Cancel = true;
    running = false;
    Console.WriteLine("\nShutting down...");
};

// 发送遥测数据循环
while (running)
{
    // 生成模拟数据
    var metrics = new DeviceMetrics
    {
        Temperature = 70 + random.NextDouble() * 20,
        Vibration = 2.0 + random.NextDouble() * 1.5,
        Pressure = 100 + random.NextDouble() * 30,
        PowerConsumption = 400 + random.NextDouble() * 200,
        RuntimeHours = (DateTime.Now - DateTime.Today).TotalHours
    };
    
    // 偶尔模拟故障 (5% 概率)
    var status = random.NextDouble() < 0.05 ? DeviceStatus.Warning : DeviceStatus.Running;
    
    // 发送数据
    await sdk.SendTelemetryAsync(metrics, status);
    
    // 显示状态
    Console.WriteLine($"[{DateTime.Now:HH:mm:ss}] Temp: {metrics.Temperature:F1}°C | " +
                      $"Vib: {metrics.Vibration:F2} | " +
                      $"Power: {metrics.PowerConsumption:F0}W | " +
                      $"Status: {status}");
    
    // 等待
    await Task.Delay(1000);
}

// 关闭
await sdk.ShutdownAsync();
Console.WriteLine("SDK shutdown complete.");
Console.WriteLine($"Cached telemetry: {sdk.CachedTelemetryCount}");

sdk.Dispose();
Console.WriteLine("Done.");