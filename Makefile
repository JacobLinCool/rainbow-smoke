GO ?= go
BINARY ?= smoke

all: fmt test build

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

build:
	$(GO) build -trimpath -o $(BINARY) ./cmd/smoke

run: build
	./$(BINARY)

benchmark-1024: build
	./$(BINARY) -width 1024 -height 1024 -seed 20260825 -benchmark

clean:
	rm -f $(BINARY)

.PHONY: all fmt test build run benchmark-1024 clean
