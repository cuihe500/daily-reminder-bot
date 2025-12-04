# 项目章程：每日提醒机器人

## 1. 项目概述
一个专为订阅用户发送每日提醒的 Telegram 机器人。内容包括本地天气、气温、穿衣建议（生活指数）以及个人待办事项。

## 2. 技术栈
基于性能、稳定性和部署便捷性进行选择。

- **编程语言**：Go (Golang) 1.23+（已在 1.25.1 测试）
    - *原因*：高性能，强大的并发处理能力以应对多用户场景，支持单二进制文件部署。
- **机器人框架**：`gopkg.in/telebot.v3` (v3.3.8)
    - *原因*：现代化的、支持中间件且类型安全的 Telegram Bot API 封装库。
- **数据库**：SQLite / MySQL（配合 GORM v1.31.1）
    - *原因*：SQLite 适合小规模部署，轻量级、无服务器架构；MySQL 适合生产环境。GORM 提供统一抽象层，便于切换数据库。
    - *支持*：`gorm.io/driver/sqlite` 和 `gorm.io/driver/mysql`
- **调度器**：`github.com/robfig/cron/v3` (v3.0.1)
    - *原因*：Go 语言中处理 cron 定时任务的行业标准，稳定可靠。
- **天气 API**：和风天气 (QWeather)
    - *原因*：对中国地区覆盖出色，提供规范所需的详细"生活指数"（穿衣、紫外线、运动等）。
- **配置管理**：`github.com/spf13/viper` (v1.19.0)
    - *原因*：配置管理领域的行业标准（支持环境变量、配置文件）。
- **农历计算**：`github.com/6tail/lunar-go` (v1.4.6)
    - *原因*：功能完善的农历、节气、节日计算库，支持阳历和农历互转。
- **日志系统**：`go.uber.org/zap` (v1.27.1)
    - *原因*：高性能结构化日志，支持多种输出格式，集成 GORM 日志适配器。
- **AI 服务**：OpenAI 兼容 API（可选）
    - *原因*：支持多种 LLM 提供商（OpenAI、DeepSeek、智谱、通义千问等），生成个性化提醒内容。
    - *配置*：支持自定义 base_url、model、temperature 等参数。
- **假期 API**：节假日 API（可选）
    - *原因*：获取中国法定节假日和调休信息。
    - *支持*：jiejiariapi.com 和 holiday.ailcc.com 两个数据源。

## 3. 项目结构（标准 Go 项目布局）
```
.
├── cmd/
│   ├── bot/            # 主程序入口（main.go）
│   └── debug_api/      # API 调试工具
├── configs/            # 配置文件
│   ├── config.example.yaml  # 配置模板
│   ├── config.yaml          # 实际配置（需自行创建）
│   └── ed25519-private.pem  # JWT 私钥（需自行生成）
├── data/               # 数据库文件目录
│   └── bot.db          # SQLite 数据库文件
├── build/              # 编译输出目录
├── internal/
│   ├── bot/            # Telegram 处理器和逻辑
│   │   ├── bot.go      # 机器人初始化
│   │   └── handlers.go # 命令处理器
│   ├── config/         # 配置加载
│   │   └── config.go   # Viper 配置管理
│   ├── migration/      # 数据库迁移
│   │   └── migrate.go  # 自动迁移逻辑
│   ├── model/          # 数据库模型
│   │   ├── user.go         # 用户模型
│   │   ├── subscription.go # 订阅模型
│   │   ├── todo.go         # 待办事项模型
│   │   └── warning_log.go  # 天气预警日志模型
│   ├── repository/     # 数据访问层
│   │   ├── user.go         # 用户数据操作
│   │   ├── subscription.go # 订阅数据操作
│   │   ├── todo.go         # 待办数据操作
│   │   └── warning_log.go  # 预警日志操作
│   └── service/        # 业务逻辑层
│       ├── scheduler.go    # 定时任务调度
│       ├── weather.go      # 天气服务
│       ├── air.go          # 空气质量服务
│       ├── warning.go      # 天气预警服务
│       ├── todo.go         # 待办服务
│       ├── calendar.go     # 日历服务（节气、节日）
│       └── ai.go           # AI 提醒生成服务
├── pkg/                # 可复用的公共包
│   ├── calendar/       # 日历计算工具
│   │   ├── calculator.go   # 农历计算
│   │   ├── festivals.go    # 节日查询
│   │   └── types.go        # 类型定义
│   ├── holiday/        # 假期 API 客户端
│   │   └── client.go   # 节假日查询客户端
│   ├── logger/         # 日志系统
│   │   ├── logger.go       # Zap 日志初始化
│   │   ├── gorm_adapter.go # GORM 日志适配器
│   │   └── sanitize.go     # 敏感信息过滤
│   ├── openai/         # OpenAI 兼容 API 客户端
│   │   ├── client.go   # API 客户端
│   │   └── types.go    # 请求/响应类型
│   └── qweather/       # 和风天气 API 客户端
│       ├── client.go   # API 客户端
│       ├── types.go    # 天气数据类型
│       ├── air.go      # 空气质量 API
│       └── warning.go  # 天气预警 API
├── go.mod              # Go 模块依赖
├── go.sum              # 依赖校验和
├── Makefile            # 构建脚本
├── README.md           # 项目文档
└── CLAUDE.md           # 项目章程
```

## 4. 核心功能模块

### 4.1 日历服务（Calendar Service）
- 阳历与农历互转（基于 lunar-go）
- 节气计算（二十四节气）
- 节日查询（阳历节日、农历节日）
- 除夕日期自动计算（处理闰月情况）

### 4.2 天气服务（Weather Service）
- 实时天气查询（和风天气 API）
- 未来天气预报
- 生活指数（穿衣、运动、紫外线等）
- 空气质量查询（AQI、PM2.5、PM10等污染物）
- 空气质量预报（未来5天）
- 天气预警信息（极端天气预警）
- 城市查询支持（支持中文城市名）

### 4.3 待办事项服务（Todo Service）
- 待办事项增删改查
- 待办状态管理（待完成/已完成）
- 按用户隔离数据

### 4.4 定时任务调度（Scheduler Service）
- 基于 cron 表达式的定时任务
- 动态添加/删除用户订阅任务
- 时区支持（默认 Asia/Shanghai）

### 4.5 AI 提醒生成（AI Service，可选）
- 基于天气、节日、待办生成个性化提醒
- 支持多种 LLM 提供商（OpenAI、DeepSeek、智谱等）
- 自动重试机制和超时控制

### 4.6 节假日查询（Holiday Service，可选）
- 中国法定节假日查询
- 调休信息获取
- 本地缓存机制（24小时 TTL）

## 5. 配置说明

### 5.1 必需配置
- `telegram.token`：Telegram Bot Token
- `telegram.api_endpoint`：Telegram Bot API 端点（可选，默认官方 API）
- `qweather.auth_mode`：认证模式（jwt 或 api_key）
- `qweather.private_key_path`：JWT 私钥路径（jwt 模式必需）
- `qweather.key_id`：凭据 ID（jwt 模式必需）
- `qweather.project_id`：项目 ID（jwt 模式必需）
- `qweather.api_key`：和风天气 API Key（api_key 模式必需）
- `qweather.base_url`：API 基础 URL
- `database.type`：数据库类型（sqlite 或 mysql）
- `scheduler.timezone`：时区设置

### 5.2 可选配置
- `openai.*`：AI 服务配置（启用个性化提醒）
- `holiday.api_url`：节假日 API 地址
- `logger.level`：日志级别（debug/info/warn/error）
- `logger.format`：日志格式（console/json）

### 5.3 数据库配置
**SQLite 模式**：
- `database.type: "sqlite"`
- `database.path`：数据库文件路径

**MySQL 模式**：
- `database.type: "mysql"`
- `database.host/port/user/password/dbname`

## 6. 开发规范
- **代码风格**：遵循标准 Go 规范（`gofmt`、`golint`）。
- **错误处理**：为错误添加上下文信息；禁止忽略错误；使用 `fmt.Errorf` 包装错误。
- **日志规范**：使用结构化日志（zap），敏感信息需过滤（如 API Key、Token）。
- **提交规范**：采用约定式提交（Conventional Commits）
  - `feat`：新功能
  - `fix`：错误修复
  - `docs`：文档更新
  - `style`：代码格式调整（不影响功能）
  - `refactor`：代码重构
  - `test`：测试相关
  - `chore`：构建工具或辅助工具变动

## 7. 机器人命令列表

### 基础命令
- `/start`：欢迎信息和用户注册
- `/help`：显示帮助信息和可用命令

### 订阅管理
- `/subscribe <城市> <时间>`：设置每日提醒（例：`/subscribe 北京 08:00`）
- `/mystatus`：查询当前订阅状态（城市、提醒时间）
- `/unsubscribe`：取消每日提醒订阅

### 功能命令
- `/weather [城市]`：获取即时天气报告（可选城市参数，默认使用订阅城市）
- `/air [城市]`：获取空气质量信息（AQI、PM2.5 等）
- `/warning [城市]`：获取天气预警信息
- `/warning_toggle`：开启/关闭天气预警推送
- `/todo`：待办事项管理
  - `/todo` - 列出所有待办
  - `/todo add <内容>` - 添加待办
  - `/todo done <编号>` - 完成待办
  - `/todo delete <编号>` - 删除待办

## 8. 数据模型

### User（用户）
- `id`：主键
- `telegram_id`：Telegram 用户 ID
- `username`：Telegram 用户名
- `first_name`：名
- `last_name`：姓
- `created_at`：创建时间
- `updated_at`：更新时间

### Subscription（订阅）
- `id`：主键
- `user_id`：用户 ID（外键）
- `city`：城市名称
- `reminder_time`：提醒时间（HH:MM 格式）
- `enabled`：是否启用
- `created_at`：创建时间
- `updated_at`：更新时间

### Todo（待办事项）
- `id`：主键
- `user_id`：用户 ID（外键）
- `content`：待办内容
- `completed`：是否完成
- `created_at`：创建时间
- `updated_at`：更新时间

### WarningLog（天气预警日志）
- `id`：主键
- `warning_id`：和风天气预警 ID（唯一索引）
- `location_id`：位置 ID
- `city`：城市名称
- `type`：预警类型
- `level`：预警级别
- `title`：预警标题
- `start_time`：预警开始时间
- `end_time`：预警结束时间
- `status`：预警状态（active/update/cancel）
- `notified_at`：通知发送时间
- `created_at`：创建时间
- `updated_at`：更新时间

注意（必须遵守）：
1. 若makefile内有，则使用make内的命令