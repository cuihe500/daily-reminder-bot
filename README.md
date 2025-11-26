# Daily Reminder Bot

一个功能完善的Telegram每日提醒机器人，提供天气播报、生活指数和待办事项管理。

## 功能特性

- 📍 **每日定时提醒**：订阅城市和时间，每天自动推送
- ☁️ **实时天气查询**：获取当前天气、温度、湿度等信息
- 👔 **生活指数**：穿衣、紫外线、运动等生活建议
- 📝 **待办事项管理**：添加、完成、删除待办项
- 🤖 **智能交互**：简单易用的命令行界面

## 技术栈

- **语言**: Go 1.23+
- **框架**: gopkg.in/telebot.v3
- **数据库**: SQLite + GORM
- **调度器**: robfig/cron
- **配置**: spf13/viper
- **天气API**: 和风天气 (QWeather)

## 项目结构

```
.
├── cmd/
│   └── bot/            # 主程序入口
├── configs/            # 配置文件
├── internal/
│   ├── bot/            # Telegram处理器
│   ├── config/         # 配置加载
│   ├── model/          # 数据库模型
│   ├── service/        # 业务逻辑
│   └── repository/     # 数据访问层
├── pkg/
│   └── qweather/       # 和风天气客户端
├── go.mod
└── README.md
```

## 快速开始

### 1. 前置要求

- Go 1.23 或更高版本
- Telegram Bot Token (从 [@BotFather](https://t.me/BotFather) 获取)
- 和风天气 API Key (从 [https://dev.qweather.com](https://dev.qweather.com) 获取)

### 2. 配置

复制配置模板并填写实际值：

```bash
cp configs/config.example.yaml configs/config.yaml
```

编辑 `configs/config.yaml`：

```yaml
telegram:
  token: "YOUR_TELEGRAM_BOT_TOKEN"

qweather:
  api_key: "YOUR_QWEATHER_API_KEY"
  base_url: "https://devapi.qweather.com/v7"

database:
  path: "./data/bot.db"

scheduler:
  timezone: "Asia/Shanghai"
```

### 3. 安装依赖

```bash
go mod download
```

### 4. 运行

```bash
go run cmd/bot/main.go
```

或构建后运行：

```bash
go build -o bot cmd/bot/main.go
./bot
```

### 5. 使用自定义配置路径

```bash
./bot -config /path/to/config.yaml
```

## 使用指南

### 基本命令

- `/register` - 注册机器人
- `/help` - 查看帮助信息
- `/subscribe 北京 08:00` - 订阅每日提醒
- ``

### 订阅每日提醒

```
/subscribe 北京 08:00
```

每天早上8点将收到北京的天气和待办提醒。

### 查询天气

```
/weather 上海
```

或者如果已订阅，直接使用：

```
/weather
```

### 待办事项管理

```
/todo                    # 列出所有待办
/todo add 买菜           # 添加待办
/todo done 1             # 完成编号为1的待办
/todo delete 2           # 删除编号为2的待办
```

## 开发指南

### 代码规范

- 遵循标准Go代码规范 (`gofmt`)
- 所有错误必须妥善处理，不可忽略
- 使用Conventional Commits规范提交代码

### 提交规范

- `feat`: 新功能
- `fix`: 错误修复
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构

### 构建

```bash
go build -o bot cmd/bot/main.go
```

### 测试

```bash
go test ./...
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
