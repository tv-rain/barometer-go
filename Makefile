.PHONY: build check format format-check integrity lint mobile-smoke mobile-android mobile-apple race test traceability

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

# Native libraries; output is ignored by Git and is not a release artifact.
MOBILE_OUT ?= build/mobile
mobile-android:
	bash scripts/mobile-build.sh android "$(MOBILE_OUT)"

mobile-apple:
	bash scripts/mobile-build.sh apple "$(MOBILE_OUT)"

check: format-check build test race lint integrity mobile-smoke

# Explicit public checkout and external output directory; no implicit fetch/install.
PYTHON ?= python3
traceability:
	@test -n "$(SPEC_CHECKOUT)" -a -n "$(REPORT_DIR)" || { echo "Set SPEC_CHECKOUT and REPORT_DIR"; exit 2; }
	$(PYTHON) -B scripts/check-traceability.py --spec-checkout "$(SPEC_CHECKOUT)" --report-dir "$(REPORT_DIR)"
