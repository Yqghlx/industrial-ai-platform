using System.Text.Json.Serialization;

namespace IndustrialAI.EdgeSDK.Models;

/// <summary>
/// 设备类型枚举
/// </summary>
public enum DeviceType
{
    CNC,        // 数控机床
    INJ,        // 注塑机
    ROB,        // 机器人
    ASM,        // 组装线
    PKG         // 包装机
}

/// <summary>
/// 设备状态
/// </summary>
public enum DeviceStatus
{
    Running,    // 运行中
    Idle,       // 待机
    Warning,    // 警告
    Fault,      // 故障
    Offline     // 离线
}

/// <summary>
/// 设备信息
/// </summary>
public class Device
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = string.Empty;
    
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;
    
    [JsonPropertyName("type")]
    public string Type { get; set; } = string.Empty;
    
    [JsonPropertyName("location")]
    public string? Location { get; set; }
    
    [JsonPropertyName("status")]
    public string Status { get; set; } = "running";
    
    [JsonPropertyName("tenant_id")]
    public string? TenantId { get; set; }
    
    [JsonPropertyName("created_at")]
    public DateTime CreatedAt { get; set; }
}

/// <summary>
/// 遥测数据点
/// </summary>
public class TelemetryData
{
    [JsonPropertyName("device_id")]
    public string DeviceId { get; set; } = string.Empty;
    
    [JsonPropertyName("device_type")]
    public string DeviceType { get; set; } = string.Empty;
    
    [JsonPropertyName("timestamp")]
    public DateTime Timestamp { get; set; } = DateTime.UtcNow;
    
    [JsonPropertyName("metrics")]
    public DeviceMetrics Metrics { get; set; } = new();
    
    [JsonPropertyName("status")]
    public string Status { get; set; } = "running";
}

/// <summary>
/// 设备指标数据
/// </summary>
public class DeviceMetrics
{
    [JsonPropertyName("temperature")]
    public double? Temperature { get; set; }
    
    [JsonPropertyName("vibration")]
    public double? Vibration { get; set; }
    
    [JsonPropertyName("pressure")]
    public double? Pressure { get; set; }
    
    [JsonPropertyName("power_consumption")]
    public double? PowerConsumption { get; set; }
    
    [JsonPropertyName("cycle_time")]
    public double? CycleTime { get; set; }
    
    [JsonPropertyName("efficiency")]
    public double? Efficiency { get; set; }
    
    [JsonPropertyName("runtime_hours")]
    public double? RuntimeHours { get; set; }
    
    // 自定义指标
    [JsonPropertyName("custom")]
    public Dictionary<string, double>? Custom { get; set; }
}

/// <summary>
/// WebSocket 消息
/// </summary>
public class WSMessage
{
    [JsonPropertyName("type")]
    public string Type { get; set; } = string.Empty; // telemetry, alert, config, command
    
    [JsonPropertyName("device_id")]
    public string? DeviceId { get; set; }
    
    [JsonPropertyName("data")]
    public object? Data { get; set; }
    
    [JsonPropertyName("timestamp")]
    public DateTime Timestamp { get; set; } = DateTime.UtcNow;
}

/// <summary>
/// 告警信息
/// </summary>
public class Alert
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = string.Empty;
    
    [JsonPropertyName("device_id")]
    public string DeviceId { get; set; } = string.Empty;
    
    [JsonPropertyName("rule_name")]
    public string RuleName { get; set; } = string.Empty;
    
    [JsonPropertyName("severity")]
    public string Severity { get; set; } = "warning"; // info, warning, critical
    
    [JsonPropertyName("message")]
    public string Message { get; set; } = string.Empty;
    
    [JsonPropertyName("triggered_at")]
    public DateTime TriggeredAt { get; set; }
}

/// <summary>
/// API 响应包装
/// </summary>
public class ApiResponse<T>
{
    [JsonPropertyName("success")]
    public bool Success { get; set; }
    
    [JsonPropertyName("data")]
    public T? Data { get; set; }
    
    [JsonPropertyName("error")]
    public string? Error { get; set; }
    
    [JsonPropertyName("message")]
    public string? Message { get; set; }
}

/// <summary>
/// SDK 配置
/// </summary>
public class EdgeSDKConfig
{
    /// <summary>
    /// 平台 API 地址
    /// </summary>
    public string BaseUrl { get; set; } = "http://localhost:8080";
    
    /// <summary>
    /// WebSocket 地址
    /// </summary>
    public string WebSocketUrl { get; set; } = "ws://localhost:8080/ws";
    
    /// <summary>
    /// 租户 ID (多租户场景)
    /// </summary>
    public string? TenantId { get; set; }
    
    /// <summary>
    /// 设备 ID
    /// </summary>
    public string DeviceId { get; set; } = string.Empty;
    
    /// <summary>
    /// 设备类型
    /// </summary>
    public DeviceType DeviceType { get; set; } = DeviceType.CNC;
    
    /// <summary>
    /// 遥测数据上报间隔 (毫秒)
    /// </summary>
    public int TelemetryIntervalMs { get; set; } = 3000;
    
    /// <summary>
    /// 最大缓存条目数 (断网时)
    /// </summary>
    public int MaxCacheSize { get; set; } = 1000;
    
    /// <summary>
    /// 重试次数
    /// </summary>
    public int MaxRetries { get; set; } = 3;
    
    /// <summary>
    /// 重试间隔 (毫秒)
    /// </summary>
    public int RetryIntervalMs { get; set; } = 1000;
    
    /// <summary>
    /// 是否启用自动心跳
    /// </summary>
    public bool EnableHeartbeat { get; set; } = true;
    
    /// <summary>
    /// 是否启用 WebSocket
    /// </summary>
    public bool EnableWebSocket { get; set; } = true;
}