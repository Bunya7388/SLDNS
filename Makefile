.PHONY: build server client clean run-server run-client deps help

help:
	@echo "DNSTT Build Targets:"
	@echo "  make build        - Build both server and client"
	@echo "  make server       - Build server only"
	@echo "  make client       - Build client only"
	@echo "  make clean        - Remove binaries"
	@echo "  make deps         - Download dependencies"
	@echo "  make run-server   - Build and run server (requires root)"
	@echo "  make run-client   - Build and run client"

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

build: deps
	@echo "Building DNSTT..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/dnstt-server ./cmd/server
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/dnstt-client ./cmd/client
	@chmod +x bin/dnstt-server bin/dnstt-client
	@echo "✓ Build complete: bin/dnstt-server bin/dnstt-client"

server: deps
	@echo "Building server..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/dnstt-server ./cmd/server
	@chmod +x bin/dnstt-server
	@echo "✓ Server built: bin/dnstt-server"

client: deps
	@echo "Building client..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/dnstt-client ./cmd/client
	@chmod +x bin/dnstt-client
	@echo "✓ Client built: bin/dnstt-client"

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "✓ Cleaned"

run-server: build
	@chmod +x run-server.sh
	@./run-server.sh

run-client: build
	@chmod +x run-client.sh
	@./run-client.sh
