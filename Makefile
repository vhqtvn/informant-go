.PHONY: build build-upx clean test install uninstall release release-upx help

# Variables
BINARY_NAME=informant
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-s -w -X main.version=${VERSION}"

# Default target
help:
	@echo "Available targets:"
	@echo "  build        Build the binary"
	@echo "  build-upx    Build and compress the binary with UPX"
	@echo "  clean        Clean build artifacts"
	@echo "  test         Run tests"
	@echo "  install      Install binary to /usr/local/bin (requires sudo)"
	@echo "  uninstall    Remove binary from /usr/local/bin (requires sudo)"
	@echo "  release      Build release binaries for multiple architectures"
	@echo "  release-upx  Build and compress release binaries with UPX"
	@echo "  help         Show this help message"

# Build the binary
build:
	go build ${LDFLAGS} -o ${BINARY_NAME} .

# Build and compress the binary with UPX
build-upx: build
	@if command -v upx >/dev/null 2>&1; then \
		echo "Original size: $$(du -h ${BINARY_NAME} | cut -f1)"; \
		upx --best --lzma ${BINARY_NAME}; \
		echo "Compressed size: $$(du -h ${BINARY_NAME} | cut -f1)"; \
	else \
		echo "UPX not found. Install with: sudo apt-get install upx-ucl"; \
		exit 1; \
	fi

# Clean build artifacts
clean:
	rm -f ${BINARY_NAME}
	rm -rf dist/

# Run tests
test:
	go test ./...

# Install binary (requires sudo)
install: build
	install -m 755 ${BINARY_NAME} /usr/local/bin/

# Uninstall binary (requires sudo)
uninstall:
	rm -f /usr/local/bin/${BINARY_NAME}

# Build release binaries
release: clean
	mkdir -p dist
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64 .
	# Linux ARM64  
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-arm64 .
	# Create checksums
	cd dist && sha256sum * > checksums.txt

# Build and compress release binaries with UPX
release-upx: clean
	@if ! command -v upx >/dev/null 2>&1; then \
		echo "UPX not found. Install with: sudo apt-get install upx-ucl"; \
		exit 1; \
	fi
	mkdir -p dist
	# Linux AMD64
	echo "Building Linux AMD64..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64 .
	echo "Original size (AMD64): $$(du -h dist/${BINARY_NAME}-linux-amd64 | cut -f1)"
	upx --best --lzma dist/${BINARY_NAME}-linux-amd64
	echo "Compressed size (AMD64): $$(du -h dist/${BINARY_NAME}-linux-amd64 | cut -f1)"
	# Linux ARM64
	echo "Building Linux ARM64..."
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-arm64 .
	echo "Original size (ARM64): $$(du -h dist/${BINARY_NAME}-linux-arm64 | cut -f1)"
	upx --best --lzma dist/${BINARY_NAME}-linux-arm64
	echo "Compressed size (ARM64): $$(du -h dist/${BINARY_NAME}-linux-arm64 | cut -f1)"
	# Create checksums
	cd dist && sha256sum * > checksums.txt 