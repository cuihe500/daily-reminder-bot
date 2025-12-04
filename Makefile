# Makefile for Daily Reminder Bot
# Go 项目构建、测试和部署管理

# 变量定义
BINARY_NAME=daily-reminder-bot
MAIN_PATH=cmd/bot/main.go
BUILD_DIR=build
CONFIG_FILE=configs/config.yaml

# Go 相关变量
GO=go
GOFLAGS=-v
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOMOD=$(GO) mod
GOFMT=gofmt

# 版本信息（从 git 获取）
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 编译标志
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)"

# 默认目标
.PHONY: all
all: clean deps fmt lint build

# 帮助信息
.PHONY: help
help:
	@echo "Daily Reminder Bot - Makefile 命令说明"
	@echo ""
	@echo "使用方式: make [目标]"
	@echo ""
	@echo "可用目标:"
	@echo "  build        - 编译项目（输出到 $(BUILD_DIR)/$(BINARY_NAME)）"
	@echo "  run          - 编译并运行项目"
	@echo "  dev          - 开发模式运行（不编译）"
	@echo "  clean        - 清理构建产物和缓存"
	@echo "  deps         - 下载和整理依赖"
	@echo "  test         - 运行所有测试"
	@echo "  test-cover   - 运行测试并生成覆盖率报告"
	@echo "  fmt          - 格式化代码"
	@echo "  fmt-check    - 检查代码格式（不自动修改）"
	@echo "  lint         - 代码静态检查"
	@echo "  vet          - Go vet 检查"
	@echo "  install      - 安装二进制文件到 GOPATH/bin"
	@echo "  release      - 构建生产版本（优化编译）"
	@echo "  docker       - 构建 Docker 镜像"
	@echo "  docker-build - 交互式构建 Docker 镜像"
	@echo "  docker-up    - 启动 Docker Compose 服务"
	@echo "  docker-down  - 停止 Docker Compose 服务"
	@echo "  docker-logs  - 查看 Docker Compose 日志"
	@echo "  docker-clean - 清理 Docker 资源"
	@echo "  ci           - CI/CD 检查（格式检查+lint+测试）"
	@echo "  help         - 显示此帮助信息"
	@echo ""

# 编译项目
.PHONY: build
build: deps fmt lint
	@echo "==> 正在编译项目..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "==> 编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 生产环境编译（优化）
.PHONY: release
release: clean deps fmt lint
	@echo "==> 正在构建生产版本..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "==> 生产版本构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 编译并运行
.PHONY: run
run: build
	@echo "==> 运行程序..."
	./$(BUILD_DIR)/$(BINARY_NAME) -config $(CONFIG_FILE)

# 开发模式运行（使用 go run）
.PHONY: dev
dev:
	@echo "==> 开发模式运行..."
	$(GO) run $(MAIN_PATH) -config $(CONFIG_FILE)

# 清理构建产物
.PHONY: clean
clean:
	@echo "==> 清理构建产物..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "==> 清理完成"

# 深度清理（包括依赖缓存）
.PHONY: clean-all
clean-all: clean
	@echo "==> 清理依赖缓存..."
	$(GO) clean -modcache
	@echo "==> 深度清理完成"

# 依赖管理
.PHONY: deps
deps:
	@echo "==> 下载依赖..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "==> 依赖准备完成"

# 更新依赖
.PHONY: deps-update
deps-update:
	@echo "==> 更新依赖..."
	$(GOMOD) get -u ./...
	$(GOMOD) tidy
	@echo "==> 依赖更新完成"

# 运行测试
.PHONY: test
test:
	@echo "==> 运行测试..."
	$(GOTEST) -v -race ./...

# 测试覆盖率
.PHONY: test-cover
test-cover:
	@echo "==> 运行测试（包含覆盖率）..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "==> 覆盖率报告已生成: coverage.html"

# 代码格式化
.PHONY: fmt
fmt:
	@echo "==> 格式化代码..."
	$(GOFMT) -s -w .
	$(GO) fmt ./...
	@echo "==> 格式化完成"

# 检查代码格式（不自动修改）
.PHONY: fmt-check
fmt-check:
	@echo "==> 检查代码格式..."
	@UNFORMATTED=$$($(GOFMT) -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "以下文件需要格式化:"; \
		echo "$$UNFORMATTED"; \
		echo "请运行: make fmt"; \
		exit 1; \
	fi
	@echo "==> 代码格式检查通过"

# 代码检查
.PHONY: vet
vet:
	@echo "==> 运行 go vet..."
	$(GO) vet ./...
	@echo "==> 检查完成"

# Lint 检查（需要安装 golangci-lint）
.PHONY: lint
lint:
	@echo "==> 运行 lint 检查..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "错误: golangci-lint 未安装"; \
		echo "安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		echo "或运行: make init"; \
		exit 1; \
	fi
	@golangci-lint run ./...
	@echo "==> lint 检查通过"

# 安装到系统
.PHONY: install
install:
	@echo "==> 安装到 GOPATH/bin..."
	$(GO) install $(LDFLAGS) $(MAIN_PATH)
	@echo "==> 安装完成"

# Docker 构建
.PHONY: docker
docker:
	@echo "===> 构建 Docker 镜像..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "===> Docker 镜像构建完成"

# Docker 交互式构建
.PHONY: docker-build
docker-build:
	@echo "=========================================="
	@echo "  Docker 镜像交互式构建"
	@echo "=========================================="
	@echo ""
	@read -p "请输入镜像名称 [$(BINARY_NAME)]: " IMAGE_NAME; \
	IMAGE_NAME=$${IMAGE_NAME:-$(BINARY_NAME)}; \
	read -p "请输入镜像标签 [$(VERSION)]: " IMAGE_TAG; \
	IMAGE_TAG=$${IMAGE_TAG:-$(VERSION)}; \
	read -p "是否添加 latest 标签? (y/n) [y]: " ADD_LATEST; \
	ADD_LATEST=$${ADD_LATEST:-y}; \
	read -p "是否推送到远程仓库? (y/n) [n]: " PUSH_IMAGE; \
	PUSH_IMAGE=$${PUSH_IMAGE:-n}; \
	if [ "$$PUSH_IMAGE" = "y" ]; then \
		read -p "请输入仓库地址 (例: docker.io/username): " REGISTRY; \
		FULL_IMAGE="$$REGISTRY/$$IMAGE_NAME"; \
	else \
		FULL_IMAGE="$$IMAGE_NAME"; \
	fi; \
	echo ""; \
	echo "===> 开始构建镜像: $$FULL_IMAGE:$$IMAGE_TAG"; \
	docker build -t $$FULL_IMAGE:$$IMAGE_TAG .; \
	if [ "$$ADD_LATEST" = "y" ]; then \
		echo "===> 添加 latest 标签"; \
		docker tag $$FULL_IMAGE:$$IMAGE_TAG $$FULL_IMAGE:latest; \
	fi; \
	if [ "$$PUSH_IMAGE" = "y" ]; then \
		echo "===> 推送镜像到仓库..."; \
		docker push $$FULL_IMAGE:$$IMAGE_TAG; \
		if [ "$$ADD_LATEST" = "y" ]; then \
			docker push $$FULL_IMAGE:latest; \
		fi; \
	fi; \
	echo "===> 构建完成!"

# Docker Compose 启动
.PHONY: docker-up
docker-up:
	@echo "===> 启动 Docker Compose 服务..."
	docker-compose up -d
	@echo "===> 服务已启动，使用 'make docker-logs' 查看日志"

# Docker Compose 停止
.PHONY: docker-down
docker-down:
	@echo "===> 停止 Docker Compose 服务..."
	docker-compose down
	@echo "===> 服务已停止"

# Docker Compose 日志
.PHONY: docker-logs
docker-logs:
	docker-compose logs -f

# Docker 清理
.PHONY: docker-clean
docker-clean:
	@echo "===> 清理 Docker 资源..."
	-docker-compose down -v
	-docker rmi $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest 2>/dev/null || true
	@echo "===> 清理完成"


# CI/CD 专用（不自动修改代码）
.PHONY: ci
ci: deps fmt-check lint test
	@echo "==> 运行 CI 检查..."
	@echo "==> 所有 CI 检查通过"

# 查看版本信息
.PHONY: version
version:
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "提交: $(COMMIT)"

# 检查配置文件
.PHONY: check-config
check-config:
	@if [ ! -f $(CONFIG_FILE) ]; then \
		echo "错误: 配置文件不存在 $(CONFIG_FILE)"; \
		echo "请从示例配置创建: cp configs/config.example.yaml $(CONFIG_FILE)"; \
		exit 1; \
	else \
		echo "配置文件存在: $(CONFIG_FILE)"; \
	fi

# 开发环境初始化
.PHONY: init
init:
	@echo "==> 初始化开发环境..."
	@if [ ! -f $(CONFIG_FILE) ] && [ -f configs/config.example.yaml ]; then \
		echo "复制配置文件模板..."; \
		cp configs/config.example.yaml $(CONFIG_FILE); \
	fi
	@echo "==> 安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "==> 初始化完成"
