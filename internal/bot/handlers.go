package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/internal/repository"
	"github.com/cuichanghe/daily-reminder-bot/internal/service"
	tele "gopkg.in/telebot.v3"
)

// Handlers holds all service dependencies for bot handlers
type Handlers struct {
	userRepo   *repository.UserRepository
	subRepo    *repository.SubscriptionRepository
	todoRepo   *repository.TodoRepository
	weatherSvc *service.WeatherService
	todoSvc    *service.TodoService
}

// NewHandlers creates a new Handlers instance
func NewHandlers(
	userRepo *repository.UserRepository,
	subRepo *repository.SubscriptionRepository,
	todoRepo *repository.TodoRepository,
	weatherSvc *service.WeatherService,
	todoSvc *service.TodoService,
) *Handlers {
	return &Handlers{
		userRepo:   userRepo,
		subRepo:    subRepo,
		todoRepo:   todoRepo,
		weatherSvc: weatherSvc,
		todoSvc:    todoSvc,
	}
}

// RegisterHandlers registers all command handlers
func (h *Handlers) RegisterHandlers(bot *tele.Bot) {
	bot.Handle("/start", h.HandleStart)
	bot.Handle("/subscribe", h.HandleSubscribe)
	bot.Handle("/mystatus", h.HandleMyStatus)
	bot.Handle("/unsubscribe", h.HandleUnsubscribe)
	bot.Handle("/weather", h.HandleWeather)
	bot.Handle("/todo", h.HandleTodo)
	bot.Handle("/help", h.HandleHelp)
}

// HandleStart handles the /start command
func (h *Handlers) HandleStart(c tele.Context) error {
	chatID := c.Sender().ID

	// Get or create user
	_, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	message := `👋 欢迎使用每日提醒机器人！

我可以帮你：
• 📍 订阅每日天气和生活指数
• ☁️ 查询实时天气
• 📝 管理待办事项

使用 /help 查看所有命令`

	return c.Send(message)
}

// HandleSubscribe handles the /subscribe command
func (h *Handlers) HandleSubscribe(c tele.Context) error {
	chatID := c.Sender().ID

	// Get or create user
	user, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	// Parse arguments: /subscribe <city> <time>
	// Example: /subscribe 北京 08:00
	args := c.Args()
	if len(args) < 2 {
		return c.Send("❌ 用法: /subscribe <城市> <时间>\n示例: /subscribe 北京 08:00")
	}

	city := args[0]
	reminderTime := args[1]

	// Validate time format (HH:MM)
	if !isValidTimeFormat(reminderTime) {
		return c.Send("❌ 时间格式错误，请使用 HH:MM 格式（如 08:00）")
	}

	// Check if user already has a subscription
	existingSub, err := h.subRepo.FindByUserID(user.ID)
	if err != nil {
		log.Printf("Error finding subscription: %v", err)
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	if existingSub != nil {
		// Update existing subscription
		existingSub.City = city
		existingSub.ReminderTime = reminderTime
		existingSub.Active = true
		if err := h.subRepo.Update(existingSub); err != nil {
			log.Printf("Error updating subscription: %v", err)
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
	} else {
		// Create new subscription
		sub := &model.Subscription{
			UserID:       user.ID,
			City:         city,
			ReminderTime: reminderTime,
			Active:       true,
		}
		if err := h.subRepo.Create(sub); err != nil {
			log.Printf("Error creating subscription: %v", err)
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
	}

	return c.Send(fmt.Sprintf("✅ 订阅成功！\n📍 城市：%s\n⏰ 时间：%s\n\n每天将在该时间为您推送天气和待办提醒。", city, reminderTime))
}

// HandleMyStatus handles the /mystatus command
func (h *Handlers) HandleMyStatus(c tele.Context) error {
	chatID := c.Sender().ID

	user, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	sub, err := h.subRepo.FindByUserID(user.ID)
	if err != nil {
		log.Printf("Error finding subscription: %v", err)
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	if sub == nil || !sub.Active {
		return c.Send("📭 您当前没有订阅每日提醒\n\n使用 /subscribe <城市> <时间> 开始订阅")
	}

	return c.Send(fmt.Sprintf("📬 您的订阅状态\n\n📍 城市：%s\n⏰ 提醒时间：%s\n✅ 状态：已激活\n\n使用 /unsubscribe 可以取消订阅", sub.City, sub.ReminderTime))
}

// HandleUnsubscribe handles the /unsubscribe command
func (h *Handlers) HandleUnsubscribe(c tele.Context) error {
	chatID := c.Sender().ID

	user, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	sub, err := h.subRepo.FindByUserID(user.ID)
	if err != nil {
		log.Printf("Error finding subscription: %v", err)
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	if sub == nil || !sub.Active {
		return c.Send("📭 您当前没有订阅每日提醒")
	}

	sub.Active = false
	if err := h.subRepo.Update(sub); err != nil {
		log.Printf("Error updating subscription: %v", err)
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	return c.Send("✅ 已成功取消订阅\n\n使用 /subscribe <城市> <时间> 可以重新订阅")
}

// HandleWeather handles the /weather command
func (h *Handlers) HandleWeather(c tele.Context) error {
	chatID := c.Sender().ID

	// Get user
	user, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	// Get city from args or subscription
	var city string
	args := c.Args()
	if len(args) > 0 {
		city = args[0]
	} else {
		// Try to get from subscription
		sub, err := h.subRepo.FindByUserID(user.ID)
		if err != nil {
			log.Printf("Error finding subscription: %v", err)
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
		if sub == nil {
			return c.Send("❌ 请指定城市或先使用 /subscribe 订阅\n用法: /weather <城市>")
		}
		city = sub.City
	}

	// Get weather report
	report, err := h.weatherSvc.GetWeatherReport(city)
	if err != nil {
		log.Printf("Error getting weather: %v", err)
		return c.Send(fmt.Sprintf("❌ 无法获取 %s 的天气信息，请检查城市名称是否正确。", city))
	}

	return c.Send(report)
}

// HandleTodo handles the /todo command
func (h *Handlers) HandleTodo(c tele.Context) error {
	chatID := c.Sender().ID

	// Get user
	user, err := h.userRepo.GetOrCreate(chatID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return c.Send("抱歉,系统出现错误,请稍后再试。")
	}

	args := c.Args()
	if len(args) == 0 {
		// List all todos
		todos, err := h.todoSvc.GetUserTodos(user.ID)
		if err != nil {
			log.Printf("Error getting todos: %v", err)
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
		return c.Send(h.todoSvc.FormatTodoList(todos))
	}

	action := args[0]
	switch action {
	case "add":
		if len(args) < 2 {
			return c.Send("❌ 用法: /todo add <内容>")
		}
		content := strings.Join(args[1:], " ")
		if err := h.todoSvc.AddTodo(user.ID, content); err != nil {
			log.Printf("Error adding todo: %v", err)
			return c.Send("抱歉,系统出现错误,请稍后再试。")
		}
		return c.Send("✅ 待办事项已添加")

	case "done":
		if len(args) < 2 {
			return c.Send("❌ 用法: /todo done <编号>")
		}
		id, err := strconv.Atoi(args[1])
		if err != nil {
			return c.Send("❌ 编号必须是数字")
		}
		if err := h.todoSvc.CompleteTodo(uint(id), user.ID); err != nil {
			log.Printf("Error completing todo: %v", err)
			return c.Send("❌ 无法完成该待办事项，请检查编号是否正确")
		}
		return c.Send("✅ 待办事项已完成")

	case "delete", "del":
		if len(args) < 2 {
			return c.Send("❌ 用法: /todo delete <编号>")
		}
		id, err := strconv.Atoi(args[1])
		if err != nil {
			return c.Send("❌ 编号必须是数字")
		}
		if err := h.todoSvc.DeleteTodo(uint(id), user.ID); err != nil {
			log.Printf("Error deleting todo: %v", err)
			return c.Send("❌ 无法删除该待办事项，请检查编号是否正确")
		}
		return c.Send("✅ 待办事项已删除")

	default:
		return c.Send("❌ 未知操作\n用法:\n/todo - 列出所有待办\n/todo add <内容> - 添加待办\n/todo done <编号> - 完成待办\n/todo delete <编号> - 删除待办")
	}
}

// HandleHelp handles the /help command
func (h *Handlers) HandleHelp(c tele.Context) error {
	message := `📖 命令帮助

/start - 开始使用机器人
/subscribe <城市> <时间> - 订阅每日提醒
  示例: /subscribe 北京 08:00
/mystatus - 查询订阅状态
/unsubscribe - 取消订阅

/weather [城市] - 查询天气
  示例: /weather 上海

/todo - 待办事项管理
  /todo - 列出所有待办
  /todo add <内容> - 添加待办
  /todo done <编号> - 完成待办
  /todo delete <编号> - 删除待办

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
