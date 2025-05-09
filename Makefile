.PHONY: build install run test clean release help wasm

APP_NAME=verilis
MAIN_FILE=cmd/main.go
BUILD_DIR=bin

# Build information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
BUILD_TIME=$(shell date +%FT%T%z)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Environment variables (can be overridden via command line)
DEBUG ?= false

# LDFLAGS for embedding version info and environment variables
LD_FLAGS=-ldflags "-X 'main.AppName=${APP_NAME}' -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)' -X 'main.Debug=$(DEBUG)'"

# 默认目标 - 显示帮助信息
default: help

# 构建应用
build:
	@echo "Building $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# 安装应用到GOPATH/bin
install: build
	@echo "Installing $(APP_NAME)..."
	@cp $(BUILD_DIR)/$(APP_NAME) $(GOPATH)/bin/
	@echo "$(APP_NAME) has been installed to $(GOPATH)/bin/$(APP_NAME)"

# 运行应用
run:
	@go run $(LD_FLAGS) $(MAIN_FILE) $(ARGS)

# 运行代码质量检查
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed"; \
	fi

# 安装依赖
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod verify

test:
	@go test ./...

# 为多平台构建发布版本
release: clean
	@echo "Building release binaries v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)/release
	@# Linux builds
	@GOOS=linux GOARCH=amd64 go build $(LD_FLAGS) -o $(BUILD_DIR)/release/$(APP_NAME)-linux-amd64 $(MAIN_FILE)
	@GOOS=linux GOARCH=arm64 go build $(LD_FLAGS) -o $(BUILD_DIR)/release/$(APP_NAME)-linux-arm64 $(MAIN_FILE)
	@# macOS builds
	@GOOS=darwin GOARCH=amd64 go build $(LD_FLAGS) -o $(BUILD_DIR)/release/$(APP_NAME)-darwin-amd64 $(MAIN_FILE)
	@GOOS=darwin GOARCH=arm64 go build $(LD_FLAGS) -o $(BUILD_DIR)/release/$(APP_NAME)-darwin-arm64 $(MAIN_FILE)
	@# Windows builds
	@GOOS=windows GOARCH=amd64 go build $(LD_FLAGS) -o $(BUILD_DIR)/release/$(APP_NAME)-windows-amd64.exe $(MAIN_FILE)
	@# Create checksums
	@cd $(BUILD_DIR)/release && shasum -a 256 * > checksums.txt
	@# Create installation script
	@./scripts/generate_install_script.sh "$(VERSION)" "$(APP_NAME)"
	@echo "Release binaries created in $(BUILD_DIR)/release"
	@echo "Installation script created: install.sh"

# 压缩发布文件
package: release
	@echo "Packaging release binaries..."
	@mkdir -p $(BUILD_DIR)/packages
	@# Create tar.gz archives for Unix systems
	@tar -czf $(BUILD_DIR)/packages/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR)/release $(APP_NAME)-linux-amd64
	@tar -czf $(BUILD_DIR)/packages/$(APP_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(BUILD_DIR)/release $(APP_NAME)-linux-arm64
	@tar -czf $(BUILD_DIR)/packages/$(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR)/release $(APP_NAME)-darwin-amd64
	@tar -czf $(BUILD_DIR)/packages/$(APP_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BUILD_DIR)/release $(APP_NAME)-darwin-arm64
	@# Create zip for Windows
	@zip -j $(BUILD_DIR)/packages/$(APP_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/release/$(APP_NAME)-windows-amd64.exe
	@# Copy install script
	@cp install.sh $(BUILD_DIR)/packages/
	@echo "Packages created in $(BUILD_DIR)/packages"
	@GOOS=darwin GOARCH=amd64 go build $(LD_FLAGS) -o $(BUILD_DIR)/release/$(APP_NAME)-darwin-amd64 $(MAIN_FILE)
	@GOOS=linux GOARCH=amd64 go build $(LD_FLAGS) -o $(BUILD_DIR)/release/$(APP_NAME)-linux-amd64 $(MAIN_FILE)
	@GOOS=windows GOARCH=amd64 go build $(LD_FLAGS) -o $(BUILD_DIR)/release/$(APP_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "Release builds complete in $(BUILD_DIR)/release/"

# 清理构建产物
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

dev: build run
