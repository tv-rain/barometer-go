.PHONY: build check format format-check lint mobile-smoke test

build:
	go build ./...

format:
	gofmt -w .

format-check:
	@test -z "$$(gofmt -l .)" || { \
		echo "Go files need formatting:"; \
		gofmt -l .; \
		exit 1; \
	}

test:
	go test ./...

lint:
	go vet ./...

mobile-smoke:
	CGO_ENABLED=0 GOOS=android GOARCH=arm64 go build ./...
	CGO_ENABLED=0 GOOS=ios GOARCH=arm64 go build ./...

check: format-check build test lint mobile-smoke
