# Daily Reminder Bot

一个功能完善的Telegram每日提醒机器人，提供天气播报、生活指数和待办事项管理。

## 功能特性

- 📍 **每日定时提醒**：订阅城市和时间，每天自动推送
- ☁️ **实时天气查询**：获取当前天气、温度、湿度等信息
- 👔 **生活指数**：穿衣、紫外线、运动等生活建议
- 🌬️ **空气质量监测**：AQI、PM2.5、PM10 等污染物指标
- ⚠️ **天气预警推送**：极端天气预警实时通知
- 📝 **待办事项管理**：添加、完成、删除待办项
- 🤖 **AI 智能提醒**：可选的 AI 个性化提醒内容（支持 OpenAI、DeepSeek 等）
- 📅 **农历日历**：节气、传统节日、法定假期信息

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
│   ├── bot/            # 主程序入口
│   └── debug_api/      # API 调试工具
├── configs/            # 配置文件
├── internal/
│   ├── bot/            # Telegram 处理器
│   ├── config/         # 配置加载
│   ├── migration/      # 数据库迁移
│   ├── model/          # 数据库模型
│   ├── repository/     # 数据访问层
│   └── service/        # 业务逻辑
├── pkg/
│   ├── calendar/       # 农历/节气计算
│   ├── holiday/        # 法定假日 API
│   ├── logger/         # 日志系统
│   ├── openai/         # AI API 客户端
│   └── qweather/       # 和风天气客户端
├── go.mod
├── Makefile            # 构建脚本
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
  auth_mode: "jwt"  # 推荐使用 jwt，也支持 api_key
  private_key_path: "./configs/ed25519-private.pem"
  key_id: "YOUR_KEY_ID"
  project_id: "YOUR_PROJECT_ID"
  base_url: "https://YOUR_HOST.qweatherapi.com"

database:
  type: "sqlite"
  path: "./data/bot.db"

scheduler:
  timezone: "Asia/Shanghai"
```

#### 和风天气 JWT 认证配置

JWT 认证比传统 API Key 更安全，推荐使用。

**步骤 1：生成 Ed25519 密钥对**

```bash
# 在项目根目录执行
openssl genpkey -algorithm ED25519 -out configs/ed25519-private.pem \
  && openssl pkey -pubout -in configs/ed25519-private.pem > configs/ed25519-public.pem
```

这将生成两个文件：
- `ed25519-private.pem` - 私钥，保存在本地，用于签名
- `ed25519-public.pem` - 公钥，需上传到和风天气控制台

**步骤 2：上传公钥到和风天气控制台**

1. 访问 [控制台-项目管理](https://console.qweather.com/project)
2. 点击你的项目，然后点击"添加凭据"
3. 选择 **JSON Web Token** 认证方式
4. 复制 `configs/ed25519-public.pem` 的全部内容粘贴到公钥文本框
5. 保存后记录下 **凭据 ID** 和 **项目 ID**

**步骤 3：获取 API Host**

访问 [控制台-设置](https://console.qweather.com/setting) 查看你的 API Host（格式如 `abc123.qweatherapi.com`）

**步骤 4：更新配置文件**

```yaml
qweather:
  auth_mode: "jwt"
  private_key_path: "./configs/ed25519-private.pem"
  key_id: "填入凭据ID"
  project_id: "填入项目ID"
  base_url: "https://你的APIHost.qweatherapi.com"
```

> **注意**：私钥文件（`ed25519-private.pem`）已在 `.gitignore` 中忽略，请妥善保管。

#### 使用传统 API Key（备选）

如果不想使用 JWT，也可以使用传统 API Key 方式：

```yaml
qweather:
  auth_mode: "api_key"
  api_key: "YOUR_QWEATHER_API_KEY"
  base_url: "https://devapi.qweather.com"
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

- `/start` - 开始使用机器人
- `/help` - 查看帮助信息
- `/subscribe <城市> <时间>` - 订阅每日提醒
- `/mystatus` - 查询订阅状态
- `/unsubscribe` - 取消订阅
- `/weather [城市]` - 查询天气
- `/air [城市]` - 查询空气质量
- `/warning [城市]` - 查询天气预警
- `/warning_toggle` - 开启/关闭天气预警推送
- `/todo` - 待办事项管理

### 订阅每日提醒

```
/subscribe 北京 08:00
```

每天早上8点将收到北京的天气和待办提醒。

### 查询订阅状态

```
/mystatus
```

查看当前的订阅信息，包括城市和提醒时间。

### 取消订阅

```
/unsubscribe
```

取消每日提醒订阅，可随时使用 `/subscribe` 重新订阅。

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

### 空气质量查询

```
/air 北京
```

获取指定城市的实时空气质量信息，包括 AQI 指数和各项污染物浓度。

### 天气预警

```
/warning 北京            # 查询北京的天气预警
/warning_toggle          # 开启/关闭预警推送
```

启用预警推送后，当订阅城市发布新预警时会自动通知。

## Docker 部署

### 使用 Docker Compose（推荐）

1. 复制环境变量模板：

```bash
cp env.example .env
```

2. 编辑 `.env` 文件，填写必要配置：

```bash
# 必填配置
TELEGRAM_TOKEN=your_telegram_bot_token
QWEATHER_AUTH_MODE=jwt
QWEATHER_KEY_ID=your_key_id
QWEATHER_PROJECT_ID=your_project_id
QWEATHER_BASE_URL=https://your-api-host.qweatherapi.com
```

3. 启动容器：

```bash
docker-compose up -d
```

4. 查看日志：

```bash
docker-compose logs -f
```

### 首次运行密钥生成

首次启动容器时，如果没有提供 `QWEATHER_PRIVATE_KEY`，系统会自动生成 Ed25519 密钥对。

1. 查看生成的公钥：

```bash
docker-compose logs | grep "Public Key Content" -A 10
# 或者
docker-compose exec daily-reminder-bot cat /app/configs/ed25519-public.pem
```

2. 将公钥上传到[和风天气控制台](https://console.qweather.com/project)
3. 获取凭据 ID 和项目 ID，更新 `.env` 文件
4. 重启容器：`docker-compose restart`

### 手动构建与运行

```bash
# 构建镜像
docker build -t daily-reminder-bot .

# 运行容器
docker run -d \
  --name daily-reminder-bot \
  -e TELEGRAM_TOKEN=your_token \
  -e QWEATHER_AUTH_MODE=jwt \
  -e QWEATHER_KEY_ID=your_key_id \
  -e QWEATHER_PROJECT_ID=your_project_id \
  -e QWEATHER_BASE_URL=https://your-host.qweatherapi.com \
  -v bot-data:/app/data \
  -v bot-configs:/app/configs \
  daily-reminder-bot
```

### 使用已有私钥

如果已有 Ed25519 私钥，可以通过环境变量注入：

```bash
# 直接使用 PEM 格式（注意处理换行符）
QWEATHER_PRIVATE_KEY=$(cat your-private-key.pem)

# 或使用 base64 编码
QWEATHER_PRIVATE_KEY=$(cat your-private-key.pem | base64)
```

### 环境变量列表

| 变量名 | 必填 | 默认值 | 说明 |
|--------|------|--------|------|
| `TELEGRAM_TOKEN` | ✓ | - | Telegram Bot Token |
| `QWEATHER_AUTH_MODE` | - | `jwt` | 认证模式 (`jwt` 或 `api_key`) |
| `QWEATHER_PRIVATE_KEY` | - | - | Ed25519 私钥（PEM 或 base64） |
| `QWEATHER_KEY_ID` | ✓ (jwt) | - | JWT 凭据 ID |
| `QWEATHER_PROJECT_ID` | ✓ (jwt) | - | 项目 ID |
| `QWEATHER_BASE_URL` | ✓ | - | API Host |
| `DATABASE_TYPE` | - | `sqlite` | 数据库类型 |
| `OPENAI_ENABLED` | - | `false` | 是否启用 AI |
| `SCHEDULER_TIMEZONE` | - | `Asia/Shanghai` | 时区 |

完整环境变量列表请参考 `env.example`。

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
