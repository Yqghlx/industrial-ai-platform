using System.Collections.Concurrent;
using Microsoft.Extensions.Logging;
using IndustrialAI.EdgeSDK.Http;
using IndustrialAI.EdgeSDK.Models;
using IndustrialAI.EdgeSDK.WebSocket;

namespace IndustrialAI.EdgeSDK;

/// <summary>
/// Industrial AI Edge SDK 主入口
/// </summary>
public class EdgeSDK : IDisposable
{
    private readonly EdgeSDKConfig _config;
    private readonly ApiClient _apiClient;
    private readonly WSClient? _wsClient;
    private readonly ILogger<EdgeSDK>? _logger;
    
    private readonly ConcurrentQueue<TelemetryData> _telemetryCache;
    private readonly Timer? _telemetryTimer;
    private readonly Timer? _heartbeatTimer;
    
    private bool _disposed;
    private Device? _device;
    
    /// <summary>
    /// SDK 是否已初始化
    /// </summary>
    public bool IsInitialized { get; private set; }
    
    /// <summary>
    /// WebSocket 是否已连接
    /// </summary>
    public bool IsWebSocketConnected => _wsClient?.IsConnected ?? false;
    
    /// <summary>
    /// 缓存的遥测数据数量
    /// </summary>
    public int CachedTelemetryCount => _telemetryCache.Count;
    
    // 事件
    public event EventHandler<Alert>? OnAlertReceived;
    public event EventHandler<WSMessage>? OnMessageReceived;
    public event EventHandler<bool>? OnConnectionChanged;
    
    /// <summary>
    /// 创建 SDK 实例
    /// </summary>
    public EdgeSDK(EdgeSDKConfig config, ILogger<EdgeSDK>? logger = null)
    {
        _config = config;
        _logger = logger;
        _telemetryCache = new ConcurrentQueue<TelemetryData>();
        
        // 创建 HTTP 客户端
        var httpClient = new HttpClient
        {
            BaseAddress = new Uri(config.BaseUrl)
        };
        _apiClient = new ApiClient(httpClient, logger);
        
        // 创建 WebSocket 客户端 (可选)
        if (config.EnableWebSocket)
        {
            _wsClient = new WSClient(config.WebSocketUrl, logger);
            _wsClient.OnAlertReceived += (s, a) => OnAlertReceived?.Invoke(this, a);
            _wsClient.OnMessageReceived += (s, m) => OnMessageReceived?.Invoke(this, m);
            _wsClient.OnConnectionChanged += (s, c) => OnConnectionChanged?.Invoke(this, c);
        }
    }
    
    /// <summary>
    /// 初始化 SDK
    /// </summary>
    public async Task<bool> InitializeAsync()
    {
        try
        {
            _logger?.LogInformation("Initializing Edge SDK for device {DeviceId}", _config.DeviceId);
            
            // 检查平台健康状态
            if (!await _apiClient.CheckHealthAsync())
            {
                _logger?.LogWarning("Platform health check failed, proceeding anyway");
            }
            
            // 注册/获取设备
            var existingDevice = await _apiClient.GetDeviceAsync(_config.DeviceId);
            if (existingDevice?.Success == true && existingDevice.Data != null)
            {
                _device = existingDevice.Data;
                _logger?.LogInformation("Device {DeviceId} already registered", _config.DeviceId);
            }
            else
            {
                // 自动注册设备
                var newDevice = new Device
                {
                    Id = _config.DeviceId,
                    Name = $"{_config.DeviceType} Device {_config.DeviceId}",
                    Type = _config.DeviceType.ToString(),
                    Status = "running",
                    TenantId = _config.TenantId
                };
                
                var registered = await _apiClient.RegisterDeviceAsync(newDevice);
                if (registered?.Success == true)
                {
                    _device = registered.Data;
                    _logger?.LogInformation("Device {DeviceId} registered successfully", _config.DeviceId);
                }
                else
                {
                    _logger?.LogWarning("Device registration failed: {Error}", registered?.Error);
                    // 使用本地设备信息
                    _device = newDevice;
                }
            }
            
            // 连接 WebSocket
            if (_config.EnableWebSocket && _wsClient != null)
            {
                await _wsClient.ConnectAsync();
            }
            
            // 启动遥测上报定时器
            if (_config.TelemetryIntervalMs > 0)
            {
                _telemetryTimer = new Timer(
                    async _ => await SendCachedTelemetryAsync(),
                    null,
                    _config.TelemetryIntervalMs,
                    _config.TelemetryIntervalMs);
            }
            
            // 启动心跳定时器
            if (_config.EnableHeartbeat)
            {
                _heartbeatTimer = new Timer(
                    async _ => await SendHeartbeatAsync(),
                    null,
                    30000,
                    30000);
            }
            
            IsInitialized = true;
            _logger?.LogInformation("Edge SDK initialized successfully");
            
            return true;
        }
        catch (Exception ex)
        {
            _logger?.LogError(ex, "SDK initialization failed");
            return false;
        }
    }
    
    /// <summary>
    /// 发送遥测数据
    /// </summary>
    public async Task SendTelemetryAsync(TelemetryData data)
    {
        if (!IsInitialized)
        {
            _logger?.LogWarning("SDK not initialized, caching telemetry");
            CacheTelemetry(data);
            return;
        }
        
        // 先缓存，防止网络问题
        CacheTelemetry(data);
        
        // 立即尝试发送
        await SendCachedTelemetryAsync();
    }
    
    /// <summary>
    /// 发送遥测数据 (简化接口)
    /// </summary>
    public async Task SendTelemetryAsync(DeviceMetrics metrics, DeviceStatus status = DeviceStatus.Running)
    {
        var data = new TelemetryData
        {
            DeviceId = _config.DeviceId,
            DeviceType = _config.DeviceType.ToString(),
            Timestamp = DateTime.UtcNow,
            Metrics = metrics,
            Status = status.ToString().ToLower()
        };
        
        await SendTelemetryAsync(data);
    }
    
    /// <summary>
    /// 缓存遥测数据
    /// </summary>
    private void CacheTelemetry(TelemetryData data)
    {
        _telemetryCache.Enqueue(data);
        
        // 限制缓存大小
        while (_telemetryCache.Count > _config.MaxCacheSize)
        {
            _telemetryCache.TryDequeue(out _);
        }
    }
    
    /// <summary>
    /// 发送缓存的遥测数据
    /// </summary>
    private async Task SendCachedTelemetryAsync()
    {
        if (_telemetryCache.IsEmpty) return;
        
        // 批量取出
        var batch = new List<TelemetryData>();
        while (_telemetryCache.TryDequeue(out var data) && batch.Count < 50)
        {
            batch.Add(data);
        }
        
        if (batch.Count == 1)
        {
            await _apiClient.SendTelemetryAsync(batch[0]);
        }
        else if (batch.Count > 1)
        {
            await _apiClient.SendTelemetryBatchAsync(batch);
        }
        
        _logger?.LogDebug("Sent {Count} telemetry data points", batch.Count);
    }
    
    /// <summary>
    /// 发送心跳
    /// </summary>
    private async Task SendHeartbeatAsync()
    {
        if (!IsInitialized) return;
        
        var heartbeat = new TelemetryData
        {
            DeviceId = _config.DeviceId,
            DeviceType = _config.DeviceType.ToString(),
            Timestamp = DateTime.UtcNow,
            Status = "running",
            Metrics = new DeviceMetrics() // 空指标表示心跳
        };
        
        await _apiClient.SendTelemetryAsync(heartbeat);
        _logger?.LogDebug("Heartbeat sent");
    }
    
    /// <summary>
    /// 手动发送所有缓存数据
    /// </summary>
    public async Task FlushAsync()
    {
        await SendCachedTelemetryAsync();
    }
    
    /// <summary>
    /// 获取设备信息
    /// </summary>
    public Device? GetDevice()
    {
        return _device;
    }
    
    /// <summary>
    /// 关闭 SDK
    /// </summary>
    public async Task ShutdownAsync()
    {
        _logger?.LogInformation("Shutting down Edge SDK");
        
        // 发送剩余数据
        await FlushAsync();
        
        // 断开 WebSocket
        if (_wsClient != null)
        {
            await _wsClient.DisconnectAsync();
        }
        
        // 停止定时器
        _telemetryTimer?.Dispose();
        _heartbeatTimer?.Dispose();
        
        IsInitialized = false;
    }
    
    public void Dispose()
    {
        if (_disposed) return;
        
        ShutdownAsync().Wait();
        
        _wsClient?.Dispose();
        _telemetryTimer?.Dispose();
        _heartbeatTimer?.Dispose();
        
        _disposed = true;
    }
}