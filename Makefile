.PHONY: all build-linux build-mac build-windows clean

BIN_DIR := build
APP_NAME := temp-cleaner

all: build-linux build-mac build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 cmd/cleaner/main.go
	gzip -9 -f $(BIN_DIR)/$(APP_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 cmd/cleaner/main.go
	gzip -9 -f $(BIN_DIR)/$(APP_NAME)-linux-arm64

build-mac:
	GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 cmd/cleaner/main.go
	gzip -9 -f $(BIN_DIR)/$(APP_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 cmd/cleaner/main.go
	gzip -9 -f $(BIN_DIR)/$(APP_NAME)-darwin-arm64

build-windows:
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe cmd/cleaner/main.go
	gzip -9 -f $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe
	GOOS=windows GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME)-windows-arm64.exe cmd/cleaner/main.go
	gzip -9 -f $(BIN_DIR)/$(APP_NAME)-windows-arm64.exe

clean:
	rm -rf $(BIN_DIR)
