package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/internal/repository"
	"github.com/cuichanghe/daily-reminder-bot/internal/service"
	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

// Handlers holds all service dependencies for bot handlers
type Handlers struct {
	userRepo   *repository.UserRepository
	subRepo    *repository.SubscriptionRepository
	todoRepo   *repository.TodoRepository
	weatherSvc *service.WeatherService
	todoSvc    *service.TodoService
	airSvc     *service.AirQualityService
	warningSvc *service.WarningService
}

// NewHandlers creates a new Handlers instance
func NewHandlers(
	userRepo *repository.UserRepository,
	subRepo *repository.SubscriptionRepository,
	todoRepo *repository.TodoRepository,
	weatherSvc *service.WeatherService,
	todoSvc *service.TodoService,
	airSvc *service.AirQualityService,
	warningSvc *service.WarningService,
) *Handlers {
	return &Handlers{
		userRepo:   userRepo,
		subRepo:    subRepo,
		todoRepo:   todoRepo,
		weatherSvc: weatherSvc,
		todoSvc:    todoSvc,
		airSvc:     airSvc,
		warningSvc: warningSvc,
	}
}

// RegisterHandlers registers all command handlers
func (h *Handlers) RegisterHandlers(bot *tele.Bot) {
	bot.Handle("/start", h.HandleStart)
	bot.Handle("/subscribe", h.HandleSubscribe)
	bot.Handle("/mystatus", h.HandleMyStatus)
	bot.Handle("/unsubscribe", h.HandleUnsubscribe)
	bot.Handle("/weather", h.HandleWeather)
	bot.Handle("/air", h.HandleAir)
	bot.Handle("/warning", h.HandleWarning)
	bot.Handle("/warning_toggle", h.HandleWarningToggle)
	bot.Handle("/todo", h.HandleTodo)
	bot.Handle("/help", h.HandleHelp)
}

// HandleStart handles the /start command
func (h *Handlers) HandleStart(c tele.Context) error {
	chatID := c.Sender().ID
	logger.Debug("Received /start command", zap.Int64("chat_id", chatID))

	// Get or create user
	_, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		logger.Error("Failed to create user",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	message := `👋 欢迎使用每日提醒机器人！

我可以帮你：
• 📍 订阅每日天气和生活指数
• ☁️ 查询实时天气
• 📝 管理待办事项

使用 /help 查看所有命令`

	logger.Info("User started bot", zap.Int64("chat_id", chatID))
	return c.Send(message)
}

// HandleSubscribe handles the /subscribe command
func (h *Handlers) HandleSubscribe(c tele.Context) error {
	chatID := c.Sender().ID
	logger.Debug("Received /subscribe command",
		zap.Int64("chat_id", chatID),
		zap.Strings("args", c.Args()))

	// Get or create user
	user, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		logger.Error("Failed to get user",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	// Parse arguments: /subscribe <city> <time>
	// Example: /subscribe 北京 08:00
	args := c.Args()
	if len(args) < 2 {
		logger.Debug("Invalid subscribe arguments",
			zap.Int64("chat_id", chatID),
			zap.Int("args_count", len(args)))
		return c.Send("❌ 用法: /subscribe <城市> <时间>\n示例: /subscribe 北京 08:00")
	}

	city := args[0]
	reminderTime := args[1]

	// Validate time format (HH:MM)
	if !isValidTimeFormat(reminderTime) {
		logger.Debug("Invalid time format",
			zap.Int64("chat_id", chatID),
			zap.String("time", reminderTime))
		return c.Send("❌ 时间格式错误，请使用 HH:MM 格式（如 08:00）")
	}

	// Check if user already has this city subscribed
	existingSub, err := h.subRepo.FindByUserAndCity(user.ID, city)
	if err != nil {
		logger.Error("Failed to find subscription",
			zap.Int64("chat_id", chatID),
			zap.Uint("user_id", user.ID),
			zap.String("city", city),
			zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	if existingSub != nil {
		// Update existing subscription for this city
		existingSub.ReminderTime = reminderTime
		existingSub.Active = true
		if err := h.subRepo.Update(existingSub); err != nil {
			logger.Error("Failed to update subscription",
				zap.Int64("chat_id", chatID),
				zap.Uint("subscription_id", existingSub.ID),
				zap.Error(err))
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
		logger.Info("Subscription updated",
			zap.Int64("chat_id", chatID),
			zap.Uint("subscription_id", existingSub.ID),
			zap.String("city", city),
			zap.String("reminder_time", reminderTime))
		return c.Send(fmt.Sprintf("✅ 订阅已更新！\n📍 城市：%s\n⏰ 新时间：%s", city, reminderTime))
	}

	// Check subscription limit (max 5)
	count, err := h.subRepo.CountActiveByUser(user.ID)
	if err != nil {
		logger.Error("Failed to count subscriptions",
			zap.Int64("chat_id", chatID),
			zap.Uint("user_id", user.ID),
			zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}
	if count >= 5 {
		logger.Warn("Subscription limit reached",
			zap.Int64("chat_id", chatID),
			zap.Uint("user_id", user.ID),
			zap.Int64("count", count))
		return c.Send("❌ 订阅数量已达上限（5个）\n请先使用 /unsubscribe <城市> 取消部分订阅")
	}

	// Create new subscription
	sub := &model.Subscription{
		UserID:       user.ID,
		City:         city,
		ReminderTime: reminderTime,
		Active:       true,
	}
	if err := h.subRepo.Create(sub); err != nil {
		logger.Error("Failed to create subscription",
			zap.Int64("chat_id", chatID),
			zap.Uint("user_id", user.ID),
			zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}
	logger.Info("Subscription created",
		zap.Int64("chat_id", chatID),
		zap.Uint("user_id", user.ID),
		zap.String("city", city),
		zap.String("reminder_time", reminderTime))

	return c.Send(fmt.Sprintf("✅ 订阅成功！\n📍 城市：%s\n⏰ 时间：%s\n\n每天将在该时间为您推送天气和待办提醒。\n\n💡 提示：您可以订阅多个城市（最多5个），每个城市的待办事项独立管理。", city, reminderTime))
}

// HandleMyStatus handles the /mystatus command
func (h *Handlers) HandleMyStatus(c tele.Context) error {
	chatID := c.Sender().ID
	logger.Debug("Received /mystatus command", zap.Int64("chat_id", chatID))

	user, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		logger.Error("Failed to get user",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	subs, err := h.subRepo.FindByUserID(user.ID)
	if err != nil {
		logger.Error("Failed to find subscriptions",
			zap.Int64("chat_id", chatID),
			zap.Uint("user_id", user.ID),
			zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	if len(subs) == 0 {
		logger.Debug("No active subscriptions found",
			zap.Int64("chat_id", chatID),
			zap.Uint("user_id", user.ID))
		return c.Send("📭 您当前没有订阅每日提醒\n\n使用 /subscribe <城市> <时间> 开始订阅")
	}

	// Build subscription list
	var status strings.Builder
	status.WriteString(fmt.Sprintf("📬 您的订阅状态（共 %d 个）\n\n", len(subs)))
	for i, sub := range subs {
		status.WriteString(fmt.Sprintf("%d. 📍 %s - ⏰ %s\n", i+1, sub.City, sub.ReminderTime))
	}
	status.WriteString("\n💡 提示：\n")
	status.WriteString("• 使用 /unsubscribe <城市> 取消指定订阅\n")
	status.WriteString("• 使用 /weather <城市> 查询天气\n")
	status.WriteString("• 使用 /todo <城市> 管理待办")

	logger.Debug("Subscription status queried",
		zap.Int64("chat_id", chatID),
		zap.Int("subscription_count", len(subs)))
	return c.Send(status.String())
}

// HandleUnsubscribe handles the /unsubscribe command
func (h *Handlers) HandleUnsubscribe(c tele.Context) error {
	chatID := c.Sender().ID
	args := c.Args()
	logger.Debug("Received /unsubscribe command",
		zap.Int64("chat_id", chatID),
		zap.Strings("args", args))

	user, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		logger.Error("Failed to get user",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	subs, err := h.subRepo.FindByUserID(user.ID)
	if err != nil {
		logger.Error("Failed to find subscriptions",
			zap.Int64("chat_id", chatID),
			zap.Uint("user_id", user.ID),
			zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	if len(subs) == 0 {
		logger.Debug("No active subscriptions to unsubscribe",
			zap.Int64("chat_id", chatID),
			zap.Uint("user_id", user.ID))
		return c.Send("📭 您当前没有订阅每日提醒")
	}

	// Case 1: City specified in arguments
	if len(args) > 0 {
		city := args[0]
		sub, err := h.subRepo.FindByUserAndCity(user.ID, city)
		if err != nil {
			logger.Error("Failed to find subscription by city",
				zap.Int64("chat_id", chatID),
				zap.String("city", city),
				zap.Error(err))
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
		if sub == nil {
			return c.Send(fmt.Sprintf("❌ 未找到 %s 的订阅", city))
		}

		if err := h.subRepo.Delete(sub.ID); err != nil {
			logger.Error("Failed to delete subscription",
				zap.Int64("chat_id", chatID),
				zap.Uint("subscription_id", sub.ID),
				zap.Error(err))
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}

		logger.Info("Subscription cancelled",
			zap.Int64("chat_id", chatID),
			zap.Uint("subscription_id", sub.ID),
			zap.String("city", city))
		return c.Send(fmt.Sprintf("✅ 已成功取消 %s 的订阅", city))
	}

	// Case 2: No city specified and only one subscription
	if len(subs) == 1 {
		if err := h.subRepo.Delete(subs[0].ID); err != nil {
			logger.Error("Failed to delete subscription",
				zap.Int64("chat_id", chatID),
				zap.Uint("subscription_id", subs[0].ID),
				zap.Error(err))
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}

		logger.Info("Subscription cancelled",
			zap.Int64("chat_id", chatID),
			zap.Uint("subscription_id", subs[0].ID))
		return c.Send(fmt.Sprintf("✅ 已成功取消 %s 的订阅", subs[0].City))
	}

	// Case 3: No city specified and multiple subscriptions
	var list strings.Builder
	list.WriteString(fmt.Sprintf("您有 %d 个订阅，请指定要取消的城市：\n\n", len(subs)))
	for i, sub := range subs {
		list.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, sub.City, sub.ReminderTime))
	}
	list.WriteString("\n💡 使用方法：/unsubscribe <城市>")

	return c.Send(list.String())
}

// HandleWeather handles the /weather command
func (h *Handlers) HandleWeather(c tele.Context) error {
	chatID := c.Sender().ID
	logger.Debug("Received /weather command",
		zap.Int64("chat_id", chatID),
		zap.Strings("args", c.Args()))

	// Get user
	user, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		logger.Error("Failed to get user",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	// Get city from args or subscription
	var city string
	args := c.Args()
	if len(args) > 0 {
		city = args[0]
		logger.Debug("City from args", zap.String("city", city))
	} else {
		// Try to get from subscriptions
		subs, err := h.subRepo.FindByUserID(user.ID)
		if err != nil {
			logger.Error("Failed to find subscriptions",
				zap.Int64("chat_id", chatID),
				zap.Uint("user_id", user.ID),
				zap.Error(err))
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
		if len(subs) == 0 {
			logger.Debug("No subscription found for weather query",
				zap.Int64("chat_id", chatID),
				zap.Uint("user_id", user.ID))
			return c.Send("❌ 请指定城市或先使用 /subscribe 订阅\n用法: /weather <城市>")
		}
		city = subs[0].City
		logger.Debug("City from subscription", zap.String("city", city))

		// If user has multiple subscriptions, hint that they can specify city
		if len(subs) > 1 {
			var hint strings.Builder
			hint.WriteString("💡 您还订阅了其他城市：")
			for i := 1; i < len(subs) && i < 3; i++ {
				hint.WriteString(fmt.Sprintf(" %s", subs[i].City))
			}
			if len(subs) > 3 {
				hint.WriteString(" ...")
			}
			hint.WriteString("\n使用 /weather <城市> 可查询指定城市天气\n\n")
			defer func(hintText string) {
				// Send hint after weather report
				if err := c.Send(hintText); err != nil {
					logger.Warn("Failed to send weather hint", zap.Error(err))
				}
			}(hint.String())
		}
	}

	// Get full weather report with warnings and air quality
	report, err := h.weatherSvc.GetFullWeatherReport(city, h.airSvc, h.warningSvc)
	if err != nil {
		logger.Error("Failed to get weather report",
			zap.Int64("chat_id", chatID),
			zap.String("city", city),
			zap.Error(err))
		return c.Send(fmt.Sprintf("❌ 无法获取 %s 的天气信息，请检查城市名称是否正确。", city))
	}

	logger.Info("Weather report sent",
		zap.Int64("chat_id", chatID),
		zap.String("city", city))
	return c.Send(report)
}

// HandleTodo handles the /todo command with multi-subscription support
func (h *Handlers) HandleTodo(c tele.Context) error {
	chatID := c.Sender().ID
	args := c.Args()
	logger.Debug("Received /todo command",
		zap.Int64("chat_id", chatID),
		zap.Strings("args", args))

	// Get user
	user, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		logger.Error("Failed to get user", zap.Int64("chat_id", chatID), zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	// Get user's subscriptions
	subs, err := h.subRepo.FindByUserID(user.ID)
	if err != nil {
		logger.Error("Failed to find subscriptions", zap.Int64("chat_id", chatID), zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}
	if len(subs) == 0 {
		return c.Send("❌ 您还没有订阅任何城市\n请先使用 /subscribe <城市> <时间> 创建订阅")
	}

	// No arguments: list all todos grouped by city
	if len(args) == 0 {
		var result strings.Builder
		totalTodos := 0
		for _, sub := range subs {
			todos, err := h.todoSvc.GetSubscriptionTodos(sub.ID)
			if err != nil {
				logger.Warn("Failed to get todos for subscription",
					zap.Uint("subscription_id", sub.ID),
					zap.Error(err))
				continue
			}
			if len(todos) > 0 {
				result.WriteString(h.todoSvc.FormatTodoListWithCity(todos, sub.City))
				result.WriteString("\n")
				totalTodos += len(todos)
			}
		}
		if totalTodos == 0 {
			return c.Send("📝 暂无待办事项\n\n💡 使用 /todo <城市> add <内容> 添加待办")
		}
		return c.Send(result.String())
	}

	// Parse arguments: first arg might be city or action
	firstArg := args[0]
	var targetSub *model.Subscription
	var action string
	var actionArgs []string

	// Check if first argument is a city name
	for i := range subs {
		if subs[i].City == firstArg {
			targetSub = &subs[i]
			if len(args) > 1 {
				action = args[1]
				actionArgs = args[2:]
			}
			break
		}
	}

	// If not a city name, treat as action (only works with single subscription)
	if targetSub == nil {
		if len(subs) == 1 {
			targetSub = &subs[0]
			action = firstArg
			actionArgs = args[1:]
		} else {
			return c.Send("❌ 您有多个订阅，请指定城市\n\n用法:\n• /todo <城市> add <内容>\n• /todo <城市> done <编号>\n• /todo <城市> delete <编号>\n\n您的订阅城市：" + h.formatCityList(subs))
		}
	}

	// If no action, list todos for the specified city
	if action == "" {
		todos, err := h.todoSvc.GetSubscriptionTodos(targetSub.ID)
		if err != nil {
			logger.Error("Failed to get todos", zap.Uint("subscription_id", targetSub.ID), zap.Error(err))
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
		return c.Send(h.todoSvc.FormatTodoListWithCity(todos, targetSub.City))
	}

	// Handle actions
	switch action {
	case "add":
		if len(actionArgs) == 0 {
			return c.Send("❌ 用法: /todo " + targetSub.City + " add <内容>")
		}
		content := strings.Join(actionArgs, " ")
		if err := h.todoSvc.AddTodo(targetSub.ID, content); err != nil {
			logger.Error("Failed to add todo", zap.Error(err))
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
		logger.Info("Todo added", zap.String("city", targetSub.City), zap.String("content", content))
		return c.Send(fmt.Sprintf("✅ 已为 %s 添加待办：%s", targetSub.City, content))

	case "done":
		if len(actionArgs) == 0 {
			return c.Send("❌ 用法: /todo " + targetSub.City + " done <编号>")
		}
		todos, err := h.todoSvc.GetSubscriptionTodos(targetSub.ID)
		if err != nil {
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
		idx, err := strconv.Atoi(actionArgs[0])
		if err != nil || idx < 1 || idx > len(todos) {
			return c.Send("❌ 编号无效，请输入 1 到 " + strconv.Itoa(len(todos)) + " 之间的数字")
		}
		todoID := todos[idx-1].ID
		if err := h.todoSvc.CompleteTodo(todoID, user.ID); err != nil {
			logger.Error("Failed to complete todo", zap.Error(err))
			return c.Send("❌ 无法完成该待办事项")
		}
		logger.Info("Todo completed", zap.Uint("todo_id", todoID))
		return c.Send("✅ 待办事项已完成")

	case "delete", "del":
		if len(actionArgs) == 0 {
			return c.Send("❌ 用法: /todo " + targetSub.City + " delete <编号>")
		}
		todos, err := h.todoSvc.GetSubscriptionTodos(targetSub.ID)
		if err != nil {
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
		idx, err := strconv.Atoi(actionArgs[0])
		if err != nil || idx < 1 || idx > len(todos) {
			return c.Send("❌ 编号无效，请输入 1 到 " + strconv.Itoa(len(todos)) + " 之间的数字")
		}
		todoID := todos[idx-1].ID
		if err := h.todoSvc.DeleteTodo(todoID, user.ID); err != nil {
			logger.Error("Failed to delete todo", zap.Error(err))
			return c.Send("❌ 无法删除该待办事项")
		}
		logger.Info("Todo deleted", zap.Uint("todo_id", todoID))
		return c.Send("✅ 待办事项已删除")

	default:
		return c.Send("❌ 未知操作: " + action + "\n\n可用操作：add, done, delete")
	}
}

// formatCityList formats a list of cities for display
func (h *Handlers) formatCityList(subs []model.Subscription) string {
	var cities []string
	for _, sub := range subs {
		cities = append(cities, sub.City)
	}
	return strings.Join(cities, "、")
}

// HandleHelp handles the /help command
func (h *Handlers) HandleHelp(c tele.Context) error {
	chatID := c.Sender().ID
	logger.Debug("Received /help command", zap.Int64("chat_id", chatID))

	message := `📖 命令帮助

🔔 订阅管理
/subscribe <城市> <时间> - 订阅每日提醒
  示例: /subscribe 北京 08:00
  💡 可订阅多个城市（最多5个），每个城市独立管理
/mystatus - 查询所有订阅状态
/unsubscribe [城市] - 取消订阅
  示例: /unsubscribe 北京
  💡 不指定城市时，单订阅直接取消，多订阅需选择

☁️ 天气查询
/weather [城市] - 查询综合天气报告（含预警和空气质量）
  示例: /weather 上海
  💡 不指定城市时使用第一个订阅

🌫️ 空气质量
/air [城市] - 查询空气质量详情
  示例: /air 北京
  💡 包含 AQI、污染物浓度、未来预报

⚠️ 天气预警
/warning [城市] - 查询当前天气预警
  示例: /warning 深圳
/warning_toggle - 开启/关闭预警主动推送
  💡 开启后会自动推送所订阅城市的新预警

📝 待办事项（按城市分组）
/todo - 列出所有待办
/todo <城市> - 列出指定城市的待办
/todo <城市> add <内容> - 添加待办
  示例: /todo 北京 add 买菜
/todo <城市> done <编号> - 完成待办
/todo <城市> delete <编号> - 删除待办
  💡 单订阅时可省略城市名

❓ 其他
/start - 开始使用机器人
/help - 显示此帮助信息`

	return c.Send(message)
}

// isValidTimeFormat validates HH:MM time format
func isValidTimeFormat(timeStr string) bool {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return false
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return false
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return false
	}

	return true
}

// HandleAir handles the /air command
func (h *Handlers) HandleAir(c tele.Context) error {
	chatID := c.Sender().ID
	logger.Debug("Received /air command",
		zap.Int64("chat_id", chatID),
		zap.Strings("args", c.Args()))

	// Get user
	user, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		logger.Error("Failed to get user",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	// Get city from args or subscription
	var city string
	args := c.Args()
	if len(args) > 0 {
		city = args[0]
		logger.Debug("City from args", zap.String("city", city))
	} else {
		// Try to get from subscriptions
		subs, err := h.subRepo.FindByUserID(user.ID)
		if err != nil {
			logger.Error("Failed to find subscriptions",
				zap.Int64("chat_id", chatID),
				zap.Uint("user_id", user.ID),
				zap.Error(err))
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
		if len(subs) == 0 {
			logger.Debug("No subscription found for air quality query",
				zap.Int64("chat_id", chatID),
				zap.Uint("user_id", user.ID))
			return c.Send("❌ 请指定城市或先使用 /subscribe 订阅\n用法: /air <城市>")
		}
		city = subs[0].City
		logger.Debug("City from subscription", zap.String("city", city))

		// If user has multiple subscriptions, hint that they can specify city
		if len(subs) > 1 {
			var hint strings.Builder
			hint.WriteString("💡 您还订阅了其他城市：")
			for i := 1; i < len(subs) && i < 3; i++ {
				hint.WriteString(fmt.Sprintf(" %s", subs[i].City))
			}
			if len(subs) > 3 {
				hint.WriteString(" ...")
			}
			hint.WriteString("\n使用 /air <城市> 可查询指定城市空气质量\n\n")
			defer func(hintText string) {
				// Send hint after air quality report
				if err := c.Send(hintText); err != nil {
					logger.Warn("Failed to send air quality hint", zap.Error(err))
				}
			}(hint.String())
		}
	}

	// Get air quality report
	report, err := h.airSvc.GetAirQualityReport(city)
	if err != nil {
		logger.Error("Failed to get air quality report",
			zap.Int64("chat_id", chatID),
			zap.String("city", city),
			zap.Error(err))
		return c.Send(fmt.Sprintf("❌ 无法获取 %s 的空气质量信息，请检查城市名称是否正确。", city))
	}

	logger.Info("Air quality report sent",
		zap.Int64("chat_id", chatID),
		zap.String("city", city))
	return c.Send(report)
}

// HandleWarning handles the /warning [city] command
func (h *Handlers) HandleWarning(c tele.Context) error {
	chatID := c.Sender().ID
	logger.Debug("Received /warning command", zap.Int64("chat_id", chatID))

	// Get user
	user, err := h.userRepo.FindByChatID(chatID)
	if err != nil || user == nil {
		logger.Error("Failed to get user", zap.Int64("chat_id", chatID), zap.Error(err))
		return c.Send("获取用户信息失败，请先使用 /start 命令注册")
	}

	// Determine city to query
	var city string
	args := c.Args()

	if len(args) > 0 {
		// Use city from arguments
		city = strings.Join(args, " ")
	} else {
		// Use city from first active subscription
		subs, err := h.subRepo.FindByUserID(user.ID)
		if err != nil || len(subs) == 0 {
			logger.Warn("No active subscriptions",
				zap.Uint("user_id", user.ID),
				zap.Error(err))
			return c.Send("请指定城市名称，例如：/warning 北京\n或先使用 /subscribe 命令订阅城市")
		}
		city = subs[0].City

		// Hint if user has multiple subscriptions
		if len(subs) > 1 {
			defer func() {
				_ = c.Send(fmt.Sprintf("💡 提示：您订阅了多个城市，默认查询 %s\n要查询其他城市，请使用：/warning 城市名", city))
			}()
		}
	}

	logger.Debug("Querying weather warnings",
		zap.Int64("chat_id", chatID),
		zap.String("city", city))

	// Get warning report
	report, err := h.warningSvc.GetWarningReport(city)
	if err != nil {
		logger.Error("Failed to get warning report",
			zap.Int64("chat_id", chatID),
			zap.String("city", city),
			zap.Error(err))
		return c.Send(fmt.Sprintf("获取 %s 的天气预警失败：%v", city, err))
	}

	logger.Info("Weather warning report sent",
		zap.Int64("chat_id", chatID),
		zap.String("city", city))
	return c.Send(report)
}

// HandleWarningToggle handles the /warning_toggle command
func (h *Handlers) HandleWarningToggle(c tele.Context) error {
	chatID := c.Sender().ID
	logger.Debug("Received /warning_toggle command", zap.Int64("chat_id", chatID))

	// Get user
	user, err := h.userRepo.FindByChatID(chatID)
	if err != nil || user == nil {
		logger.Error("Failed to get user", zap.Int64("chat_id", chatID), zap.Error(err))
		return c.Send("获取用户信息失败，请先使用 /start 命令注册")
	}

	// Get all active subscriptions
	subs, err := h.subRepo.FindByUserID(user.ID)
	if err != nil || len(subs) == 0 {
		logger.Warn("No active subscriptions",
			zap.Uint("user_id", user.ID),
			zap.Error(err))
		return c.Send("您还没有订阅任何城市，请先使用 /subscribe 命令订阅")
	}

	// Toggle warning notification for all subscriptions
	var response strings.Builder
	response.WriteString("⚙️ 预警通知设置\n\n")

	allEnabled := true
	for _, sub := range subs {
		if !sub.EnableWarning {
			allEnabled = false
			break
		}
	}

	// Determine the new state (toggle all to opposite of current state)
	newState := !allEnabled

	// Update all subscriptions
	for i := range subs {
		subs[i].EnableWarning = newState
		if err := h.subRepo.Update(&subs[i]); err != nil {
			logger.Error("Failed to update subscription",
				zap.Uint("subscription_id", subs[i].ID),
				zap.Error(err))
			return c.Send(fmt.Sprintf("更新订阅 %s 失败：%v", subs[i].City, err))
		}
	}

	if newState {
		response.WriteString("✅ 已为所有订阅开启预警通知\n")
	} else {
		response.WriteString("🔕 已为所有订阅关闭预警通知\n")
	}

	response.WriteString("\n影响的订阅：\n")
	for _, sub := range subs {
		response.WriteString(fmt.Sprintf("   • %s\n", sub.City))
	}

	logger.Info("Warning notification toggled",
		zap.Uint("user_id", user.ID),
		zap.Bool("new_state", newState),
		zap.Int("subscription_count", len(subs)))

	return c.Send(response.String())
}
