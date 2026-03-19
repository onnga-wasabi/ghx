.PHONY: build install test lint clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS  = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)

build:
	go build -ldflags "$(LDFLAGS)" -o ghx ./cmd/ghx/main.go

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/ghx/main.go

test:
	go test -count=1 ./...

lint:
	golangci-lint run ./...

clean:
	rm -f ghx
