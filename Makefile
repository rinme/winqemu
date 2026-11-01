.PHONY: all build test run clean setup-pkgconfig

BINARY=bin/winqemu
PKG_DIR := $(shell pwd)/.pkgconfig
export PKG_CONFIG_PATH := $(PKG_DIR):$(HOME)/.local/lib/pkgconfig:$(PKG_CONFIG_PATH)

all: build

setup-pkgconfig:
	@mkdir -p $(PKG_DIR)
	@if ! pkg-config --exists webkit2gtk-4.0 2>/dev/null && pkg-config --exists webkit2gtk-4.1 2>/dev/null; then \
		ln -sf $$(pkg-config --variable=pcfiledir webkit2gtk-4.1)/webkit2gtk-4.1.pc $(PKG_DIR)/webkit2gtk-4.0.pc; \
	fi

build: setup-pkgconfig
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/winqemu

test:
	go test -v ./...

run: build
	./$(BINARY)

clean:
	rm -rf bin/ $(PKG_DIR)
