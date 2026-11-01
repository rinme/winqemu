.PHONY: all build test run clean

BINARY=bin/winqemu

all: build

build:
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/winqemu

test:
	go test -v ./...

run: build
	./$(BINARY)

clean:
	rm -rf bin/
