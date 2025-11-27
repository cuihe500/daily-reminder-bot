# 项目章程：每日提醒机器人

## 1. 项目概述
一个专为订阅用户发送每日提醒的 Telegram 机器人。内容包括本地天气、气温、穿衣建议（生活指数）以及个人待办事项。

## 2. 技术栈
基于性能、稳定性和部署便捷性进行选择。

- **编程语言**：Go (Golang) 1.23+
    - *原因*：高性能，强大的并发处理能力以应对多用户场景，支持单二进制文件部署。
- **机器人框架**：`gopkg.in/telebot.v3`
    - *原因*：现代化的、支持中间件且类型安全的 Telegram Bot API 封装库。
- **数据库**：SQLite（配合 GORM）
    - *原因*：轻量级、无服务器架构、易于备份。GORM 提供抽象层，便于未来迁移至 PostgreSQL。
- **调度器**：`github.com/robfig/cron/v3`
    - *原因*：Go 语言中处理 cron 定时任务的行业标准，稳定可靠。
- **天气 API**：和风天气 (QWeather)
    - *原因*：对中国地区覆盖出色，提供规范所需的详细"生活指数"（穿衣、紫外线、运动等）。
- **配置管理**：`github.com/spf13/viper`
    - *原因*：配置管理领域的行业标准（支持环境变量、配置文件）。

## 3. 项目结构（标准 Go 项目布局）
```
.
├── cmd/
│   └── bot/            # 主程序入口
├── configs/            # 配置文件
├── internal/
│   ├── bot/            # Telegram 处理器和逻辑
│   ├── config/         # 配置加载
│   ├── model/          # 数据库模型
│   ├── service/        # 业务逻辑（天气、待办、调度）
│   └── repository/     # 数据访问层
├── pkg/
│   └── qweather/       # QWeather API 客户端
├── go.mod
└── CLAUDE.md
```

## 4. 开发规范
- **代码风格**：遵循标准 Go 规范（`gofmt`）。
- **错误处理**：为错误添加上下文信息；禁止忽略错误。
- **提交规范**：采用约定式提交（Conventional Commits）（feat、fix、docs、style、refactor）。

## 5. 命令
- `/start`：欢迎信息和注册。
- `/subscribe`：设置每日提醒的位置和时间。
- `/mystatus`：查询当前订阅状态。
- `/unsubscribe`：取消每日提醒订阅。
- `/weather`：获取即时天气报告。
- `/todo`：管理待办事项列表。
- `/help`：显示帮助信息。