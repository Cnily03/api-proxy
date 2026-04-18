APP_NAME := api-proxy
BIN_DIR := bin
PKG := ./cmd/server

ifeq ($(OS),Windows_NT)
	EXE := .exe
	MKDIR_BIN := if not exist $(BIN_DIR) mkdir $(BIN_DIR)
	RMDIR_BIN := if exist $(BIN_DIR) rmdir /s /q $(BIN_DIR)
else
	EXE :=
	MKDIR_BIN := mkdir -p $(BIN_DIR)
	RMDIR_BIN := rm -rf $(BIN_DIR)
endif

.PHONY: all help deps build clean

fmt:
	go fmt ./...

all: deps build

help:
	@echo "Available targets:"
	@echo "  all    - run full flow (deps + build)"
	@echo "  deps   - tidy and download go modules"
	@echo "  build  - build binary into $(BIN_DIR)/$(APP_NAME)$(EXE)"
	@echo "  clean  - remove build output"

deps:
	go mod tidy
	go mod download

build:
	$(MKDIR_BIN)
	go build -o $(BIN_DIR)/$(APP_NAME)$(EXE) $(PKG)

clean:
	$(RMDIR_BIN)
