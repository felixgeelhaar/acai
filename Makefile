VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build test test-integration test-all test-race lint clean release

build:
	CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o bin/acai ./cmd/acai

test:
	CGO_ENABLED=1 go test ./... -count=1

test-integration:
	CGO_ENABLED=1 go test -tags=integration -v -count=1 ./internal/integration/...

test-all:
	CGO_ENABLED=1 go test -tags=integration -count=1 ./...

test-race:
	CGO_ENABLED=1 go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/ coverage.out .cover/
