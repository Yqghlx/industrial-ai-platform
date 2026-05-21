using System;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;

class EdgeSimulator
{
    private static readonly HttpClient httpClient = new HttpClient();
    private static readonly Random random = new Random();
    
    // Configuration
    private static readonly int DeviceCount = GetEnvInt("DEVICE_COUNT", 5);
    private static readonly int ReportIntervalMs = GetEnvInt("REPORT_INTERVAL_MS", 3000);
    private static readonly string ApiBaseUrl = GetEnvString("API_BASE_URL", "http://localhost:8080");
    
    // Device state for fault injection
    private static readonly Dictionary<string, bool> faultStates = new Dictionary<string, bool>();
    private static readonly Dictionary<string, DateTime> faultStartTimes = new Dictionary<string, DateTime>();
    
    // Color codes for console output
    private const string ResetColor = "\x1b[0m";
    private const string GreenColor = "\x1b[32m";
    private const string YellowColor = "\x1b[33m";
    private const string RedColor = "\x1b[31m";
    private const string CyanColor = "\x1b[36m";
    private const string GrayColor = "\x1b[37m";

    static async Task Main(string[] args)
    {
        Console.WriteLine("\n╔═══════════════════════════════════════════════════════════╗");
        Console.WriteLine("║       Industrial AI Platform - Edge Simulator             ║");
        Console.WriteLine("║       Version 1.0.0                                       ║");
        Console.WriteLine("╚═══════════════════════════════════════════════════════════╝\n");
        
        Console.WriteLine($"{CyanColor}Configuration:{ResetColor}");
        Console.WriteLine($"  Device Count:      {DeviceCount}");
        Console.WriteLine($"  Report Interval:   {ReportIntervalMs}ms");
        Console.WriteLine($"  API Base URL:      {ApiBaseUrl}");
        Console.WriteLine();

        // Initialize devices
        var devices = InitializeDevices();
        
        Console.WriteLine($"{CyanColor}Simulated Devices:{ResetColor}");
        foreach (var device in devices)
        {
            Console.WriteLine($"  {GreenColor}{device.Id}{ResetColor} - {device.Name} ({device.Type})");
        }
        Console.WriteLine();
        Console.WriteLine($"{CyanColor}Starting simulation...{ResetColor}");
        Console.WriteLine($"{GrayColor}Press Ctrl+C to stop{ResetColor}\n");

        // Run simulation loop
        var tasks = new List<Task>();
        foreach (var device in devices)
        {
            tasks.Add(SimulateDevice(device));
        }

        // Wait for all tasks
        await Task.WhenAll(tasks);
    }

    static List<Device> InitializeDevices()
    {
        var devices = new List<Device>();
        var types = new[] { "CNC", "INJ", "ROB", "ASM", "CNV" };
        var names = new Dictionary<string, string>
        {
            { "CNC", "数控机床" },
            { "INJ", "注塑机" },
            { "ROB", "工业机器人" },
            { "ASM", "装配线" },
            { "CNV", "传送带" }
        };

        for (int i = 1; i <= DeviceCount; i++)
        {
            var typeIndex = (i - 1) % types.Length;
            var type = types[typeIndex];
            var id = $"{type}-{i.ToString("D3")}";
            
            devices.Add(new Device
            {
                Id = id,
                Name = $"{names[type]} {i}",
                Type = names[type],
                Location = "车间A"
            });
        }

        return devices;
    }

    static async Task SimulateDevice(Device device)
    {
        while (true)
        {
            try
            {
                // Check fault state
                var inFault = faultStates.TryGetValue(device.Id, out var fault) && fault;
                
                // Inject random fault (5% chance)
                if (!inFault && random.NextDouble() < 0.05)
                {
                    StartFault(device.Id);
                }

                // Generate telemetry data
                var telemetry = GenerateTelemetry(device, inFault);

                // Log the telemetry
                LogTelemetry(device, telemetry, inFault);

                // Send to API
                await SendTelemetryAsync(telemetry);

                // Check if fault should end (5-15 seconds)
                if (inFault && faultStartTimes.TryGetValue(device.Id, out var startTime))
                {
                    var faultDuration = (DateTime.Now - startTime).TotalSeconds;
                    var maxDuration = random.Next(5, 15);
                    if (faultDuration > maxDuration)
                    {
                        EndFault(device.Id);
                    }
                }

                // Wait for next interval
                await Task.Delay(ReportIntervalMs);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"{RedColor}[ERROR] {device.Id}: {ex.Message}{ResetColor}");
                await Task.Delay(5000);
            }
        }
    }

    static TelemetryData GenerateTelemetry(Device device, bool inFault)
    {
        var now = DateTime.UtcNow;

        // Normal ranges
        var temperature = 60 + random.NextDouble() * 30; // 60-90°C
        var vibration = 0.5 + random.NextDouble() * 1.5; // 0.5-2.0 mm/s
        var pressure = 80 + random.NextDouble() * 40; // 80-120 bar
        var power = 3 + random.NextDouble() * 5; // 3-8 kW

        // Inject fault conditions
        if (inFault)
        {
            // High temperature: 120°C+
            if (random.NextDouble() < 0.5)
            {
                temperature = 120 + random.NextDouble() * 30;
            }
            // High vibration: 5.0+
            if (random.NextDouble() < 0.5)
            {
                vibration = 5.0 + random.NextDouble() * 3;
            }
        }

        var status = "normal";
        var message = "";
        
        if (temperature > 120 || vibration > 5.0)
        {
            status = "fault";
            message = "设备故障";
        }
        else if (temperature > 100 || vibration > 3.0)
        {
            status = "warning";
            message = "运行异常";
        }

        return new TelemetryData
        {
            DeviceId = device.Id,
            Timestamp = now.ToString("O"),
            Temperature = Math.Round(temperature, 2),
            Pressure = Math.Round(pressure, 1),
            Vibration = Math.Round(vibration, 2),
            Humidity = Math.Round(40 + random.NextDouble() * 20, 1),
            Power = Math.Round(power, 2),
            Status = status,
            Message = message
        };
    }

    static void LogTelemetry(Device device, TelemetryData telemetry, bool inFault)
    {
        var timestamp = DateTime.Now.ToString("HH:mm:ss");
        var color = telemetry.Status == "normal" ? GreenColor :
                    telemetry.Status == "warning" ? YellowColor :
                    RedColor;

        Console.WriteLine($"{GrayColor}[{timestamp}]{ResetColor} {color}{device.Id}{ResetColor} " +
            $"温度:{telemetry.Temperature}°C 振动:{telemetry.Vibration}mm/s " +
            $"压力:{telemetry.Pressure}bar 功率:{telemetry.Power}kW " +
            $"{color}[{telemetry.Status}]{ResetColor}");
        
        if (inFault && telemetry.Status == "fault")
        {
            Console.WriteLine($"{RedColor}          ⚠️ 故障状态: {telemetry.Message}{ResetColor}");
        }
    }

    static async Task SendTelemetryAsync(TelemetryData telemetry)
    {
        var json = JsonSerializer.Serialize(telemetry);
        var content = new StringContent(json, Encoding.UTF8, "application/json");

        var response = await httpClient.PostAsync($"{ApiBaseUrl}/api/v1/devices/telemetry", content);
        
        if (!response.IsSuccessStatusCode)
        {
            Console.WriteLine($"{RedColor}[ERROR] Failed to send telemetry: {response.StatusCode}{ResetColor}");
        }
    }

    static void StartFault(string deviceId)
    {
        faultStates[deviceId] = true;
        faultStartTimes[deviceId] = DateTime.Now;
        Console.WriteLine($"{YellowColor}[FAULT INJECTED] {deviceId} - Fault started{ResetColor}");
    }

    static void EndFault(string deviceId)
    {
        faultStates[deviceId] = false;
        faultStartTimes.Remove(deviceId);
        Console.WriteLine($"{GreenColor}[FAULT ENDED] {deviceId} - Normal operation restored{ResetColor}");
    }

    static int GetEnvInt(string key, int defaultValue)
    {
        var value = Environment.GetEnvironmentVariable(key);
        return int.TryParse(value, out var result) ? result : defaultValue;
    }

    static string GetEnvString(string key, string defaultValue)
    {
        var value = Environment.GetEnvironmentVariable(key);
        return string.IsNullOrEmpty(value) ? defaultValue : value;
    }
}

// Data models
class Device
{
    public string Id { get; set; } = "";
    public string Name { get; set; } = "";
    public string Type { get; set; } = "";
    public string Location { get; set; } = "";
}

class TelemetryData
{
    public string DeviceId { get; set; } = "";
    public string Timestamp { get; set; } = "";
    public double Temperature { get; set; }
    public double Pressure { get; set; }
    public double Vibration { get; set; }
    public double Humidity { get; set; }
    public double Power { get; set; }
    public string Status { get; set; } = "normal";
    public string Message { get; set; } = "";
}