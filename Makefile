SOURCE_DIR = src
GO = go

all: fmt clean build run

fmt:
	$(GO) fmt $(SOURCE_DIR)/*.go

build:
	GO111MODULE=off $(GO) build -o smoke $(SOURCE_DIR)/*.go

run:
	./smoke -cpu=cpu.profile -mem=mem.profile
	@$(GO) tool pprof -text cpu.profile
	@$(GO) tool pprof -text mem.profile

clean:
	rm -f smoke *.profile

setup:
	# GO111MODULE=on go install golang.org/dl/$(GO)@latest
	# $(GO) download
	GO111MODULE=off $(GO) get github.com/lucasb-eyer/go-colorful

.PHONY: all fmt build run clean setup
