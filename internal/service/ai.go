package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"github.com/cuichanghe/daily-reminder-bot/pkg/openai"
	"github.com/cuichanghe/daily-reminder-bot/pkg/qweather"
	"go.uber.org/zap"
)

// AIService handles AI-powered content generation
type AIService struct {
	client     *openai.Client
	maxRetries int
	enabled    bool
}

// NewAIService creates a new AIService
func NewAIService(client *openai.Client, maxRetries int, enabled bool) *AIService {
	return &AIService{
		client:     client,
		maxRetries: maxRetries,
		enabled:    enabled,
	}
}

// IsEnabled returns whether the AI service is enabled
func (s *AIService) IsEnabled() bool {
	return s.enabled && s.client != nil
}

// ReminderData holds the data needed to generate a reminder
type ReminderData struct {
	City         string
	Date         string
	Weather      *qweather.CurrentWeather
	LifeIndices  []qweather.LifeIndex
	Todos        []model.Todo
	CalendarInfo string                       // Formatted calendar info including lunar date, festivals, solar terms
	AirQuality   *qweather.AirQualityResponse // Air quality data (optional)
	Warnings     []qweather.Warning           // Weather warnings (optional)
}

// GenerateReminder generates a daily reminder using AI with retry logic
// Returns the generated content and a boolean indicating success
func (s *AIService) GenerateReminder(ctx context.Context, data ReminderData) (string, bool) {
	if !s.IsEnabled() {
		return "", false
	}

	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(data)

	var lastErr error
	for i := 0; i < s.maxRetries; i++ {
		content, err := s.client.GetContent(ctx, systemPrompt, userPrompt)
		if err == nil {
			logger.Debug("AI generated reminder successfully", zap.Int("attempt", i+1))
			return content, true
		}

		lastErr = err
		logger.Warn("AI generation failed, retrying...",
			zap.Int("attempt", i+1),
			zap.Int("max_retries", s.maxRetries),
			zap.Error(err))

		// Exponential backoff
		if i < s.maxRetries-1 {
			time.Sleep(time.Duration(1<<i) * time.Second)
		}
	}

	logger.Error("AI service unavailable after retries",
		zap.Int("attempts", s.maxRetries),
		zap.Error(lastErr))

	return "", false
}

// buildSystemPrompt builds the system prompt for AI generation
func buildSystemPrompt() string {
	return `你是一个友善的每日提醒助手。你的任务是根据提供的日期、天气数据和待办事项，生成一条温馨、自然的提醒消息。

要求：
1. 开头根据现在的时间给予问候（比如早上好、中午好等），展示今日日期（公历和农历），如有节日或节气要特别提及
2. 如果临近重要节日/假期，给予温馨提示（如"还有X天就放假啦"）
3. 如果有天气预警，必须在开头用醒目的方式提醒用户注意，说明预警类型、等级和简要建议
4. 详细解读天气状况：
   - 重点关注实际温度与体感温度的差异，如果相差较大需特别说明原因（风力、湿度等）
   - 根据风力等级和风速给出具体影响提示（如3级以上建议注意防风）
   - 结合湿度说明体感舒适度（如高湿度闷热、低湿度干燥）
   - 如果天气有特殊情况（高温、低温、大风、高湿度等）需重点提醒
5. 充分利用生活指数给出实用建议：
   - 穿衣指数：具体建议穿什么类型的衣物
   - 紫外线指数：说明是否需要防晒措施
   - 运动指数：建议适合的运动类型或是否适宜户外活动
6. 根据空气质量给出健康建议：
   - 如果空气质量差，提醒减少户外活动或佩戴口罩
7. 自然地提及今日待办事项，如有多项可按重要程度排序提醒
8. 根据天气、节日、待办事项的综合情况给出贴心的生活建议
9. 保持积极正面、温暖友善的语气
10. 使用适当的 emoji 增加亲和力和可读性
11. 总长度控制在 400 字以内
12. 使用中文回复`
}

// buildUserPrompt builds the user prompt with weather and todo data
func buildUserPrompt(data ReminderData) string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	// Calculate temperature difference for AI analysis
	tempDiff := ""
	if data.Weather.Temp != "" && data.Weather.FeelsLike != "" {
		// Note: This is for display purposes; actual calculation would need parsing
		tempDiff = fmt.Sprintf("（温差：实际温度与体感温度相差 %s°C - %s°C）", data.Weather.Temp, data.Weather.FeelsLike)
	}

	// Format weather information with more details
	weatherInfo := fmt.Sprintf(`城市: %s
日期: %s
时间: %s
实际温度: %s°C
体感温度: %s°C %s
天气状况: %s
相对湿度: %s%%
风向风力: %s %s级 (风速 %s km/h)`,
		data.City,
		data.Date,
		now.Format("15:04"),
		data.Weather.Temp,
		data.Weather.FeelsLike,
		tempDiff,
		data.Weather.Text,
		data.Weather.Humidity,
		data.Weather.WindDir,
		data.Weather.WindScale,
		data.Weather.WindSpeed,
	)

	// Format life indices with more details
	var indicesInfo string
	indicesMap := make(map[string]qweather.LifeIndex)
	for _, idx := range data.LifeIndices {
		indicesMap[idx.Type] = idx
	}

	// Prioritize important indices: dressing (3), UV (5), sports (1)
	importantTypes := []string{"3", "5", "1"}
	for _, typ := range importantTypes {
		if idx, exists := indicesMap[typ]; exists {
			indicesInfo += fmt.Sprintf("• %s：等级 %s，%s\n  详细建议：%s\n",
				idx.Name, idx.Level, idx.Category, idx.Text)
		}
	}

	// Add other available indices
	for _, idx := range data.LifeIndices {
		// Skip already processed indices
		if idx.Type == "1" || idx.Type == "3" || idx.Type == "5" {
			continue
		}
		indicesInfo += fmt.Sprintf("• %s：%s\n  %s\n", idx.Name, idx.Category, idx.Text)
	}

	if indicesInfo == "" {
		indicesInfo = "暂无生活指数数据"
	}

	// Format todos
	var todosInfo string
	if len(data.Todos) == 0 {
		todosInfo = "今日暂无待办事项"
	} else {
		for i, todo := range data.Todos {
			todosInfo += fmt.Sprintf("%d. %s\n", i+1, todo.Content)
		}
	}

	// Format air quality
	var airQualityInfo string
	if data.AirQuality != nil && len(data.AirQuality.Indexes) > 0 {
		// Find primary index (prefer "qaqi" for China, or "us-epa", or first available)
		var mainIndex qweather.AirQualityIndex
		foundIndex := false
		for _, idx := range data.AirQuality.Indexes {
			if idx.Code == "qaqi" {
				mainIndex = idx
				foundIndex = true
				break
			}
		}
		if !foundIndex {
			mainIndex = data.AirQuality.Indexes[0]
		}

		airQualityInfo = fmt.Sprintf(`• AQI：%.0f
• 等级：%s
• 类别：%s`,
			mainIndex.Aqi,
			mainIndex.Level,
			mainIndex.Category)
		if mainIndex.PrimaryPollutant.Name != "" {
			airQualityInfo += fmt.Sprintf("\n• 主要污染物：%s", mainIndex.PrimaryPollutant.Name)
		}
	} else {
		airQualityInfo = "暂无空气质量数据"
	}

	// Format calendar info
	calendarInfo := data.CalendarInfo
	if calendarInfo == "" {
		calendarInfo = fmt.Sprintf("日期: %s", data.Date)
	}

	// Format warnings
	warningsInfo := formatWarningsForAI(data.Warnings)

	return fmt.Sprintf(`请根据以下信息生成今日提醒：

【日期信息】
%s

【天气预警】
%s

【天气信息】
%s

【空气质量】
%s

【生活指数】
%s

【待办事项】
%s

请特别注意：
1. 如果有天气预警，必须在开头醒目提醒，说明预警内容和应对建议
2. 如果实际温度与体感温度相差较大（≥3°C），请重点说明并解释原因
3. 根据风速和风力等级判断是否需要提醒防风
4. 根据湿度水平说明体感舒适度（<30%%干燥，>70%%潮湿闷热）
5. 根据AQI等级给出健康建议（优：无需特殊措施，良：敏感人群减少户外，轻度污染以上：减少户外活动，佩戴口罩）
6. 充分利用生活指数的详细建议，给出具体可行的行动指导
7. 如果有待办事项，要自然地融入提醒中，不要生硬列举`, calendarInfo, warningsInfo, weatherInfo, airQualityInfo, indicesInfo, todosInfo)
}

// formatWarningsForAI formats weather warnings for AI prompt
func formatWarningsForAI(warnings []qweather.Warning) string {
	if len(warnings) == 0 {
		return "当前无天气预警"
	}

	var result string
	for i, w := range warnings {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("• 预警类型：%s\n  级别：%s\n  颜色：%s\n  内容：%s",
			w.TypeName, w.Level, w.SeverityColor, w.Title)
		if w.Text != "" {
			result += fmt.Sprintf("\n  详情：%s", w.Text)
		}
	}
	return result
}
