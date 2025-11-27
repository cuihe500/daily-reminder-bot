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
	CalendarInfo string // Formatted calendar info including lunar date, festivals, solar terms
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
	return `你是一个友善的每日提醒助手。你的任务是根据提供的日期、天气数据和待办事项，生成一条温馨、自然的早间提醒消息。

要求：
1. 开头展示今日日期（公历和农历），如有节日或节气要特别提及
2. 如果临近重要节日/假期，给予温馨提示（如"还有X天就放假啦"）
3. 简洁地总结天气和穿衣建议
4. 自然地提及今日待办事项
5. 根据天气和节日给出贴心建议
6. 保持积极正面的语气
7. 使用适当的 emoji 增加亲和力
8. 总长度控制在 350 字以内
9. 使用中文回复`
}

// buildUserPrompt builds the user prompt with weather and todo data
func buildUserPrompt(data ReminderData) string {
	// Format weather information
	weatherInfo := fmt.Sprintf(`城市: %s
日期: %s
温度: %s°C (体感 %s°C)
天气: %s
湿度: %s%%
风向: %s %s级`,
		data.City,
		data.Date,
		data.Weather.Temp,
		data.Weather.FeelsLike,
		data.Weather.Text,
		data.Weather.Humidity,
		data.Weather.WindDir,
		data.Weather.WindScale,
	)

	// Format life indices
	var indicesInfo string
	for _, idx := range data.LifeIndices {
		// Filter important indices: sports (1), dressing (3), UV (5)
		if idx.Type == "1" || idx.Type == "3" || idx.Type == "5" {
			indicesInfo += fmt.Sprintf("- %s: %s (%s)\n", idx.Name, idx.Category, idx.Text)
		}
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

	// Format calendar info
	calendarInfo := data.CalendarInfo
	if calendarInfo == "" {
		calendarInfo = fmt.Sprintf("日期: %s", data.Date)
	}

	return fmt.Sprintf(`请根据以下信息生成今日提醒：

【日期信息】
%s

【天气信息】
%s

【生活指数】
%s

【待办事项】
%s`, calendarInfo, weatherInfo, indicesInfo, todosInfo)
}
