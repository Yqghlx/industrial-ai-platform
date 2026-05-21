using System.Net.Http.Json;
using System.Text.Json;
using Microsoft.Extensions.Logging;
using IndustrialAI.EdgeSDK.Models;
using Polly;

namespace IndustrialAI.EdgeSDK.Http;

/// <summary>
/// HTTP API 客户端
/// </summary>
public class ApiClient
{
    private readonly HttpClient _httpClient;
    private readonly ILogger<ApiClient>? _logger;
    private readonly JsonSerializerOptions _jsonOptions;
    private readonly IAsyncPolicy<HttpResponseMessage> _retryPolicy;
    
    public ApiClient(HttpClient httpClient, ILogger<ApiClient>? logger = null)
    {
        _httpClient = httpClient;
        _logger = logger;
        
        _jsonOptions = new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
            PropertyNameCaseInsensitive = true
        };
        
        // 配置重试策略
        _retryPolicy = Policy<HttpResponseMessage>
            .Handle<HttpRequestException>()
            .OrResult(r => !r.IsSuccessStatusCode && r.StatusCode != System.Net.HttpStatusCode.BadRequest)
            .WaitAndRetryAsync(3, retryAttempt => TimeSpan.FromSeconds(retryAttempt));
    }
    
    /// <summary>
    /// 发送遥测数据
    /// </summary>
    public async Task<ApiResponse<object>?> SendTelemetryAsync(TelemetryData data)
    {
        try
        {
            var response = await _retryPolicy.ExecuteAsync(async () =>
            {
                return await _httpClient.PostAsJsonAsync(
                    "/api/v1/devices/telemetry",
                    data,
                    _jsonOptions);
            });
            
            if (response.IsSuccessStatusCode)
            {
                _logger?.LogDebug("Telemetry sent successfully for device {DeviceId}", data.DeviceId);
                return await response.Content.ReadFromJsonAsync<ApiResponse<object>>(_jsonOptions);
            }
            
            _logger?.LogWarning("Telemetry send failed: {StatusCode}", response.StatusCode);
            return new ApiResponse<object>
            {
                Success = false,
                Error = $"HTTP {response.StatusCode}"
            };
        }
        catch (Exception ex)
        {
            _logger?.LogError(ex, "Error sending telemetry");
            return new ApiResponse<object>
            {
                Success = false,
                Error = ex.Message
            };
        }
    }
    
    /// <summary>
    /// 批量发送遥测数据
    /// </summary>
    public async Task<ApiResponse<object>?> SendTelemetryBatchAsync(IEnumerable<TelemetryData> data)
    {
        try
        {
            var response = await _retryPolicy.ExecuteAsync(async () =>
            {
                return await _httpClient.PostAsJsonAsync(
                    "/api/v1/devices/telemetry/batch",
                    data,
                    _jsonOptions);
            });
            
            return await response.Content.ReadFromJsonAsync<ApiResponse<object>>(_jsonOptions);
        }
        catch (Exception ex)
        {
            _logger?.LogError(ex, "Error sending telemetry batch");
            return new ApiResponse<object> { Success = false, Error = ex.Message };
        }
    }
    
    /// <summary>
    /// 获取设备详情
    /// </summary>
    public async Task<ApiResponse<Device>?> GetDeviceAsync(string deviceId)
    {
        try
        {
            var response = await _httpClient.GetAsync($"/api/v1/devices/{deviceId}");
            return await response.Content.ReadFromJsonAsync<ApiResponse<Device>>(_jsonOptions);
        }
        catch (Exception ex)
        {
            _logger?.LogError(ex, "Error getting device");
            return new ApiResponse<Device> { Success = false, Error = ex.Message };
        }
    }
    
    /// <summary>
    /// 注册设备
    /// </summary>
    public async Task<ApiResponse<Device>?> RegisterDeviceAsync(Device device)
    {
        try
        {
            var response = await _httpClient.PostAsJsonAsync(
                "/api/v1/devices",
                device,
                _jsonOptions);
            
            return await response.Content.ReadFromJsonAsync<ApiResponse<Device>>(_jsonOptions);
        }
        catch (Exception ex)
        {
            _logger?.LogError(ex, "Error registering device");
            return new ApiResponse<Device> { Success = false, Error = ex.Message };
        }
    }
    
    /// <summary>
    /// 获取健康状态
    /// </summary>
    public async Task<bool> CheckHealthAsync()
    {
        try
        {
            var response = await _httpClient.GetAsync("/health");
            return response.IsSuccessStatusCode;
        }
        catch
        {
            return false;
        }
    }
}