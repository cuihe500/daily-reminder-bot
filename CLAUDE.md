# Project Charter: Daily Reminder Bot

## 1. Project Overview
A Telegram bot designed to send daily reminders to subscribed users. Content includes local weather, temperature, clothing advice (life indices), and personal todo items.

## 2. Technology Stack
Selected for performance, stability, and ease of deployment.

- **Language**: Go (Golang) 1.23+
    - *Reason*: High performance, strong concurrency for handling multiple users, single binary deployment.
- **Bot Framework**: `gopkg.in/telebot.v3`
    - *Reason*: Modern, middleware-friendly, and type-safe wrapper for the Telegram Bot API.
- **Database**: SQLite (with GORM)
    - *Reason*: Lightweight, serverless, easy to back up. GORM provides an abstraction layer allowing future migration to PostgreSQL if needed.
- **Scheduler**: `github.com/robfig/cron/v3`
    - *Reason*: Robust standard for cron-style job scheduling in Go.
- **Weather API**: QWeather (和风天气)
    - *Reason*: Excellent coverage for China, provides detailed "Life Indices" (clothing, UV, sports) required by the spec.
- **Configuration**: `github.com/spf13/viper`
    - *Reason*: Industry standard for configuration management (supports env vars, config files).

## 3. Project Structure (Standard Go Layout)
```
.
├── cmd/
│   └── bot/            # Main entry point
├── configs/            # Configuration files
├── internal/
│   ├── bot/            # Telegram handlers and logic
│   ├── config/         # Config loading
│   ├── model/          # Database models
│   ├── service/        # Business logic (Weather, Todo, Scheduler)
│   └── repository/     # Data access layer
├── pkg/
│   └── qweather/       # QWeather API client
├── go.mod
└── CLAUDE.md
```

## 4. Development Guidelines
- **Code Style**: Follow standard Go conventions (`gofmt`).
- **Error Handling**: Wrap errors with context; do not ignore errors.
- **Commits**: Conventional Commits (feat, fix, docs, style, refactor).

## 5. Commands
- `/start`: Welcome and registration.
- `/subscribe`: Set location and time for daily reminders.
- `/weather`: Get instant weather report.
- `/todo`: Manage todo list.
- `/help`: Show help message.
