APP_NAME := homeass-wol
SRC_DIR := .
BUILD_DIR := ./bin
BIN := $(BUILD_DIR)/$(APP_NAME)

.PHONY: all build clean docker-build

all: build

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BIN) $(SRC_DIR)


clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)


docker-build:
	docker build -t $(APP_NAME) .
