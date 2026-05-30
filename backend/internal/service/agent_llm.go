package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/errors"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// LLM配置常量
// P2-FIX: 将魔法数字定义为常量，提高代码可读性和可维护性
const (
	// DefaultMaxTokens LLM响应的最大token数
	// 较小的值可以加快响应速度，适合工业AI平台的实时查询场景
	DefaultMaxTokens = 512

	// DefaultTemperature LLM生成的温度参数
	// 0.7 提供了创造性和一致性的良好平衡
	DefaultTemperature = 0.7
)

func (s *AgentService) determineAgent(query string) string {
	queryLower := strings.ToLower(query)

	if strings.Contains(queryLower, "设备") || strings.Contains(queryLower, "device") ||
		strings.Contains(queryLower, "温度") || strings.Contains(queryLower, "temperature") ||
		strings.Contains(queryLower, "振动") || strings.Contains(queryLower, "vibration") {
		return "设备专家"
	}

	if strings.Contains(queryLower, "维护") || strings.Contains(queryLower, "maintenance") ||
		strings.Contains(queryLower, "工单") || strings.Contains(queryLower, "repair") {
		return "维护专家"
	}

	if strings.Contains(queryLower, "预测") || strings.Contains(queryLower, "predict") ||
		strings.Contains(queryLower, "故障") || strings.Contains(queryLower, "fault") {
		return "预测专家"
	}

	if strings.Contains(queryLower, "优化") || strings.Contains(queryLower, "optimize") ||
		strings.Contains(queryLower, "效率") || strings.Contains(queryLower, "efficiency") {
		return "优化专家"
	}

	return "通用智能体"
}

// callLLM calls the actual LLM API (Bailian/GLM-5)
func (s *AgentService) callLLM(ctx context.Context, query string, contextData map[string]interface{}, agent string) (string, error) {
	// Build system prompt based on agent type
	systemPrompt := s.buildSystemPrompt(agent)

	// Build user message with context
	userMessage := query
	if contextData != nil && len(contextData) > 0 {
		contextJSON, _ := json.Marshal(contextData)
		userMessage = fmt.Sprintf("当前设备上下文数据:\n%s\n\n用户问题: %s", string(contextJSON), query)
	}

	// Build request body (OpenAI-compatible format)
	// OPT-001: Reduce max_tokens for faster response (see DefaultMaxTokens constant)
	reqBody := map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userMessage},
		},
		"max_tokens":  DefaultMaxTokens, // P2-FIX: 使用常量替代魔法数字
		"temperature": DefaultTemperature, // P2-FIX: 使用常量替代魔法数字
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", errors.NewInternalError(fmt.Sprintf("Failed to marshal request: %v", err))
	}

	// Create HTTP request
	url := s.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", errors.NewInternalError(fmt.Sprintf("Failed to create request: %v", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// FIX-019: 使用共享 HTTP Client 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", errors.NewInternalError(fmt.Sprintf("Failed to send request: %v", err))
	}
	defer resp.Body.Close()

	// Read response
	// LLM 响应体最大 1MB（与 middleware.MaxBodySize 保持一致）
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", errors.NewInternalError(fmt.Sprintf("Failed to read response: %v", err))
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		logger.L().Error("LLM API error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return "", errors.NewAppError(errors.ErrCodeService, fmt.Sprintf("LLM API returned status %d: %s", resp.StatusCode, string(body)), "")
	}

	// Parse response
	var llmResp LLMResponse
	if err := json.Unmarshal(body, &llmResp); err != nil {
		return "", errors.NewInternalError(fmt.Sprintf("Failed to parse response: %v", err))
	}

	// Extract content
	if len(llmResp.Choices) == 0 {
		return "", errors.NewAppError(errors.ErrCodeService, "No choices in LLM response", "")
	}

	content := llmResp.Choices[0].Message.Content
	logger.L().Info("LLM response received",
		zap.String("model", llmResp.Model),
		zap.Int("tokens_used", llmResp.Usage.TotalTokens),
	)

	return content, nil
}

// buildSystemPrompt builds system prompt for different agents
func (s *AgentService) buildSystemPrompt(agent string) string {
	prompts := map[string]string{
		"设备专家": `你是工业AI平台的设备专家智能体。你的职责是：
- 分析设备运行状态和遥测数据
- 诊断设备故障和异常
- 提供设备维护建议
- 解答设备相关问题

回答时请：
1. 使用结构化的格式（标题、列表等）
2. 提供具体的数据分析
3. 给出可操作的建议
4. 使用中文回答，保持专业但友好的语气`,
		"维护专家": `你是工业AI平台的维护专家智能体。你的职责是：
- 制定预防性维护计划
- 评估维护优先级
- 管理维护工单
- 提供维护最佳实践

回答时请：
1. 按优先级排序建议
2. 提供时间预估
3. 考虑成本效益
4. 使用中文回答，保持专业但友好的语气`,
		"预测专家": `你是工业AI平台的预测专家智能体。你的职责是：
- 分析设备故障概率
- 预测剩余使用寿命
- 识别潜在风险设备
- 提供预警建议

回答时请：
1. 提供具体的风险概率
2. 解释预测依据
3. 给出预防措施
4. 使用中文回答，保持专业但友好的语气`,
		"优化专家": `你是工业AI平台的优化专家智能体。你的职责是：
- 分析生产效率数据
- 提供优化建议
- 计算ROI和成本节约
- 规划改进方案

回答时请：
1. 提供具体的数值分析
2. 计算预期收益
3. 分步骤实施方案
4. 使用中文回答，保持专业但友好的语气`,
		"通用智能体": `你是工业AI平台的通用智能体。你可以帮助用户：
- 分析设备状态和遥测数据
- 提供维护建议和工单管理
- 进行故障预测和风险评估
- 优化生产效率

请根据用户的问题，选择合适的专家角色来回答。使用中文，保持专业但友好的语气。`,
	}

	if prompt, ok := prompts[agent]; ok {
		return prompt
	}
	return prompts["通用智能体"]
}

// LLMResponse represents LLM API response
type LLMResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// mockResponse generates a mock response (fallback)
func (s *AgentService) mockResponse(query string, agent string) string {
	responses := map[string][]string{
		"设备专家": {
			"根据您的问题，我分析了相关设备数据：\n\n**设备状态分析**\n- 当前设备运行正常，温度在正常范围内\n- 振动数据稳定，无异常波动\n- 建议继续监控，每30分钟记录一次数据\n\n**建议措施**\n1. 保持当前运行参数\n2. 定期检查设备润滑状态\n3. 关注温度变化趋势",
			"设备数据分析结果：\n\n**当前状态**\n- 设备温度: 75°C (正常)\n- 振动幅度: 1.2mm/s (正常)\n- 负载率: 85% (良好)\n\n**运行建议**\n设备运行状态良好，建议继续按照当前参数运行。",
		},
		"维护专家": {
			"根据设备运行数据，我提供以下维护建议：\n\n**预防性维护计划**\n1. **每周检查**: 润滑油液位、密封件状态\n2. **每月检查**: 电气连接、传感器校准\n3. **季度保养**: 轴承更换检查、整机清洁\n\n**优先级建议**\n- 立即: 无紧急维护项目\n- 本周: 检查CNC-001设备轴承",
			"维护工单分析：\n\n**当前待处理工单**: 3件\n- 工单#1234: CNC-001 主轴检查 (优先级: 高)\n- 工单#1235: INJ-002 液压系统维护 (优先级: 中)\n- 工单#1236: ROB-003 关节润滑 (优先级: 低)\n\n**建议处理顺序**: 按优先级从高到低处理",
		},
		"预测专家": {
			"基于机器学习模型的预测分析：\n\n**故障预测结果**\n- CNC-001: 未来7天故障概率 5% (低风险)\n- INJ-002: 未来7天故障概率 25% (中风险)\n- ROB-003: 未来7天故障概率 2% (低风险)\n\n**风险设备清单**\n1. INJ-002 注塑机 - 建议提前检查液压系统\n\n**建议措施**\n- 对INJ-002进行预防性维护\n- 增加监控频率至每5分钟",
			"预测性维护分析报告：\n\n**剩余使用寿命预测**\n- 主轴轴承: 剩余约2000小时\n- 液压泵: 剩余约1500小时\n- 伺服电机: 剩余约3000小时\n\n**关键指标趋势**\n- 温度趋势: 稳定\n- 振动趋势: 轻微上升，需关注\n- 能耗趋势: 正常范围",
		},
		"优化专家": {
			"生产优化建议：\n\n**效率提升方案**\n1. **参数优化**: 建议调整加工速度提升5%\n2. **能耗优化**: 通过智能调度可节省约10%能源\n3. **良品率提升**: 调整温度控制参数可提高良品率2%\n\n**实施建议**\n- 优先实施能耗优化方案\n- 预计ROI: 3个月收回成本",
			"产能优化分析：\n\n**当前产能利用率**: 78%\n\n**优化机会识别**\n1. 减少换线时间: 可提升产能5%\n2. 优化生产排程: 可提升产能3%\n3. 设备预测维护: 减少停机时间10%\n\n**预期收益**\n- 月产能提升: 约15%\n- 成本节约: 约¥50,000/月",
		},
		"通用智能体": {
			"您好！我是工业AI智能助手，我可以帮助您：\n\n**核心能力**\n1. 📊 设备状态分析\n2. 🔧 维护建议与工单管理\n3. 🔮 故障预测与预警\n4. ⚡ 生产优化建议\n\n**示例问题**\n- \"分析CNC-001的运行状态\"\n- \"给我维护建议\"\n- \"预测哪些设备可能出问题\"\n- \"如何优化生产效率\"\n\n请问有什么可以帮您？",
			"欢迎使用工业AI智能平台！\n\n**快速开始**\n您可以问我关于：\n- 设备监控与分析\n- 预测性维护\n- 生产优化\n- 故障诊断\n\n我会基于实时数据和机器学习模型为您提供专业建议。",
		},
	}

	// Get random response for the agent
	agentResponses, ok := responses[agent]
	if !ok {
		agentResponses = responses["通用智能体"]
	}

	// FIX-006: 使用 crypto/rand 替代 math/rand
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(agentResponses))))
	if err != nil {
		// fallback to first response if random fails
		return agentResponses[0]
	}
	return agentResponses[n.Int64()]
}

// GetTaskLogs retrieves recent task logs
func (s *AgentService) GetTaskLogs(ctx context.Context, limit int) ([]model.AgentTaskLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.taskLogRepo.List(ctx, limit)
}

// generateSessionID generates a random session ID using crypto/rand
// FIX-006: 使用 crypto/rand 替代 math/rand
func generateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// fallback to timestamp if crypto/rand fails (should not happen)
		return fmt.Sprintf("session_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("session_%x", b)
}

// GetDeviceContext gathers device context for AI queries
func (s *AgentService) GetDeviceContext(ctx context.Context, deviceID string) (*model.DeviceContext, error) {
	contextData := &model.DeviceContext{}

	if deviceID != "" {
		// Get device info
		device, err := s.deviceRepo.GetByID(ctx, deviceID)
		if err == nil {
			contextData.Device = device
		}

		// Get latest telemetry
		telemetry, err := s.telemetryRepo.GetByDeviceID(ctx, deviceID, time.Now().Add(-1*time.Hour), time.Now(), 10)
		if err == nil {
			contextData.Telemetry = telemetry
		}
	}

	return contextData, nil
}

// AnalyzeQuery analyzes a query to extract entities and intent
func (s *AgentService) AnalyzeQuery(query string) map[string]interface{} {
	analysis := make(map[string]interface{})

	// Extract device IDs (pattern: XXX-001) using regex for Chinese text support
	// Device ID pattern: 3 uppercase letters + hyphen + digits (e.g., CNC-001, INJ-999)
	deviceIDPattern := regexp.MustCompile(`[A-Z]{3}-\d{3,}`)
	matches := deviceIDPattern.FindAllString(query, -1)
	if len(matches) > 0 {
		analysis["possible_device_id"] = matches[0]
	}

	// Extract intent
	queryLower := strings.ToLower(query)
	if strings.Contains(queryLower, "分析") || strings.Contains(queryLower, "analyze") {
		analysis["intent"] = "analyze"
	} else if strings.Contains(queryLower, "预测") || strings.Contains(queryLower, "predict") {
		analysis["intent"] = "predict"
	} else if strings.Contains(queryLower, "维护") || strings.Contains(queryLower, "maintain") {
		analysis["intent"] = "maintain"
	} else if strings.Contains(queryLower, "优化") || strings.Contains(queryLower, "optimize") {
		analysis["intent"] = "optimize"
	} else {
		analysis["intent"] = "query"
	}

	return analysis
}
