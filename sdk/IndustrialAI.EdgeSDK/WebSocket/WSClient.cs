using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Microsoft.Extensions.Logging;
using IndustrialAI.EdgeSDK.Models;

namespace IndustrialAI.EdgeSDK.WebSocket;

/// <summary>
/// WebSocket 客户端管理器
/// </summary>
public class WSClient : IDisposable
{
    private ClientWebSocket? _ws;
    private readonly string _wsUrl;
    private readonly ILogger<WSClient>? _logger;
    private readonly JsonSerializerOptions _jsonOptions;
    private readonly CancellationTokenSource _cts;
    
    private bool _isConnected;
    private bool _disposed;
    
    // 事件
    public event EventHandler<WSMessage>? OnMessageReceived;
    public event EventHandler<Alert>? OnAlertReceived;
    public event EventHandler<bool>? OnConnectionChanged;
    
    public bool IsConnected => _isConnected;
    
    public WSClient(string wsUrl, ILogger<WSClient>? logger = null)
    {
        _wsUrl = wsUrl;
        _logger = logger;
        _cts = new CancellationTokenSource();
        
        _jsonOptions = new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
            PropertyNameCaseInsensitive = true
        };
    }
    
    /// <summary>
    /// 连接 WebSocket
    /// </summary>
    public async Task ConnectAsync()
    {
        if (_isConnected) return;
        
        try
        {
            _ws = new ClientWebSocket();
            await _ws.ConnectAsync(new Uri(_wsUrl), _cts.Token);
            
            _isConnected = true;
            OnConnectionChanged?.Invoke(this, true);
            _logger?.LogInformation("WebSocket connected to {Url}", _wsUrl);
            
            // 开始接收消息
            _ = ReceiveMessagesAsync();
            
            // 开始心跳
            _ = SendHeartbeatAsync();
        }
        catch (Exception ex)
        {
            _logger?.LogError(ex, "WebSocket connection failed");
            _isConnected = false;
            OnConnectionChanged?.Invoke(this, false);
            
            // 尝试重连
            _ = ReconnectAsync();
        }
    }
    
    /// <summary>
    /// 断开连接
    /// </summary>
    public async Task DisconnectAsync()
    {
        if (!_isConnected || _ws == null) return;
        
        try
        {
            await _ws.CloseAsync(WebSocketCloseStatus.NormalClosure, "Closing", CancellationToken.None);
            _isConnected = false;
            OnConnectionChanged?.Invoke(this, false);
            _logger?.LogInformation("WebSocket disconnected");
        }
        catch (Exception ex)
        {
            _logger?.LogError(ex, "Error closing WebSocket");
        }
    }
    
    /// <summary>
    /// 发送消息
    /// </summary>
    public async Task SendAsync(WSMessage message)
    {
        if (_ws == null || !_isConnected) return;
        
        try
        {
            var json = JsonSerializer.Serialize(message, _jsonOptions);
            var bytes = Encoding.UTF8.GetBytes(json);
            await _ws.SendAsync(new ArraySegment<byte>(bytes), WebSocketMessageType.Text, true, _cts.Token);
        }
        catch (Exception ex)
        {
            _logger?.LogError(ex, "Error sending WebSocket message");
        }
    }
    
    /// <summary>
    /// 接收消息循环
    /// </summary>
    private async Task ReceiveMessagesAsync()
    {
        if (_ws == null) return;
        
        var buffer = new byte[8192];
        
        while (_ws.State == WebSocketState.Open && !_cts.Token.IsCancellationRequested)
        {
            try
            {
                var result = await _ws.ReceiveAsync(new ArraySegment<byte>(buffer), _cts.Token);
                
                if (result.MessageType == WebSocketMessageType.Text)
                {
                    var json = Encoding.UTF8.GetString(buffer, 0, result.Count);
                    var message = JsonSerializer.Deserialize<WSMessage>(json, _jsonOptions);
                    
                    if (message != null)
                    {
                        OnMessageReceived?.Invoke(this, message);
                        
                        // 如果是告警消息，触发告警事件
                        if (message.Type == "alert")
                        {
                            var alertJson = JsonSerializer.Serialize(message.Data);
                            var alert = JsonSerializer.Deserialize<Alert>(alertJson, _jsonOptions);
                            if (alert != null)
                            {
                                OnAlertReceived?.Invoke(this, alert);
                            }
                        }
                    }
                }
                else if (result.MessageType == WebSocketMessageType.Close)
                {
                    _logger?.LogInformation("WebSocket closed by server");
                    _isConnected = false;
                    OnConnectionChanged?.Invoke(this, false);
                    break;
                }
            }
            catch (OperationCanceledException)
            {
                break;
            }
            catch (WebSocketException ex)
            {
                _logger?.LogError(ex, "WebSocket error");
                _isConnected = false;
                OnConnectionChanged?.Invoke(this, false);
                _ = ReconnectAsync();
                break;
            }
        }
    }
    
    /// <summary>
    /// 心跳循环
    /// </summary>
    private async Task SendHeartbeatAsync()
    {
        while (_isConnected && !_cts.Token.IsCancellationRequested)
        {
            try
            {
                await Task.Delay(30000, _cts.Token); // 30秒心跳
                
                if (_ws?.State == WebSocketState.Open)
                {
                    await SendAsync(new WSMessage { Type = "heartbeat" });
                    _logger?.LogDebug("WebSocket heartbeat sent");
                }
            }
            catch (OperationCanceledException)
            {
                break;
            }
        }
    }
    
    /// <summary>
    /// 重连逻辑
    /// </summary>
    private async Task ReconnectAsync()
    {
        for (int i = 0; i < 5; i++)
        {
            await Task.Delay(5000 * (i + 1)); // 递增延迟
            
            _logger?.LogInformation("Attempting WebSocket reconnect (attempt {Attempt})", i + 1);
            
            try
            {
                _ws?.Dispose();
                _ws = new ClientWebSocket();
                await _ws.ConnectAsync(new Uri(_wsUrl), _cts.Token);
                
                _isConnected = true;
                OnConnectionChanged?.Invoke(this, true);
                _logger?.LogInformation("WebSocket reconnected");
                
                _ = ReceiveMessagesAsync();
                _ = SendHeartbeatAsync();
                return;
            }
            catch (Exception ex)
            {
                _logger?.LogWarning(ex, "Reconnect attempt {Attempt} failed", i + 1);
            }
        }
        
        _logger?.LogError("WebSocket reconnect failed after 5 attempts");
    }
    
    public void Dispose()
    {
        if (_disposed) return;
        
        _cts.Cancel();
        _ws?.Dispose();
        _cts.Dispose();
        
        _disposed = true;
    }
}