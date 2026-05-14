.PHONY: build build-strategic build-scanner build-web test clean

GO := go
GOFLAGS := -ldflags="-s -w"
BIN_DIR := bin
BINARY := $(BIN_DIR)/swarm

build: build-web build-strategic build-scanner

build-web:
	cd web && npm install && npm run build

build-strategic: build-web
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY) ./cmd/swarm
	@echo "strategic node -> $(BINARY)"

build-scanner: build-web
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY) ./cmd/swarm
	@echo "scanner node -> $(BINARY)"

build-windows:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BIN_DIR)/swarm.exe ./cmd/swarm
	@echo "windows node -> $(BIN_DIR)/swarm.exe"

test:
	$(GO) test ./...

clean:
	rm -rf $(BIN_DIR)
