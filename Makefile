.PHONY: build check format format-check integrity lint mobile-smoke mobile-smoke-android mobile-smoke-apple race test

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

race:
	go test -race ./...

lint:
	go vet ./...

integrity:
	go mod tidy -diff
	go mod verify
	git diff --check

mobile-smoke:
	CGO_ENABLED=0 GOOS=android GOARCH=arm64 go build ./...
	CGO_ENABLED=0 GOOS=ios GOARCH=arm64 go build ./...

mobile-smoke-android:
	bash scripts/mobile-smoke.sh android

mobile-smoke-apple:
	bash scripts/mobile-smoke.sh apple

check: format-check build test race lint integrity mobile-smoke
