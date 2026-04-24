
CLI_BIN = k0re
DAEMON_BIN = k0red

CLI_DIR = ./cmd/cli
DAEMON_DIR = ./cmd/daemon

.PHONY: all build run test test-e2e clean install

all: build

build:
	@echo "Building K0re binaries..."
	@go build -o $(CLI_BIN) $(CLI_DIR)/main.go
	@go build -o $(DAEMON_BIN) $(DAEMON_DIR)/main.go
	@echo "Build complete. Binaries created: ./$(CLI_BIN), ./$(DAEMON_BIN)"

run: build
	@echo "Starting k0red daemon..."
	@./$(DAEMON_BIN)

test:
	@echo "Running unit tests..."
	@go test -v ./...

test-e2e:
	@echo "Running end-to-end integration tests..."
	@./test/integration.sh

clean:
	@echo "Cleaning up build artifacts and test data..."
	@rm -f $(CLI_BIN) $(DAEMON_BIN)
	@rm -rf ./k0re-data/*
	@rm -f /tmp/k0re_opts_*
	@rm -f /tmp/test-server.yaml
	@echo "Clean complete."

install: build
	@echo "Installing $(CLI_BIN) to /usr/local/bin..."
	@sudo cp $(CLI_BIN) /usr/local/bin/k0re
	@echo "Installation complete."