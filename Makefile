# 基础配置
APP_NAME = myapp
BUILD_DIR = build

# 默认构建 Linux AMD64
.DEFAULT_GOAL := build

.PHONY: build
build: build-linux-amd64

.PHONY: build-linux-amd64
build-linux-amd64:
	@echo "Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)-linux-amd64"

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build              - Build for Linux AMD64"
	@echo "  build-linux-amd64  - Build specifically for Linux AMD64"
	@echo "  clean              - Clean build directory"
	@echo "  help               - Show this help"