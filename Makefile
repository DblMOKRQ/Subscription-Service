APP_NAME=effective_mobile_app
BUILD_DIR=./bin
SWAG_DIR=./docs

.PHONY: all build run test clean swagger

all: build

build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd
	@echo "Build complete."

run:
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(SWAG_DIR)/docs.go $(SWAG_DIR)/swagger.json $(SWAG_DIR)/swagger.yaml
	@echo "Clean complete."

swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g ./cmd/main.go -o $(SWAG_DIR)
	@echo "Swagger documentation generated."


